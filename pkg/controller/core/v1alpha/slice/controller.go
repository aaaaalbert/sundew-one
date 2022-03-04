/*
Copyright 2021 Contributors to the EdgeNet project.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package slice

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"time"

	corev1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	clientset "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/scheme"
	edgenetscheme "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/scheme"
	informers "github.com/EdgeNet-project/edgenet/pkg/generated/informers/externalversions/core/v1alpha"
	listers "github.com/EdgeNet-project/edgenet/pkg/generated/listers/core/v1alpha"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const controllerAgentName = "slice-controller"

// Definitions of the state of the slice resource
const (
	successSynced         = "Synced"
	messageResourceSynced = "Slice synced successfully"
	successBound          = "Bound"
	messageBound          = "Slice is bound successfully"
	successReserved       = "Reserved"
	messageReserved       = "Desired resources are reserved"
	failureSlice          = "Slice Failed"
	messageSliceFailed    = "There are no adequate resources to slice"
	failurePatch          = "Patch Failed"
	messagePatchFailed    = "Node patch operation has failed"
	failure               = "Failure"
	reserved              = "Reserved"
	bound                 = "Bound"
)

// Controller is the controller implementation for Slice resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// edgenetclientset is a clientset for the EdgeNet API groups
	edgenetclientset clientset.Interface

	sliceClaimsLister listers.SliceClaimLister
	sliceClaimsSynced cache.InformerSynced

	slicesLister listers.SliceLister
	slicesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new controller
func NewController(
	kubeclientset kubernetes.Interface,
	edgenetclientset clientset.Interface,
	sliceClaimInformer informers.SliceClaimInformer,
	sliceInformer informers.SliceInformer) *Controller {

	utilruntime.Must(edgenetscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		edgenetclientset:  edgenetclientset,
		sliceClaimsLister: sliceClaimInformer.Lister(),
		sliceClaimsSynced: sliceClaimInformer.Informer().HasSynced,
		slicesLister:      sliceInformer.Lister(),
		slicesSynced:      sliceInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Slices"),
		recorder:          recorder,
	}

	klog.V(4).Infoln("Setting up event handlers")
	// Set up an event handler for when Slice resources change
	sliceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSlice,
		UpdateFunc: func(old, new interface{}) {
			newSlice := new.(*corev1alpha.Slice)
			oldSlice := old.(*corev1alpha.Slice)
			if (oldSlice.Status.Expiry == nil && newSlice.Status.Expiry != nil) ||
				!oldSlice.Status.Expiry.Time.Equal(newSlice.Status.Expiry.Time) {
				controller.enqueueSliceAfter(newSlice, time.Until(newSlice.Status.Expiry.Time))
			}
			controller.enqueueSlice(new)
		},
		DeleteFunc: func(obj interface{}) {
			slice := obj.(*corev1alpha.Slice)
			if nodeRaw, err := controller.kubeclientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("edge-net.io/slice=%s", slice.GetName())}); err != nil {
				for _, nodeRow := range nodeRaw.Items {
					patch := []byte(`"metadata": {"labels": {"edge-net.io~access": "private", "edge-net.io~slice":  "none"}}`)
					// Patch the node
					_, err := controller.kubeclientset.CoreV1().Nodes().Patch(context.TODO(), nodeRow.GetName(), types.StrategicMergePatchType, patch, metav1.PatchOptions{})
					if err != nil {
						log.Println(err.Error())
						panic(err.Error())
					}
				}
			}
			controller.edgenetclientset.CoreV1alpha().SliceClaims(slice.Spec.ClaimRef.Namespace).Delete(context.TODO(), slice.Spec.ClaimRef.Name, metav1.DeleteOptions{})
		},
	})

	sliceClaimInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for the types of slice and node, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.V(4).Infoln("Starting Slice controller")

	klog.V(4).Infoln("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh,
		c.sliceClaimsSynced,
		c.slicesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.V(4).Infoln("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.V(4).Infoln("Started workers")
	<-stopCh
	klog.V(4).Infoln("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.V(4).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Slice
// resource with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	slice, err := c.slicesLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("slice '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	c.processSlice(slice.DeepCopy())

	c.recorder.Event(slice, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

// enqueueSlice takes a Slice resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Slice.
func (c *Controller) enqueueSlice(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// enqueueSliceAfter takes a Slice resource and converts it into a namespace/name
// string which is then put onto the work queue after the expiry date of a claim/drop to delete the so-said claim/drop.
// This method should *not* be passed resources of any type other than Slice.
func (c *Controller) enqueueSliceAfter(obj interface{}, after time.Duration) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddAfter(key, after)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Slice resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Slice resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Slice" {
			return
		}

		slice, err := c.slicesLister.Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of slice '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueSlice(slice)
		return
	}
}

func (c *Controller) processSlice(sliceCopy *corev1alpha.Slice) {
	oldStatus := sliceCopy.Status
	statusUpdate := func() {
		if !reflect.DeepEqual(oldStatus, sliceCopy.Status) {
			if _, err := c.edgenetclientset.CoreV1alpha().Slices().UpdateStatus(context.TODO(), sliceCopy, metav1.UpdateOptions{}); err != nil {
				klog.V(4).Infoln(err)
			}
		}
	}
	defer statusUpdate()

	if sliceCopy.Spec.ClaimRef != nil {
		if sliceClaim, err := c.edgenetclientset.CoreV1alpha().SliceClaims(sliceCopy.Spec.ClaimRef.Namespace).Get(context.TODO(), sliceCopy.Spec.ClaimRef.Name, metav1.GetOptions{}); err != nil && errors.IsGone(err) {
			c.edgenetclientset.CoreV1alpha().Slices().Delete(context.TODO(), sliceCopy.GetName(), metav1.DeleteOptions{})
			return
		} else {
			isOwned := false
			if ownerRef := metav1.GetControllerOf(sliceClaim); ownerRef != nil {
				if ownerRef.Kind == "Slice" {
					if ownerRef.UID != sliceCopy.GetUID() {
						c.edgenetclientset.CoreV1alpha().Slices().Delete(context.TODO(), sliceCopy.GetName(), metav1.DeleteOptions{})
						return
					} else {
						isOwned = true
						if sliceCopy.Status.State == reserved {
							c.recorder.Event(sliceCopy, corev1.EventTypeNormal, successBound, messageBound)
							sliceCopy.Status.State = bound
							sliceCopy.Status.Message = messageBound
							return
						}
					}
				}
			}
			if !isOwned {
				defer func() {
					ownerReferences := SetAsOwnerReference(sliceCopy)
					sliceClaimCopy := sliceClaim.DeepCopy()
					sliceClaimCopy.SetOwnerReferences(ownerReferences)
					c.edgenetclientset.CoreV1alpha().SliceClaims(sliceClaimCopy.GetNamespace()).Update(context.TODO(), sliceClaimCopy, metav1.UpdateOptions{})
				}()
			}
		}
	}

	if sliceCopy.Status.State != reserved && sliceCopy.Status.State != bound {
		c.reserveNode(sliceCopy)
	}
}

func (c *Controller) reserveNode(sliceCopy *corev1alpha.Slice) {
	for _, nodeSelectorTerm := range sliceCopy.Spec.NodeSelector.Selector.NodeSelectorTerms {
		var labelSelector string
		var fieldSelector string
		for _, matchExpression := range nodeSelectorTerm.MatchExpressions {
			if labelSelector != "" {
				labelSelector = labelSelector + ","
			}
			if matchExpression.Operator == "In" || matchExpression.Operator == "NotIn" {
				labelSelector = fmt.Sprintf("%s%s %s (%s)", labelSelector, matchExpression.Key, strings.ToLower(string(matchExpression.Operator)), strings.Join(matchExpression.Values, ","))
			} else if matchExpression.Operator == "Exists" {
				labelSelector = fmt.Sprintf("%s%s", labelSelector, matchExpression.Key)
			} else if matchExpression.Operator == "DoesNotExist" {
				labelSelector = fmt.Sprintf("%s!%s", labelSelector, matchExpression.Key)
			} else {
				// TO-DO: Handle Gt and Lt operaters later.
				continue
			}
		}
		for _, matchField := range nodeSelectorTerm.MatchFields {
			if fieldSelector != "" {
				fieldSelector = fieldSelector + ","
			}
			if matchField.Operator == "In" || matchField.Operator == "NotIn" {
				fieldSelector = fmt.Sprintf("%s%s %s (%s)", fieldSelector, matchField.Key, strings.ToLower(string(matchField.Operator)), strings.Join(matchField.Values, ","))
			} else if matchField.Operator == "Exists" {
				fieldSelector = fmt.Sprintf("%s%s", fieldSelector, matchField.Key)
			} else if matchField.Operator == "DoesNotExist" {
				fieldSelector = fmt.Sprintf("%s!%s", fieldSelector, matchField.Key)
			} else {
				// TO-DO: Handle Gt and Lt operaters later.
				continue
			}
		}

		var nodeList []string
		if nodeRaw, err := c.kubeclientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector}); err != nil {
			for _, nodeRow := range nodeRaw.Items {
				nodeLabels := nodeRow.GetLabels()
				if nodeLabels["edge-net.io/access"] != "private" || nodeLabels["edge-net.io/slice"] != "none" {
					continue
				}

				match := true
				for key, value := range sliceCopy.Spec.NodeSelector.Resources.Limits {
					if value.Cmp(nodeRow.Status.Capacity[key]) == 1 || value.Cmp(nodeRow.Status.Capacity[key]) == 0 {
						match = false
					}
				}
				if match {
					nodeList = append(nodeList, nodeRow.GetName())
				}
			}
		}

		if len(nodeList) < sliceCopy.Spec.NodeSelector.Count {
			c.recorder.Event(sliceCopy, corev1.EventTypeWarning, failureSlice, messageSliceFailed)
			sliceCopy.Status.State = failure
			sliceCopy.Status.Message = messageSliceFailed
		} else {
			var pickedNodeList []string
			for i := 0; i < sliceCopy.Spec.NodeSelector.Count; i++ {
				rand.Seed(time.Now().Unix())
				pickedNodeList = append(pickedNodeList, nodeList[rand.Intn(len(nodeList))])
			}
			isPatched := true
			for i := 0; i < len(pickedNodeList); i++ {
				if err := c.patchNode("slice", sliceCopy.GetName(), pickedNodeList[i]); err != nil {
					c.recorder.Event(sliceCopy, corev1.EventTypeWarning, failurePatch, messagePatchFailed)
					sliceCopy.Status.State = failure
					sliceCopy.Status.Message = messagePatchFailed
					isPatched = false
					break
				}
			}
			if !isPatched {
				for i := 0; i < len(pickedNodeList); i++ {
					if err := c.patchNode("return", sliceCopy.GetName(), pickedNodeList[i]); err != nil {
						c.recorder.Event(sliceCopy, corev1.EventTypeWarning, failurePatch, messagePatchFailed)
					}
				}
				return
			}
		}
	}
	c.recorder.Event(sliceCopy, corev1.EventTypeNormal, successReserved, messageReserved)
	sliceCopy.Status.State = reserved
	sliceCopy.Status.Message = messageReserved
}

func (c *Controller) patchNode(kind, slice, node string) error {
	var err error
	patch := []byte(fmt.Sprintf(`"metadata": {"labels": {"edge-net.io~access": "private", "edge-net.io~slice":  %s}}`, slice))
	if kind == "return" {
		patch = []byte(`"metadata": {"labels": {"edge-net.io~access": "public", "edge-net.io~slice":  "none"}}`)
	}
	_, err = c.kubeclientset.CoreV1().Nodes().Patch(context.TODO(), node, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		klog.V(4).Infoln(err.Error())
	}
	return err
}

// SetAsOwnerReference returns the slice as owner
func SetAsOwnerReference(slice *corev1alpha.Slice) []metav1.OwnerReference {
	// The following section makes slice become the owner
	ownerReferences := []metav1.OwnerReference{}
	newRef := *metav1.NewControllerRef(slice, corev1alpha.SchemeGroupVersion.WithKind("Slice"))
	takeControl := true
	newRef.Controller = &takeControl
	ownerReferences = append(ownerReferences, newRef)
	return ownerReferences
}
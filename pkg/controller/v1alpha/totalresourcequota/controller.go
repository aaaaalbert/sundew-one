/*
Copyright 2020 Sorbonne Université

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

package totalresourcequota

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"edgenet/pkg/apis/apps/v1alpha"
	apps_v1alpha "edgenet/pkg/apis/apps/v1alpha"
	"edgenet/pkg/client/clientset/versioned"
	appsinformer_v1 "edgenet/pkg/client/informers/externalversions/apps/v1alpha"
	"edgenet/pkg/node"

	log "github.com/Sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// The main structure of controller
type controller struct {
	logger       *log.Entry
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	nodeInformer cache.SharedIndexInformer
	handler      HandlerInterface
}

// The main structure of informerEvent
type informerevent struct {
	key      string
	function string
	change   fields
}

// This contains the fields to check whether they are updated
type fields struct {
	expiry bool
	spec   bool
}

// Constant variables for events
const create = "create"
const update = "update"
const delete = "delete"
const failure = "Pulled off"
const success = "Applied"
const trueStr = "True"
const falseStr = "False"
const unknownStr = "Unknown"

// Dictionary of status messages
var statusDict = map[string]string{
	"TRQ-created":       "Total resource quota created",
	"TRQ-failed":        "Couldn't create total resource quota in %s: %s",
	"authority-disable": "Authority disabled",
	"TRQ-disabled":      "Total resource quota disabled",
	"TRQ-applied":       "Total resource quota applied",
	"TRQ-appliedFail":   "Total resource quota couldn't be applied",
}

// Start function is entry point of the controller
func Start(kubernetes kubernetes.Interface, edgenet versioned.Interface) {
	var err error
	clientset := kubernetes
	edgenetClientset := edgenet

	TRQHandler := &Handler{}
	// Create the TRQ informer which was generated by the code generator to list and watch TRQ resources
	informer := appsinformer_v1.NewTotalResourceQuotaInformer(
		edgenetClientset,
		0,
		cache.Indexers{},
	)
	// Create a work queue which contains a key of the resource to be handled by the handler
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var event informerevent
	// Event handlers deal with events of resources. Here, there are three types of events as Add, Update, and Delete
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// Put the resource object into a key
			event.key, err = cache.MetaNamespaceKeyFunc(obj)
			event.function = create
			log.Infof("Add TRQ: %s", event.key)
			if err == nil {
				// Add the key to the queue
				queue.Add(event)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			event.key, err = cache.MetaNamespaceKeyFunc(newObj)
			event.function = update
			event.change.expiry = false
			event.change.spec = false
			oldExists := CheckExpiryDate(oldObj.(*apps_v1alpha.TotalResourceQuota))
			newExists := CheckExpiryDate(newObj.(*apps_v1alpha.TotalResourceQuota))
			if oldExists == false && newExists == true {
				event.change.expiry = true
			}
			if !reflect.DeepEqual(oldObj.(*apps_v1alpha.TotalResourceQuota).Spec, newObj.(*apps_v1alpha.TotalResourceQuota).Spec) {
				event.change.spec = true
			}
			log.Infof("Update TRQ: %s", event.key)
			if err == nil {
				queue.Add(event)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// DeletionHandlingMetaNamsespaceKeyFunc helps to check the existence of the object while it is still contained in the index.
			// Put the resource object into a key
			event.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			event.function = delete
			log.Infof("Delete TRQ: %s", event.key)
			if err == nil {
				queue.Add(event)
			}
		},
	})
	// The total resource quota objects are reconfigured according to node events in this section
	nodeInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			// The main purpose of listing is to attach geo labels to whole nodes at the beginning
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return clientset.CoreV1().Nodes().List(options)
			},
			// This function watches all changes/updates of nodes
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.CoreV1().Nodes().Watch(options)
			},
		},
		&corev1.Node{},
		0,
		cache.Indexers{},
	)
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			nodeObj := obj.(*corev1.Node)
			for key, _ := range nodeObj.Labels {
				if key == "node-role.kubernetes.io/master" {
					return
				}
			}
			ready := node.GetConditionReadyStatus(nodeObj)
			if ready == trueStr {
				for _, owner := range nodeObj.GetOwnerReferences() {
					if owner.Kind == "Authority" {
						TRQCopy, err := edgenetClientset.AppsV1alpha().TotalResourceQuotas().Get(owner.Name, metav1.GetOptions{})
						if err == nil {
							exists := false
							CPUAward := resource.NewQuantity(int64(float64(nodeObj.Status.Capacity.Cpu().Value())*1.5), resource.BinarySI).DeepCopy()
							memoryAward := resource.NewQuantity(int64(float64(nodeObj.Status.Capacity.Memory().Value())*1.3), resource.BinarySI).DeepCopy()
							claims := TRQCopy.Spec.Claim
							for i, claimRow := range TRQCopy.Spec.Claim {
								if claimRow.Name == "Reward" {
									exists = true
									rewardedCPU := resource.MustParse(claimRow.CPU)
									rewardedCPU.Add(CPUAward)
									rewardedMemory := resource.MustParse(claimRow.Memory)
									rewardedMemory.Add(memoryAward)
									claimRow.CPU = rewardedCPU.String()
									claimRow.Memory = rewardedMemory.String()
									claims = append(claims[:i], claims[i+1:]...)
									claims = append(claims, claimRow)
								}
							}
							if !exists {
								claim := v1alpha.TotalResourceDetails{}
								claim.Name = "Reward"
								claim.CPU = CPUAward.String()
								claim.Memory = memoryAward.String()
								TRQCopy.Spec.Claim = append(TRQCopy.Spec.Claim, claim)
							} else {
								TRQCopy.Spec.Claim = claims
							}
							edgenetClientset.AppsV1alpha().TotalResourceQuotas().Update(TRQCopy)
						}
					}
				}
			}
		},
		UpdateFunc: func(old, new interface{}) {
			oldObj := old.(*corev1.Node)
			newObj := new.(*corev1.Node)
			oldReady := node.GetConditionReadyStatus(oldObj)
			newReady := node.GetConditionReadyStatus(newObj)
			if (oldReady == falseStr && newReady == trueStr) ||
				(oldReady == unknownStr && newReady == trueStr) {
				for _, owner := range newObj.GetOwnerReferences() {
					if owner.Kind == "Authority" {
						TRQCopy, err := edgenetClientset.AppsV1alpha().TotalResourceQuotas().Get(owner.Name, metav1.GetOptions{})
						if err == nil {
							exists := false
							CPUAward := resource.NewQuantity(int64(float64(newObj.Status.Capacity.Cpu().Value())*1.5), resource.BinarySI).DeepCopy()
							memoryAward := resource.NewQuantity(int64(float64(newObj.Status.Capacity.Memory().Value())*1.3), resource.BinarySI).DeepCopy()
							claims := TRQCopy.Spec.Claim
							for i, claimRow := range TRQCopy.Spec.Claim {
								if claimRow.Name == "Reward" {
									exists = true
									rewardedCPU := resource.MustParse(claimRow.CPU)
									rewardedCPU.Add(CPUAward)
									rewardedMemory := resource.MustParse(claimRow.Memory)
									rewardedMemory.Add(memoryAward)
									claimRow.CPU = rewardedCPU.String()
									claimRow.Memory = rewardedMemory.String()
									claims = append(claims[:i], claims[i+1:]...)
									claims = append(claims, claimRow)
								}
							}
							if !exists {
								claim := v1alpha.TotalResourceDetails{}
								claim.Name = "Reward"
								claim.CPU = CPUAward.String()
								claim.Memory = memoryAward.String()
								TRQCopy.Spec.Claim = append(TRQCopy.Spec.Claim, claim)
							} else {
								TRQCopy.Spec.Claim = claims
							}
							edgenetClientset.AppsV1alpha().TotalResourceQuotas().Update(TRQCopy)
						}
					}
				}
			} else if (oldReady == trueStr && newReady == falseStr) ||
				(oldReady == trueStr && newReady == unknownStr) {
				for _, owner := range newObj.GetOwnerReferences() {
					if owner.Kind == "Authority" {
						TRQCopy, err := edgenetClientset.AppsV1alpha().TotalResourceQuotas().Get(owner.Name, metav1.GetOptions{})
						if err == nil {
							CPUAward := resource.NewQuantity(int64(float64(newObj.Status.Capacity.Cpu().Value())*1.5), resource.BinarySI).DeepCopy()
							memoryAward := resource.NewQuantity(int64(float64(newObj.Status.Capacity.Memory().Value())*1.3), resource.BinarySI).DeepCopy()
							claims := TRQCopy.Spec.Claim
							for i, claimRow := range TRQCopy.Spec.Claim {
								if claimRow.Name == "Reward" {
									rewardedCPU := resource.MustParse(claimRow.CPU)
									rewardedCPU.Sub(CPUAward)
									rewardedMemory := resource.MustParse(claimRow.Memory)
									rewardedMemory.Sub(memoryAward)
									claimRow.CPU = rewardedCPU.String()
									claimRow.Memory = rewardedMemory.String()
									claims = append(claims[:i], claims[i+1:]...)
									claims = append(claims, claimRow)
								}
							}
							TRQCopy.Spec.Claim = claims
							edgenetClientset.AppsV1alpha().TotalResourceQuotas().Update(TRQCopy)
						}
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			log.Println("Node Deleted Event")
			nodeObj := obj.(*corev1.Node)
			ready := node.GetConditionReadyStatus(nodeObj)
			if ready == trueStr {
				for _, owner := range nodeObj.GetOwnerReferences() {
					if owner.Kind == "Authority" {
						TRQCopy, err := edgenetClientset.AppsV1alpha().TotalResourceQuotas().Get(owner.Name, metav1.GetOptions{})
						if err == nil {
							CPUAward := resource.NewQuantity(int64(float64(nodeObj.Status.Capacity.Cpu().Value())*1.5), resource.BinarySI).DeepCopy()
							memoryAward := resource.NewQuantity(int64(float64(nodeObj.Status.Capacity.Memory().Value())*1.3), resource.BinarySI).DeepCopy()
							claims := TRQCopy.Spec.Claim
							for i, claimRow := range TRQCopy.Spec.Claim {
								if claimRow.Name == "Reward" {
									rewardedCPU := resource.MustParse(claimRow.CPU)
									rewardedCPU.Sub(CPUAward)
									rewardedMemory := resource.MustParse(claimRow.Memory)
									rewardedMemory.Sub(memoryAward)
									claimRow.CPU = rewardedCPU.String()
									claimRow.Memory = rewardedMemory.String()
									claims = append(claims[:i], claims[i+1:]...)
									claims = append(claims, claimRow)
								}
							}
							TRQCopy.Spec.Claim = claims
							edgenetClientset.AppsV1alpha().TotalResourceQuotas().Update(TRQCopy)
						}
					}
				}
			}
		},
	})
	controller := controller{
		logger:       log.NewEntry(log.New()),
		informer:     informer,
		nodeInformer: nodeInformer,
		queue:        queue,
		handler:      TRQHandler,
	}

	// A channel to terminate elegantly
	stopCh := make(chan struct{})
	defer close(stopCh)
	// Run the controller loop as a background task to start processing resources
	go controller.run(stopCh, clientset, edgenetClientset)
	// A channel to observe OS signals for smooth shut down
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}

// Run starts the controller loop
func (c *controller) run(stopCh <-chan struct{}, clientset kubernetes.Interface, edgenetClientset versioned.Interface) {
	// A Go panic which includes logging and terminating
	defer utilruntime.HandleCrash()
	// Shutdown after all goroutines have done
	defer c.queue.ShutDown()
	c.logger.Info("run: initiating")
	c.handler.Init(clientset, edgenetClientset)
	// Run the informer to list and watch resources
	go c.informer.Run(stopCh)
	go c.nodeInformer.Run(stopCh)

	// Synchronization to settle resources one
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced, c.nodeInformer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	c.logger.Info("run: cache sync complete")
	// Operate the runWorker
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
}

// To process new objects added to the queue
func (c *controller) runWorker() {
	log.Info("runWorker: starting")
	// Run processNextItem for all the changes
	for c.processNextItem() {
		log.Info("runWorker: processing next item")
	}

	log.Info("runWorker: completed")
}

// This function deals with the queue and sends each item in it to the specified handler to be processed.
func (c *controller) processNextItem() bool {
	log.Info("processNextItem: start")
	// Fetch the next item of the queue
	event, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(event)
	// Get the key string
	keyRaw := event.(informerevent).key
	// Use the string key to get the object from the indexer
	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		if c.queue.NumRequeues(event.(informerevent).key) < 5 {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, retrying", event.(informerevent).key, err)
			c.queue.AddRateLimited(event.(informerevent).key)
		} else {
			c.logger.Errorf("Controller.processNextItem: Failed processing item with key %s with error %v, no more retries", event.(informerevent).key, err)
			c.queue.Forget(event.(informerevent).key)
			utilruntime.HandleError(err)
		}
	}

	if !exists {
		if event.(informerevent).function == delete {
			c.logger.Infof("Controller.processNextItem: object deleted detected: %s", keyRaw)
			c.handler.ObjectDeleted(item)
		}
	} else {
		if event.(informerevent).function == create {
			c.logger.Infof("Controller.processNextItem: object created detected: %s", keyRaw)
			c.handler.ObjectCreated(item)
		} else if event.(informerevent).function == update {
			log.Println(event.(informerevent).key)
			c.logger.Infof("Controller.processNextItem: object updated detected: %s", keyRaw)
			c.handler.ObjectUpdated(item, event.(informerevent).change)
		}
	}
	c.queue.Forget(event.(informerevent).key)

	return true
}

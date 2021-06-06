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

package subnamespace

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	"github.com/EdgeNet-project/edgenet/pkg/controller/core/v1alpha/tenantresourcequota"
	"github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// HandlerInterface interface contains the methods that are required
type HandlerInterface interface {
	Init(kubernetes kubernetes.Interface, edgenet versioned.Interface)
	ObjectCreatedOrUpdated(obj interface{})
	ObjectDeleted(obj interface{})
	RunExpiryController()
}

// Handler implementation
type Handler struct {
	clientset        kubernetes.Interface
	edgenetClientset versioned.Interface
}

// Init handles any handler initialization
func (t *Handler) Init(kubernetes kubernetes.Interface, edgenet versioned.Interface) {
	log.Info("SubNamespaceHandler.Init")
	t.clientset = kubernetes
	t.edgenetClientset = edgenet
}

// ObjectCreatedOrUpdated is called when an object is created
func (t *Handler) ObjectCreatedOrUpdated(obj interface{}) {
	log.Info("SubNamespaceHandler.ObjectCreatedOrUpdated")
	// Make a copy of the subNamespace object to make changes on it
	subNamespace := obj.(*corev1alpha.SubNamespace).DeepCopy()

	tenantEnabled := false
	parentNamespace, _ := t.clientset.CoreV1().Namespaces().Get(context.TODO(), subNamespace.GetNamespace(), metav1.GetOptions{})
	labels := parentNamespace.GetLabels()
	if labels != nil && labels["edge-net.io/tenant"] != "" {
		if tenant, err := t.edgenetClientset.CoreV1alpha().Tenants().Get(context.TODO(), labels["edge-net.io/tenant"], metav1.GetOptions{}); err == nil {
			tenantEnabled = tenant.Spec.Enabled
		}
	} else {
		return
	}
	// Check if the tenant is active
	if tenantEnabled {
		defer func() {
			if !reflect.DeepEqual(obj.(*corev1alpha.SubNamespace).Status, subNamespace.Status) {
				if _, err := t.edgenetClientset.CoreV1alpha().SubNamespaces(subNamespace.GetNamespace()).UpdateStatus(context.TODO(), subNamespace, metav1.UpdateOptions{}); err != nil {
					// TO-DO: Provide more information on error
					log.Println(err)
				}
			}
		}()

		tenantResourceQuota, _ := t.edgenetClientset.CoreV1alpha().TenantResourceQuotas().Get(context.TODO(), labels["edge-net.io/tenant"], metav1.GetOptions{})
		trqHandler := tenantresourcequota.Handler{}
		trqHandler.Init(t.clientset, t.edgenetClientset)
		cpuResource := resource.MustParse(subNamespace.Spec.Resources.CPU)
		cpuDemand := cpuResource.Value()
		memoryResource := resource.MustParse(subNamespace.Spec.Resources.Memory)
		memoryDemand := memoryResource.Value()
		_, quotaExceeded, cpuDecline, memoryDecline := trqHandler.ResourceConsumptionControl(tenantResourceQuota, cpuDemand, memoryDemand)
		childNamespace, err := t.clientset.CoreV1().Namespaces().Get(context.TODO(), fmt.Sprintf("%s-%s", labels["edge-net.io/tenant"], subNamespace.GetName()), metav1.GetOptions{})
		if err == nil {
			childNamespaceLabels := childNamespace.GetLabels()
			if childNamespaceLabels["edge-net.io/owner"] != fmt.Sprintf("%s-%s", subNamespace.GetNamespace(), subNamespace.GetName()) {
				// TODO: Error handling
				subNamespace.Status.State = failure
				subNamespace.Status.Message = []string{fmt.Sprintf(statusDict["namespace-exists"], fmt.Sprintf("%s-%s", subNamespace.GetNamespace(), subNamespace.GetName()))}
				return
			}
		} else {
			if !quotaExceeded {
				childNamespaceObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", labels["edge-net.io/tenant"], subNamespace.GetName())}}
				namespaceLabels := map[string]string{"edge-net.io/generated": "true", "edge-net.io/kind": "sub", "edge-net.io/tenant": labels["edge-net.io/tenant"], "edge-net.io/owner": fmt.Sprintf("%s-%s", subNamespace.GetNamespace(), subNamespace.GetName())}
				childNamespaceObj.SetLabels(namespaceLabels)
				childNamespace, err = t.clientset.CoreV1().Namespaces().Create(context.TODO(), childNamespaceObj, metav1.CreateOptions{})
				if err != nil {
					// TODO: Error handling
					subNamespace.Status.State = failure
					subNamespace.Status.Message = []string{statusDict["subnamespace-failure"]}
				}
			} else {
				// TODO: Error handling
				subNamespace.Status.State = failure
				subNamespace.Status.Message = []string{statusDict["quota-exceeded"]}
				return
			}
		}

		coreQuotaCPU := -cpuDemand
		coreQuotaMemory := -memoryDemand
		if subResourceQuota, err := t.clientset.CoreV1().ResourceQuotas(childNamespace.GetName()).Get(context.TODO(), "sub-quota", metav1.GetOptions{}); err == nil {
			if subResourceQuota.Spec.Hard.Cpu().Value() >= cpuDecline && subResourceQuota.Spec.Hard.Memory().Value() >= memoryDecline {
				quotaExceeded = false
			} else {
				return
			}
			coreQuotaCPU += subResourceQuota.Spec.Hard.Cpu().Value()
			coreQuotaMemory += subResourceQuota.Spec.Hard.Memory().Value()
			subResourceQuota.Spec.Hard["cpu"] = *resource.NewQuantity(cpuDemand, resource.DecimalSI)
			subResourceQuota.Spec.Hard["memory"] = *resource.NewQuantity(memoryDemand, resource.BinarySI)
			t.clientset.CoreV1().ResourceQuotas(childNamespace.GetName()).Update(context.TODO(), subResourceQuota, metav1.UpdateOptions{})
		} else {
			subResourceQuota := &corev1.ResourceQuota{}
			subResourceQuota.Name = "sub-quota"
			subResourceQuota.Spec = corev1.ResourceQuotaSpec{
				Hard: map[corev1.ResourceName]resource.Quantity{
					"cpu":    cpuResource,
					"memory": memoryResource,
				},
			}
			t.clientset.CoreV1().ResourceQuotas(childNamespace.GetName()).Create(context.TODO(), subResourceQuota, metav1.CreateOptions{})
		}

		if !quotaExceeded {
			coreResourceQuota, _ := t.clientset.CoreV1().ResourceQuotas(labels["edge-net.io/tenant"]).Get(context.TODO(), "core-quota", metav1.GetOptions{})
			coreQuotaCPU += coreResourceQuota.Spec.Hard.Cpu().Value()
			coreQuotaMemory += coreResourceQuota.Spec.Hard.Memory().Value()
			coreResourceQuota.Spec.Hard["cpu"] = *resource.NewQuantity(coreQuotaCPU, resource.DecimalSI)
			coreResourceQuota.Spec.Hard["memory"] = *resource.NewQuantity(coreQuotaMemory, resource.BinarySI)
			t.clientset.CoreV1().ResourceQuotas(labels["edge-net.io/tenant"]).Update(context.TODO(), coreResourceQuota, metav1.UpdateOptions{})

			if roleRaw, err := t.clientset.RbacV1().Roles(subNamespace.GetNamespace()).List(context.TODO(), metav1.ListOptions{}); err == nil && subNamespace.Spec.Inheritance.RBAC {
				// TODO: Provide err information at the status
				for _, roleRow := range roleRaw.Items {
					role := roleRow.DeepCopy()
					role.SetNamespace(childNamespace.GetName())
					role.SetUID(types.UID(uuid.New().String()))
					t.clientset.RbacV1().Roles(childNamespace.GetName()).Create(context.TODO(), role, metav1.CreateOptions{})
				}
			}
			if roleBindingRaw, err := t.clientset.RbacV1().RoleBindings(subNamespace.GetNamespace()).List(context.TODO(), metav1.ListOptions{}); err == nil && subNamespace.Spec.Inheritance.RBAC {
				// TODO: Provide err information at the status
				for _, roleBindingRow := range roleBindingRaw.Items {
					roleBinding := roleBindingRow.DeepCopy()
					roleBinding.SetNamespace(childNamespace.GetName())
					roleBinding.SetUID(types.UID(uuid.New().String()))
					t.clientset.RbacV1().RoleBindings(childNamespace.GetName()).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
				}
			}
			if networkPolicyRaw, err := t.clientset.NetworkingV1().NetworkPolicies(subNamespace.GetNamespace()).List(context.TODO(), metav1.ListOptions{}); err == nil && subNamespace.Spec.Inheritance.NetworkPolicy {
				// TODO: Provide err information at the status
				for _, networkPolicyRow := range networkPolicyRaw.Items {
					networkPolicy := networkPolicyRow.DeepCopy()
					networkPolicy.SetNamespace(childNamespace.GetName())
					networkPolicy.SetUID(types.UID(uuid.New().String()))
					t.clientset.NetworkingV1().NetworkPolicies(childNamespace.GetName()).Create(context.TODO(), networkPolicy, metav1.CreateOptions{})
				}
			}
			subNamespace.Status.State = established
			subNamespace.Status.Message = []string{statusDict["subnamespace-ok"]}
		} else {
			// TODO: Error handling
			subNamespace.Status.State = failure
			subNamespace.Status.Message = []string{statusDict["quota-exceeded"]}
		}
	}
}

// ObjectDeleted is called when an object is deleted
func (t *Handler) ObjectDeleted(obj interface{}) {
	log.Info("SubNamespaceHandler.ObjectDeleted")
	// Delete the namespace created by subsidiary namespace, TBD.
}

// SetAsOwnerReference returns the subsidiary namespace as owner
func SetAsOwnerReference(tenant *corev1alpha.SubNamespace) []metav1.OwnerReference {
	// The following section makes subnamespace become the owner
	ownerReferences := []metav1.OwnerReference{}
	newSubNamespaceRef := *metav1.NewControllerRef(tenant, corev1alpha.SchemeGroupVersion.WithKind("SubNamespace"))
	takeControl := false
	newSubNamespaceRef.Controller = &takeControl
	ownerReferences = append(ownerReferences, newSubNamespaceRef)
	return ownerReferences
}

// RunExpiryController puts a procedure in place to turn accepted policies into not accepted
func (t *Handler) RunExpiryController() {
	var closestExpiry time.Time
	terminated := make(chan bool)
	newExpiry := make(chan time.Time)
	defer close(terminated)
	defer close(newExpiry)

	watchSubNamespace, err := t.edgenetClientset.CoreV1alpha().SubNamespaces("").Watch(context.TODO(), metav1.ListOptions{})
	if err == nil {
		watchEvents := func(watchSubNamespace watch.Interface, newExpiry *chan time.Time) {
			// Watch the events of subsidiary namespace object
			// Get events from watch interface
			for subNamespaceEvent := range watchSubNamespace.ResultChan() {
				// Get updated subsidiary namespace object
				updatedSubNamespace, status := subNamespaceEvent.Object.(*corev1alpha.SubNamespace)
				if status {
					if subNamespaceEvent.Type == "DELETED" {
						parentNamespace, _ := t.clientset.CoreV1().Namespaces().Get(context.TODO(), updatedSubNamespace.GetNamespace(), metav1.GetOptions{})
						parentNamespaceLabels := parentNamespace.GetLabels()
						if parentNamespaceLabels != nil && parentNamespaceLabels["edge-net.io/tenant"] != "" {
							if childNamespaceObj, err := t.clientset.CoreV1().Namespaces().Get(context.TODO(), fmt.Sprintf("%s-%s", parentNamespaceLabels["edge-net.io/tenant"], updatedSubNamespace.GetName()), metav1.GetOptions{}); err == nil {
								childNamespaceObjLabels := childNamespaceObj.GetLabels()
								if childNamespaceObjLabels != nil && childNamespaceObjLabels["edge-net.io/generated"] == "true" && childNamespaceObjLabels["edge-net.io/owner"] == fmt.Sprintf("%s-%s", updatedSubNamespace.GetNamespace(), updatedSubNamespace.GetName()) {
									coreResourceQuota, _ := t.clientset.CoreV1().ResourceQuotas(childNamespaceObjLabels["edge-net.io/tenant"]).Get(context.TODO(), "core-quota", metav1.GetOptions{})
									coreQuotaCPU := coreResourceQuota.Spec.Hard.Cpu().Value()
									coreQuotaMemory := coreResourceQuota.Spec.Hard.Memory().Value()
									if subResourceQuota, err := t.clientset.CoreV1().ResourceQuotas(childNamespaceObj.GetName()).Get(context.TODO(), "sub-quota", metav1.GetOptions{}); err == nil {
										coreQuotaCPU += subResourceQuota.Spec.Hard.Cpu().Value()
										coreQuotaMemory += subResourceQuota.Spec.Hard.Memory().Value()
									}
									coreResourceQuota.Spec.Hard["cpu"] = *resource.NewQuantity(coreQuotaCPU, resource.DecimalSI)
									coreResourceQuota.Spec.Hard["memory"] = *resource.NewQuantity(coreQuotaMemory, resource.BinarySI)
									t.clientset.CoreV1().ResourceQuotas(childNamespaceObjLabels["edge-net.io/tenant"]).Update(context.TODO(), coreResourceQuota, metav1.UpdateOptions{})

									log.Println(updatedSubNamespace.GetName())
									t.clientset.CoreV1().Namespaces().Delete(context.TODO(), childNamespaceObj.GetName(), metav1.DeleteOptions{})
								}
							}
						}
						continue
					}

					if updatedSubNamespace.Spec.Expiry != nil {
						*newExpiry <- updatedSubNamespace.Spec.Expiry.Time
					}
				}
			}
		}
		go watchEvents(watchSubNamespace, &newExpiry)
	} else {
		go t.RunExpiryController()
		terminated <- true
	}

infiniteLoop:
	for {
		// Wait on multiple channel operations
		select {
		case timeout := <-newExpiry:
			if closestExpiry.Sub(timeout) > 0 {
				closestExpiry = timeout
				log.Printf("ExpiryController: Closest expiry date is %v", closestExpiry)
			}
		case <-time.After(time.Until(closestExpiry)):
			subNamespaceRaw, err := t.edgenetClientset.CoreV1alpha().SubNamespaces("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				// TO-DO: Provide more information on error
				log.Println(err)
			}
			for _, subNamespaceRow := range subNamespaceRaw.Items {
				if subNamespaceRow.Spec.Expiry != nil && subNamespaceRow.Spec.Expiry.Time.Sub(time.Now()) <= 0 {
					t.edgenetClientset.CoreV1alpha().SubNamespaces(subNamespaceRow.GetNamespace()).Delete(context.TODO(), subNamespaceRow.GetName(), metav1.DeleteOptions{})

					parentNamespace, _ := t.clientset.CoreV1().Namespaces().Get(context.TODO(), subNamespaceRow.GetNamespace(), metav1.GetOptions{})
					parentNamespaceLabels := parentNamespace.GetLabels()
					if parentNamespaceLabels != nil && parentNamespaceLabels["edge-net.io/tenant"] != "" {
						if childNamespaceObj, err := t.clientset.CoreV1().Namespaces().Get(context.TODO(), fmt.Sprintf("%s-%s", parentNamespaceLabels["edge-net.io/tenant"], subNamespaceRow.GetName()), metav1.GetOptions{}); err == nil {
							childNamespaceObjLabels := childNamespaceObj.GetLabels()
							if childNamespaceObjLabels != nil && childNamespaceObjLabels["edge-net.io/generated"] == "true" && childNamespaceObjLabels["edge-net.io/owner"] == fmt.Sprintf("%s-%s", subNamespaceRow.GetNamespace(), subNamespaceRow.GetName()) {
								coreResourceQuota, _ := t.clientset.CoreV1().ResourceQuotas(childNamespaceObjLabels["edge-net.io/tenant"]).Get(context.TODO(), "core-quota", metav1.GetOptions{})
								coreQuotaCPU := coreResourceQuota.Spec.Hard.Cpu().Value()
								coreQuotaMemory := coreResourceQuota.Spec.Hard.Memory().Value()
								if subResourceQuota, err := t.clientset.CoreV1().ResourceQuotas(childNamespaceObj.GetName()).Get(context.TODO(), "sub-quota", metav1.GetOptions{}); err == nil {
									coreQuotaCPU += subResourceQuota.Spec.Hard.Cpu().Value()
									coreQuotaMemory += subResourceQuota.Spec.Hard.Memory().Value()
								}
								coreResourceQuota.Spec.Hard["cpu"] = *resource.NewQuantity(coreQuotaCPU, resource.DecimalSI)
								coreResourceQuota.Spec.Hard["memory"] = *resource.NewQuantity(coreQuotaMemory, resource.BinarySI)
								t.clientset.CoreV1().ResourceQuotas(childNamespaceObjLabels["edge-net.io/tenant"]).Update(context.TODO(), coreResourceQuota, metav1.UpdateOptions{})

								log.Println(subNamespaceRow.GetName())
								t.clientset.CoreV1().Namespaces().Delete(context.TODO(), childNamespaceObj.GetName(), metav1.DeleteOptions{})
							}
						}
					}
				}
			}
		case <-terminated:
			watchSubNamespace.Stop()
			break infiniteLoop
		}
	}
}
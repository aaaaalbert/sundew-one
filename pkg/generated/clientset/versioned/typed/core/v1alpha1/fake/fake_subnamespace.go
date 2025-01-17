/*
Copyright 2023 Contributors to the EdgeNet project.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSubNamespaces implements SubNamespaceInterface
type FakeSubNamespaces struct {
	Fake *FakeCoreV1alpha1
	ns   string
}

var subnamespacesResource = v1alpha1.SchemeGroupVersion.WithResource("subnamespaces")

var subnamespacesKind = v1alpha1.SchemeGroupVersion.WithKind("SubNamespace")

// Get takes name of the subNamespace, and returns the corresponding subNamespace object, and an error if there is any.
func (c *FakeSubNamespaces) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SubNamespace, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(subnamespacesResource, c.ns, name), &v1alpha1.SubNamespace{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubNamespace), err
}

// List takes label and field selectors, and returns the list of SubNamespaces that match those selectors.
func (c *FakeSubNamespaces) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SubNamespaceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(subnamespacesResource, subnamespacesKind, c.ns, opts), &v1alpha1.SubNamespaceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SubNamespaceList{ListMeta: obj.(*v1alpha1.SubNamespaceList).ListMeta}
	for _, item := range obj.(*v1alpha1.SubNamespaceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested subNamespaces.
func (c *FakeSubNamespaces) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(subnamespacesResource, c.ns, opts))

}

// Create takes the representation of a subNamespace and creates it.  Returns the server's representation of the subNamespace, and an error, if there is any.
func (c *FakeSubNamespaces) Create(ctx context.Context, subNamespace *v1alpha1.SubNamespace, opts v1.CreateOptions) (result *v1alpha1.SubNamespace, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(subnamespacesResource, c.ns, subNamespace), &v1alpha1.SubNamespace{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubNamespace), err
}

// Update takes the representation of a subNamespace and updates it. Returns the server's representation of the subNamespace, and an error, if there is any.
func (c *FakeSubNamespaces) Update(ctx context.Context, subNamespace *v1alpha1.SubNamespace, opts v1.UpdateOptions) (result *v1alpha1.SubNamespace, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(subnamespacesResource, c.ns, subNamespace), &v1alpha1.SubNamespace{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubNamespace), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSubNamespaces) UpdateStatus(ctx context.Context, subNamespace *v1alpha1.SubNamespace, opts v1.UpdateOptions) (*v1alpha1.SubNamespace, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(subnamespacesResource, "status", c.ns, subNamespace), &v1alpha1.SubNamespace{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubNamespace), err
}

// Delete takes name of the subNamespace and deletes it. Returns an error if one occurs.
func (c *FakeSubNamespaces) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(subnamespacesResource, c.ns, name, opts), &v1alpha1.SubNamespace{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSubNamespaces) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(subnamespacesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SubNamespaceList{})
	return err
}

// Patch applies the patch and returns the patched subNamespace.
func (c *FakeSubNamespaces) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SubNamespace, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(subnamespacesResource, c.ns, name, pt, data, subresources...), &v1alpha1.SubNamespace{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.SubNamespace), err
}

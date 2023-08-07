/*
Copyright The Kubernetes Authors.

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

// FakeSlices implements SliceInterface
type FakeSlices struct {
	Fake *FakeCoreV1alpha1
}

var slicesResource = v1alpha1.SchemeGroupVersion.WithResource("slices")

var slicesKind = v1alpha1.SchemeGroupVersion.WithKind("Slice")

// Get takes name of the slice, and returns the corresponding slice object, and an error if there is any.
func (c *FakeSlices) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Slice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(slicesResource, name), &v1alpha1.Slice{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Slice), err
}

// List takes label and field selectors, and returns the list of Slices that match those selectors.
func (c *FakeSlices) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SliceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(slicesResource, slicesKind, opts), &v1alpha1.SliceList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SliceList{ListMeta: obj.(*v1alpha1.SliceList).ListMeta}
	for _, item := range obj.(*v1alpha1.SliceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested slices.
func (c *FakeSlices) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(slicesResource, opts))
}

// Create takes the representation of a slice and creates it.  Returns the server's representation of the slice, and an error, if there is any.
func (c *FakeSlices) Create(ctx context.Context, slice *v1alpha1.Slice, opts v1.CreateOptions) (result *v1alpha1.Slice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(slicesResource, slice), &v1alpha1.Slice{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Slice), err
}

// Update takes the representation of a slice and updates it. Returns the server's representation of the slice, and an error, if there is any.
func (c *FakeSlices) Update(ctx context.Context, slice *v1alpha1.Slice, opts v1.UpdateOptions) (result *v1alpha1.Slice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(slicesResource, slice), &v1alpha1.Slice{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Slice), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSlices) UpdateStatus(ctx context.Context, slice *v1alpha1.Slice, opts v1.UpdateOptions) (*v1alpha1.Slice, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(slicesResource, "status", slice), &v1alpha1.Slice{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Slice), err
}

// Delete takes name of the slice and deletes it. Returns an error if one occurs.
func (c *FakeSlices) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(slicesResource, name, opts), &v1alpha1.Slice{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSlices) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(slicesResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SliceList{})
	return err
}

// Patch applies the patch and returns the patched slice.
func (c *FakeSlices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Slice, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(slicesResource, name, pt, data, subresources...), &v1alpha1.Slice{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Slice), err
}

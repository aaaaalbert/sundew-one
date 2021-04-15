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

package v1alpha

import (
	"context"
	"time"

	v1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/apps/v1alpha"
	scheme "github.com/EdgeNet-project/edgenet/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// UserRegistrationRequestsGetter has a method to return a UserRegistrationRequestInterface.
// A group's client should implement this interface.
type UserRegistrationRequestsGetter interface {
	UserRegistrationRequests(namespace string) UserRegistrationRequestInterface
}

// UserRegistrationRequestInterface has methods to work with UserRegistrationRequest resources.
type UserRegistrationRequestInterface interface {
	Create(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.CreateOptions) (*v1alpha.UserRegistrationRequest, error)
	Update(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.UpdateOptions) (*v1alpha.UserRegistrationRequest, error)
	UpdateStatus(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.UpdateOptions) (*v1alpha.UserRegistrationRequest, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha.UserRegistrationRequest, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha.UserRegistrationRequestList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha.UserRegistrationRequest, err error)
	UserRegistrationRequestExpansion
}

// userRegistrationRequests implements UserRegistrationRequestInterface
type userRegistrationRequests struct {
	client rest.Interface
	ns     string
}

// newUserRegistrationRequests returns a UserRegistrationRequests
func newUserRegistrationRequests(c *AppsV1alphaClient, namespace string) *userRegistrationRequests {
	return &userRegistrationRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the userRegistrationRequest, and returns the corresponding userRegistrationRequest object, and an error if there is any.
func (c *userRegistrationRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha.UserRegistrationRequest, err error) {
	result = &v1alpha.UserRegistrationRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of UserRegistrationRequests that match those selectors.
func (c *userRegistrationRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha.UserRegistrationRequestList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha.UserRegistrationRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested userRegistrationRequests.
func (c *userRegistrationRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a userRegistrationRequest and creates it.  Returns the server's representation of the userRegistrationRequest, and an error, if there is any.
func (c *userRegistrationRequests) Create(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.CreateOptions) (result *v1alpha.UserRegistrationRequest, err error) {
	result = &v1alpha.UserRegistrationRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(userRegistrationRequest).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a userRegistrationRequest and updates it. Returns the server's representation of the userRegistrationRequest, and an error, if there is any.
func (c *userRegistrationRequests) Update(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.UpdateOptions) (result *v1alpha.UserRegistrationRequest, err error) {
	result = &v1alpha.UserRegistrationRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		Name(userRegistrationRequest.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(userRegistrationRequest).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *userRegistrationRequests) UpdateStatus(ctx context.Context, userRegistrationRequest *v1alpha.UserRegistrationRequest, opts v1.UpdateOptions) (result *v1alpha.UserRegistrationRequest, err error) {
	result = &v1alpha.UserRegistrationRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		Name(userRegistrationRequest.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(userRegistrationRequest).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the userRegistrationRequest and deletes it. Returns an error if one occurs.
func (c *userRegistrationRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *userRegistrationRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("userregistrationrequests").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched userRegistrationRequest.
func (c *userRegistrationRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha.UserRegistrationRequest, err error) {
	result = &v1alpha.UserRegistrationRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("userregistrationrequests").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

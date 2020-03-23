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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha

import (
	internalinterfaces "headnode/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// AcceptableUsePolicies returns a AcceptableUsePolicyInformer.
	AcceptableUsePolicies() AcceptableUsePolicyInformer
	// Authorities returns a AuthorityInformer.
	Authorities() AuthorityInformer
	// AuthorityRequests returns a AuthorityRequestInformer.
	AuthorityRequests() AuthorityRequestInformer
	// EmailVerifications returns a EmailVerificationInformer.
	EmailVerifications() EmailVerificationInformer
	// Logins returns a LoginInformer.
	Logins() LoginInformer
	// NodeContributions returns a NodeContributionInformer.
	NodeContributions() NodeContributionInformer
	// SelectiveDeployments returns a SelectiveDeploymentInformer.
	SelectiveDeployments() SelectiveDeploymentInformer
	// Slices returns a SliceInformer.
	Slices() SliceInformer
	// Teams returns a TeamInformer.
	Teams() TeamInformer
	// Users returns a UserInformer.
	Users() UserInformer
	// UserRegistrationRequests returns a UserRegistrationRequestInformer.
	UserRegistrationRequests() UserRegistrationRequestInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// AcceptableUsePolicies returns a AcceptableUsePolicyInformer.
func (v *version) AcceptableUsePolicies() AcceptableUsePolicyInformer {
	return &acceptableUsePolicyInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Authorities returns a AuthorityInformer.
func (v *version) Authorities() AuthorityInformer {
	return &authorityInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// AuthorityRequests returns a AuthorityRequestInformer.
func (v *version) AuthorityRequests() AuthorityRequestInformer {
	return &authorityRequestInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// EmailVerifications returns a EmailVerificationInformer.
func (v *version) EmailVerifications() EmailVerificationInformer {
	return &emailVerificationInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Logins returns a LoginInformer.
func (v *version) Logins() LoginInformer {
	return &loginInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// NodeContributions returns a NodeContributionInformer.
func (v *version) NodeContributions() NodeContributionInformer {
	return &nodeContributionInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// SelectiveDeployments returns a SelectiveDeploymentInformer.
func (v *version) SelectiveDeployments() SelectiveDeploymentInformer {
	return &selectiveDeploymentInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Slices returns a SliceInformer.
func (v *version) Slices() SliceInformer {
	return &sliceInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Teams returns a TeamInformer.
func (v *version) Teams() TeamInformer {
	return &teamInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Users returns a UserInformer.
func (v *version) Users() UserInformer {
	return &userInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// UserRegistrationRequests returns a UserRegistrationRequestInformer.
func (v *version) UserRegistrationRequests() UserRegistrationRequestInformer {
	return &userRegistrationRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

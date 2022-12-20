//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterCache) DeepCopyInto(out *ClusterCache) {
	*out = *in
	if in.Characteristics != nil {
		in, out := &in.Characteristics, &out.Characteristics
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterCache.
func (in *ClusterCache) DeepCopy() *ClusterCache {
	if in == nil {
		return nil
	}
	out := new(ClusterCache)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterList.
func (in *ClusterList) DeepCopy() *ClusterList {
	if in == nil {
		return nil
	}
	out := new(ClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPreferences) DeepCopyInto(out *ClusterPreferences) {
	*out = *in
	if in.Allowlist != nil {
		in, out := &in.Allowlist, &out.Allowlist
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Denylist != nil {
		in, out := &in.Denylist, &out.Denylist
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPreferences.
func (in *ClusterPreferences) DeepCopy() *ClusterPreferences {
	if in == nil {
		return nil
	}
	out := new(ClusterPreferences)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	in.Preferences.DeepCopyInto(&out.Preferences)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Hierarchy) DeepCopyInto(out *Hierarchy) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Hierarchy.
func (in *Hierarchy) DeepCopy() *Hierarchy {
	if in == nil {
		return nil
	}
	out := new(Hierarchy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagerCache) DeepCopyInto(out *ManagerCache) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagerCache.
func (in *ManagerCache) DeepCopy() *ManagerCache {
	if in == nil {
		return nil
	}
	out := new(ManagerCache)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ManagerCache) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagerCacheList) DeepCopyInto(out *ManagerCacheList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ManagerCache, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagerCacheList.
func (in *ManagerCacheList) DeepCopy() *ManagerCacheList {
	if in == nil {
		return nil
	}
	out := new(ManagerCacheList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ManagerCacheList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagerCacheSpec) DeepCopyInto(out *ManagerCacheSpec) {
	*out = *in
	out.Hierarchy = in.Hierarchy
	if in.Cluster != nil {
		in, out := &in.Cluster, &out.Cluster
		*out = make([]ClusterCache, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagerCacheSpec.
func (in *ManagerCacheSpec) DeepCopy() *ManagerCacheSpec {
	if in == nil {
		return nil
	}
	out := new(ManagerCacheSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagerCacheStatus) DeepCopyInto(out *ManagerCacheStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagerCacheStatus.
func (in *ManagerCacheStatus) DeepCopy() *ManagerCacheStatus {
	if in == nil {
		return nil
	}
	out := new(ManagerCacheStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginReference) DeepCopyInto(out *OriginReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginReference.
func (in *OriginReference) DeepCopy() *OriginReference {
	if in == nil {
		return nil
	}
	out := new(OriginReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelectiveDeploymentAnchor) DeepCopyInto(out *SelectiveDeploymentAnchor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelectiveDeploymentAnchor.
func (in *SelectiveDeploymentAnchor) DeepCopy() *SelectiveDeploymentAnchor {
	if in == nil {
		return nil
	}
	out := new(SelectiveDeploymentAnchor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SelectiveDeploymentAnchor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelectiveDeploymentAnchorList) DeepCopyInto(out *SelectiveDeploymentAnchorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SelectiveDeploymentAnchor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelectiveDeploymentAnchorList.
func (in *SelectiveDeploymentAnchorList) DeepCopy() *SelectiveDeploymentAnchorList {
	if in == nil {
		return nil
	}
	out := new(SelectiveDeploymentAnchorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SelectiveDeploymentAnchorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelectiveDeploymentAnchorSpec) DeepCopyInto(out *SelectiveDeploymentAnchorSpec) {
	*out = *in
	out.OriginRef = in.OriginRef
	if in.ClusterAffinity != nil {
		in, out := &in.ClusterAffinity, &out.ClusterAffinity
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.FederationManagerNames != nil {
		in, out := &in.FederationManagerNames, &out.FederationManagerNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelectiveDeploymentAnchorSpec.
func (in *SelectiveDeploymentAnchorSpec) DeepCopy() *SelectiveDeploymentAnchorSpec {
	if in == nil {
		return nil
	}
	out := new(SelectiveDeploymentAnchorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelectiveDeploymentAnchorStatus) DeepCopyInto(out *SelectiveDeploymentAnchorStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelectiveDeploymentAnchorStatus.
func (in *SelectiveDeploymentAnchorStatus) DeepCopy() *SelectiveDeploymentAnchorStatus {
	if in == nil {
		return nil
	}
	out := new(SelectiveDeploymentAnchorStatus)
	in.DeepCopyInto(out)
	return out
}

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha

import (
	v1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/apps/v1alpha"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SelectiveDeploymentLister helps list SelectiveDeployments.
// All objects returned here must be treated as read-only.
type SelectiveDeploymentLister interface {
	// List lists all SelectiveDeployments in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha.SelectiveDeployment, err error)
	// SelectiveDeployments returns an object that can list and get SelectiveDeployments.
	SelectiveDeployments(namespace string) SelectiveDeploymentNamespaceLister
	SelectiveDeploymentListerExpansion
}

// selectiveDeploymentLister implements the SelectiveDeploymentLister interface.
type selectiveDeploymentLister struct {
	indexer cache.Indexer
}

// NewSelectiveDeploymentLister returns a new SelectiveDeploymentLister.
func NewSelectiveDeploymentLister(indexer cache.Indexer) SelectiveDeploymentLister {
	return &selectiveDeploymentLister{indexer: indexer}
}

// List lists all SelectiveDeployments in the indexer.
func (s *selectiveDeploymentLister) List(selector labels.Selector) (ret []*v1alpha.SelectiveDeployment, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.SelectiveDeployment))
	})
	return ret, err
}

// SelectiveDeployments returns an object that can list and get SelectiveDeployments.
func (s *selectiveDeploymentLister) SelectiveDeployments(namespace string) SelectiveDeploymentNamespaceLister {
	return selectiveDeploymentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SelectiveDeploymentNamespaceLister helps list and get SelectiveDeployments.
// All objects returned here must be treated as read-only.
type SelectiveDeploymentNamespaceLister interface {
	// List lists all SelectiveDeployments in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha.SelectiveDeployment, err error)
	// Get retrieves the SelectiveDeployment from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha.SelectiveDeployment, error)
	SelectiveDeploymentNamespaceListerExpansion
}

// selectiveDeploymentNamespaceLister implements the SelectiveDeploymentNamespaceLister
// interface.
type selectiveDeploymentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SelectiveDeployments in the indexer for a given namespace.
func (s selectiveDeploymentNamespaceLister) List(selector labels.Selector) (ret []*v1alpha.SelectiveDeployment, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.SelectiveDeployment))
	})
	return ret, err
}

// Get retrieves the SelectiveDeployment from the indexer for a given namespace and name.
func (s selectiveDeploymentNamespaceLister) Get(name string) (*v1alpha.SelectiveDeployment, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha.Resource("selectivedeployment"), name)
	}
	return obj.(*v1alpha.SelectiveDeployment), nil
}

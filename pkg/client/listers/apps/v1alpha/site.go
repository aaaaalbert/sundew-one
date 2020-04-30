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
	v1alpha "headnode/pkg/apis/apps/v1alpha"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SiteLister helps list Sites.
type SiteLister interface {
	// List lists all Sites in the indexer.
	List(selector labels.Selector) (ret []*v1alpha.Site, err error)
	// Sites returns an object that can list and get Sites.
	Sites(namespace string) SiteNamespaceLister
	SiteListerExpansion
}

// siteLister implements the SiteLister interface.
type siteLister struct {
	indexer cache.Indexer
}

// NewSiteLister returns a new SiteLister.
func NewSiteLister(indexer cache.Indexer) SiteLister {
	return &siteLister{indexer: indexer}
}

// List lists all Sites in the indexer.
func (s *siteLister) List(selector labels.Selector) (ret []*v1alpha.Site, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.Site))
	})
	return ret, err
}

// Sites returns an object that can list and get Sites.
func (s *siteLister) Sites(namespace string) SiteNamespaceLister {
	return siteNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SiteNamespaceLister helps list and get Sites.
type SiteNamespaceLister interface {
	// List lists all Sites in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha.Site, err error)
	// Get retrieves the Site from the indexer for a given namespace and name.
	Get(name string) (*v1alpha.Site, error)
	SiteNamespaceListerExpansion
}

// siteNamespaceLister implements the SiteNamespaceLister
// interface.
type siteNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Sites in the indexer for a given namespace.
func (s siteNamespaceLister) List(selector labels.Selector) (ret []*v1alpha.Site, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.Site))
	})
	return ret, err
}

// Get retrieves the Site from the indexer for a given namespace and name.
func (s siteNamespaceLister) Get(name string) (*v1alpha.Site, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha.Resource("site"), name)
	}
	return obj.(*v1alpha.Site), nil
}

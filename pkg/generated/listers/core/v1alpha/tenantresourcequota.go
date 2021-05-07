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
	v1alpha "github.com/EdgeNet-project/edgenet/pkg/apis/core/v1alpha"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// TenantResourceQuotaLister helps list TenantResourceQuotas.
// All objects returned here must be treated as read-only.
type TenantResourceQuotaLister interface {
	// List lists all TenantResourceQuotas in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha.TenantResourceQuota, err error)
	// Get retrieves the TenantResourceQuota from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha.TenantResourceQuota, error)
	TenantResourceQuotaListerExpansion
}

// tenantResourceQuotaLister implements the TenantResourceQuotaLister interface.
type tenantResourceQuotaLister struct {
	indexer cache.Indexer
}

// NewTenantResourceQuotaLister returns a new TenantResourceQuotaLister.
func NewTenantResourceQuotaLister(indexer cache.Indexer) TenantResourceQuotaLister {
	return &tenantResourceQuotaLister{indexer: indexer}
}

// List lists all TenantResourceQuotas in the indexer.
func (s *tenantResourceQuotaLister) List(selector labels.Selector) (ret []*v1alpha.TenantResourceQuota, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha.TenantResourceQuota))
	})
	return ret, err
}

// Get retrieves the TenantResourceQuota from the index for a given name.
func (s *tenantResourceQuotaLister) Get(name string) (*v1alpha.TenantResourceQuota, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha.Resource("tenantresourcequota"), name)
	}
	return obj.(*v1alpha.TenantResourceQuota), nil
}
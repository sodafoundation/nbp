// Copyright 2018 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"strings"

	"github.com/sodafoundation/nbp/service-broker/pkg/model"
	"github.com/sodafoundation/nbp/service-broker/pkg/store/etcd"
)

// NewStore function can perform some initialization work of different databases.
func NewStore(driver, endpoint string) Store {
	switch driver {
	case "etcd":
		return etcd.NewStore(strings.Split(endpoint, ","))
	default:
		return etcd.NewStore(strings.Split(endpoint, ","))
	}
}

// Store is an interface to be implemented by types capable of handling
// persistence for other broker-related types
type Store interface {
	// SetServiceClass persists the given service class to the underlying storage
	SetServiceClass(service *model.ServiceClassSpec) error
	// GetServiceClass retrieves a persisted instance from the underlying storage by
	// instance id
	GetServiceClass(serviceID string) (*model.ServiceClassSpec, bool, error)
	// ListServiceClasses retrieves all persisted service classes from the underlying storage
	ListServiceClasses() ([]*model.ServiceClassSpec, error)
	// DeleteServiceClass deletes a persisted service class from the underlying storage by
	// service id
	DeleteServiceClass(serviceID string) (bool, error)

	// SetInstance persists the given instance to the underlying storage
	SetInstance(instance *model.ServiceInstanceSpec) error
	// GetInstance retrieves a persisted instance from the underlying storage by
	// instance id
	GetInstance(instanceID string) (*model.ServiceInstanceSpec, bool, error)
	// ListInstances retrieves all persisted instances from the underlying storage
	ListInstances() ([]*model.ServiceInstanceSpec, error)
	// DeleteInstance deletes a persisted instance from the underlying storage by
	// instance id
	DeleteInstance(instanceID string) (bool, error)

	// SetBinding persists the given binding to the underlying storage
	SetBinding(binding *model.ServiceBindingSpec) error
	// GetBinding retrieves a persisted instance from the underlying storage by
	// binding id and instance id
	GetBinding(bindingID, instanceID string) (*model.ServiceBindingSpec, bool, error)
	// ListBindings retrieves all persisted instance bindings from the underlying
	// storage
	ListBindings(instanceID string) ([]*model.ServiceBindingSpec, error)
	// DeleteBinding deletes a persisted binding from the underlying storage by
	// binding id and instance id
	DeleteBinding(bindingID, instanceID string) (bool, error)
}

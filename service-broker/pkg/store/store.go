// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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

	"github.com/opensds/nbp/service-broker/pkg/model"
	"github.com/opensds/nbp/service-broker/pkg/store/etcd"
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
	// SetInstance persists the given instance to the underlying storage
	SetInstance(instance *model.ServiceInstance) error
	// GetInstance retrieves a persisted instance from the underlying storage by
	// instance id
	GetInstance(instanceID string) (*model.ServiceInstance, bool, error)
	// ListInstances retrieves all persisted instances from the underlying storage
	ListInstances() ([]*model.ServiceInstance, error)
	// DeleteInstance deletes a persisted instance from the underlying storage by
	// instance id
	DeleteInstance(instanceID string) (bool, error)
	// SetBinding persists the given binding to the underlying storage
	SetBinding(binding *model.ServiceBinding) error
	// GetBinding retrieves a persisted instance from the underlying storage by
	// binding id and instance id
	GetBinding(bindingID, instanceID string) (*model.ServiceBinding, bool, error)
	// ListBindings retrieves all persisted instance bindings from the underlying
	// storage
	ListBindings(instanceID string) ([]*model.ServiceBinding, error)
	// DeleteBinding deletes a persisted binding from the underlying storage by
	// binding id and instance id
	DeleteBinding(bindingID, instanceID string) (bool, error)
}

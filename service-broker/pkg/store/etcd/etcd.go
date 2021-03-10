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

package etcd

import (
	"encoding/json"
	"errors"

	log "github.com/golang/glog"
	client "github.com/sodafoundation/api/pkg/db/drivers/etcd"
	"github.com/sodafoundation/nbp/service-broker/pkg/model"
)

// NewStore
func NewStore(edps []string) *etcdStore {
	return &etcdStore{
		clientInterface: client.Init(edps),
	}
}

// etcdStore
type etcdStore struct {
	clientInterface
}

type clientInterface interface {
	Create(req *client.Request) *client.Response

	Get(req *client.Request) *client.Response

	List(req *client.Request) *client.Response

	Update(req *client.Request) *client.Response

	Delete(req *client.Request) *client.Response
}

// SetServiceClass persists the given service class to the etcd storage
func (es *etcdStore) SetServiceClass(service *model.ServiceClassSpec) error {
	if service.ID == "" {
		return errors.New("Service id can NOT be empty!")
	}
	reqBody, err := json.Marshal(service)
	if err != nil {
		return err
	}

	dbReq := &client.Request{
		Url:     "/v2/service_classes/" + service.ID,
		Content: string(reqBody),
	}
	dbRes := es.Create(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When create service class in db:", dbRes.Error)
		return errors.New(dbRes.Error)
	}

	return nil
}

// GetServiceClass retrieves a persisted service class from the etcd storage by
// service id
func (es *etcdStore) GetServiceClass(serviceID string) (*model.ServiceClassSpec, bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_classes/" + serviceID,
	}
	dbRes := es.Get(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When get service class in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	var service = &model.ServiceClassSpec{}
	if err := json.Unmarshal([]byte(dbRes.Message[0]), service); err != nil {
		log.Error("When parsing service class in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	return service, true, nil
}

// ListServiceClasses retrieves all persisted service classes from the etcd storage
func (es *etcdStore) ListServiceClasses() ([]*model.ServiceClassSpec, error) {
	dbReq := &client.Request{
		Url: "/v2/service_classes",
	}
	dbRes := es.List(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When list service classes in db:", dbRes.Error)
		return nil, errors.New(dbRes.Error)
	}

	var services = []*model.ServiceClassSpec{}
	if len(dbRes.Message) == 0 {
		return services, nil
	}
	for _, msg := range dbRes.Message {
		var service = &model.ServiceClassSpec{}
		if err := json.Unmarshal([]byte(msg), service); err != nil {
			log.Error("When parsing service class in db:", dbRes.Error)
			return nil, errors.New(dbRes.Error)
		}
		services = append(services, service)
	}

	return services, nil
}

// DeleteServiceClass deletes a persisted service class from the etcd storage by
// service id
func (es *etcdStore) DeleteServiceClass(serviceID string) (bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_classes/" + serviceID,
	}
	dbRes := es.Delete(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When delete service class in db:", dbRes.Error)
		return false, errors.New(dbRes.Error)
	}

	return true, nil
}

// SetInstance persists the given instance to the etcd storage
func (es *etcdStore) SetInstance(instance *model.ServiceInstanceSpec) error {
	if instance.ID == "" {
		return errors.New("Instance id can NOT be empty!")
	}
	reqBody, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	dbReq := &client.Request{
		Url:     "/v2/service_instances/" + instance.ID,
		Content: string(reqBody),
	}
	dbRes := es.Create(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When create instance in db:", dbRes.Error)
		return errors.New(dbRes.Error)
	}

	return nil
}

// GetInstance retrieves a persisted instance from the etcd storage by
// instance id
func (es *etcdStore) GetInstance(instanceID string) (*model.ServiceInstanceSpec, bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances/" + instanceID,
	}
	dbRes := es.Get(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When get instance in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	var instance = &model.ServiceInstanceSpec{}
	if err := json.Unmarshal([]byte(dbRes.Message[0]), instance); err != nil {
		log.Error("When parsing instance in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	return instance, true, nil
}

// ListInstances retrieves all persisted instances from the etcd storage
func (es *etcdStore) ListInstances() ([]*model.ServiceInstanceSpec, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances",
	}
	dbRes := es.List(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When list instances in db:", dbRes.Error)
		return nil, errors.New(dbRes.Error)
	}

	var instances = []*model.ServiceInstanceSpec{}
	if len(dbRes.Message) == 0 {
		return instances, nil
	}
	for _, msg := range dbRes.Message {
		var instance = &model.ServiceInstanceSpec{}
		if err := json.Unmarshal([]byte(msg), instance); err != nil {
			log.Error("When parsing instance in db:", dbRes.Error)
			return nil, errors.New(dbRes.Error)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// DeleteInstance deletes a persisted instance from the etcd storage by
// instance id
func (es *etcdStore) DeleteInstance(instanceID string) (bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances/" + instanceID,
	}
	dbRes := es.Delete(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When delete instance in db:", dbRes.Error)
		return false, errors.New(dbRes.Error)
	}

	return true, nil
}

// CreateBinding persists the given binding to the etcd storage
func (es *etcdStore) SetBinding(binding *model.ServiceBindingSpec) error {
	if binding.ID == "" || binding.InstanceID == "" {
		return errors.New("Instance id or binding id can NOT be empty!")
	}
	reqBody, err := json.Marshal(binding)
	if err != nil {
		return err
	}

	dbReq := &client.Request{
		Url:     "/v2/service_instances/" + binding.InstanceID + "/service_bindings/" + binding.ID,
		Content: string(reqBody),
	}
	dbRes := es.Create(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When create binding in db:", dbRes.Error)
		return errors.New(dbRes.Error)
	}

	return nil
}

// GetBinding retrieves a persisted instance from the etcd storage by
// binding id
func (es *etcdStore) GetBinding(bindingID, instanceID string) (*model.ServiceBindingSpec, bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances/" + instanceID + "/service_bindings/" + bindingID,
	}
	dbRes := es.Get(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When get binding in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	var binding = &model.ServiceBindingSpec{}
	if err := json.Unmarshal([]byte(dbRes.Message[0]), binding); err != nil {
		log.Error("When parsing binding in db:", dbRes.Error)
		return nil, false, errors.New(dbRes.Error)
	}

	return binding, true, nil
}

// ListBindings retrieves all persisted instance bindings from the etcd
// storage
func (es *etcdStore) ListBindings(instanceID string) ([]*model.ServiceBindingSpec, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances/" + instanceID + "/service_bindings",
	}
	dbRes := es.List(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When list bindings in db:", dbRes.Error)
		return nil, errors.New(dbRes.Error)
	}

	var bindings = []*model.ServiceBindingSpec{}
	if len(dbRes.Message) == 0 {
		return bindings, nil
	}
	for _, msg := range dbRes.Message {
		var binding = &model.ServiceBindingSpec{}
		if err := json.Unmarshal([]byte(msg), binding); err != nil {
			log.Error("When parsing binding in db:", dbRes.Error)
			return nil, errors.New(dbRes.Error)
		}
		bindings = append(bindings, binding)
	}

	return bindings, nil
}

// DeleteBinding deletes a persisted binding from the etcd storage by
// binding id
func (es *etcdStore) DeleteBinding(bindingID, instanceID string) (bool, error) {
	dbReq := &client.Request{
		Url: "/v2/service_instances/" + instanceID + "/service_bindings/" + bindingID,
	}
	dbRes := es.Delete(dbReq)
	if dbRes.Status != "Success" {
		log.Error("When delete binding in db:", dbRes.Error)
		return false, errors.New(dbRes.Error)
	}

	return true, nil
}

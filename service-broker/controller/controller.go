/*
Copyright 2016 The Kubernetes Authors.
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

package controller

import (
	"fmt"
	"log"
	"sync"

	"errors"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/controller"
	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/brokerapi"
	sdsController "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/opensds/pkg/model"
)

type openSDSServiceInstance struct {
	Name       string
	Credential *brokerapi.Credential
}

type openSDSController struct {
	Endpoint string

	rwMutex     sync.RWMutex
	instanceMap map[string]*openSDSServiceInstance
}

// CreateController creates an instance of an OpenSDS service broker controller.
func CreateController(edp string) controller.Controller {
	var instanceMap = make(map[string]*openSDSServiceInstance)

	return &openSDSController{
		Endpoint:    edp,
		instanceMap: instanceMap,
	}
}

func (c *openSDSController) Catalog() (*brokerapi.Catalog, error) {
	prfs, err := sdsController.GetClient(c.Endpoint).ListProfiles()
	if err != nil {
		return nil, err
	}

	var plans = []brokerapi.ServicePlan{}
	for _, prf := range prfs {
		plan := brokerapi.ServicePlan{
			Name:        prf.GetName(),
			ID:          prf.GetId(),
			Description: prf.GetDescription(),
			Metadata:    prf.Extra,
			Free:        true,
		}
		plans = append(plans, plan)
	}

	return &brokerapi.Catalog{
		Services: []*brokerapi.Service{
			{
				Name:        "opensds-service",
				ID:          "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
				Description: "Policy based storage service",
				Plans:       plans,
				Bindable:    true,
			},
		},
	}, nil
}

func (c *openSDSController) GetServiceInstanceLastOperation(
	instanceID, serviceID, planID, operation string,
) (*brokerapi.LastOperationResponse, error) {
	return nil, fmt.Errorf("Not implemented!")
}

func (c *openSDSController) CreateServiceInstance(
	instanceID string,
	req *brokerapi.CreateServiceInstanceRequest,
) (*brokerapi.CreateServiceInstanceResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	fmt.Printf("instanceId: %s, Bind req:%v", instanceID, *req)

	var in = new(model.VolumeSpec)
	if nameInterface, ok := req.Parameters["name"]; ok {
		in.Name = nameInterface.(string)
	}
	if despInterface, ok := req.Parameters["description"]; ok {
		in.Description = despInterface.(string)
	}
	if capInterface, ok := req.Parameters["capacity"]; ok {
		in.Size = int64(capInterface.(float64))
	}

	if instance, ok := c.instanceMap[instanceID]; ok {
		_, ok := (*instance.Credential)["volumeId"]
		if ok {
			log.Printf("Instance %s already exist!", instanceID)
			return &brokerapi.CreateServiceInstanceResponse{}, nil
		}
	}

	vol, err := sdsController.GetClient(c.Endpoint).CreateVolume(in)
	if err != nil {
		return nil, err
	}

	c.instanceMap[instanceID] = &openSDSServiceInstance{
		Name: instanceID,
		Credential: &brokerapi.Credential{
			"volumeId": vol.GetId(),
			"image":    "OPENSDS:" + vol.GetName() + ":" + vol.GetId(),
		},
	}

	log.Printf("Created User Provided Service Instance:\n%v\n",
		c.instanceMap[instanceID])
	return &brokerapi.CreateServiceInstanceResponse{}, nil
}

func (c *openSDSController) RemoveServiceInstance(
	instanceID, serviceID, planID string,
	acceptsIncomplete bool,
) (*brokerapi.DeleteServiceInstanceResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	instance, ok := c.instanceMap[instanceID]
	if !ok {
		msg := fmt.Sprintf("No such instance %s exited!", instanceID)
		return nil, errors.New(msg)
	}
	volInterface, ok := (*instance.Credential)["volumeId"]
	if !ok {
		return &brokerapi.DeleteServiceInstanceResponse{}, nil
	}

	if err := sdsController.GetClient(c.Endpoint).
		DeleteVolume(volInterface.(string), nil); err != nil {
		return nil, err
	}
	delete(c.instanceMap, instanceID)

	return &brokerapi.DeleteServiceInstanceResponse{}, nil
}

func (c *openSDSController) Bind(
	instanceID, bindingID string,
	req *brokerapi.BindingRequest,
) (*brokerapi.CreateServiceBindingResponse, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()
	fmt.Printf("instanceId: %s, bindingId: %s, Bind req:%v", instanceID, bindingID, *req)
	instance, ok := c.instanceMap[instanceID]
	if !ok {
		return nil, fmt.Errorf("No such instance %s exited!", instanceID)
	}
	cred := instance.Credential
	return &brokerapi.CreateServiceBindingResponse{Credentials: *cred}, nil
}

func (c *openSDSController) UnBind(
	instanceID, bindingID, serviceID, planID string,
) error {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	return nil
}


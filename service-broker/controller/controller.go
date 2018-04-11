/*
Copyright 2016 The Kubernetes Authors.
Copyright (c) 2016 Huawei Technologies Co., Ltd. All Rights Reserved.

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
	"net/http"
	"sync"

	"github.com/golang/glog"
	sdsClient "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

type opensdsServiceInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

type opensdsController struct {
	Endpoint    string
	async       bool
	rwMutex     sync.RWMutex
	instanceMap map[string]*opensdsServiceInstance
}

// NewController creates an instance of an OpenSDS service broker controller.
func NewController(edp string) *opensdsController {
	var instanceMap = make(map[string]*opensdsServiceInstance)

	return &opensdsController{
		Endpoint:    edp,
		instanceMap: instanceMap,
	}
}

func truePtr() *bool {
	a := true
	return &a
}

func falsePtr() *bool {
	b := false
	return &b
}

func (c *opensdsController) GetCatalog(ctx *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog business logic goes here
	response := &broker.CatalogResponse{}

	prfs, err := sdsClient.NewClient(&sdsClient.Config{Endpoint: c.Endpoint}).ListProfiles()
	if err != nil {
		return nil, err
	}

	var plans = []osb.Plan{}
	for _, prf := range prfs {
		plan := osb.Plan{
			Name:        prf.Name,
			ID:          prf.Id,
			Description: prf.Description,
			Metadata:    prf.Extras,
			Free:        truePtr(),
		}
		plans = append(plans, plan)
	}

	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "opensds-service",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
				Description:   "Policy based storage service",
				Bindable:      true,
				PlanUpdatable: falsePtr(),
				Plans:         plans,
			},
		},
	}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (c *opensdsController) LastOperation(
	request *osb.LastOperationRequest,
	ctx *broker.RequestContext,
) (*broker.LastOperationResponse, error) {
	return nil, fmt.Errorf("Not implemented!")
}

func (c *opensdsController) Provision(
	request *osb.ProvisionRequest,
	ctx *broker.RequestContext,
) (*broker.ProvisionResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	response := broker.ProvisionResponse{}

	var in = new(model.VolumeSpec)
	if nameInterface, ok := request.Parameters["name"]; ok {
		in.Name = nameInterface.(string)
	}
	if despInterface, ok := request.Parameters["description"]; ok {
		in.Description = despInterface.(string)
	}
	if capInterface, ok := request.Parameters["capacity"]; ok {
		in.Size = int64(capInterface.(float64))
	}

	if _, ok := c.instanceMap[request.InstanceID]; ok {
		glog.Infof("Instance %s already exist!\n", request.InstanceID)
		return &response, nil
	}

	vol, err := sdsClient.NewClient(&sdsClient.Config{Endpoint: c.Endpoint}).CreateVolume(in)
	if err != nil {
		return nil, err
	}

	c.instanceMap[request.InstanceID] = &opensdsServiceInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}
	c.instanceMap[request.InstanceID].Params["volumeID"] = vol.Id

	glog.Infof("Created OpenSDS Service Instance:\n%v\n",
		c.instanceMap[request.InstanceID])

	if request.AcceptsIncomplete {
		response.Async = c.async
	}

	return &response, nil
}

func (c *opensdsController) Update(
	request *osb.UpdateInstanceRequest,
	ctx *broker.RequestContext,
) (*broker.UpdateInstanceResponse, error) {
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = c.async
	}

	return &response, nil
}

func (c *opensdsController) Deprovision(
	request *osb.DeprovisionRequest,
	ctx *broker.RequestContext,
) (*broker.DeprovisionResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	response := broker.DeprovisionResponse{}

	if _, ok := c.instanceMap[request.InstanceID]; !ok {
		return nil, fmt.Errorf("No such instance %s exited!", request.InstanceID)
	}

	if err := sdsClient.NewClient(&sdsClient.Config{Endpoint: c.Endpoint}).
		DeleteVolume(request.InstanceID, nil); err != nil {
		return nil, err
	}
	delete(c.instanceMap, request.InstanceID)

	if request.AcceptsIncomplete {
		response.Async = c.async
	}

	return &response, nil
}

func (c *opensdsController) Bind(
	request *osb.BindRequest,
	ctx *broker.RequestContext,
) (*broker.BindResponse, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	instance, ok := c.instanceMap[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: instance.Params,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = c.async
	}

	return &response, nil
}

func (c *opensdsController) Unbind(
	request *osb.UnbindRequest,
	ctx *broker.RequestContext,
) (*broker.UnbindResponse, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (c *opensdsController) ValidateBrokerAPIVersion(version string) error {
	return nil
}

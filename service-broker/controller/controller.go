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
	"github.com/opensds/nbp/csi/util"
	sdsClient "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

const (
	defaultSnapshotPlan    = "787c9322-3d92-11e8-8cb3-4f1353df06c1"
	defaultReplicationPlan = "b466e4ce-6cb2-11e8-a580-bb2b8754c9de"
)

const (
	noauthAuthType   = "noauth"
	keystoneAuthType = "keystone"
)

const (
	secondaryPrefix = "secondary-"
)

var (
	supportedPlanList []string
)

type opensdsServiceInstance struct {
	ID, ServiceID, PlanID string
	Params                map[string]interface{}
}

type opensdsServiceBinding struct {
	ID, InstanceID, ServiceID, PlanID string
	BindResource                      *osb.BindResource
	Params                            map[string]interface{}
}

type opensdsController struct {
	*sdsClient.Client

	async       bool
	rwMutex     sync.RWMutex
	instanceMap map[string]*opensdsServiceInstance
	bindingMap  map[string]*opensdsServiceBinding
}

// NewController creates an instance of an OpenSDS service broker controller.
func NewController(edp, auth string) *opensdsController {
	var authOption sdsClient.AuthOptions
	switch auth {
	case keystoneAuthType:
		authOption = LoadKeystoneAuthOptionsFromEnv()
		break
	default:
		authOption = LoadNoAuthOptionsFromEnv()
	}

	cli := sdsClient.NewClient(&sdsClient.Config{
		Endpoint:    edp,
		AuthOptions: authOption,
	})
	return &opensdsController{
		Client:      cli,
		instanceMap: make(map[string]*opensdsServiceInstance),
		bindingMap:  make(map[string]*opensdsServiceBinding),
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

func validatePlanID(planID string) bool {
	for _, v := range supportedPlanList {
		if v == planID {
			return true
		}
	}
	return false
}

func (c *opensdsController) GetCatalog(ctx *broker.RequestContext) (*broker.CatalogResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	// Clean the supportedPlanList and insert default snapshot plan.
	supportedPlanList = supportedPlanList[:0]
	supportedPlanList = append(supportedPlanList, defaultSnapshotPlan)

	response := &broker.CatalogResponse{}

	prfs, err := c.Client.ListProfiles()
	if err != nil {
		errMsg := fmt.Sprint("Broker error:", err)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: &errMsg,
		}
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
		supportedPlanList = append(supportedPlanList, prf.Id)
	}

	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "volume-service",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
				Description:   "Policy based volume provision service",
				Bindable:      true,
				PlanUpdatable: falsePtr(),
				Plans:         plans,
			},
			{
				Name:          "volume-snapshot-service",
				ID:            "434ba788-3d92-11e8-8712-1740dc7b3f46",
				Description:   "Policy based volume snapshot service",
				Bindable:      false,
				PlanUpdatable: falsePtr(),
				Plans: []osb.Plan{
					{
						Name:        "default-snapshot-plan",
						ID:          defaultSnapshotPlan,
						Description: "This is the default snapshot plan",
						Free:        truePtr(),
					},
				},
			},
			{
				Name:          "volume-replication-service",
				ID:            "ef136c64-6cb2-11e8-802e-3fdf92f7654d",
				Description:   "Policy based volume replication service",
				Bindable:      true,
				PlanUpdatable: falsePtr(),
				Plans: []osb.Plan{
					{
						Name:        "default-replication-plan",
						ID:          defaultReplicationPlan,
						Description: "This is the default replication plan",
						Free:        truePtr(),
					},
				},
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
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	return &broker.LastOperationResponse{}, nil
}

func (c *opensdsController) Provision(
	request *osb.ProvisionRequest,
	ctx *broker.RequestContext,
) (*broker.ProvisionResponse, error) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	response := broker.ProvisionResponse{}

	if _, ok := c.instanceMap[request.InstanceID]; ok {
		glog.Infof("Instance %s already exist!\n", request.InstanceID)
		response.Exists = true
		return &response, nil
	}
	if !validatePlanID(request.PlanID) {
		errMsg := fmt.Sprintf("PlanID (%s) is not supported!", request.PlanID)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}
	instance := &opensdsServiceInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}
	switch request.PlanID {
	case defaultSnapshotPlan:
		volInterface, ok := request.Parameters["volumeID"]
		if !ok {
			errMsg := fmt.Sprint("volumeID not found in provision request params!")
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: &errMsg,
			}
		}

		var in = &model.VolumeSnapshotSpec{
			VolumeId: volInterface.(string),
		}
		if nameInterface, ok := request.Parameters["name"]; ok {
			in.Name = nameInterface.(string)
		}
		if despInterface, ok := request.Parameters["description"]; ok {
			in.Description = despInterface.(string)
		}

		snp, err := c.Client.CreateVolumeSnapshot(in)
		if err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}

		instance.Params["snapshotID"] = snp.Id
	case defaultReplicationPlan:
		volInterface, ok := request.Parameters["volumeID"]
		if !ok {
			errMsg := fmt.Sprint("volumeID not found in provision request params!")
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: &errMsg,
			}
		}

		// Step 1: Check if the primary volume exists
		vol, err := c.Client.GetVolume(volInterface.(string))
		if err != nil || vol.Status != "available" {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
		// Step 2: Create a secondary volume
		volumeBody := &model.VolumeSpec{
			Name:             secondaryPrefix + vol.Name,
			Size:             vol.Size,
			AvailabilityZone: util.OpensdsDefaultSecondaryAZ,
		}
		sVol, err := c.Client.CreateVolume(volumeBody)
		if err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
		// Step 3: Create a replication
		replicaBody := &model.ReplicationSpec{
			PrimaryVolumeId:   vol.Id,
			SecondaryVolumeId: sVol.Id,
			ReplicationMode:   model.ReplicationModeSync,
			ReplicationPeriod: 0,
		}
		if nameInterface, ok := request.Parameters["name"]; ok {
			replicaBody.Name = nameInterface.(string)
		}
		if despInterface, ok := request.Parameters["description"]; ok {
			replicaBody.Description = despInterface.(string)
		}
		replicaResp, err := c.Client.CreateReplication(replicaBody)
		if err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}

		instance.Params["secondaryVolumeID"] = sVol.Id
		instance.Params["replicationID"] = replicaResp.Id
	default:
		capInterface, ok := request.Parameters["capacity"]
		if !ok {
			errMsg := fmt.Sprint("capacity not found in provision request params!")
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: &errMsg,
			}
		}

		var in = &model.VolumeSpec{
			ProfileId: request.PlanID,
			Size:      int64(capInterface.(float64)),
		}
		if nameInterface, ok := request.Parameters["name"]; ok {
			in.Name = nameInterface.(string)
		}
		if despInterface, ok := request.Parameters["description"]; ok {
			in.Description = despInterface.(string)
		}

		vol, err := c.Client.CreateVolume(in)
		if err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}

		instance.Params["volumeID"] = vol.Id
	}
	// Store instance info into instance map.
	c.instanceMap[request.InstanceID] = instance

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
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

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

	instance, ok := c.instanceMap[request.InstanceID]
	if !ok {
		return &response, nil
	}
	if request.PlanID != "" {
		if !validatePlanID(request.PlanID) {
			errMsg := fmt.Sprintf("PlanID (%s) is not supported!", request.PlanID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: &errMsg,
			}
		}
	}

	switch instance.PlanID {
	case defaultSnapshotPlan:
		snpInterface, ok := instance.Params["snapshotID"]
		if !ok {
			errMsg := fmt.Sprintf("snapshotID not found in instance (%s) params!", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusNotFound,
				ErrorMessage: &errMsg,
			}
		}
		if err := c.Client.DeleteVolumeSnapshot(snpInterface.(string), nil); err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
	case defaultReplicationPlan:
		replicaInterface, ok := instance.Params["replicationID"]
		if !ok {
			errMsg := fmt.Sprintf("replicationID not found in instance (%s) params!", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusNotFound,
				ErrorMessage: &errMsg,
			}
		}
		sVolInterface, ok := instance.Params["secondaryVolumeID"]
		if !ok {
			errMsg := fmt.Sprintf("secondaryVolumeID not found in instance (%s) params!", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusNotFound,
				ErrorMessage: &errMsg,
			}
		}
		if err := c.Client.DeleteReplication(replicaInterface.(string), nil); err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
		if err := c.Client.DeleteVolume(sVolInterface.(string), nil); err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
	default:
		volInterface, ok := instance.Params["volumeID"]
		if !ok {
			errMsg := fmt.Sprintf("volumeID not found in instance (%s) params!", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusNotFound,
				ErrorMessage: &errMsg,
			}
		}
		if err := c.Client.DeleteVolume(volInterface.(string), nil); err != nil {
			errMsg := fmt.Sprint("Broker error:", err)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: &errMsg,
			}
		}
	}
	delete(c.instanceMap, request.InstanceID)

	glog.Infof("Deleted OpenSDS Service Instance:\n%s\n", request.InstanceID)

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

	response := &broker.BindResponse{}

	if request.InstanceID == "" {
		errMsg := fmt.Sprintf("Instance (%s) is not supported!", request.InstanceID)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}
	if _, ok := c.bindingMap[request.BindingID]; ok {
		glog.Infof("Binding %s already exist!\n", request.BindingID)
		response.Exists = true
		return response, nil
	}

	instance, ok := c.instanceMap[request.InstanceID]
	if !ok {
		errMsg := fmt.Sprintf("Instance (%s) not found in instance map!", request.InstanceID)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}
	volInterface, ok := instance.Params["volumeID"]
	if !ok {
		volInterface, ok = instance.Params["secondaryVolumeID"]
		if !ok {
			errMsg := fmt.Sprintf("volumeID not found in instance (%s) params!", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:   http.StatusNotFound,
				ErrorMessage: &errMsg,
			}
		}
	}
	nodeInterface, ok := request.Parameters["nodeID"]
	if !ok {
		errMsg := fmt.Sprint("nodeID not found in bind request params!")
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: &errMsg,
		}
	}

	devResp, err := c.volumeAttachHandler(volInterface.(string), nodeInterface.(string))
	if err != nil {
		errMsg := fmt.Sprint("Broker error:", err)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: &errMsg,
		}
	}

	// Insert credential info into opensds service binding map.
	c.bindingMap[request.BindingID] = &opensdsServiceBinding{
		ID:           request.BindingID,
		InstanceID:   request.InstanceID,
		ServiceID:    request.ServiceID,
		PlanID:       request.PlanID,
		BindResource: request.BindResource,
		Params:       request.Parameters,
	}
	c.bindingMap[request.BindingID].Params["attachmentID"] = devResp["attachmentID"]
	c.bindingMap[request.BindingID].Params["device"] = devResp["device"]

	glog.Infof("Created OpenSDS Service Binding:\n%v\n",
		c.bindingMap[request.BindingID])

	// Generate service binding credentials.
	creds := make(map[string]interface{})
	creds["device"] = devResp["device"]
	osbResponse := &osb.BindResponse{
		Credentials: creds,
	}

	if request.AcceptsIncomplete {
		response.Async = c.async
	}
	response.BindResponse = *osbResponse
	return response, nil
}

func (c *opensdsController) Unbind(
	request *osb.UnbindRequest,
	ctx *broker.RequestContext,
) (*broker.UnbindResponse, error) {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	// Your unbind business logic goes here
	response := broker.UnbindResponse{}

	binding, ok := c.bindingMap[request.BindingID]
	if !ok {
		return &response, nil
	}
	atcInterface, ok := binding.Params["attachmentID"]
	if !ok {
		errMsg := fmt.Sprintf("attachmentID not found in bind (%s) params!", request.BindingID)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusNotFound,
			ErrorMessage: &errMsg,
		}
	}

	if err := c.volumeDetachHandler(atcInterface.(string)); err != nil {
		errMsg := fmt.Sprint("Broker error:", err)
		return nil, osb.HTTPStatusCodeError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: &errMsg,
		}
	}

	delete(c.bindingMap, request.BindingID)

	glog.Infof("Deleted OpenSDS Service Binding:\n%s\n", request.BindingID)

	if request.AcceptsIncomplete {
		response.Async = c.async
	}
	return &response, nil
}

func (c *opensdsController) ValidateBrokerAPIVersion(version string) error {
	return nil
}

func (c *opensdsController) volumeAttachHandler(volID, nodeID string) (DeviceSpec, error) {
	dck, err := discoverAttacherDock(c.Client, nodeID)
	if err != nil {
		return nil, err
	}

	in := &model.VolumeAttachmentSpec{
		VolumeId: volID,
		HostInfo: model.HostInfo{
			Platform:  dck.Metadata["Platform"],
			OsType:    dck.Metadata["OsType"],
			Ip:        dck.Metadata["HostIp"],
			Host:      dck.NodeId,
			Initiator: dck.Metadata["Initiator"],
		},
	}
	// Step 1: Create volume attachment.
	atcResp, err := c.Client.CreateVolumeAttachment(in)
	if err != nil {
		glog.Errorf("failed to create volume(%s) attachment: %v", in.VolumeId, err)
		return nil, fmt.Errorf("failed to create volume(%s) attachment: %v", in.VolumeId, err)
	}
	// Step 2: Check the status of volume attachment.
	atc, err := c.Client.GetVolumeAttachment(atcResp.Id)
	if err != nil || atc.Status != "available" {
		glog.Errorf("failed to get volume attachment(%s): %v", atcResp.Id, err)
		return nil, fmt.Errorf("failed to get volume attachment(%s): %v", atcResp.Id, err)
	}
	// Step 3: Attach volume to the host.
	devResp, err := AttachVolume(c.Client, atc)
	if err != nil {
		glog.Errorf("failed to attach volume to host: %v", err)
		return nil, fmt.Errorf("failed to attach volume to host: %v", err)
	}

	devResp["attachmentID"] = atc.Id
	return devResp, nil
}

func (c *opensdsController) volumeDetachHandler(atcId string) error {
	// Step 1: Check the status of volume attachment.
	atc, err := c.Client.GetVolumeAttachment(atcId)
	if err != nil || atc.Status != "available" {
		glog.Errorf("failed to get volume attachment(%s): %v", atcId, err)
		return fmt.Errorf("failed to get volume attachment(%s): %v", atcId, err)
	}
	// Step 2: Detach volume from host.
	if err := DetachVolume(c.Client, atc); err != nil {
		glog.Errorf("failed to detach volume from host: %v", err)
		return fmt.Errorf("failed to detach volume from host: %v", err)
	}
	// Step 3: Delete volume attachment.
	return c.Client.DeleteVolumeAttachment(atc.Id, nil)
}

/*
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

package client

import (
	"fmt"
	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	"github.com/opensds/opensds/pkg/utils/constants"
	"os"
	"strconv"
)

const (
	// OpenSDSEndPoint environment variable name
	OpenSDSEndPoint = "OPENSDS_ENDPOINT"
	
	// OpenSDSAuthStrategy environment variable name
	OpenSDSAuthStrategy = "OPENSDS_AUTH_STRATEGY"
)

const (
	KVolumeName       = "kubernetes.io/volumeName"
	KVolumeSize       = "kubernetes.io/size"
	KAvailabilityZone = "kubernetes.io/availabilityZone"
	KVolumeId         = "volumeId"
	KFsType           = "kubernetes.io/type"
)

type SdsClient struct {
	client *client.Client
}

type WarpOpensdsClient interface {
	Provision(opts map[string]string) (string, error)
	Delete(volumeId string) error
}

var _ WarpOpensdsClient = &SdsClient{}

func NewSdsClient(endpoint string, authStrategy string) WarpOpensdsClient {
	client := getSdsClient(endpoint, authStrategy)
	return &SdsClient{
		client: client,
	}
}

func (c *SdsClient) Provision(opts map[string]string) (string, error) {
	err := optionCheck([]string{KVolumeName, KVolumeSize}, opts)
	if err != nil {
		return "", err
	}

	size, _ := strconv.ParseInt(opts[KVolumeSize], 10, 0)
	volSpec := &model.VolumeSpec{
		Name: opts[KVolumeName],
		Size: size,
	}

	if zone, exist := opts[KAvailabilityZone]; exist {
		volSpec.AvailabilityZone = zone
	}

	vol, errCreate := c.client.CreateVolume(volSpec)
	if errCreate != nil {
		return "", nil
	}

	return vol.Id, nil
}

func (c *SdsClient) Delete(volumeId string) error {
	return c.client.DeleteVolume(volumeId, &model.VolumeSpec{})
}

// getSdsClient return OpenSDS Client
func getSdsClient(endpoint string, authStrategy string) *client.Client {
	if endpoint == "" {
		// Get endpoint from environment
		endpoint = os.Getenv(OpenSDSEndPoint)
	}

	if endpoint == "" {
		// Using default endpoint
		endpoint = "http://localhost:50040"
	}

	if authStrategy == "" {
		// Get auth strategy from environment
		authStrategy = os.Getenv(OpenSDSAuthStrategy)
	}

	if authStrategy == "" {
		// Using default auth strategy
		authStrategy = "noauth"
	}
	
	cfg := &client.Config{Endpoint: endpoint}
	
	switch authStrategy {
	case client.Keystone:
		cfg.AuthOptions = client.LoadKeystoneAuthOptionsFromEnv()
	case client.Noauth:
		cfg.AuthOptions = client.LoadNoAuthOptionsFromEnv()
	default:
		cfg.AuthOptions = client.NewNoauthOptions(constants.DefaultTenantId)
	}

	return client.NewClient(cfg)
}

func optionCheck(optCheckList []string, opts map[string]string) error {
	for _, value := range optCheckList {
		if _, exist := opts[value]; !exist {
			return fmt.Errorf("option %s not specified", value)
		}
	}

	return nil
}

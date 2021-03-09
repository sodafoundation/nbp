// Copyright 2018 The OpenSDS Authors.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package controller

import (
	"github.com/Masterminds/semver"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sodafoundation/api/pkg/utils"
	"github.com/sodafoundation/nbp/service-broker/pkg/store"
)

var (
	minSupportedVersion = "2.13"
	versionConstraint   = ">= " + minSupportedVersion
)

func validateBrokerAPIVersion(version string) bool {
	c, _ := semver.NewConstraint(versionConstraint)
	v, _ := semver.NewVersion(version)
	// Check if the version meets the constraints. The a variable will be true.
	return c.Check(v)
}

const (
	ServiceInstanceType = "service_instance"
	ServiceBindingType  = "service_binding"

	OperationCreate = "create"
	OperationUpdate = "update"
)

func validateCatalogSchema(
	serviceID, planID string,
	params map[string]interface{},
	schemaType, operation string,
	dbStore store.Store,
) bool {
	service, ok, _ := dbStore.GetServiceClass(serviceID)
	if !ok {
		return false
	}
	plan := func() *osb.Plan {
		for _, v := range service.Service.Plans {
			if planID == v.ID {
				return &v
			}
		}
		return nil
	}()
	if plan == nil {
		return false
	}
	schemas := plan.Schemas

	mapKeyContained := func(obj interface{}, tgt map[string]interface{}) bool {
		mapObj := obj.(map[string]interface{})
		for k := range mapObj {
			if !utils.Contained(k, tgt) {
				return false
			}
		}
		return true
	}

	switch schemaType {
	case ServiceInstanceType:
		instanceSchema := schemas.ServiceInstance
		if operation == OperationCreate {
			if mapKeyContained(instanceSchema.Create.Parameters, params) {
				return true
			}
		} else if operation == OperationUpdate {
			if mapKeyContained(instanceSchema.Update.Parameters, params) {
				return true
			}
		}
		break
	case ServiceBindingType:
		bindingSchema := schemas.ServiceBinding
		if operation == OperationCreate {
			if mapKeyContained(bindingSchema.Create.Parameters, params) {
				return true
			}
		}
		break
	default:
		break
	}
	return false
}

const (
	defaultVolumeService      = "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468"
	defaultSnapshotService    = "434ba788-3d92-11e8-8712-1740dc7b3f46"
	defaultReplicationService = "ef136c64-6cb2-11e8-802e-3fdf92f7654d"

	defaultSnapshotPlan    = "787c9322-3d92-11e8-8cb3-4f1353df06c1"
	defaultReplicationPlan = "b466e4ce-6cb2-11e8-a580-bb2b8754c9de"
)

var (
	supportedServiceList = []string{
		defaultVolumeService, defaultSnapshotService, defaultReplicationService,
	}
	supportedPlanList = []string{
		defaultSnapshotPlan, defaultReplicationPlan,
	}
)

// initializePlanList would clean the supportedPlanList and insert default
// snapshot and replication plan.
func initializePlanList() []string {
	supportedPlanList = supportedPlanList[:0]
	supportedPlanList = append(supportedPlanList, defaultSnapshotPlan,
		defaultReplicationPlan)

	return supportedPlanList
}

func validateServiceID(serviceID string) bool {
	for _, v := range supportedServiceList {
		if v == serviceID {
			return true
		}
	}
	return false
}

func validatePlanID(planID string) bool {
	for _, v := range supportedPlanList {
		if v == planID {
			return true
		}
	}
	return false
}

func truePtr() *bool {
	a := true
	return &a
}

func falsePtr() *bool {
	b := false
	return &b
}

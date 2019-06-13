// Copyright 2019 The OpenSDS Authors.
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

package opensds

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/opensds/nbp/csi/util"
)

func getSize(capacityRange *csi.CapacityRange) int64 {
	var size int64
	allocationUnitBytes := util.GiB

	if capacityRange != nil {
		sizeBytes := int64(capacityRange.GetRequiredBytes())
		size = (sizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
		if size < 1 {
			//Using default size
			size = 1
		}
	} else {
		//Using default size
		size = 1
	}

	return size
}

func waitForStatusStable(id string, f func(string) (interface{}, error)) (interface{}, error) {

	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(5 * time.Minute)

	defer ticker.Stop()
	validStatus := []string{"error", "error_deleting", "error_restoring", "error_extending", "available", "in-use"}

	for {
		select {
		case <-ticker.C:
			o, err := f(id)
			if err != nil {
				return nil, err
			}

			status, err := findStatusFiledFromStruct(o)
			if err != nil {
				return nil, err
			}

			if util.Contained(status, validStatus) {
				if status == "available" {
					return o, nil
				}
				return nil, fmt.Errorf("status is %s but not available", status)

			}
			return nil, fmt.Errorf("invalid status: %s", status)

		case <-timeout:
			return nil, fmt.Errorf("timeout occured waiting for checking status of %s", id)
		}
	}
}

func findStatusFiledFromStruct(o interface{}) (string, error) {
	v := reflect.ValueOf(o)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", errors.New("input cannot be nil")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", errors.New("input must be struct")
	}

	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Name == "Status" {
			return v.Field(i).String(), nil
		}
	}

	return "", fmt.Errorf("cannot find status from struct %v", o)
}

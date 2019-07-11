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

package sanity

import (
	"testing"

	"github.com/opensds/opensds/pkg/model"
)

type Volume struct {
	Name string
	Size int
}

func TestStructListCopy(t *testing.T) {
	var vTgt []*Volume
	var vSrc = []*Volume{
		&Volume{
			Name: "test",
			Size: 4,
		},
	}

	t.Run("struct list copy successfully", func(t *testing.T) {
		structListCopy(vSrc, &vTgt)
		assertTestResult(t, vSrc, vTgt)
	})

	t.Run("target is unaddressable", func(t *testing.T) {
		err := structListCopy(vSrc, vTgt)
		assertTestResult(t, "target is unaddressable", err.Error())
	})

	t.Run("target is unaddressable", func(t *testing.T) {

		pools := []*model.StoragePoolSpec{
			&model.StoragePoolSpec{
				BaseModel: &model.BaseModel{
					Id: "084bf71e-a102-11e7-88a8-e31fe6d52248",
				},
				Name:          "sample-pool-01",
				Description:   "This is the first sample storage pool for testing",
				StorageType:   "block",
				TotalCapacity: 100,
				FreeCapacity:  90,
				DockId:        "b7602e18-771e-11e7-8f38-dbd6d291f4e0",
			},
		}
		var vTgt []*model.StoragePoolSpec
		structListCopy(pools, &vTgt)
		assertTestResult(t, pools, vTgt)
	})
}

func TestStructCopy(t *testing.T) {
	var vTgt Volume
	var vSrc = &Volume{
		Name: "test",
		Size: 4,
	}

	t.Run("struct copy successfully", func(t *testing.T) {
		structCopy(vSrc, &vTgt)
		assertTestResult(t, vSrc, &vTgt)
	})

	t.Run("target is unaddressable", func(t *testing.T) {
		err := structListCopy(vSrc, vTgt)
		assertTestResult(t, "target is unaddressable", err.Error())
	})

	t.Run("target is unaddressable", func(t *testing.T) {
		type test struct {
			Name string
			id   string
		}

		var vSrc = &test{
			Name: "test-case",
			id:   "123",
		}

		var vTgt *test
		err := structCopy(vSrc, vTgt)
		assertTestResult(t, "target is unaddressable", err.Error())
	})

	t.Run("id field is unaddressable", func(t *testing.T) {
		type test struct {
			Name string
			id   string
		}

		var vSrc = &test{
			Name: "test-case",
			id:   "123",
		}

		var vTgt test
		err := structCopy(vSrc, &vTgt)
		assertTestResult(t, "id field is unaddressable", err.Error())
	})
}

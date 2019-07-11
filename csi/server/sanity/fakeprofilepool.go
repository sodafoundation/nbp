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
	"strings"

	"github.com/opensds/opensds/pkg/model"
)

type fakeProfile struct{}

func (v *fakeProfile) GetProfile(profID string) (*model.ProfileSpec, error) {
	return &model.ProfileSpec{
		BaseModel: &model.BaseModel{
			Id: "1106b972-66ef-11e7-b172-db03f3689c9c",
		},
		Name:        "default",
		Description: "default policy",
		StorageType: "block",
	}, nil
}

func (v *fakeProfile) Recv(url, method string, input, output interface{}) error {
	switch strings.ToUpper(method) {
	case "GET":
		out, _ := v.GetProfile("")
		return structCopy(out, output)
	}

	return nil
}

type fakePool struct{}

var pool = &model.StoragePoolSpec{
	BaseModel: &model.BaseModel{
		Id: "084bf71e-a102-11e7-88a8-e31fe6d52248",
	},
	Name:          "sample-pool-01",
	Description:   "This is the first sample storage pool for testing",
	StorageType:   "block",
	TotalCapacity: 100,
	FreeCapacity:  90,
	DockId:        "b7602e18-771e-11e7-8f38-dbd6d291f4e0",
	Extras: model.StoragePoolExtraSpec{
		IOConnectivity: model.IOConnectivityLoS{
			AccessProtocol: "sample",
		},
	},
}

func (p *fakePool) ListPools() ([]*model.StoragePoolSpec, error) {
	pools := []*model.StoragePoolSpec{
		pool,
	}
	return pools, nil
}

func (p *fakePool) GetPool(poolId string) (*model.StoragePoolSpec, error) {
	return pool, nil
}

func (v *fakePool) Recv(url, method string, input, output interface{}) error {
	switch strings.ToUpper(method) {
	case "GET":
		switch output.(type) {
		case *model.StoragePoolSpec:
			pool, _ := v.GetPool("")
			return structCopy(pool, output)

		case *[]*model.StoragePoolSpec:
			pools, _ := v.ListPools()
			return structListCopy(pools, output)
		}
	}

	return nil
}

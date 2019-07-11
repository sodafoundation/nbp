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

	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
)

type fakeReplication struct{}

func (*fakeReplication) CreateReplication(req c.ReplicationBuilder) (
	*model.ReplicationSpec, error) {
	return nil, nil
}

func (*fakeReplication) GetReplication(replicaId string) (
	*model.ReplicationSpec, error) {
	return nil, nil
}

func (*fakeReplication) ListReplications() ([]*model.ReplicationSpec, error) {
	return nil, nil
}

func (*fakeReplication) DeleteReplication(replicaId string,
	req c.ReplicationBuilder) error {
	return nil
}

func (*fakeReplication) UpdateReplication(replicaId string,
	req c.ReplicationBuilder) (
	*model.ReplicationSpec, error) {
	return nil, nil
}

func (r *fakeReplication) Recv(url, method string, input, output interface{}) error {
	switch strings.ToUpper(method) {
	case "POST":
		r.CreateReplication(nil)
	case "PUT":
		r.UpdateReplication("", nil)
	case "GET":
		r.GetReplication("")
	case "DELETE":
		return r.DeleteReplication("", nil)
	}

	return nil
}

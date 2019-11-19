// Copyright 2016 The OpenSDS Authors.
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

package client

type FakeClient struct {
}

var _ WarpOpensdsClient = &FakeClient{}

func NewFakeClient(endpoint string, authStrategy string) WarpOpensdsClient {
	return &FakeClient{}
}

func (c *FakeClient) Provision(opts map[string]string) (string, error) {
	return "volume-opendsds-nbp-privisioner", nil
}

func (c *FakeClient) Delete(volumeId string) error {
	return nil
}

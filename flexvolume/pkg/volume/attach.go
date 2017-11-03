// Copyright (c) 2016 Huawei Technologies Co., Ltd. All Rights Reserved.
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

/*
This module implements a standard SouthBound interface of volume resource to
storage plugins.

*/

package volume

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/astaxie/beego/httplib"

	"github.com/opensds/nbp/flexvolume/pkg/api"
)

const (
	URL_PREFIX string = "http://192.168.0.9:50040"
)

func CreateVolumeAttachment(volID string, prop *api.ConnectorProperties) (*api.VolumeAttachment, error) {
	url := URL_PREFIX + "/api/v1/volumes/" + volID + "/attachments"
	vr := &api.VolumeRequest{
		Schema: &api.VolumeOperationSchema{
			DoLocalAttach: prop.DoLocalAttach,
			MultiPath:     prop.MultiPath,
			HostInfo: api.HostInfo{
				Platform:  prop.Platform,
				OsType:    prop.OsType,
				Ip:        prop.Ip,
				Host:      prop.Host,
				Initiator: prop.Initiator,
			},
		},
	}

	// fmt.Println("Start POST request to create volume attachment, url =", url)
	req := httplib.Post(url).SetTimeout(100*time.Second, 50*time.Second)
	req.JSONBody(vr)

	resp, err := req.Response()
	if err != nil {
		return nil, err
	}
	err = CheckHTTPResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var atc = &api.VolumeAttachment{}
	if err = json.Unmarshal(rbody, atc); err != nil {
		return nil, err
	}
	return atc, nil
}

func GetVolumeAttachment(id, volID string) (*api.VolumeAttachment, error) {
	url := URL_PREFIX + "/api/v1/volumes/" + volID + "/attachments/" + id

	// fmt.Println("Start GET request to get volume attachment, url =", url)
	req := httplib.Get(url).SetTimeout(100*time.Second, 50*time.Second)

	resp, err := req.Response()
	if err != nil {
		return nil, err
	}
	err = CheckHTTPResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var atc = &api.VolumeAttachment{}
	if err = json.Unmarshal(rbody, atc); err != nil {
		return nil, err
	}
	return atc, nil
}

func ListVolumeAttachments(volID string) (*[]api.VolumeAttachment, error) {
	url := URL_PREFIX + "/api/v1/volumes/" + volID + "/attachments"

	// fmt.Println("Start GET request to list volume attachments, url =", url)
	req := httplib.Get(url).SetTimeout(100*time.Second, 50*time.Second)

	resp, err := req.Response()
	if err != nil {
		return nil, err
	}
	err = CheckHTTPResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var atcs = &[]api.VolumeAttachment{}
	if err = json.Unmarshal(rbody, atcs); err != nil {
		return nil, err
	}
	return atcs, nil
}

func UpdateVolumeAttachment(id, volID, mountpoint string, hostInfo api.HostInfo) (*api.VolumeAttachment, error) {
	url := URL_PREFIX + "/api/v1/volumes/" + volID + "/attachments/" + id
	vr := &api.VolumeRequest{
		Schema: &api.VolumeOperationSchema{
			HostInfo:   hostInfo,
			Mountpoint: mountpoint,
		},
	}

	// fmt.Println("Start PUT request to update volume attachment, url =", url)
	req := httplib.Put(url).SetTimeout(100*time.Second, 50*time.Second)
	req.JSONBody(vr)

	resp, err := req.Response()
	if err != nil {
		return nil, err
	}
	err = CheckHTTPResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var atc = &api.VolumeAttachment{}
	if err = json.Unmarshal(rbody, atc); err != nil {
		return nil, err
	}
	return atc, nil
}

func DeleteVolumeAttachment(id, volID string) (*api.VolumeResponse, error) {
	url := URL_PREFIX + "/api/v1/volumes/" + volID + "/attachments/" + id

	// fmt.Println("Start DELETE request to delete volume attachment, url =", url)
	req := httplib.Delete(url).SetTimeout(100*time.Second, 50*time.Second)

	resp, err := req.Response()
	if err != nil {
		return nil, err
	}
	err = CheckHTTPResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var volumeResponse = &api.VolumeResponse{}
	err = json.Unmarshal(rbody, volumeResponse)
	if err != nil {
		return nil, err
	}
	return volumeResponse, nil
}

// CheckHTTPResponseStatusCode compares http response header StatusCode against expected
// statuses. Primary function is to ensure StatusCode is in the 20x (return nil).
// Ok: 200. Created: 201. Accepted: 202. No Content: 204. Partial Content: 206.
// Otherwise return error message.
func CheckHTTPResponseStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case 200, 201, 202, 204, 206:
		return nil
	case 400:
		return errors.New("Error: response == 400 bad request")
	case 401:
		return errors.New("Error: response == 401 unauthorised")
	case 403:
		return errors.New("Error: response == 403 forbidden")
	case 404:
		return errors.New("Error: response == 404 not found")
	case 405:
		return errors.New("Error: response == 405 method not allowed")
	case 409:
		return errors.New("Error: response == 409 conflict")
	case 413:
		return errors.New("Error: response == 413 over limit")
	case 415:
		return errors.New("Error: response == 415 bad media type")
	case 422:
		return errors.New("Error: response == 422 unprocessable")
	case 429:
		return errors.New("Error: response == 429 too many request")
	case 500:
		return errors.New("Error: response == 500 instance fault / server err")
	case 501:
		return errors.New("Error: response == 501 not implemented")
	case 503:
		return errors.New("Error: response == 503 service unavailable")
	}
	return errors.New("Error: unexpected response status code")
}


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
	"errors"
	"fmt"
	"strings"
	"time"

	c "github.com/opensds/opensds/client"
	"github.com/opensds/opensds/pkg/model"
	constants "github.com/opensds/opensds/pkg/utils/constants"
	uuid "github.com/satori/go.uuid"
)

type fakeVolume struct{}

// used as a database
var volumeList []*model.VolumeSpec

func (v *fakeVolume) CreateVolume(req c.VolumeBuilder) (*model.VolumeSpec, error) {
	vol := &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: uuid.NewV4().String(),
		},
		Name:             req.Name,
		UserId:           req.UserId,
		Description:      req.Description,
		Size:             req.Size,
		Status:           "available",
		AvailabilityZone: req.AvailabilityZone,
		ProfileId:        req.ProfileId,
	}
	volumeList = append(volumeList, vol)

	return vol, nil
}

func (v *fakeVolume) GetVolume(volID string) (*model.VolumeSpec, error) {
	for _, vol := range volumeList {
		if vol.Id == volID {
			return getNewVolume(vol), nil
		}
	}

	return nil, fmt.Errorf("volume %s cannot be found", volID)
}

func getNewVolume(vol *model.VolumeSpec) *model.VolumeSpec {
	// Return a new pointer object in order not to affect the metadata
	return &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: vol.Id,
		},
		Name:             vol.Name,
		UserId:           vol.UserId,
		Description:      vol.Description,
		Size:             vol.Size,
		Status:           vol.Status,
		AvailabilityZone: vol.AvailabilityZone,
		ProfileId:        vol.ProfileId,
	}
}

func (v *fakeVolume) ListVolumes() ([]*model.VolumeSpec, error) {
	var volumeListNew []*model.VolumeSpec

	for _, vol := range volumeList {
		volumeListNew = append(volumeListNew, getNewVolume(vol))
	}

	return volumeListNew, nil
}

func (v *fakeVolume) DeleteVolume(volID string, req c.VolumeBuilder) error {
	if _, err := v.GetVolume(volID); err != nil {
		return err
	}

	for i, vol := range volumeList {
		if vol.Id == volID {
			volumeList = append(volumeList[:i], volumeList[i+1:]...)
			break
		}
	}

	return nil
}

func (v *fakeVolume) UpdateVolume(volID string, req c.VolumeBuilder) (
	*model.VolumeSpec, error) {
	for _, vol := range volumeList {
		if vol.Id == req.Id {
			vol.Name = req.Name
			vol.UserId = req.UserId
			vol.Description = req.Description
			vol.Size = req.Size
			vol.AvailabilityZone = req.AvailabilityZone
			vol.ProfileId = req.ProfileId
			vol.Status = req.Status

			return getNewVolume(vol), nil
		}
	}

	return nil, fmt.Errorf("volume %s cannot be found", volID)
}

func (v *fakeVolume) ExtendVolume(volID string, req c.VolumeBuilder) (
	*model.VolumeSpec, error) {
	newSize := req.Size
	for _, vol := range volumeList {
		if vol.Id == volID {
			vol.Size = newSize
			return getNewVolume(vol), nil
		}
	}

	return nil, fmt.Errorf("volume %s cannot be found", volID)
}

// used as a database
var attachments []*model.VolumeAttachmentSpec

func (v *fakeVolume) CreateVolumeAttachment(req c.VolumeAttachmentBuilder) (
	*model.VolumeAttachmentSpec, error) {
	atcm := &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{
			Id: uuid.NewV4().String(),
		},
		VolumeId:       req.VolumeId,
		Mountpoint:     req.Mountpoint,
		Status:         "available",
		Metadata:       req.Metadata,
		HostInfo:       req.HostInfo,
		ConnectionInfo: req.ConnectionInfo,
		AccessProtocol: req.AccessProtocol,
		AttachMode:     req.AttachMode,
	}
	attachments = append(attachments, atcm)

	return atcm, nil
}

func getNewVolumeAttachment(req *model.VolumeAttachmentSpec) *model.VolumeAttachmentSpec {
	// Return a new pointer object in order not to affect the metadata
	return &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{
			Id: req.Id,
		},
		VolumeId:       req.VolumeId,
		Mountpoint:     req.Mountpoint,
		Status:         req.Status,
		Metadata:       req.Metadata,
		HostInfo:       req.HostInfo,
		ConnectionInfo: req.ConnectionInfo,
		AccessProtocol: req.AccessProtocol,
		AttachMode:     req.AttachMode,
	}
}

func (v *fakeVolume) UpdateVolumeAttachment(atcID string, req c.VolumeAttachmentBuilder) (
	*model.VolumeAttachmentSpec, error) {
	for _, atcm := range attachments {
		if atcm.Id == req.Id {
			atcm.VolumeId = req.VolumeId
			atcm.Mountpoint = req.Mountpoint
			atcm.Status = req.Status
			atcm.Metadata = req.Metadata
			atcm.HostInfo = req.HostInfo
			atcm.ConnectionInfo = req.ConnectionInfo
			atcm.AccessProtocol = req.AccessProtocol
			atcm.AttachMode = req.AttachMode

			return getNewVolumeAttachment(req), nil
		}
	}

	return nil, fmt.Errorf("volume attachment %s cannot be found", atcID)
}

func (v *fakeVolume) GetVolumeAttachment(atcID string) (*model.VolumeAttachmentSpec, error) {
	for _, atcm := range attachments {
		if atcm.Id == atcID {
			// Return a new pointer object in order not to affect the metadata
			return getNewVolumeAttachment(atcm), nil
		}
	}

	return nil, fmt.Errorf("volume attachment %s cannot be found", atcID)
}

func (v *fakeVolume) ListVolumeAttachments() ([]*model.VolumeAttachmentSpec, error) {
	var attachmentsNew []*model.VolumeAttachmentSpec

	for _, atcm := range attachments {
		attachmentsNew = append(attachmentsNew, getNewVolumeAttachment(atcm))
	}

	return attachmentsNew, nil
}

func (v *fakeVolume) DeleteVolumeAttachment(atcID string, req c.VolumeAttachmentBuilder) error {
	if _, err := v.GetVolumeAttachment(atcID); err != nil {
		return err
	}

	for i, atcm := range attachments {
		if atcm.Id == atcID {
			attachments = append(attachments[:i], attachments[i+1:]...)
			break
		}
	}

	return nil
}

var snapshots []*model.VolumeSnapshotSpec

func (v *fakeVolume) CreateVolumeSnapshot(req c.VolumeSnapshotBuilder) (
	*model.VolumeSnapshotSpec, error) {
	snp := &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id:        uuid.NewV4().String(),
			CreatedAt: time.Now().Format(constants.TimeFormat),
		},
		Name:        req.Name,
		Description: req.Description,
		ProfileId:   req.ProfileId,
		Size:        req.Size,
		Status:      "available",
		VolumeId:    req.VolumeId,
		Metadata:    req.Metadata,
	}
	snapshots = append(snapshots, snp)

	return snp, nil
}

func getNewSnapshot(snp *model.VolumeSnapshotSpec) *model.VolumeSnapshotSpec {
	// Return a new pointer object in order not to affect the metadata
	return &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id:        snp.Id,
			CreatedAt: snp.CreatedAt,
		},
		Name:        snp.Name,
		Description: snp.Description,
		ProfileId:   snp.ProfileId,
		Size:        snp.Size,
		Status:      snp.Status,
		VolumeId:    snp.VolumeId,
		Metadata:    snp.Metadata,
	}
}

func (v *fakeVolume) GetVolumeSnapshot(snpID string) (*model.VolumeSnapshotSpec, error) {
	for _, snp := range snapshots {
		if snp.Id == snpID {
			return getNewSnapshot(snp), nil
		}
	}

	return nil, fmt.Errorf("snapshot %s cannot be found", snpID)
}

func (v *fakeVolume) ListVolumeSnapshots() ([]*model.VolumeSnapshotSpec, error) {
	var snapshotListNew []*model.VolumeSnapshotSpec

	for _, snp := range snapshots {
		snapshotListNew = append(snapshotListNew, getNewSnapshot(snp))
	}

	return snapshotListNew, nil
}

func (v *fakeVolume) DeleteVolumeSnapshot(snpID string, snp c.VolumeSnapshotBuilder) error {
	if _, err := v.GetVolumeSnapshot(snpID); err != nil {
		return err
	}

	for i, snp := range snapshots {
		if snp.Id == snpID {
			snapshots = append(snapshots[:i], snapshots[i+1:]...)
			break
		}
	}

	return nil
}

func (v *fakeVolume) UpdateVolumeSnapshot(snpID string, req c.VolumeSnapshotBuilder) (
	*model.VolumeSnapshotSpec, error) {
	for _, snp := range snapshots {
		if snp.Id == req.Id {
			snp.Name = req.Name
			snp.Description = req.Description
			snp.ProfileId = req.ProfileId
			snp.Size = req.Size
			snp.Status = req.Status
			snp.VolumeId = req.VolumeId
			snp.Metadata = req.Metadata

			return getNewSnapshot(snp), nil
		}
	}

	return nil, fmt.Errorf("volume snapshot %s cannot be found", snpID)
}

func (v *fakeVolume) Recv(url, method string, input, output interface{}) error {
	urlList := strings.Split(url, "/")
	id := urlList[len(urlList)-1]

	switch strings.ToUpper(method) {
	case "POST":
		return v.post(input, output)
	case "PUT":
		return v.put(input, output)
	case "GET":
		return v.get(input, output, id)
	case "DELETE":
		return v.delete(input, id)
	}
	return nil
}

func (v *fakeVolume) post(input, output interface{}) error {
	switch output.(type) {
	case *model.VolumeSpec:
		out, _ := v.CreateVolume(input.(c.VolumeBuilder))
		return structCopy(out, output)

	case *model.VolumeSnapshotSpec:
		out, _ := v.CreateVolumeSnapshot(input.(c.VolumeSnapshotBuilder))
		return structCopy(out, output)

	case *model.VolumeAttachmentSpec:
		out, _ := v.CreateVolumeAttachment(input.(c.VolumeAttachmentBuilder))
		return structCopy(out, output)

	default:
		return errors.New("output format not supported")
	}
}

func (v *fakeVolume) put(input, output interface{}) error {
	switch output.(type) {
	case *model.VolumeSpec:
		out, err := v.UpdateVolume("", input.(c.VolumeBuilder))
		if err != nil {
			return err
		}
		return structCopy(out, output)

	case *model.VolumeSnapshotSpec:
		out, err := v.UpdateVolumeSnapshot("", input.(c.VolumeSnapshotBuilder))
		if err != nil {
			return err
		}
		return structCopy(out, output)

	case *model.VolumeAttachmentSpec:
		out, err := v.UpdateVolumeAttachment("", input.(c.VolumeAttachmentBuilder))
		if err != nil {
			return err
		}
		return structCopy(out, output)

	default:
		return errors.New("output format not supported")
	}
}

func (v *fakeVolume) get(input, output interface{}, id string) error {
	switch output.(type) {
	case *[]*model.VolumeSpec:
		out, _ := v.ListVolumes()
		return structListCopy(out, output)

	case *[]*model.VolumeSnapshotSpec:
		out, _ := v.ListVolumeSnapshots()
		return structListCopy(out, output)

	case *[]*model.VolumeAttachmentSpec:
		out, _ := v.ListVolumeAttachments()
		return structListCopy(out, output)

	case *model.VolumeSpec:
		out, err := v.GetVolume(id)
		if err != nil {
			return err
		}
		return structCopy(out, output)

	case *model.VolumeSnapshotSpec:
		out, err := v.GetVolumeSnapshot(id)
		if err != nil {
			return err
		}
		return structCopy(out, output)

	case *model.VolumeAttachmentSpec:
		out, err := v.GetVolumeAttachment(id)
		if err != nil {
			return err
		}
		return structCopy(out, output)

	default:
		return errors.New("output format not supported")
	}
}

func (v *fakeVolume) delete(input interface{}, id string) error {
	switch input.(type) {
	case c.VolumeBuilder:
		return v.DeleteVolume(id, nil)
	case c.VolumeSnapshotBuilder:
		return v.DeleteVolumeSnapshot(id, nil)
	case c.VolumeAttachmentBuilder:
		return v.DeleteVolumeAttachment(id, nil)
	default:
		return errors.New("input format not supported")
	}
}

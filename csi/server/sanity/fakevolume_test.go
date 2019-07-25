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
	"fmt"
	"reflect"
	"testing"

	"github.com/opensds/opensds/pkg/model"
)

var assertTestResult = func(t *testing.T, expected, got interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %#v, got: %#v\n", expected, got)
	}
}

func reset() {
	volumeList = []*model.VolumeSpec{}
	attachments = []*model.VolumeAttachmentSpec{}
	snapshots = []*model.VolumeSnapshotSpec{}
}

func TestVolume(t *testing.T) {
	fakeVol := &fakeVolume{}

	req := &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		Name:             "volume_test",
		UserId:           "2",
		Status:           "available",
		Description:      "volume for test",
		Size:             3,
		AvailabilityZone: "defaultZone",
		ProfileId:        "bd5b12a8-a101-11e7-941e-d77981b584d8",
	}

	t.Run("get volume failed", func(t *testing.T) {
		reset()
		volId := "17b880d2f537"
		_, err := fakeVol.GetVolume(volId)
		assertTestResult(t, fmt.Sprintf("volume %s cannot be found", volId), err.Error())
	})

	t.Run("normal case", func(t *testing.T) {
		reset()
		vol, _ := fakeVol.CreateVolume(req)
		assertTestResult(t, vol.Name, req.Name)

		volNew, _ := fakeVol.GetVolume(vol.Id)
		assertTestResult(t, vol, volNew)

		volListExpected := []*model.VolumeSpec{vol}
		volListGot, _ := fakeVol.ListVolumes()
		assertTestResult(t, volListExpected, volListGot)

		fakeVol.DeleteVolume(vol.Id, nil)
		volListGot, _ = fakeVol.ListVolumes()
		assertTestResult(t, 0, len(volListGot))
	})

	t.Run("delete volume failed", func(t *testing.T) {
		reset()
		volId := "1"
		err := fakeVol.DeleteVolume(volId, nil)
		assertTestResult(t, fmt.Sprintf("volume %s cannot be found", volId), err.Error())
	})

	t.Run("update volume successfully", func(t *testing.T) {
		reset()
		volGot, _ := fakeVol.CreateVolume(req)
		assertTestResult(t, req.Name, volGot.Name)

		newVolName := "volume_update"
		volGot.Name = newVolName
		volNew, _ := fakeVol.UpdateVolume(volGot.Id, volGot)
		assertTestResult(t, newVolName, volNew.Name)
	})

	t.Run("update volume failed", func(t *testing.T) {
		reset()
		newVolName := "volume_update"
		req.Name = newVolName
		_, err := fakeVol.UpdateVolume(req.Id, req)
		assertTestResult(t, fmt.Sprintf("volume %s cannot be found", req.Id), err.Error())
	})

	t.Run("extend volume successfully", func(t *testing.T) {
		reset()
		volGot, _ := fakeVol.CreateVolume(req)
		assertTestResult(t, req.Name, volGot.Name)

		newSize := int64(6)
		volGot.Size = newSize
		volNew, _ := fakeVol.ExtendVolume(volGot.Id, volGot)
		assertTestResult(t, int64(newSize), volNew.Size)
	})

	t.Run("extend volume failed", func(t *testing.T) {
		reset()
		newSize := int64(6)
		req.Size = newSize
		_, err := fakeVol.ExtendVolume(req.Id, req)
		assertTestResult(t, fmt.Sprintf("volume %s cannot be found", req.Id), err.Error())
	})
}

func TestAttachment(t *testing.T) {
	fakeVol := &fakeVolume{}

	req := &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{
			Id: "3bfaf2cc-a102-11e7-8ecb-63aea739d755",
		},
		VolumeId:       "3769855c-a102-11e7-b772-17b880d2f537",
		Mountpoint:     "/mnt",
		Status:         "available",
		AccessProtocol: "iscsi",
		AttachMode:     "ro",
	}

	t.Run("get attachment failed", func(t *testing.T) {
		reset()
		atcmId := "17b880d2f537"
		_, err := fakeVol.GetVolumeAttachment(atcmId)
		assertTestResult(t, fmt.Sprintf("volume attachment %s cannot be found", atcmId), err.Error())
	})

	t.Run("normal case", func(t *testing.T) {
		reset()
		atcm, _ := fakeVol.CreateVolumeAttachment(req)
		assertTestResult(t, atcm.DriverVolumeType, req.DriverVolumeType)

		atcmNew, _ := fakeVol.GetVolumeAttachment(atcm.Id)
		assertTestResult(t, atcm, atcmNew)

		atcmListExpected := []*model.VolumeAttachmentSpec{atcm}
		atcmListGot, _ := fakeVol.ListVolumeAttachments()
		assertTestResult(t, atcmListExpected, atcmListGot)

		fakeVol.DeleteVolumeAttachment(atcm.Id, nil)
		atcmListGot, _ = fakeVol.ListVolumeAttachments()
		assertTestResult(t, 0, len(atcmListGot))
	})

	t.Run("delete attachment failed", func(t *testing.T) {
		reset()
		atcmId := "1"
		err := fakeVol.DeleteVolumeAttachment(atcmId, nil)
		assertTestResult(t, fmt.Sprintf("volume attachment %s cannot be found", atcmId), err.Error())
	})

	t.Run("update attachment successfully", func(t *testing.T) {
		reset()
		atcmGot, _ := fakeVol.CreateVolumeAttachment(req)
		assertTestResult(t, req.DriverVolumeType, atcmGot.DriverVolumeType)

		newMountpoint := "/tmp"
		atcmGot.Mountpoint = newMountpoint
		atcmNew, _ := fakeVol.UpdateVolumeAttachment(atcmGot.Id, atcmGot)
		assertTestResult(t, newMountpoint, atcmNew.Mountpoint)
	})

	t.Run("update attachment failed", func(t *testing.T) {
		reset()
		newMountpoint := "/tmp"
		req.Mountpoint = newMountpoint
		_, err := fakeVol.UpdateVolumeAttachment(req.Id, req)
		assertTestResult(t, fmt.Sprintf("volume attachment %s cannot be found", req.Id), err.Error())
	})
}

func TestSnapshot(t *testing.T) {
	fakeVol := &fakeVolume{}

	req := &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3bfaf2cc-a102-11e7-8ecb-63aea739d755",
		},
		Name:        "snapshot_test",
		Description: "snapshot for test",
		ProfileId:   "3769855c-a102-11e7-b772-17b880d2f537",
		Size:        4,
		Status:      "available",
		VolumeId:    "bd5b12a8-a101-11e7-941e-d77981b584d8",
	}

	t.Run("get volume snapshot failed", func(t *testing.T) {
		reset()
		snpId := "17b880d2f537"
		_, err := fakeVol.GetVolumeSnapshot(snpId)
		assertTestResult(t, fmt.Sprintf("snapshot %s cannot be found", snpId), err.Error())
	})

	t.Run("normal case", func(t *testing.T) {
		reset()
		snp, _ := fakeVol.CreateVolumeSnapshot(req)
		assertTestResult(t, snp.Name, req.Name)

		snpNew, _ := fakeVol.GetVolumeSnapshot(snp.Id)
		assertTestResult(t, snp, snpNew)

		snpListExpected := []*model.VolumeSnapshotSpec{snp}
		snpListGot, _ := fakeVol.ListVolumeSnapshots()
		assertTestResult(t, snpListExpected, snpListGot)

		fakeVol.DeleteVolumeSnapshot(snp.Id, nil)
		snpListGot, _ = fakeVol.ListVolumeSnapshots()
		assertTestResult(t, 0, len(snpListGot))
	})

	t.Run("delete volume snapshot failed", func(t *testing.T) {
		reset()
		snpId := "1"
		err := fakeVol.DeleteVolumeSnapshot(snpId, nil)
		assertTestResult(t, fmt.Sprintf("snapshot %s cannot be found", snpId), err.Error())
	})

	t.Run("update volume snapshot successfully", func(t *testing.T) {
		reset()
		snpGot, _ := fakeVol.CreateVolumeSnapshot(req)
		assertTestResult(t, req.Name, snpGot.Name)

		newDescription := "volume snapshot for update"
		snpGot.Description = newDescription
		snpNew, _ := fakeVol.UpdateVolumeSnapshot(snpGot.Id, snpGot)
		assertTestResult(t, newDescription, snpNew.Description)
	})

	t.Run("update volume snapshot failed", func(t *testing.T) {
		reset()
		newDescription := "volume snapshot for update"
		req.Description = newDescription
		_, err := fakeVol.UpdateVolumeSnapshot(req.Id, req)
		assertTestResult(t, fmt.Sprintf("volume snapshot %s cannot be found", req.Id), err.Error())
	})
}

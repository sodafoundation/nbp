// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
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
	"fmt"

	"google.golang.org/grpc/codes"

	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/opensds/nbp/client/iscsi"
	sdscontroller "github.com/opensds/nbp/client/opensds"
	"github.com/opensds/nbp/driver"
	"github.com/opensds/opensds/pkg/model"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
)

////////////////////////////////////////////////////////////////////////////////
//                            Node Service                                    //
////////////////////////////////////////////////////////////////////////////////

func (p *Plugin) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (p *Plugin) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodePublishVolume implementation
func (p *Plugin) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	glog.Info("start to NodePublishVolume")
	defer glog.Info("end to NodePublishVolume")

	volId := req.VolumeId
	attachId := req.PublishInfo[KPublishAttachId]
	client := sdscontroller.GetClient("", "")
	if r := getReplicationByVolume(volId); r != nil {
		if r.ReplicationStatus == model.ReplicationFailover {
			volId = r.SecondaryVolumeId
			attachId = req.PublishInfo[KPublishSecondaryAttachId]
		}
		if r.Metadata == nil {
			r.Metadata = make(map[string]string)
		}
		r.Metadata[KAttachedVolumeId] = volId
		if _, err := client.UpdateReplication(r.Id, r); err != nil {
			msg := fmt.Sprintf("update replication(%s) failed, %v", r.Id, err)
			glog.Error(msg)
			return nil, status.Error(codes.FailedPrecondition, msg)
		}
	}

	//check volume is exist
	volSpec, errVol := client.GetVolume(volId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", volId)
		return nil, status.Error(codes.NotFound, msg)
	}

	atc, atcErr := client.GetVolumeAttachment(attachId)
	if atcErr != nil || atc == nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}

	var targetPaths []string
	if tps, exist := atc.Metadata[KTargetPath]; exist && len(tps) != 0 {
		targetPaths = strings.Split(tps, ";")
		for _, tp := range targetPaths {
			if req.TargetPath == tp {
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}

		// if volume don't have MULTI_NODE capability, just termination.
		mode := req.VolumeCapability.AccessMode.Mode
		if mode != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY &&
			mode != csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER {
			msg := fmt.Sprintf("the volume %s has been published to this node.", volId)
			return nil, status.Error(codes.Aborted, msg)
		}
	}

	// if not attach before, attach first.
	if len(atc.Mountpoint) == 0 || atc.Mountpoint == "-" {
		volDriver := driver.NewVolumeDriver(atc.DriverVolumeType)
		if volDriver == nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("Unsupport driverVolumeType: %s", atc.DriverVolumeType))
		}

		device, err := volDriver.Attach(atc.ConnectionData)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		atc.Mountpoint = device

		_, err = client.UpdateVolumeAttachment(atc.Id, atc)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
	}

	// obtain attachments to decide if can format.
	atcs, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}
	format := true
	for _, attachSpec := range atcs {
		if attachSpec.VolumeId == volId {
			if _, exist := attachSpec.Metadata[KTargetPath]; exist {
				// The device is formatted, can't be reformat for shared storage.
				format = false
				break
			}
		}
	}

	// Format and Mount
	glog.Infof("[NodePublishVolume] device:%s TargetPath:%s", atc.Mountpoint, req.TargetPath)
	if format {
		err = iscsi.FormatAndMount(atc.Mountpoint, "", req.TargetPath)
	} else {
		err = iscsi.Mount(atc.Mountpoint, req.TargetPath)
	}
	if err != nil {
		return nil, err
	}

	targetPaths = append(targetPaths, req.TargetPath)
	atc.Metadata[KTargetPath] = strings.Join(targetPaths, ";")
	_, err = client.UpdateVolumeAttachment(atc.Id, atc)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}
	volSpec.Status = model.VolumeInUse
	_, err = client.UpdateVolume(volSpec.Id, volSpec)
	if err != nil {
		glog.Error("Error: update volume status failed")
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume implementation
func (p *Plugin) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	glog.Info("start to NodeUnpublishVolume")
	defer glog.Info("end to NodeUnpublishVolume")

	volId := req.VolumeId
	client := sdscontroller.GetClient("", "")
	if r := getReplicationByVolume(volId); r != nil {
		volId = r.Metadata[KAttachedVolumeId]
	}

	//check volume is exist
	volSpec, errVol := client.GetVolume(volId)
	if errVol != nil || volSpec == nil {
		msg := fmt.Sprintf("the volume %s is not exist", volId)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := client.ListVolumeAttachments()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to NodeUnpublish volume.")
	}

	var atc *model.VolumeAttachmentSpec
	// hostname, _ := os.Hostname()
	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}

	for _, attachSpec := range attachments {
		if attachSpec.VolumeId == volId && attachSpec.Host == localIqn {
			atc = attachSpec
			break
		}
	}

	if atc == nil {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	if _, exist := atc.Metadata[KTargetPath]; !exist {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	var modifyTargetPaths []string
	tpExist := false
	targetPaths := strings.Split(atc.Metadata[KTargetPath], ";")
	for index, path := range targetPaths {
		if path == req.TargetPath {
			modifyTargetPaths = append(targetPaths[:index], targetPaths[index+1:]...)
			tpExist = true
			break
		}
	}
	if !tpExist {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// Umount
	glog.Infof("[NodeUnpublishVolume] TargetPath:%s", req.TargetPath)
	err = iscsi.Umount(req.TargetPath)
	if err != nil {
		glog.Errorf("unmount %v", err)
		return nil, err
	}

	if len(modifyTargetPaths) == 0 {
		volDriver := driver.NewVolumeDriver(atc.DriverVolumeType)
		if volDriver == nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("Unsupport driverVolumeType: %s", atc.DriverVolumeType))
		}

		err := volDriver.Detach(atc.ConnectionData)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "%s", err.Error())
		}
		atc.Mountpoint = "-"
	}

	atc.Metadata[KTargetPath] = strings.Join(modifyTargetPaths, ";")
	_, err = client.UpdateVolumeAttachment(atc.Id, atc)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Failed to NodeUnpublish volume.")
	}
	volSpec.Status = model.VolumeAvailable
	_, err = client.UpdateVolume(volSpec.Id, volSpec)
	if err != nil {
		glog.Info("Error: update volume status failed")
		return nil, status.Error(codes.FailedPrecondition, "Failed to publish node.")
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// GetNodeID implementation
func (p *Plugin) NodeGetId(
	ctx context.Context,
	req *csi.NodeGetIdRequest) (
	*csi.NodeGetIdResponse, error) {

	glog.Info("start to GetNodeID")
	defer glog.Info("end to GetNodeID")

	iqns, _ := iscsi.GetInitiator()
	localIqn := ""
	if len(iqns) > 0 {
		localIqn = iqns[0]
	}

	return &csi.NodeGetIdResponse{
		NodeId: localIqn,
	}, nil
}

// NodeGetInfo
func (p *Plugin) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetCapabilities implementation
func (p *Plugin) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	glog.Info("start to NodeGetCapabilities")
	defer glog.Info("end to NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			&csi.NodeServiceCapability{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

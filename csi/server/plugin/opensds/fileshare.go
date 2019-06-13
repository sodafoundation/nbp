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
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/client"
	"github.com/opensds/opensds/contrib/connector"
	"github.com/opensds/opensds/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	FileshareStorageType = "file"
)

type FileShare struct {
	Client *client.Client
}

func NewFileshare(c *client.Client) *FileShare {
	return &FileShare{Client: c}
}

func (f *FileShare) CreateFileShare(req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	var profileId, name, availabilityZone string
	var size int64
	// fileshare name
	name = strings.Replace(req.GetName(), "-", "_", -1)

	var attachMode = "read,write"
	for k, v := range req.GetParameters() {
		switch k {
		case ParamProfile:
			if v == "" {
				msg := "profile id cannot be empty"
				glog.Error(msg)
				return nil, status.Error(codes.InvalidArgument, msg)
			}
			// profile id
			profileId = v
		case PublishAttachMode:
			if strings.ToLower(v) == "read" {
				// attach mode
				attachMode = "read"
			} else {
				glog.Infof("use default attach mode: %s", attachMode)
			}
		}
	}

	// check profile is valid
	prf, err := f.Client.GetProfile(profileId)
	if err != nil {
		msg := fmt.Sprintf("get profile %s failed", profileId)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if FileshareStorageType != prf.StorageType {
		msg := fmt.Sprintf("the input storage type %s and storage type %s in profile %s are inconsistent",
			FileshareStorageType, prf.StorageType, profileId)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	// size
	size = getSize(req.GetCapacityRange())

	// availability zone
	if req.GetAccessibilityRequirements() != nil {
		availabilityZone = getZone(req.GetAccessibilityRequirements())
	}

	glog.Infof("find if fileshare %s has been created successfully", name)
	shareExist, err := f.FindFileshare(name)
	if err != nil {
		return nil, err
	}

	if shareExist == nil {
		glog.Infof("the fileshare %s does not exist and now create it", name)

		filesharebody := &model.FileShareSpec{
			Name:             name,
			AvailabilityZone: availabilityZone,
			ProfileId:        profileId,
			Size:             size,
		}

		shareExist, err = f.Client.CreateFileShare(filesharebody)
		if err != nil {
			msg := fmt.Sprintf("create file share failed: %v", err)
			glog.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}

	glog.Info("wait for the fileshare to be created successfully")

	shareStable, err := waitForStatusStable(shareExist.Id, func(id string) (interface{}, error) {
		return f.Client.GetFileShare(id)
	})

	if err != nil {
		msg := fmt.Sprintf("failed to create fileshare %s: %v", name, err)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	share := shareStable.(*model.FileShareSpec)

	shareinfo := &csi.Volume{
		CapacityBytes: share.Size * util.GiB,
		VolumeId:      share.Id,
		VolumeContext: map[string]string{
			ShareName:         share.Name,
			ShareAZ:           share.AvailabilityZone,
			ShareStatus:       share.Status,
			SharePoolId:       share.PoolId,
			ShareProfileId:    share.ProfileId,
			ShareProtocol:     share.Protocols[0],
			PublishAttachMode: attachMode,
			ExportLocations:   strings.Join(share.ExportLocations, ","),
		},

		AccessibleTopology: []*csi.Topology{
			{
				Segments: map[string]string{
					TopologyZoneKey: share.AvailabilityZone,
				},
			},
		},
	}

	return &csi.CreateVolumeResponse{
		Volume: shareinfo,
	}, nil
}

func (f *FileShare) FindFileshare(fileshareName string) (*model.FileShareSpec, error) {
	shares, err := f.Client.ListFileShares()
	if err != nil {
		msg := fmt.Sprintf("list file shares failed: %v", err)
		glog.Error(msg)
		return nil, errors.New(msg)
	}

	for _, share := range shares {
		if share.Name == fileshareName {
			return share, nil
		}
	}

	return nil, nil
}

func (f *FileShare) DeleteFileShare(shareID string) (*csi.DeleteVolumeResponse, error) {
	share, _ := f.Client.GetFileShare(shareID)
	if share == nil {
		return &csi.DeleteVolumeResponse{}, nil
	}

	err := f.Client.DeleteFileShare(shareID)
	if err != nil {
		msg := fmt.Sprintf("delete share failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (f *FileShare) ControllerPublishFileShare(req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	attachMode, ok := req.VolumeContext[PublishAttachMode]
	if !ok {
		glog.Info("attach mode will use default value: read,write")
		attachMode = "read,write"
	}

	accessTo := strings.Split(req.GetNodeId(), ",")[IpIdx]
	// check if fileshare exists
	shareSpec, err := f.Client.GetFileShare(req.VolumeId)
	if err != nil || shareSpec == nil {
		msg := fmt.Sprintf("the fileshare %s does not exist: %v", req.VolumeId, err)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	_, err = f.getProtoFromPool(shareSpec.PoolId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	attachReq := &model.FileShareAclSpec{
		FileShareId: shareSpec.Id,
		// Only support ip based mode
		Type:             "ip",
		AccessCapability: strings.Split(attachMode, ","),
		AccessTo:         accessTo,
		ProfileId:        shareSpec.ProfileId,
	}

	mode := req.VolumeCapability.AccessMode.Mode
	canAtMultiNode := false

	if csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY == mode ||
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER == mode {
		canAtMultiNode = true
	}

	err = f.isFileshareCanBePublished(canAtMultiNode, attachReq)
	if err != nil {
		return nil, err
	}

	newAttachment, errAttach := f.Client.CreateFileShareAcl(attachReq)
	if errAttach != nil {
		msg := fmt.Sprintf("the fileshare %s is failed to be published to node %s, error info: %v",
			req.VolumeId, req.NodeId, errAttach)
		glog.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	resp := &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			PublishHostIp:     attachReq.AccessTo,
			PublishAttachId:   newAttachment.Id,
			PublishAttachMode: attachMode,
			ExportLocations:   req.GetVolumeContext()[ExportLocations],
			FileShareName:     shareSpec.Name,
		},
	}

	return resp, nil
}

func (f *FileShare) getProtoFromPool(poolId string) (string, error) {
	// get protocol from pool
	pool, err := f.Client.GetPool(poolId)
	if err != nil || pool == nil {
		msg := fmt.Sprintf("the pool %s does not exist: %v", poolId, err)
		glog.Error(msg)
		return "", errors.New(msg)
	}

	var protocol = strings.ToLower(pool.Extras.IOConnectivity.AccessProtocol)

	glog.V(5).Infof("the fileshare protocol is %s", protocol)

	if protocol != strings.ToLower(NFS) && protocol != NFS {
		return "", errors.New("only support nfs protocol")
	}

	return protocol, nil
}

func (f *FileShare) ControllerUnpublishFileShare(req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	//check volume is exist
	shareSpec, err := f.Client.GetFileShare(req.VolumeId)
	if err != nil || shareSpec == nil {
		msg := fmt.Sprintf("the fileshare %s does not exist: %v", req.VolumeId, err)
		glog.Error(msg)
		return nil, status.Error(codes.NotFound, msg)
	}

	attachments, err := f.Client.ListFileSharesAcl()
	if err != nil {
		msg := fmt.Sprintf("list fileshare access clients failed: %v", err)
		glog.Error(msg)
		fmt.Println(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	accessTo := strings.Split(req.GetNodeId(), ",")[IpIdx]

	for _, attachSpec := range attachments {
		if attachSpec.FileShareId == shareSpec.Id && attachSpec.AccessTo == accessTo {
			if ok := UnpublishAttachmentList.isExist(attachSpec.Id); !ok {
				glog.Infof("add attachment id %s into unpublish attachment list", attachSpec.Id)
				UnpublishAttachmentList.Add(attachSpec)
				UnpublishAttachmentList.PrintList(FileshareStorageType)
			}
			break
		}
	}

	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func extractNFSInitiatorFromNodeInfo(nodeInfo string) string {
	return strings.Split(nodeInfo, ",")[IpIdx]
}

func (f *FileShare) isFileshareCanBePublished(canAtMultiNode bool, attachReq *model.FileShareAclSpec) error {
	attachments, err := f.Client.ListFileSharesAcl()
	if err != nil {
		msg := fmt.Sprintf("list fileshare access clients failed: %v", err)
		glog.Error(msg)
		return status.Error(codes.FailedPrecondition, msg)
	}

	glog.V(5).Infof("access clients are %#v", attachments)

	for _, attachSpec := range attachments {
		if attachSpec.FileShareId == attachReq.FileShareId && attachSpec.AccessTo != attachReq.AccessTo {
			if !canAtMultiNode {
				msg := fmt.Sprintf("the fileshare %s has been published to the node %s and kubernetes does not have MULTI_NODE volume capability",
					attachReq.FileShareId, attachSpec.AccessTo)
				glog.Error(msg)
				return status.Error(codes.FailedPrecondition, msg)
			}
		}
	}

	glog.Infof("the fileshare %s can be published", attachReq.FileShareId)
	return nil
}

func (f *FileShare) ListFileShares(req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	shares, err := f.Client.ListFileShares()
	if err != nil {
		return nil, err
	}

	ens := []*csi.ListVolumesResponse_Entry{}
	for _, v := range shares {
		if v != nil {
			shareinfo := &csi.Volume{
				CapacityBytes: v.Size,
				VolumeId:      v.Id,
				VolumeContext: map[string]string{
					"Name":             v.Name,
					"AvailabilityZone": v.AvailabilityZone,
					"PoolId":           v.PoolId,
					"ProfileId":        v.ProfileId,
				},
			}

			ens = append(ens, &csi.ListVolumesResponse_Entry{
				Volume: shareinfo,
			})
		}
	}

	return &csi.ListVolumesResponse{
		Entries: ens,
	}, nil
}

func (f *FileShare) NodeStageFileShare(req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	mountpoint := req.GetStagingTargetPath()
	if mountpoint == "" {
		return nil, status.Error(codes.InvalidArgument, "staging target path cannot be empty")
	}

	ctx := req.GetPublishContext()
	if ctx == nil {
		return nil, status.Error(codes.InvalidArgument, "publish context cannot be nil")
	}

	exportLocations := strings.Split(ctx[ExportLocations], ",")

	fsConnector := connector.NewConnector(NFS)
	if nil == fsConnector {
		msg := fmt.Sprintf("unsupport file share driver type: %s", NFS)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	connectData := map[string]interface{}{ExportLocations: exportLocations}

	device, err := fsConnector.Attach(connectData)
	if nil != err || device == "" {
		msg := fmt.Sprintf("failed to find device: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	mnt := req.VolumeCapability.GetMount()
	glog.V(5).Infof("fileshare capability Mount=%+v", mnt)

	// Mount
	mounted, err := connector.IsMounted(mountpoint)
	if err != nil {
		msg := fmt.Sprintf("failed to check mounted: %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	if mounted {
		glog.Info("fileshare is already mounted.")
		return &csi.NodeStageVolumeResponse{}, nil
	}

	glog.Info("mounting...")

	err = connector.Mount(device, mountpoint, strings.ToLower(NFS), mnt.MountFlags)
	if err != nil {
		msg := fmt.Sprintf("failed to mount: %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.FailedPrecondition, "%s", msg)
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (f *FileShare) NodeUnstageFileShare(req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	// check volume is unmounted
	stagingTargetPath := req.GetStagingTargetPath()
	if stagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "staging target path cannot be empty")
	}

	mounted, err := connector.IsMounted(stagingTargetPath)
	if err != nil {
		msg := fmt.Sprintf("check volume is unmounted failed: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	if !mounted {
		glog.Info("target path is already unmounted")
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	// Umount
	glog.Infof("node unpublish volume mountpoint: %s", stagingTargetPath)
	err = connector.Umount(stagingTargetPath)
	if err != nil {
		msg := fmt.Sprintf("failed to umount, %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (f *FileShare) NodePublishFileShare(req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	device := req.GetStagingTargetPath()
	mountpoint := req.GetTargetPath()

	if device == "" || mountpoint == "" {
		return nil, status.Error(codes.FailedPrecondition, "staging target path and target path cannot be empty")
	}

	// Bind mount
	mnt := req.GetVolumeCapability().GetMount()
	mountFlags := append(mnt.GetMountFlags(), "bind")

	glog.V(5).Infof("read only: %v", req.GetReadonly())

	if req.GetReadonly() {
		mountFlags = append(mountFlags, "ro")
	}

	glog.V(5).Infof("mount flags: %v", mountFlags)

	// Mount
	mounted, err := connector.IsMounted(mountpoint)
	if err != nil {
		msg := fmt.Sprintf("failed to check mounted: %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	if mounted {
		glog.Infof("fileshare is already mounted to %s", mountpoint)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	glog.Info("mounting...")

	err = connector.Mount(device, mountpoint, strings.ToLower(NFS), mountFlags)
	if err != nil {
		msg := fmt.Sprintf("failed to mount: %v", err)
		glog.Errorf(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (f *FileShare) NodeUnpublishFileShare(req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	// check volume is unmounted
	targetPath := req.GetTargetPath()
	if targetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "target path cannot be empty")
	}

	mounted, err := connector.IsMounted(targetPath)
	if !mounted {
		glog.Info("target path is already unmounted")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	// Umount
	glog.V(5).Infof("mountpoint: %s", targetPath)
	err = connector.Umount(targetPath)
	if err != nil {
		msg := fmt.Sprintf("failed to umount: %v", err)
		glog.Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

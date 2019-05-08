/*
Copyright 2016 The Kubernetes Authors.
Copyright (c) 2016 Huawei Technologies Co., Ltd. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package volume

import (
	"strconv"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/opensds/nbp/opensds-provisioner/pkg/client"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/volume"
)

const (
	// are we allowed to set this? else make up our own
	annCreatedBy = "kubernetes.io/createdby"
	createdBy    = "opensds-provisioner"

	// A PV annotation for the identity of the s3fsProvisioner that provisioned it
	annProvisionerID = "opensds-provisioner"
)

// NewProvisioner creates a new provisioner
func NewOpensdsProvisioner(client kubernetes.Interface, sdsclient client.WarpOpensdsClient) controller.Provisioner {
	return newProvisionerInternal(client, sdsclient)
}

func newProvisionerInternal(client kubernetes.Interface, sdsclient client.WarpOpensdsClient) *opensdsProvisioner {
	var identity types.UID = "opensds-provisioner"

	provisioner := &opensdsProvisioner{
		client:    client,
		identity:  identity,
		sdsclient: sdsclient,
	}

	return provisioner
}

type opensdsProvisioner struct {
	client    kubernetes.Interface
	identity  types.UID
	sdsclient client.WarpOpensdsClient
}

var _ controller.Provisioner = &opensdsProvisioner{}

// Provision creates a volume i.e. the storage asset and returns a PV object for
// the volume.
func (p *opensdsProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	volId, err := p.createVolume(options)
	if err != nil {
		return nil, err
	}

	annotations := make(map[string]string)
	annotations[annCreatedBy] = createdBy

	annotations[annProvisionerID] = string(p.identity)

	fstype := "ext4"
	if _, exist := options.Parameters[client.KFsType]; exist {
		fstype = options.Parameters[client.KFsType]
	}
	/*
		This PV won't work since there's nothing backing it.  the flex script
		is in flex/flex/flex  (that many layers are required for the flex volume plugin)
	*/
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        volId,
			Labels:      map[string]string{},
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{

				FlexVolume: &v1.FlexVolumeSource{
					Driver: "opensds.io/opensds",
					FSType: fstype,
					Options: map[string]string{
						client.KVolumeId: volId,
					},
					ReadOnly: false,
				},
			},
		},
	}

	return pv, nil
}

func (p *opensdsProvisioner) createVolume(volumeOptions controller.VolumeOptions) (string, error) {
	capacity := volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestBytes := capacity.Value()
	//requestBytes := volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)].Value()
	size := volume.RoundUpSize(requestBytes, 1024*1024*1024)

	opts := map[string]string{
		client.KVolumeName: volumeOptions.PVName,
		client.KVolumeSize: strconv.FormatInt(size, 10),
	}
	for key, value := range volumeOptions.Parameters {
		opts[key] = value
	}

	volId, err := p.sdsclient.Provision(opts)
	if err != nil {
		glog.Errorf("Failed to create volume %s, error: %s", volumeOptions, err.Error())
		return "", err
	}

	return volId, nil
}

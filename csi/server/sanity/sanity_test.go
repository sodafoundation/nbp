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
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"testing"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	"github.com/opensds/nbp/csi/server/plugin/opensds"
	"github.com/opensds/nbp/csi/util"
	"github.com/opensds/opensds/client"
	c "github.com/opensds/opensds/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/opensds/opensds/contrib/connector/samplefortest"
)

//start sanity test for driver
func TestDriver(t *testing.T) {
	// Initialize csi plugin
	client := &client.Client{
		VolumeMgr:      &c.VolumeMgr{Receiver: &fakeVolume{}},
		ReplicationMgr: &c.ReplicationMgr{Receiver: &fakeReplication{}},
		ProfileMgr:     &c.ProfileMgr{Receiver: &fakeProfile{}},
		PoolMgr:        &c.PoolMgr{Receiver: &fakePool{}},
	}

	fakePlugin := &opensds.Plugin{
		Client: client,
		VolumeClient: &opensds.Volume{
			Client:  client,
			Mounter: getFakeMounter(),
		},
		PluginStorageType: "block",
		Mounter:           getFakeMounter(),
	}

	go fakePlugin.UnpublishRoutine()

	// New Grpc Server
	s := grpc.NewServer()

	// Register CSI Service
	csi.RegisterIdentityServer(s, fakePlugin)
	csi.RegisterControllerServer(s, fakePlugin)
	csi.RegisterNodeServer(s, fakePlugin)

	// Register reflection Service
	reflection.Register(s)

	// Get CSI Endpoint Listener
	lis, err := util.GetCSIEndPointListener(util.CSIDefaultEndpoint)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	// Remove sock file
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		for sig := range sigs {
			if sig == syscall.SIGKILL ||
				sig == syscall.SIGQUIT ||
				sig == syscall.SIGHUP ||
				sig == syscall.SIGTERM ||
				sig == syscall.SIGINT {
				t.Log("exit to serve")
				if lis.Addr().Network() == "unix" {
					sockfile := lis.Addr().String()
					os.RemoveAll(sockfile)
					t.Logf("remove sock file: %s", sockfile)
				}
			}
		}
	}()

	// Serve Plugin Server
	go func() {
		t.Logf("start to serve: %s", lis.Addr())
		s.Serve(lis)
	}()

	// Initialize the sanity
	mountDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mountDir)

	mountStageDir, err := ioutil.TempDir("", "temp-stage")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(mountStageDir)

	config := &sanity.Config{
		TargetPath:           mountDir,
		StagingPath:          mountStageDir,
		Address:              lis.Addr().String(),
		TestVolumeParameters: map[string]string{opensds.ParamProfile: "1106b972-66ef-11e7-b172-db03f3689c9c"},
	}
	// start test
	sanity.Test(t, config)
}

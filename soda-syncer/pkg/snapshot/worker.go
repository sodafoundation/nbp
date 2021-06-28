// Copyright 2021 The SodaFoundation Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http:#www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"net/http"
	"os/exec"
)

type SnapshotProfile struct {
	AwsAccesskey     string `json:"AWS_ACCESS_KEY_ID"`
	AwsSecretkey     string `json:"AWS_SECRET_ACCESS_KEY"`
	ResticRepository string `json:"RESTIC_REPOSITORY"`
	ResticPassword   string `json:"RESTIC_PASSWORD"`
	TimeInterval     int    `json:"timeInterval"`
	ResticSourceRepo string `json:"resticSourceRepo"`
}

func CreateSnapshot(w http.ResponseWriter, r *http.Request) {

	// TODO Remove this Sleep time after changing provisioner to send request after PVC Attachment event
	time.Sleep(20 * time.Second)
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("---------------------Getting Source Mount Point ---------------------")
	//TODO Configure grep based on the pvc name recieved in this request
	cmdToGetPVCMountPoint := "df -h --output=target | grep csi"
	cmdToGetFileSystemMountPoint, errno := exec.Command("bash", "-c", cmdToGetPVCMountPoint).Output()
	if errno != nil {
		fmt.Sprintf("Failed to execute command: %s", cmdToGetFileSystemMountPoint)
	}
	mountPointToBeBackedUp := string(cmdToGetFileSystemMountPoint)
	fmt.Println("The MountPoint to be backed up : ", mountPointToBeBackedUp)

	fmt.Println("---------------------Backing Up ---------------------")
	var snapshotProfile SnapshotProfile
	_ = json.NewDecoder(r.Body).Decode(&snapshotProfile)
	fmt.Println(snapshotProfile)
	os.Setenv("AWS_ACCESS_KEY_ID", snapshotProfile.AwsAccesskey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", snapshotProfile.AwsSecretkey)
	os.Setenv("RESTIC_REPOSITORY", snapshotProfile.ResticRepository)
	os.Setenv("RESTIC_PASSWORD", snapshotProfile.ResticPassword)

	timeD := time.Duration(snapshotProfile.TimeInterval)
	ticker := time.NewTicker(timeD * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Backup Started at", t)
				cmdToDoBackup := "restic backup " + mountPointToBeBackedUp
				cmdOutputForBackup, errBck := exec.Command("bash", "-c", cmdToDoBackup).Output()
				if errBck != nil {
					fmt.Sprintf("Failed to execute command: %s", string(cmdOutputForBackup))
				}
				fmt.Println("Backup Success : ", string(cmdOutputForBackup))

				fmt.Println("---------------------SnapShots---------------------")
				cmd := exec.Command("restic", "snapshots")
				cmd.Stderr = os.Stdout
				cmd.Stdout = os.Stdout

				if err := cmd.Run(); err != nil {
					fmt.Println(err)
					json.NewEncoder(w).Encode(err)
					return
				}
			}
		}
	}()
	json.NewEncoder(w).Encode("Back Up Done Successfully")
}

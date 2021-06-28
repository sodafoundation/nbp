package snapshot

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"

	"encoding/json"
	"github.com/sodafoundation/nbp/client/opensds"
	"net/http"
	"os"
)

//type CustomPropertiesSpec map[string]interface{}
//
//func (cps CustomPropertiesSpec) GetSnapshotProfile() (string, string, string, string, int) {
//	var timeInterval int
//	timeFloat := cps["TimeInterval"].(float64)
//	timeInterval = int(timeFloat)
//	return fmt.Sprintf("%v", cps["AWS_ACCESS_KEY_ID"]), fmt.Sprintf("%v", cps["AWS_SECRET_ACCESS_KEY"]),
//	fmt.Sprintf("%v", cps["RESTIC_REPOSITORY"]), fmt.Sprintf("%v", cps["RESTIC_PASSWORD"]), timeInterval
//}

type SnapshotProfile struct {
	AwsAccesskey     string `json:"AWS_ACCESS_KEY_ID"`
	AwsSecretkey     string `json:"AWS_SECRET_ACCESS_KEY"`
	ResticRepository string `json:"RESTIC_REPOSITORY"`
	ResticPassword   string `json:"RESTIC_PASSWORD"`
	TimeInterval     int    `json:"timeInterval"`
	ResticSourceRepo string `json:"resticSourceRepo"`
}

func CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	opensdsEndpoint, errLookUp := os.LookupEnv("OPENSDS_ENDPOINT")
	if !errLookUp {
		fmt.Println("No env variables found for endpoint, switching to default")
		opensdsEndpoint = "http://127.0.0.1:50040"
	}
	client, err := opensds.GetClient(opensdsEndpoint, "keystone")
	if client == nil || err != nil {
		fmt.Printf("get opensds client failed: %v", err)
	}
	profile, errosds := client.GetProfile(key)
	if errosds != nil {
		fmt.Printf("got error in GetProfile  : %s", errosds.Error())
	}

	var timeInterval int
	timeFloat := profile.CustomProperties["TimeInterval"].(float64)
	timeInterval = int(timeFloat)

	snapshotProfile := SnapshotProfile{AwsAccesskey: fmt.Sprintf("%v", profile.CustomProperties["AWS_ACCESS_KEY_ID"]), AwsSecretkey: fmt.Sprintf("%v", profile.CustomProperties["AWS_SECRET_ACCESS_KEY"]), ResticRepository: fmt.Sprintf("%v", profile.CustomProperties["RESTIC_REPOSITORY"]), ResticPassword: fmt.Sprintf("%v", profile.CustomProperties["RESTIC_PASSWORD"]), TimeInterval: timeInterval, ResticSourceRepo: ""}

	postBody, _ := json.Marshal(snapshotProfile)
	fmt.Println(postBody)
	requestBody := bytes.NewBuffer(postBody)
	syncerEndpoint := os.Getenv("NODE_IP")

	response, err := http.Post("http://" +syncerEndpoint+":50030/snapshot", "application/json", requestBody)
	fmt.Println(response.Body)
	json.NewEncoder(w).Encode(response.Body)
}

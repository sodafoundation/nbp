package profile

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sodafoundation/nbp/client/opensds"
	"net/http"
	"os"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(profile.CustomProperties)
}

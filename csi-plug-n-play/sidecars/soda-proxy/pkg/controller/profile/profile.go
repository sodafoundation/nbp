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

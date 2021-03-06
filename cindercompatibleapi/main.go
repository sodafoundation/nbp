// Copyright 2018 The OpenSDS Authors.
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

/*
This module implements a entry into the OpenSDS REST service.

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sodafoundation/api/pkg/utils/logs"
	"github.com/sodafoundation/nbp/cindercompatibleapi/api"
)

func main() {
	flag.Parse()
	logs.InitLogs(5 * time.Second)
	defer logs.FlushLogs()

	cinderEndpoint, ok := os.LookupEnv("CINDER_ENDPOINT")
	if !ok {
		fmt.Println("ERROR: You must provide the cinder endpoint by setting " +
			"the environment variable CINDER_ENDPOINT")
		return
	}

	api.Run(cinderEndpoint)
}

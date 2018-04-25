/*
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
package main

import (
	"context"
	gflag "flag"
	"fmt"
	"os"
	"path"

	"github.com/kubernetes-incubator/service-catalog/contrib/pkg/broker/server"
	"github.com/kubernetes-incubator/service-catalog/pkg"
	"github.com/opensds/nbp/service-broker/controller"
)

var options struct {
	Port     string
	Endpoint string
}

func init() {
	gflag.StringVar(&options.Port, "port", ":8005", "use '--port' option to specify the port for broker to listen on")
	gflag.StringVar(&options.Endpoint, "endpoint", "http://127.0.0.1:50040", "use '--endpoint' option to specify the client endpoint for broker to connect the backend")
	gflag.Parse()
}

func main() {
	if gflag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), pkg.VERSION)
		return
	}

	server.Run(context.Background(),
		options.Port, controller.CreateController(options.Endpoint))
}

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

package util

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/opensds/opensds/contrib/connector"
	"google.golang.org/grpc"
)

// getProtoandAdd return protocal and address
func getProtoAndAdd(target string) (string, string) {
	reg := `(?i)^((?:(?:tcp|udp|ip)[46]?)|` + `(?:unix(?:gram|packet)?))://(.+)$`
	t := regexp.MustCompile(reg).FindStringSubmatch(target)
	return t[1], t[2]
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetCSIEndPoint from environment variable
func GetCSIEndPoint(csiEndpoint string) (string, error) {
	// example: CSI_ENDPOINT=unix://path/to/unix/domain/socket.sock
	csiEndpoint = strings.TrimSpace(csiEndpoint)

	if csiEndpoint == "" {
		err := errors.New("csi endpoint is empty")
		log.Fatalf("%v", err)
		return csiEndpoint, err
	}

	return csiEndpoint, nil
}

// GetCSIEndPointListener from endpoint
func GetCSIEndPointListener(csiEndpoint string) (net.Listener, error) {
	target, err := GetCSIEndPoint(csiEndpoint)
	if err != nil {
		return nil, err
	}
	proto, addr := getProtoAndAdd(target)

	log.Printf("proto: %s addr: %s", proto, addr)
	if strings.HasPrefix(proto, "unix") {
		// clean up previous sock file.
		os.RemoveAll(addr)
		log.Printf("remove sock file: %s", addr)
		// Need to make directory at the first time the csi service runs.
		dir := path.Dir(addr)
		if exist, _ := PathExists(dir); !exist {
			os.MkdirAll(dir, 0755)
		}
	}

	return net.Listen(proto, addr)
}

// GetCSIClientConn from endpoint
func GetCSIClientConn(csiEndpoint string) (*grpc.ClientConn, error) {
	// Get parameters for grpc
	ctx := context.Background()
	target, err := GetCSIEndPoint(csiEndpoint)
	if err != nil {
		return nil, err
	}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(
			func(target string, timeout time.Duration) (net.Conn, error) {
				proto, addr := getProtoAndAdd(target)
				return net.DialTimeout(proto, addr, timeout)
			}),
	}

	// Set up a connection to the server
	return grpc.DialContext(ctx, target, dialOpts...)
}

// IsSupportPotocol ...
func IsSupportPotocol(protocol string) bool {
	supportPotocols := [...]string{connector.FcDriver, connector.IscsiDriver}
	for _, v := range supportPotocols {
		if strings.ToLower(protocol) == v {
			return true
		}
	}
	return false
}

// IsSupportFstype ...
func IsSupportFstype(fstype string) bool {
	supportFstypes := [...]string{"ext2", "ext3", "ext4", "cramfs", "minix", ""}
	for _, v := range supportFstypes {
		if strings.ToLower(fstype) == v {
			return true
		}
	}
	return false
}

// Contained ...
func Contained(obj, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	default:
		return false
	}
	return false
}

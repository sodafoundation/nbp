# Copyright 2018 The OpenSDS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL=/bin/bash
BASE_DIR := $(shell pwd)
BUILD_DIR := $(BASE_DIR)/build/out
IMAGE_TAG := latest
DIST_DIR := $(BASE_DIR)/build/dist
VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
                 git describe --match=$(git rev-parse --short=8 HEAD) \
		 --always --dirty --abbrev=8)
BUILD_TGT := opensds-sushi-$(VERSION)-linux-amd64
#.PHONY: all build prebuild csi.block.opensds csi.file.opensds flexvolume.server.opensds service-broker cindercompatibleapi docker clean
.PHONY: all build prebuild csi.block.opensds csi.file.opensds cindercompatibleapi docker clean

all: build
#build: csi.block.opensds csi.file.opensds flexvolume.server.opensds service-broker cindercompatibleapi
build: csi.block.opensds csi.file.opensds cindercompatibleapi

prebuild:
	mkdir -p  $(BUILD_DIR)

csi.block.opensds: prebuild
	go build -ldflags '-w -s' -o $(BUILD_DIR)/csi.block.opensds github.com/opensds/nbp/csi/cmd/block
	wget https://github.com/linux-nvme/nvme-cli/archive/v1.8.1.tar.gz -O ./nvmecli-1.8.1.tar.gz
	tar -zxf ./nvmecli-1.8.1.tar.gz -C ./
	cd ./nvme-cli-1.8.1 && sudo make && sudo make install
	cd ..
	cp -a ./nvme-cli-1.8.1 ./csi/

csi.file.opensds: prebuild
	go build -ldflags '-w -s' -o $(BUILD_DIR)/csi.file.opensds github.com/opensds/nbp/csi/cmd/file
	wget https://github.com/linux-nvme/nvme-cli/archive/v1.8.1.tar.gz -O ./nvmecli-1.8.1.tar.gz
	tar -zxf ./nvmecli-1.8.1.tar.gz -C ./
	cd ./nvme-cli-1.8.1 && sudo make && sudo make install
	cd ..
	cp -a ./nvme-cli-1.8.1 ./csi/

#flexvolume.server.opensds: prebuild
#	go build -ldflags '-w -s' -o $(BUILD_DIR)/flexvolume.server.opensds github.com/opensds/nbp/flexvolume/cmd/flex-plugin

#service-broker: prebuild
#	go build -ldflags '-w -s' -o $(BUILD_DIR)/service-broker github.com/opensds/nbp/service-broker/cmd/service-broker

cindercompatibleapi: prebuild
	go build -ldflags '-w -s' -o $(BUILD_DIR)/cindercompatibleapi github.com/opensds/nbp/cindercompatibleapi

docker: build
	cp $(BUILD_DIR)/csi.block.opensds ./csi/
	cp $(BUILD_DIR)/csi.file.opensds ./csi/
#	cp $(BUILD_DIR)/service-broker ./service-broker/cmd/service-broker
	docker build -f csi/cmd/block/Dockerfile -t opensdsio/csiplugin-block:$(IMAGE_TAG) csi
	docker build -f csi/cmd/file/Dockerfile -t opensdsio/csiplugin-file:$(IMAGE_TAG) csi
#	docker build service-broker/cmd/service-broker -t opensdsio/service-broker:$(IMAGE_TAG)

goimports:
	goimports -w $(shell go list -f {{.Dir}} ./... |grep -v /vendor/)

clean:
	rm -rf $(BUILD_DIR) ./csi/csi.block.opensds ./csi/csi.file.opensds \
		./service-broker/cmd/service-broker/service-broker \
		./csi/nvme-cli-1.8.1

version:
	@echo ${VERSION}

.PHONY: dist
dist: build
	( \
	    rm -fr $(DIST_DIR) && mkdir $(DIST_DIR) && \
	    cd $(DIST_DIR) && \
#	    mkdir -p $(BUILD_TGT)/{csi,flexvolume,provisioner,service-broker} && \
	    mkdir -p $(BUILD_TGT)/{csi,provisioner} && \
	    cp -r $(BUILD_DIR) $(BUILD_TGT)/bin/ && \
	    cp -r $(BASE_DIR)/csi/deploy $(BUILD_TGT)/csi/ && \
	    cp -r $(BASE_DIR)/csi/examples $(BUILD_TGT)/csi/ && \
#	    cp -r $(BASE_DIR)/flexvolume/examples $(BUILD_TGT)/flexvolume/ && \
	    cp -r $(BASE_DIR)/opensds-provisioner/deploy $(BUILD_TGT)/provisioner/ && \
	    cp -r $(BASE_DIR)/opensds-provisioner/examples $(BUILD_TGT)/provisioner/ && \
#	    cp -r $(BASE_DIR)/service-broker/examples $(BUILD_TGT)/service-broker/ && \
	    cp $(BASE_DIR)/LICENSE $(BUILD_TGT)/ && \
	    zip -r $(DIST_DIR)/$(BUILD_TGT).zip $(BUILD_TGT) && \
	    tar zcvf $(DIST_DIR)/$(BUILD_TGT).tar.gz $(BUILD_TGT) && \
	    tree \
	)

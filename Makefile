.PHONY: all build package csi.server.opensds csi.client.opensds docker clean

all:build

build:csi.server.opensds csi.client.opensds

package:
	go get github.com/opensds/nbp/csi/server
	go get github.com/opensds/nbp/csi/client

csi.server.opensds:package
	mkdir -p  ./.output/
	go build -o ./.output/csi.server.opensds github.com/opensds/nbp/csi/server

csi.client.opensds:package
	mkdir -p  ./.output/
	go build -o ./.output/csi.client.opensds github.com/opensds/nbp/csi/client

docker:build
	cp ./.output/csi.server.opensds ./csi/server
	cp ./.output/csi.client.opensds ./csi/client
	docker build csi/server -t csi.server.opensds:0.0.1
	docker build csi/client -t csi.client.opensds:0.0.1	
	# csi.server.opensds docker run usage:
	# docker run -it -v /var/lib/docker:/tmp/ -e CSI_ENDPOINT=$CSI_ENDPOINT -e OPENSDS_ENDPOINT=$OPENSDS_ENDPOINT csi.server.opensds:0.0.1
	# csi.client.opensds docker run usage:
	# docker run -it -v /var/lib/docker:/tmp/ -e CSI_ENDPOINT=$CSI_ENDPOINT csi.client.opensds:0.0.1
clean:
	rm -rf ./.output/* ./csi/server/csi.server.opensds ./csi/client/csi.client.opensds

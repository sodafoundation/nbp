# Docker build usage: docker build . -t csi.client.opensds:0.0.1
# Docker run usage:	docker run -it -v /var/lib/docker:/tmp/
#                                  -e CSI_ENDPOINT=$CSI_ENDPOINT
#                                  csi.client.opensds:0.0.1

FROM golang:1.9
MAINTAINER Edison Xiang <xiang.edison@gmail.com>

ARG CSI_ENDPOINT

ENV CSI_ENDPOINT $CSI_ENDPOINT

COPY csi.client.opensds /usr/bin

# Define default command.
CMD ["/usr/bin/csi.client.opensds"]

# Docker build usage: docker build . -t csi.server.opensds:0.0.1
# Docker run usage:	docker run -it -v /var/lib/docker:/tmp/
#                                  -e CSI_ENDPOINT=$CSI_ENDPOINT
#                                  -e OPENSDS_ENDPOINT=$OPENSDS_ENDPOINT
#                                  csi.server.opensds:0.0.1

FROM golang:1.9
MAINTAINER Edison Xiang <xiang.edison@gmail.com>

ARG CSI_ENDPOINT
ARG OPENSDS_ENDPOINT

ENV CSI_ENDPOINT $CSI_ENDPOINT
ENV OPENSDS_ENDPOINT $OPENSDS_ENDPOINT

COPY csi.server.opensds /usr/bin

# Define default command.
CMD ["/usr/bin/csi.server.opensds"]

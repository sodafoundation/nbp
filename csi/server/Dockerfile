# Based on ubuntu
FROM ubuntu
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin"

# Copy opensdsplugin from build directory
COPY csi.server.opensds /csi.server.opensds

# Install iscsi
RUN apt-get update

RUN apt-get -y install open-iscsi
RUN apt-get -y install kmod
RUN apt-get -y install ceph-common

# Define default command
ENTRYPOINT ["/csi.server.opensds"]

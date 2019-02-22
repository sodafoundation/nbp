# Based on ubuntu
FROM ubuntu:xenial
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin"

# Copy opensdsplugin from build directory
COPY csi.server.opensds /csi.server.opensds

# Install iscsi
RUN apt-get update && \
    apt-get -y install open-iscsi \
    sysfsutils \
    kmod \
    ceph-common \
    iputils-ping && \
    rm -rf /var/lib/apt/lists/*

# Define default command
ENTRYPOINT ["/csi.server.opensds"]

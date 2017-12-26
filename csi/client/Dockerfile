# Based on centos
FROM centos:7.4.1708
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin Client"

# Copy opensdsplugin client from build directory
COPY csi.client.opensds /csi.client.opensds

# Install e4fsprogs for format
RUN yum -y install e4fsprogs

# Define default command
ENTRYPOINT ["/csi.client.opensds"]

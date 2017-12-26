# Based on centos
FROM centos:7.4.1708
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin"

# Copy opensdsplugin from build directory
COPY csi.server.opensds /csi.server.opensds

# Install e4fsprogs for format
RUN yum -y install e4fsprogs

# Define default command
ENTRYPOINT ["/csi.server.opensds"]

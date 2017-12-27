# Based on ubuntu
FROM ubuntu
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin"

# Copy opensdsplugin from build directory
COPY csi.server.opensds /csi.server.opensds

# Install iscsi
RUN apt-get update

RUN apt-get -y install open-iscsi

# Define default command
ENTRYPOINT ["/csi.server.opensds"]

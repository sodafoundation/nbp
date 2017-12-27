# Based on ubuntu
FROM ubuntu
LABEL maintainers="Edison Xiang <xiang.edison@gmail.com>"
LABEL description="OpenSDS CSI Plugin Client"

# Copy opensdsplugin client from build directory
COPY csi.client.opensds /csi.client.opensds

# Install iscsi
RUN apt-get update

RUN apt-get -y install open-iscsi

# Define default command
ENTRYPOINT ["/csi.client.opensds"]

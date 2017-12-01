FROM bitnami/minideb:latest
ENV PATH=$PATH:/usr/bin
ADD service-broker opt/services/

ENTRYPOINT ["/opt/services/service-broker", "-logtostderr"]

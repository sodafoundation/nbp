# Kubernetes external opensds provisioner

This is an example external provisioner for kubernetes meant for use with opensds.

**How to use**
After building. you can use binary to start opensds-provsioner service. For example:
./opensds-provisioner --master http://127.0.0.1:8080 --endpoint http://192.168.56.100:50040 --authstrategy noauth


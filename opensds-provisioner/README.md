# Kubernetes external opensds provisioner

This is an example external provisioner for kubernetes meant for use with opensds.

**First Steps**

Before building and packaging this, you need to include the shell script which flex will use for provisioning.  The shell script path must match what's in the provisioning container.

The current example is in flex/deploy/docker and is specified in examples/pod-provisioner.yaml here:
*- "-execCommand=/opt/storage/flex-provision.sh"*
If you copy in a new file or change the path, update the flag in the pod yaml.

**To Build**

```bash
make
```

**To Deploy**

You can use the example provisioner pod to deploy ```kubectl create -f examples/pod-provisioner.yaml```


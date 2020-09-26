# soda-csi-provisioner

`soda-csi-provisioner` is a side car which can be used by vendor csi-plugins to provision the storage using the soda-provisioner for hetrogeneous storages.

### Build
The docker images are already build and published [here](https://hub.docker.com/repository/docker/sodafoundation/soda-csi-provisioner) , still if you want to build the latest image then you can follow the below steps:
#### Step1: 
Download the code
```
git clone https://github.com/sodafoundation/nbp

cd nbp

git checkout csipnp_dev

mkdir -p $GOPATH:/src/github.com/kubernetes-csi

cd csi-plug-n-play/sidecars/soda-csi-provisioner/

cp -r external-provisioner/ $GOPATH/src/github.com/kubernetes-csi

```

#### Step2:
Build the and make the docker image
```

cd $GOPATH/src/github.com/kubernetes-csi/external-provisioner

go mod vendor
go mod download

make all

docker build -t sodafoundation/soda-csi-provisioner:v1.6.0 .


```

Now you can use this image along with vendor csi plugins to provision heterogeneous storage dynamically.



***Note***: Currently soda-csi-provisioner is built upon a code base of [csi-provisioner](https://github.com/kubernetes-csi/external-provisioner) v1.6.0 . Since the changes are relevant only for SODA so the code is maintained here, if there is a wide adoption of soda-csi-provisioner then a proposal can be made to upstream the changes. Till then all the credit and rights goes to the original authors of https://github.com/kubernetes-csi/external-provisioner.

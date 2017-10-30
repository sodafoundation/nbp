# OpenSDS Installation with Kubernetes Service Catalog
In this tutorial, we will show how OpenSDS provide Ceph storage as a service for Kubernetes users through Service Catalog. Now hope you enjoy the trip!

## Pre-configuration

### Check it out about your os (very important)
Please NOTICE that the installation tutorial is tested on Ubuntu17.04, and we SUGGEST you follow our styles and use Ubuntu16.04+.

### Download and install Golang
```
wget https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz
tar xvf go1.9.linux-amd64.tar.gz -C /usr/local/
mkdir -p $HOME/gopath/src
mkdir -p $HOME/gopath/bin
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/gopath/bin' >> /etc/profile
echo 'export GOPATH=$HOME/gopath' >> /etc/profile
source /etc/profile
go version (check if go has been installed)
```

### Install some package dependencies
```
apt-get install gcc make libc-dev docker.io
```
If the docker command doesn't work, try to restart it:
```
sudo service docker stop
sudo nohup docker daemon -H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock &
```

### Install containerized Ceph cluster (If your machine has already a ceph cluster installed, just skip this step)
Since this is a test, we choose to install a simple containerized ceph single-node cluster:
```
sudo docker run -d --net=host -v /etc/ceph:/etc/ceph -e MON_IP=your_host_ip -e CEPH_PUBLIC_NETWORK=your_host_ip/24 ceph/demo
```
NOTICE that ```your_host_ip``` means the real ip address of your machine.

After this container has been running, Add ```rbd_default_features = 1``` one line in ```/etc/ceph/ceph.conf``` file:
```
echo 'rbd_default_features = 1' >> /etc/ceph/ceph.conf
```

## Local Kubernetes Cluster Setup

### Download and unpackage k8s and etcd source code
```
wget https://github.com/kubernetes/kubernetes/archive/v1.6.0.tar.gz
wget https://github.com/coreos/etcd/releases/download/v3.2.0/etcd-v3.2.0-linux-amd64.tar.gz
tar -zxvf etcd-v3.2.0-linux-amd64.tar.gz
cp etcd-v3.2.0-linux-amd64/etcd /usr/local/bin
tar -zxvf v1.6.0.tar.gz
```

### Setup your cluster (suggest run as ```root```)
```
source /etc/profile
go get -u github.com/cloudflare/cfssl/cmd/...
KUBE_ENABLE_CLUSTER_DNS=true kubernetes-1.6.0/hack/local-up-cluster.sh -O
```
The setup process may run for a while, and after it done, some tips will occur to help you configure your cluster correctly.

### Check the status of cluster (open another terminal)
Follow the tips and configure your cluster, and then add another path ```your_path_to/kubernetes-1.6.0/hack/local-up-cluster.sh``` in the ```PATH``` variable configured in ```/etc/profile``` file.

After that, run ```kubectl.sh get po``` to check the status of kubernetes cluster:
```
source /etc/profile
kubectl.sh get po
```

## Service Catalog Setup

### Install helm (from scipt)
```
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get > get_helm.sh
chmod 700 get_helm.sh
./get_helm.sh
helm init
kubectl.sh get po -n kube-system (check if the till-deploy pod is running)
```

If your kubernetes cluster has RBAC enabled, you must ensure that the tiller pod has cluster-admin access. By default, helm init installs the tiller pod into kube-system namespace, with tiller configured to use the default service account.
```
kubectl create clusterrolebinding tiller-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```
cluster-admin access is required in order for helm to work correctly in clusters with RBAC enabled. If you used the --tiller-namespace or --service-account flags when running helm init, the --serviceaccount flag in the previous command needs to be adjusted to reference the appropriate namespace and ServiceAccount name.

### Set up kube-dns service (optional)
To avoid network barrieres, please make sure kube-dns service be set up:
```
mkdir -p gopath/src/github.com/leonwanghui
git clone https://github.com/leonwanghui/opensds-broker.git gopath/src/github.com/leonwanghui

kubectl.sh create -f gopath/src/github.com/leonwanghui/opensds-broker/examples/kube-dns.yaml
```

### Service Catalog download and install
```
mkdir -p gopath/src/github.com/kubernetes-incubator
wget https://github.com/kubernetes-incubator/service-catalog/archive/v0.1.0.tar.gz
tar xvf service-catalog-0.1.0.tar.gz -C gopath/src/github.com/kubernetes-incubator
helm install gopath/src/github.com/kubernetes-incubator/service-catalog-0.1.0/charts/catalog --name catalog --namespace catalog

kubectl.sh get po -n catalog (check if the catalog api-server and controller-manager pod are running)
```

## OpenSDS Service Broker Setup

### OpenSDS cluster install
1. Before the system starts, you should configure the pool and backend information. Here are some examples for testing:
- OpenSDS global configuration info (stored in ```opensds.conf```):
```conf
[osdslet]
api_endpoint = localhost:50040
graceful = True
log_file = /var/log/opensds/osdslet.log
socket_order = inc

[osdsdock]
api_endpoint = localhost:50050
log_file = /var/log/opensds/osdsdock.log

# Enabled backend types, such as sample, ceph, cinder, etc.
enabled_backends = ceph

# If backend needs config file, specify the path here.
ceph_config = /etc/opensds/driver/ceph.yaml

[sample]
name = sample
description = Sample backend for testing
driver_name = default

[ceph]
name = ceph
description = Ceph Test
driver_name = ceph

[database]
credential = opensds:password@127.0.0.1:3306/dbname
endpoint = localhost:2379,localhost:2380
driver = etcd
```
- Ceph configuration info (stored in ```ceph.yaml```):
```yaml
configFile: /etc/ceph/ceph.conf
pool:
  "rbd":
    diskType: SSD
    iops: 1000
    bandwidth: 1000
  "test":
    diskType: SAS
    iops: 800
    bandwidth: 800
```
Now put them in the right place:
```
mkdir -p /etc/opensds/driver && mkdir -p /var/log/opensds
mv opensds.conf /etc/opensds/
mv ceph.yaml /etc/opensds/driver/
```

2. Download and start opensds service daemon:
To ensure service broker connecting to OpenSDS api-service, you probably need to configure your service ip:
```
docker run -d --net=host -v /var/log/opensds:/var/log/opensds -v /etc/opensds:/etc/opensds -v /etc/ceph:/etc/ceph opensds/osdsdock:v1alpha
docker run -it --net=host -v /var/log/opensds:/var/log/opensds -v /etc/opensd:/etc/opensds opensds/osdslet:v1alpha

curl -X POST "http://your_host_ip:50040/v1alpha/profiles" -H "Content-Type: application/json" -d '{"name": "default", "description": "default policy", "extra": {"capacity": 5}}'
curl -X POST "http://your_host_ip:50040/v1alpha/profiles" -H "Content-Type: application/json" -d '{"name": "silver", "description": "silver policy", "extra": {"iops": 300, "bandwidth": 500, "diskType":"SAS", "capacity": 5}}'
```

### OpenSDS service broker install
Firstly, you need to modify the value of ```argEndpoint``` field in ```broker-deployment.yaml```, just change it to ```your_host_ip:50040```:
```
vim gopath/src/github.com/opensds/nbp/service-broker/charts/deployment.yaml (modify this file)
```
Then you can install service broker via helm:
```
cd gopath/src/github.com/nbp/service-broker
helm install charts/ --name service-broker --namespace service-broker

kubectl.sh get po -n service-broker (check if opensds broker pod is running)
```

### Configure Kubectl context
```
kubectl.sh config set-cluster service-catalog --server=https://127.0.0.1:30443 --insecure-skip-tls-verify=true
kubectl.sh config set-context service-catalog --cluster=service-catalog

kubectl.sh --context=service-catalog get clusterservicebrokers,serviceinstances,servicebindings
```

## Start to work

1. Create service broker

```
kubectl.sh --context=service-catalog create -f examples/service-broker.yaml
kubectl.sh --context=service-catalog get clusterservicebrokers,clusterserviceclasses,clusterserviceplans
```

2. Create service instance

```
kubectl.sh create ns opensds

kubectl.sh --context=service-catalog create -f examples/service-instance.yaml -n opensds
kubectl.sh --context=service-catalog get serviceinstances -n opensds
```

3. Create service instance binding

```
kubectl.sh --context=service-catalog create -f examples/service-binding.yaml -n opensds
kubectl.sh --context=service-catalog get servicebindings -n opensds

kubectl.sh get secrets -n opensds
kubectl.sh get secrets opensds-secret -o yaml -n opensds
```

4. Creat service wordpress for testing
From the secret ```opensds-instance-secret``` shown above, you can find a field called ```image``` in data structure, just decode it using ```encoding/base64``` package.

Then update it in ```Wordpress.yaml``` file:

```
vim examples/Wordpress.yaml (modify image field)
kubectl.sh create -f examples/Wordpress.yaml -n opensds
kubectl.sh get po -n opensds
kubectl.sh get service -n opensds
```

After all things done, you can visit your own blog by searching: ```http://service_cluster_ip:8084```!

## Clean it up

1. Delete service wordpress

```
kubectl.sh delete -f examples/Wordpress.yaml -n opensds
```

2. Delete service instance binding

```
kubectl.sh --context=service-catalog delete servicebindings service-binding -n opensds
```

3. Delete service instance

```
kubectl.sh --context=service-catalog delete serviceinstances service-instance -n opensds
```

4. Delete service broker

```
kubectl.sh --context=service-catalog delete clusterservicebrokers service-broker
```

5. Uninstall service broker pod

```
helm delete --purge service-broker
```

6. Uninstall service catalog pods

```
helm delete --purge catalog
```

## Ending

That's all the tutorial, thank you for watching it!

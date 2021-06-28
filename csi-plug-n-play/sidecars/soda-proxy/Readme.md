## soda-proxy

soda-proxy is a simple http server which helps to call the api's of SodaFoundation Components in a simpler way, it uses the authentication and directly helps to connect the osdsapi-server using the client provided by sodafoundation/api.

### Quick Start Guide

Simplest way to use soda-proxy is to get the latest image from [here](https://hub.docker.com/repository/docker/sodafoundation/soda-proxy)

After getting the image please download the deployment yaml for soda-porxy from [here](https://github.com/sodafoundation/nbp/blob/master/csi-plug-n-play/sidecars/soda-proxy/deploy/sodaProxy.yaml)
```go
wget https://github.com/sodafoundation/nbp/blob/master/csi-plug-n-play/sidecars/soda-proxy/deploy/sodaProxy.yaml

```
Please edit the Sodafoundation env variables in the above yaml, for reference you can visit this [link](https://docs.sodafoundation.io/soda-gettingstarted/installation-using-ansible/#how-to-test-soda-projects-cluster)

After editing you can follow the below guide to start the soda-proxy
```go
kubectl create -f sodaProxy.yaml
```

### Build the image through code
If you want to try our latest and greatest one then you can follow the below guide

```go
go get github.com/sodafoundation/nbp

cd $GOPATH/src/github.com/sodafoundation/nbp/csi-plug-n-play/sidecars/soda-proxy

go build -o soda-proxy cmd/proxy.go
```
 Once the code is build, then next step will be to build the docker images

```go

docker build -t sodafoundation/soda-proxy:dev .
```

Once the image is build then you can  follow the above Quick start guide to run the soda-proxy

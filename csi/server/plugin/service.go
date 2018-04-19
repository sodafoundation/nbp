package plugin

import (
	csi "github.com/container-storage-interface/spec/lib/go/csi/v0"
)

// Service Define CSI Interface
type Service interface {
	csi.IdentityServer
	csi.ControllerServer
	csi.NodeServer
}

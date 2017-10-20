package plugin

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

// Service Define CSI Interface
type Service interface {
	csi.IdentityServer
	csi.ControllerServer
	csi.NodeServer
}

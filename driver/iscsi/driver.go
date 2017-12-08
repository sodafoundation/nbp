package iscsi

import (
	"github.com/opensds/nbp/client/iscsi"
	"github.com/opensds/nbp/driver"
	"strconv"
)

var (
	ISCSI_DRIVER = "iscsi"
)

type Iscsi struct{}

func init() {
	driver.RegisterDriver(ISCSI_DRIVER, &Iscsi{})
}

func (isc *Iscsi) Attach(conn map[string]interface{}) (string, error) {
	iscsiCon := iscsi.ParseIscsiConnectInfo(conn)

	return iscsi.Connect(iscsiCon.TgtPortal, iscsiCon.TgtIQN, strconv.Itoa(iscsiCon.TgtLun))
}

func (isc *Iscsi) Detach(conn map[string]interface{}) error {
	iscsiCon := iscsi.ParseIscsiConnectInfo(conn)

	return iscsi.Disconnect(iscsiCon.TgtPortal, iscsiCon.TgtIQN)
}

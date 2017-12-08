package dirver

import (
	"fmt"
)

type VolumeDriver interface {
	Attach(map[string]interface{}) (string, error)
	Detach(map[string]interface{}) error
}

var drivers map[string]VolumeDriver

func RegisterDriver(driverType string, driver VolumeDriver) error {
	if _, exist := drivers[driverType]; exist {
		return fmt.Errorf("Driver %s already exist.", driverType)
	}

	drivers[driverType] = driver
	return nil
}

func UnregisterDriver(driverType string) {
	if _, exist := drivers[driverType]; !exist {
		return
	}

	delete(drivers, driverType)
	return
}

func NewVolumeDriver(driverType string) VolumeDriver {
	if volumeDriver, exist := drivers[driverType]; exist {
		return volumeDriver
	}

	return nil
}

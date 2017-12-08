package rbd

import (
	"fmt"
	"github.com/opensds/nbp/driver"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	rbdBusPath    = "/sys/bus/rbd"
	rbdDevicePath = path.Join(rbdBusPath, "devices")
	rbdDev        = "/dev/rbd"

	RBD_DRIVER = "rbd"
)

type RBD struct{}

var _ driver.VolumeDriver = &RBD{}

func init() {
	driver.RegisterDriver(RBD_DRIVER, &RBD{})
}

func (rbd *RBD) Attach(conn map[string]interface{}) (string, error) {
	if _, ok := conn["name"]; !ok {
		return "", os.ErrInvalid
	}

	name := conn["name"].(string)
	fields := strings.Split(name, "/")
	if len(fields) != 2 {
		return "", os.ErrInvalid
	}

	if _, ok := conn["hosts"].([]string); !ok {
		return "", os.ErrInvalid
	}
	hosts := conn["hosts"].([]string)

	if _, ok := conn["ports"].([]string); !ok {
		return "", os.ErrInvalid
	}
	ports := conn["ports"].([]string)

	poolName, imageName := fields[0], fields[1]
	device, err := mapDevice(poolName, imageName, hosts, ports)
	if err != nil {
		return "", err
	}

	return device, nil
}

func (rbd *RBD) Detach(conn map[string]interface{}) error {
	if _, ok := conn["name"]; !ok {
		return os.ErrInvalid
	}

	name := conn["name"].(string)
	fields := strings.Split(name, "/")
	if len(fields) != 2 {
		return os.ErrInvalid
	}

	poolName, imageName := fields[0], fields[1]
	device, err := findDevice(poolName, imageName, 1)
	if err != nil {
		return err
	}

	_, err = exec.Command("rbd", "unmap", device).CombinedOutput()
	return err
}

func mapDevice(poolName, imageName string, hosts, ports []string) (string, error) {
	devName, err := findDevice(poolName, imageName, 1)
	if err == nil {
		return devName, nil
	}

	// modprobe
	_, err = exec.Command("modprobe", "rbd").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("rbd: failed to load rbd kernel module:%v", err)
	}

	for i := 0; i < len(hosts); i++ {
		host := fmt.Sprintf("%s:%s", hosts[i], ports[0])
		_, err = exec.Command("rbd", "map", imageName, "--pool", poolName, "-m", host).CombinedOutput()
		if err == nil {
			break
		}
	}

	devName, err = findDevice(poolName, imageName, 10)
	if err != nil {
		return "", err
	}

	return devName, nil
}

func findDevice(poolName, imageName string, retries int) (string, error) {
	for i := 0; i < retries; i++ {
		if name, err := findDeviceTree(poolName, imageName); err == nil {
			if _, err := os.Stat(rbdDev + name); err != nil {
				return "", err
			}

			return rbdDev + name, nil
		}

		time.Sleep(time.Second)
	}

	return "", os.ErrNotExist
}

func findDeviceTree(poolName, imageName string) (string, error) {
	fi, err := ioutil.ReadDir(rbdDevicePath)
	if err != nil && err != os.ErrNotExist {
		return "", err
	} else if err == os.ErrNotExist {
		return "", fmt.Errorf("Could not locate devices directory")
	}

	for _, f := range fi {
		namePath := filepath.Join(rbdDevicePath, f.Name(), "name")
		content, err := ioutil.ReadFile(namePath)
		if err != nil {
			return "", err
		}

		if strings.TrimSpace(string(content)) == imageName {
			poolPath := filepath.Join(rbdDevicePath, f.Name(), "pool")
			content, err := ioutil.ReadFile(poolPath)
			if err != nil {
				return "", err
			}

			if strings.TrimSpace(string(content)) == poolName {
				return f.Name(), err
			}
		}
	}

	return "", os.ErrNotExist
}

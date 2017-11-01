package iscsi

import (
	"log"
	"os/exec"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
//      Refer some codes from: https://github.com/j-griffith/csi-cinder       //
////////////////////////////////////////////////////////////////////////////////

// GetInitiator returns all the ISCSI Initiator Name
func GetInitiator() ([]string, error) {
	res, err := exec.Command("sudo", "cat", "/etc/iscsi/initiatorname.iscsi").CombinedOutput()
	log.Printf("result from cat: %s", res)
	if err != nil {
		log.Fatalf("Error encountered gathering initiator names: %v", err)
		return nil, err
	}

	iqns := []string{}
	lines := strings.Split(string(res), "\n")
	for _, l := range lines {
		log.Printf("Inspect line: %s", l)
		if strings.Contains(l, "InitiatorName=") {
			iqns = append(iqns, strings.Split(l, "=")[1])
		}
	}

	log.Printf("Found the following iqns: %s", iqns)
	return iqns, nil
}

// Discovery ISCSI Target
func Discovery(portal string) error {
	log.Printf("Discovery portal: %s", portal)
	_, err := exec.Command("sudo", "iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", portal).CombinedOutput()
	if err != nil {
		log.Fatalf("Error encountered in sendtargets: %v", err)
		return err
	}
	return nil
}

// Login ISCSI Target
func Login(portal string, targetiqn string) error {
	log.Printf("Login portal: %s targetiqn: %s", portal, targetiqn)
	_, err := exec.Command("sudo", "iscsiadm", "-m", "node", "-p", portal, "-T", targetiqn, "--login").CombinedOutput()
	if err != nil {
		log.Fatalf("Received error on login attempt: %v", err)
		return err
	}
	return nil
}

// Logout ISCSI Target
func Logout(portal string, targetiqn string) error {
	log.Printf("Logout portal: %s targetiqn: %s", portal, targetiqn)
	_, err := exec.Command("sudo", "iscsiadm", "-m", "node", "-p", portal, "-T", targetiqn, "--logout").CombinedOutput()
	if err != nil {
		log.Fatalf("Received error on logout attempt: %v", err)
		return err
	}
	return nil
}

// GetFSType returns the File System Type of device
func GetFSType(device string) string {
	log.Printf("GetFSType: %s", device)
	fsType := ""
	res, err := exec.Command("blkid", device).CombinedOutput()
	if err != nil {
		log.Fatalf("failed to GetFSType: %v", err)
		return fsType
	}

	if strings.Contains(string(res), "TYPE=") {
		for _, v := range strings.Split(string(res), " ") {
			if strings.Contains(v, "TYPE=") {
				fsType = strings.Split(v, "=")[1]
				fsType = strings.Replace(fsType, "\"", "", -1)
			}
		}
	}
	return fsType
}

// Format device by File System Type
func Format(device string, fstype string) error {
	log.Printf("Format device: %s fstype: %s", device, fstype)

	// Get current File System Type
	curFSType := GetFSType(device)
	if curFSType == "" {
		// Default File Sysem Type is ext4
		if fstype == "" {
			fstype = "ext4"
		}
		_, err := exec.Command("mkfs.", fstype, "-F", device).CombinedOutput()
		if err != nil {
			log.Fatalf("failed to Format: %v", err)
			return err
		}
	} else {
		log.Printf("Device: %s has been formatted yet. fsType: %s", device, curFSType)
	}
	return nil
}

// Mount device into mount point
func Mount(device string, mountpoint string) error {
	log.Printf("Mount device: %s mountpoint: %s", device, mountpoint)

	_, err := exec.Command("mkdir", mountpoint).CombinedOutput()
	if err != nil {
		log.Fatalf("failed to mkdir: %v", err)
	}
	_, err = exec.Command("mount", device, mountpoint).CombinedOutput()
	if err != nil {
		log.Fatalf("failed to mount: %v", err)
		return err
	}
	return nil
}

// Umount from mountpoint
func Umount(mountpoint string) error {
	log.Printf("Umount mountpoint: %s", mountpoint)

	_, err := exec.Command("umount", mountpoint).CombinedOutput()
	if err != nil {
		log.Fatalf("failed to Umount: %v", err)
		return err
	}
	return nil
}

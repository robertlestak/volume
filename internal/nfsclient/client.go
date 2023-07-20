package nfsclient

import (
	"net"
	"os"
	"os/exec"
	"strings"

	"git.shdw.tech/shdw.tech/sproxy/pkg/sproxy"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Address   string
	ServerDir string
	MountPath string
	Auth      *sproxy.ClientAuth
}

func (n *Client) Unmount() error {
	l := log.WithFields(log.Fields{
		"fn":  "Unmount",
		"app": "webfs",
		"pkg": "nfsclient",
	})
	l.Infof("Unmounting %s", n.MountPath)
	// unmount the nfs
	unmountCmd := exec.Command("sudo", "umount", n.MountPath)
	unmountCmd.Stdout = os.Stdout
	unmountCmd.Stderr = os.Stderr
	if err := unmountCmd.Run(); err != nil {
		// fail to unmount, try to force unmount
		l.WithError(err).Debug("Failed to unmount, trying to force unmount")
		forceUnmountCmd := exec.Command("sudo", "umount", "-f", n.MountPath)
		forceUnmountCmd.Stdout = os.Stdout
		forceUnmountCmd.Stderr = os.Stderr
		if err := forceUnmountCmd.Run(); err != nil {
			l.WithError(err).Info("Failed to automatically unmount, please unmount manually")
		}
		return err
	}
	return nil
}

func (n *Client) Mount() error {
	l := log.WithFields(log.Fields{
		"fn":  "Mount",
		"app": "webfs",
		"pkg": "nfsclient",
	})
	l.Infof("Mounting %s:%s to %s", n.Address, n.ServerDir, n.MountPath)
	// check if dir exists
	if _, err := os.Stat(n.MountPath); os.IsNotExist(err) {
		l.Infof("Creating directory %s", n.MountPath)
		if err := os.MkdirAll(n.MountPath, 0755); err != nil {
			l.WithError(err).Error("Failed to create directory")
			return err
		}
	}
	// mount the nfs
	// split the address into host and port
	var host, port string
	host = n.Address
	var err error
	if strings.Contains(n.Address, ":") {
		host, port, err = net.SplitHostPort(n.Address)
		if err != nil {
			l.WithError(err).Error("Failed to split host and port")
			return err
		}
	}
	cmdStr := []string{"sudo", "mount", "-t", "nfs"}
	if port != "" {
		cmdStr = append(cmdStr, "-o", "port="+port)
		cmdStr = append(cmdStr, "-o", "mountport="+port)
	}
	cmdStr = append(cmdStr, host+":"+n.ServerDir, n.MountPath)
	l.Debugf("Running command: %s", strings.Join(cmdStr, " "))
	mountCmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	mountCmd.Stdout = os.Stdout
	mountCmd.Stderr = os.Stderr
	if err := mountCmd.Run(); err != nil {
		l.WithError(err).Error("Failed to mount")
		return err
	}
	return nil
}

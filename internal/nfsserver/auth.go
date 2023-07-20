package nfsserver

import (
	"context"
	"net"

	"github.com/go-git/go-billy/v5"
	"github.com/willscott/go-nfs"
)

// NewCIDRAuthHandler creates a handler for the provided filesystem
func NewCIDRAuthHandler(fs billy.Filesystem, allowCIDRs []string) nfs.Handler {
	return &CIDRAuthHandler{fs, allowCIDRs}
}

// CIDRAuthHandler returns a NFS backing that exposes a given file system in response to all mount requests.
type CIDRAuthHandler struct {
	fs         billy.Filesystem
	AllowCIDRs []string
}

// Mount backs Mount RPC Requests, allowing for access control policies.
func (h *CIDRAuthHandler) Mount(ctx context.Context, conn net.Conn, req nfs.MountRequest) (status nfs.MountStatus, hndl billy.Filesystem, auths []nfs.AuthFlavor) {
	if len(h.AllowCIDRs) > 0 {
		remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP
		allowed := false
		for _, cidr := range h.AllowCIDRs {
			_, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if ipnet.Contains(remoteIP) {
				allowed = true
				break
			}
		}
		if !allowed {
			status = nfs.MountStatusErrPerm
			return
		}
	}
	status = nfs.MountStatusOk
	hndl = h.fs
	auths = []nfs.AuthFlavor{nfs.AuthFlavorDES}
	return
}

// Change provides an interface for updating file attributes.
func (h *CIDRAuthHandler) Change(fs billy.Filesystem) billy.Change {
	if c, ok := h.fs.(billy.Change); ok {
		return c
	}
	return nil
}

// FSStat provides information about a filesystem.
func (h *CIDRAuthHandler) FSStat(ctx context.Context, f billy.Filesystem, s *nfs.FSStat) error {
	return nil
}

// ToHandle handled by CachingHandler
func (h *CIDRAuthHandler) ToHandle(f billy.Filesystem, s []string) []byte {
	return []byte{}
}

// FromHandle handled by CachingHandler
func (h *CIDRAuthHandler) FromHandle([]byte) (billy.Filesystem, []string, error) {
	return nil, []string{}, nil
}

// HandleLImit handled by cachingHandler
func (h *CIDRAuthHandler) HandleLimit() int {
	return -1
}

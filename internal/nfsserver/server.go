package nfsserver

import (
	"encoding/json"
	"net"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	log "github.com/sirupsen/logrus"
	nfs "github.com/willscott/go-nfs"
	nfshelper "github.com/willscott/go-nfs/helpers"
	"gopkg.in/yaml.v3"
)

type Auth struct {
	Enabled    bool     `json:"enabled" yaml:"enabled"`
	AllowCIDRs []string `json:"allow_cidrs" yaml:"allow_cidrs"`
}

type Server struct {
	Dir     string `json:"dir" yaml:"dir"`
	Address string `json:"address" yaml:"address"`
	Auth    Auth   `json:"auth" yaml:"auth"`
}

func ParseConfig(f string) (*Server, error) {
	l := log.WithFields(log.Fields{
		"fn":  "ParseConfig",
		"app": "webfs",
		"pkg": "server",
	})
	l.Debugf("Parsing config file: %s", f)
	var s Server
	bd, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	bd = []byte(os.ExpandEnv(string(bd)))
	if err := yaml.Unmarshal(bd, &s); err != nil {
		// try json
		if err := json.Unmarshal(bd, &s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *Server) Start() error {
	l := log.WithFields(log.Fields{
		"fn":  "Start",
		"app": "webfs",
		"pkg": "nfsserver",
		"dir": s.Dir,
	})
	l.Infof("starting NFS server on %s", s.Address)
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		l.Errorf("Failed to listen: %v", err)
		return err
	}
	bfs := osfs.New(s.Dir)
	bfsPlusChange := NewChangeOSFS(bfs)

	handler := NewCIDRAuthHandler(bfsPlusChange, s.Auth.AllowCIDRs)
	cacheHelper := nfshelper.NewCachingHandler(handler, 1024)
	if err := nfs.Serve(listener, cacheHelper); err != nil {
		l.Errorf("Failed to serve: %v", err)
		return err
	}
	return nil
}

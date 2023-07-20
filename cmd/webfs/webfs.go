package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"time"

	"git.shdw.tech/shdw.tech/sproxy/pkg/sproxy"
	"git.shdw.tech/shdw.tech/webfs/internal/nfsclient"
	"git.shdw.tech/shdw.tech/webfs/internal/nfsserver"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func replaceHomeDir(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return strings.Replace(path, "~", homeDir, 1)
}

func cmdClient() {
	l := log.WithFields(log.Fields{
		"fn":  "cmdClient",
		"app": "webfs",
	})
	l.Debug("Starting client")
	// usage:
	// webfs mount [options] <address> <mount-dir>
	var address, mountDir *string
	clientFlags := flag.NewFlagSet("client", flag.ExitOnError)
	authToken := clientFlags.String("token", "", "Auth token")
	authTokenCmd := clientFlags.String("token-cmd", "", "Auth token command")
	tcpAddress := clientFlags.String("tcp-address", "127.0.0.1:6049", "Local TCP listener address")
	clientFlags.Parse(os.Args[2:])
	args := clientFlags.Args()
	if len(args) > 0 {
		address = &args[0]
	}
	if len(args) > 1 {
		mountDir = &args[1]
	}
	if *address == "" {
		l.Fatal("Missing server address")
	}
	c := &sproxy.Client{
		Listen:  *tcpAddress,
		Connect: *address,
	}
	if *authToken != "" || *authTokenCmd != "" {
		c.Auth = &sproxy.ClientAuth{
			Token:    *authToken,
			TokenCmd: *authTokenCmd,
		}
		l.Debug("Using auth token")
	}
	go func() {
		if err := c.Run(); err != nil {
			l.WithError(err).Fatal("Error while running client")
		}
	}()
	// give some time for the client to start
	time.Sleep(1 * time.Second)
	hr := replaceHomeDir(*mountDir)
	mountDir = &hr
	if *mountDir == "" {
		l.Fatal("Missing mount directory")
	}
	l.Infof("Mounting NFS server on %s", *mountDir)
	nc := &nfsclient.Client{
		Address:   *tcpAddress,
		ServerDir: "/",
		MountPath: *mountDir,
	}
	if err := nc.Mount(); err != nil {
		l.WithError(err).Fatal("Error while mounting NFS server")
	}
	l.Infof("NFS server mounted on %s", *mountDir)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for sig := range ch {
			// sig is a ^C, handle it
			l.Debugf("Received %s", sig)
			l.Infof("Unmounting NFS server on %s", *mountDir)
			if err := nc.Unmount(); err != nil {
				l.WithError(err).Debug("Error while unmounting NFS server")
			}
			os.Exit(0)
		}
	}()
	select {}
}

func cmdServer() {
	l := log.WithFields(log.Fields{
		"fn":  "cmdServer",
		"app": "webfs",
	})
	l.Debug("Starting server")
	// usage:
	// webfs serve [options] <dir>
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	address := serverFlags.String("address", "0.0.0.0:8049", "Address of the server")
	metricsAddress := serverFlags.String("metrics-address", "0.0.0.0:9090", "Address of the metrics server")
	tcpAddress := serverFlags.String("tcp-address", "127.0.0.1:7049", "Address of the server")
	localNFSEnabled := serverFlags.Bool("local-nfs", true, "Enable local NFS server")
	localNFSHost := serverFlags.String("nfs-address", "127.0.0.1:7049", "Address of the local NFS server")
	serverFlags.Parse(os.Args[2:])
	var localNFSDir *string
	args := serverFlags.Args()
	if len(args) > 0 {
		localNFSDir = &args[0]
	}
	if *localNFSEnabled {
		if *localNFSDir == "" {
			l.Fatal("Missing local NFS directory")
		}
		hr := replaceHomeDir(*localNFSDir)
		localNFSDir = &hr
		l.Debug("Starting local NFS server")
		ns := &nfsserver.Server{
			Address: *localNFSHost,
			Dir:     *localNFSDir,
		}
		go func() {
			if err := ns.Start(); err != nil {
				l.Errorf("Failed to start local NFS server: %v", err)
			}
		}()
	}
	s := &sproxy.Server{
		MetricsAddress: *metricsAddress,
		Listen:         *address,
		Connect:        *tcpAddress,
	}
	if err := s.Run(); err != nil {
		l.Errorf("Failed to start sproxy server: %v", err)
	}
}

func usage() {
	log.Printf("Usage: %s <command> [options]", os.Args[0])
	log.Printf("Commands:")
	log.Printf("  mount [options] <address> <mount-dir>")
	log.Printf("  serve [options] <dir>")
}

func main() {
	l := log.WithFields(log.Fields{
		"fn":  "main",
		"app": "webfs",
	})
	l.Debug("Starting webfs server")
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		usage()
		os.Exit(0)
	}
	switch os.Args[1] {
	case "mount":
		cmdClient()
	case "serve":
		cmdServer()
	default:
		l.Fatalf("Unknown command: %s", os.Args[1])
	}
}

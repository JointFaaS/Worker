package controller

import (
	"context"
	"net"
	dc "github.com/docker/docker/client"
)

type task struct {
	funcName string
	args string
	res chan string
	ctx context.Context
}

type funcState string
const (
	running funcState = "running"
	creating funcState = "creating"
	cold funcState = "cold"
)

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	dockerClient *dc.Client

	unixListener *net.UnixListener

	tasks chan *task

	funcStateMap map[string]funcState

	containerMap map[string][]containerMeta
}

// Config is used to initialize controller client
// It supports adjusting the resource limits 
type Config struct {

}

// NewClient initializes a new API client
func NewClient(config *Config) (*Client, error){
	unixAddr, err := net.ResolveUnixAddr("unix", "/var/run/worker.sock")
	if err != nil {
		return nil, err
	}
	unixListener, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		return nil, err
	}
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return nil, err
	}
	c := &Client{
		dockerClient: dockerClient,
		tasks: make(chan * task),
		unixListener: unixListener,
		funcStateMap: make(map[string]funcState),
		containerMap: make(map[string][]containerMeta),
	}
	go c.work()
	return c, nil
}
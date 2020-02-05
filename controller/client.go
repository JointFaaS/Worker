package controller

import (
	"context"
	"runtime"
	"net"
	dc "github.com/docker/docker/client"
)
const (
	socketPath string = "/var/run/worker.sock"
)
type task struct {
	funcName string
	args string
	res chan []byte
	id uint64
	ctx context.Context
}

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	dockerClient *dc.Client

	unixListener *net.UnixListener

	tasks chan *task

	containerRegistration chan *registerBody

	createContainerResponse chan *containerMeta

	funcStateMap map[string]funcState

	containerMap map[string][]containerMeta

	subTasks map[string]chan *task

	ctx context.Context

	cancel context.CancelFunc

	config *Config
}

// Config is used to initialize controller client
// It supports adjusting the resource limits 
type Config struct {
	SocketPath string
}

// NewClient initializes a new API client
func NewClient(config *Config) (*Client, error){
	unixAddr, err := net.ResolveUnixAddr("unix", config.SocketPath)
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
	ctx, cancel := context.WithCancel(context.TODO())
	c := &Client{
		dockerClient: dockerClient,
		tasks: make(chan * task, 256),
		containerRegistration: make(chan *registerBody),
		createContainerResponse: make(chan *containerMeta),
		unixListener: unixListener,
		funcStateMap: make(map[string]funcState),
		containerMap: make(map[string][]containerMeta),
		subTasks: make(map[string]chan *task),
		ctx: ctx,
		cancel: cancel,
		config: config,
	}
	runtime.SetFinalizer(c, clientFinalizer)
	go c.workForExternalRequest(ctx)
	go c.workForContainerRegistration()
	return c, nil
}

func clientFinalizer(c *Client) {
	c.cancel()
	c.dockerClient.Close()
	c.unixListener.Close()
}
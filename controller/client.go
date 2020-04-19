package controller

import (
	"context"
	"net"
	"sync"

	dc "github.com/docker/docker/client"
)

const (
	socketPath string = "/var/run/worker.sock"
)

// Response is used in async ret
type Response struct {
	Err  error
	Body *[]byte
}
type task struct {
	funcName string
	args     []byte
	res      chan *Response
	ctx      context.Context
}

type initTask struct {
	funcName string
	image    string
	codeURI  string
	res      chan *Response
	ctx      context.Context
}

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	dockerClient *dc.Client

	unixListener *net.UnixListener

	tasks chan *task

	initTasks chan *initTask

	containerRegistration chan *containerMeta

	funcStateMap map[string]funcState

	funcResourceMap map[string]*funcResource

	funcContainerMap map[string][]*containerMeta

	containerIDMap *sync.Map

	subTasks map[string]chan *task

	ctx context.Context

	cancel context.CancelFunc

	config *Config

	wg *sync.WaitGroup
}

// Config is used to initialize controller client
// It supports adjusting the resource limits
type Config struct {
	SocketPath string

	ContainerEnvVariables []string
}

// NewClient initializes a new API client
func NewClient(config *Config) (*Client, error) {
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
		dockerClient:          dockerClient,
		tasks:                 make(chan *task, 256),
		initTasks:             make(chan *initTask, 8),
		containerRegistration: make(chan *containerMeta),
		unixListener:          unixListener,
		funcStateMap:          make(map[string]funcState),
		funcResourceMap:       make(map[string]*funcResource),
		funcContainerMap:      make(map[string][]*containerMeta),
		containerIDMap: 		new(sync.Map),
		subTasks:              make(map[string]chan *task),
		ctx:                   ctx,
		cancel:                cancel,
		config:                config,
		wg:                    new(sync.WaitGroup),
	}
	go c.workForExternalRequest()
	go c.workForContainerRegistration()
	return c, nil
}

// Close stops the client and wait for all the resource release
func (c *Client) Close() {
	c.cancel()
	c.dockerClient.Close()
	c.unixListener.Close()
	c.wg.Wait()
}

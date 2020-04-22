package controller

import (
	"context"
	"sync"

	dc "github.com/docker/docker/client"
	wc "github.com/JointFaaS/Worker/container"
	wpb "github.com/JointFaaS/Worker/pb/worker"
)

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	dockerClient *dc.Client

	funcResourceMap map[string]*funcResource

	funcContainerMap map[string][]*wc.ContainerMeta

	ctx context.Context

	cancel context.CancelFunc

	config *Config

	wg *sync.WaitGroup
}

// Config is used to initialize controller client
// It supports adjusting the resource limits
type Config struct {
	ContainerEnvVariables []string
}

// NewClient initializes a new API client
func NewClient(config *Config) (*Client, error) {
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.TODO())

	c := &Client{
		dockerClient:          dockerClient,
		funcResourceMap:       make(map[string]*funcResource),
		funcContainerMap:      make(map[string][]*wc.ContainerMeta),
		ctx:                   ctx,
		cancel:                cancel,
		config:                config,
		wg:                    new(sync.WaitGroup),
	}

	return c, nil
}

// Close stops the client and wait for all the resource release
func (c *Client) Close() {
	c.cancel()
	c.dockerClient.Close()
	c.wg.Wait()
}

func (c *Client) Invoke(context.Context, *wpb.InvokeRequest) (*wpb.InvokeReply, error) {
	return nil, nil
}

func (c *Client) Register(context.Context, *wpb.RegisterRequest) (*wpb.RegisterReply, error) {
	return nil, nil
}

func (c *Client) InitFunction(context.Context, *wpb.InitFunctionRequest) (*wpb.InitFunctionReply, error) {
	return nil, nil
}

func (c *Client) Metrics(context.Context, *wpb.MetricsRequest) (*wpb.MetricsReply, error) {
	return nil, nil
}
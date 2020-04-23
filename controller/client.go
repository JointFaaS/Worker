package controller

import (
	"context"
	"sync"

	wc "github.com/JointFaaS/Worker/container"
	wpb "github.com/JointFaaS/Worker/pb/worker"
	dc "github.com/docker/docker/client"
)

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	config Config

	dockerClient *dc.Client

	// funcName to CodeURI and Image
	funcResourceMap map[string]*FuncResource

	resourceRWMu sync.RWMutex

	// funcName to Container
	funcContainerMap map[string][]*wc.Meta

	// the key is memory size of the container
	idleContainerMap map[int64][]*wc.Meta

	containerMu sync.Mutex

	ctx context.Context

	cancel context.CancelFunc

	wg *sync.WaitGroup
}

// Config is used to initialize controller client
// It supports adjusting the resource limits
type Config struct {
	ContainerEnvVariables []string
}

// NewClient initializes a new API client
func NewClient(config Config) (*Client, error) {
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.TODO())

	c := &Client{
		containerMu:      sync.Mutex{},
		resourceRWMu:     sync.RWMutex{},
		dockerClient:     dockerClient,
		funcResourceMap:  make(map[string]*FuncResource),
		funcContainerMap: make(map[string][]*wc.Meta),
		idleContainerMap: make(map[int64][]*wc.Meta),
		ctx:              ctx,
		cancel:           cancel,
		config:           config,
		wg:               new(sync.WaitGroup),
	}

	return c, nil
}

// Close stops the client and wait for all the resource release
func (c *Client) Close() {
	c.cancel()
	c.dockerClient.Close()
	c.wg.Wait()
}

// Invoke exec a function with name and payload
func (c *Client) Invoke(ctx context.Context, req *wpb.InvokeRequest) (*wpb.InvokeResponse, error) {
	c.containerMu.Lock()
	containers, isPresent := c.funcContainerMap[req.GetName()]
	if isPresent == false {
		containers = make([]*wc.Meta, 3)
		c.funcContainerMap[req.GetName()] = containers
	}
	c.containerMu.Unlock()
	for _, container := range containers {
		// if the connection is broken, someone will reset the container as nil
		// so here we just skip that container.
		if container != nil {
			output, err := container.InvokeFunc(ctx, req.GetName(), req.GetPayload())
			if err == nil {
				return &wpb.InvokeResponse{Code: wpb.InvokeResponse_OK, Output: output}, nil
			}
			switch err.(type) {
			case *wc.ExceedConcurrencyLimit:
				continue
			default:
				return &wpb.InvokeResponse{Code: wpb.InvokeResponse_RUNTIME_ERROR, Output: []byte(err.Error())}, err
			}
		}
		// no idle container
	}
	err := c.addSpeifiedFuncContainer(req.GetName(), len(containers)+1)
	if err != nil {
		return &wpb.InvokeResponse{Code: wpb.InvokeResponse_NO_SUCH_FUNCTION, Output: nil}, nil
	}
	return &wpb.InvokeResponse{Code: wpb.InvokeResponse_RETRY, Output: nil}, nil
}

// Register is for a function exec env register itself into Worker env-list
func (c *Client) Register(ctx context.Context, req *wpb.RegisterRequest) (res *wpb.RegisterResponse, err error) {
	res = &wpb.RegisterResponse{
		Code: wpb.RegisterResponse_OK,
		Msg:  "",
	}
	var container *wc.Meta
	container, err = wc.NewMeta(req.GetAddr(), req.GetFuncName(), req.GetRuntime())
	if err != nil {
		res.Code = wpb.RegisterResponse_ERROR
		res.Msg = err.Error()
		return
	}
	memorySize := req.GetMemory()
	memorySize = memorySize - memorySize%128
	if memorySize < 128 || memorySize > 4096 {
		res.Code = wpb.RegisterResponse_ERROR
		res.Msg = "Invalid Memory, it should be in [128, 4096]"
		return
	}
	c.containerMu.Lock()
	containers, isPresent := c.idleContainerMap[memorySize]
	if isPresent == false {
		containers = make([]*wc.Meta, 3)
	}
	c.idleContainerMap[memorySize] = append(containers, container)
	c.containerMu.Unlock()
	return
}

// InitFunction the Manager pass function info to Worker
func (c *Client) InitFunction(ctx context.Context, req *wpb.InitFunctionRequest) (*wpb.InitFunctionResponse, error) {
	resource := &FuncResource{
		FuncName:   req.GetFuncName(),
		Image:      req.GetImage(),
		CodeURL:    req.GetCodeURI(),
		Runtime:	req.GetRuntime(),
		Timeout:    req.GetTimeout(),
		MemorySize: req.GetMemorySize(),
	}

	c.resourceRWMu.Lock()
	c.funcResourceMap[resource.FuncName] = resource
	c.resourceRWMu.Unlock()

	return &wpb.InitFunctionResponse{
		Code: wpb.InitFunctionResponse_OK,
		Msg:  "",
	}, nil
}

// Metrics is used to retrive worker's overhead and resource utilization
func (c *Client) Metrics(ctx context.Context, req *wpb.MetricsRequest) (*wpb.MetricsResponse, error) {
	return nil, nil
}

// Reset moves the func container to idle-list
func (c *Client) Reset(ctx context.Context, req *wpb.ResetRequest) (*wpb.ResetResponse, error) {
	return nil, nil
}

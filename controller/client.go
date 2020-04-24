package controller

import (
	"container/list"
	"context"
	"sync"
	"time"

	wc "github.com/JointFaaS/Worker/container"
	wpb "github.com/JointFaaS/Worker/pb/worker"
	dc "github.com/docker/docker/client"
)

// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	config Config

	dockerClient *dc.Client

	// funcName to CreatingContainerNum
	creatingContainerNumMap map[string]int64
	creatingContainerMu sync.Mutex

	// funcName to CodeURI and Image
	funcResourceMap map[string]*FuncResource

	resourceMu sync.RWMutex

	// funcName to []Container
	funcContainerMap map[string]*list.List

	// the key is memorySize of the container
	idleContainerMap map[int64]*list.List

	containerMu sync.RWMutex

	idleContainerMu sync.Mutex

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
		containerMu:      sync.RWMutex{},
		resourceMu:       sync.RWMutex{},
		creatingContainerMu: sync.Mutex{},
		idleContainerMu: sync.Mutex{},
		dockerClient:     dockerClient,
		creatingContainerNumMap: make(map[string]int64),
		funcResourceMap:  make(map[string]*FuncResource),
		funcContainerMap: make(map[string]*list.List),
		idleContainerMap: make(map[int64]*list.List),
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
	for i := 0; i < 3; i++ {
		c.containerMu.RLock()
		containers, isPresent := c.funcContainerMap[req.GetName()]
		if isPresent {
			for e := containers.Front(); e != nil; e = e.Next() {
				// if the connection is broken, someone will reset the container
				output, err := e.Value.(*wc.Meta).InvokeFunc(ctx, req.GetName(), req.GetPayload())
				if err == nil {
					c.containerMu.RUnlock()
					return &wpb.InvokeResponse{Code: wpb.InvokeResponse_OK, Output: output}, nil
				}
				switch err.(type) {
				case *wc.ExceedConcurrencyLimit:
					continue
				default:
					c.containerMu.RUnlock()
					return &wpb.InvokeResponse{Code: wpb.InvokeResponse_RUNTIME_ERROR, Output: []byte(err.Error())}, err
				}
			}
		}
		c.containerMu.RUnlock()
		// no idle container
		err := c.addSpecifiedContainer(req.GetName())
		if err != nil {
			return &wpb.InvokeResponse{Code: wpb.InvokeResponse_NO_SUCH_FUNCTION, Output: nil}, nil
		}
		// TODO
		// I never find the best way to handle the async container creating
		// sleep is a simple solution, just retry and ensure there is no deadlock
		time.Sleep(time.Millisecond * 500)
	}
	return &wpb.InvokeResponse{Code: wpb.InvokeResponse_RETRY, Output: nil}, nil
}

// Register is for a function exec env register itself into Worker env-list
func (c *Client) Register(ctx context.Context, req *wpb.RegisterRequest) (res *wpb.RegisterResponse, err error) {
	res = &wpb.RegisterResponse{
		Code: wpb.RegisterResponse_OK,
		Msg:  "",
	}
	memorySize := req.GetMemory()
	memorySize = memorySize - memorySize%128
	if memorySize < 128 || memorySize > 4096 {
		res.Code = wpb.RegisterResponse_ERROR
		res.Msg = "Invalid Memory, it should be in [128, 4096]"
		return
	}
	var container *wc.Meta
	container, err = wc.NewMeta(req.GetAddr(), req.GetFuncName(), req.GetRuntime())
	if err != nil {
		res.Code = wpb.RegisterResponse_ERROR
		res.Msg = err.Error()
		return
	}

	if req.GetFuncName() == "" {
		c.idleContainerMu.Lock()
		containers, isPresent := c.idleContainerMap[memorySize]
		if isPresent == false {
			containers = list.New()
			c.idleContainerMap[memorySize] = containers
		}
		containers.PushBack(container)
		c.idleContainerMu.Unlock()
	} else {
		c.containerMu.Lock()
		containers, isPresent := c.funcContainerMap[req.GetFuncName()]
		if isPresent == false {
			containers = list.New()
			c.funcContainerMap[req.GetFuncName()] = containers
		}
		containers.PushBack(container)
		c.containerMu.Unlock()

		c.creatingContainerMu.Lock()
		c.creatingContainerNumMap[req.GetFuncName()] = 0
		c.creatingContainerMu.Unlock()
	}
	return
}

// InitFunction the Manager pass function info to Worker
func (c *Client) InitFunction(ctx context.Context, req *wpb.InitFunctionRequest) (*wpb.InitFunctionResponse, error) {
	resource := &FuncResource{
		FuncName:   req.GetFuncName(),
		Image:      req.GetImage(),
		CodeURL:    req.GetCodeURI(),
		Runtime:    req.GetRuntime(),
		Timeout:    req.GetTimeout(),
		MemorySize: req.GetMemorySize(),
	}

	c.resourceMu.Lock()
	c.funcResourceMap[resource.FuncName] = resource
	c.resourceMu.Unlock()

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

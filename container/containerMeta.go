package container

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	cpb "github.com/JointFaaS/Worker/pb/container"
	"google.golang.org/grpc"
)

// ExceedConcurrencyLimit means the number of waitting task beyonds the concurrencyLimit
type ExceedConcurrencyLimit struct {
}

func (e *ExceedConcurrencyLimit) Error() string {
	return "Exceed the concurrency limit"
}

var (
	ecl ExceedConcurrencyLimit = ExceedConcurrencyLimit{}
)

// Meta is an abstract handler of a specified env container
type Meta struct {
	mu sync.Mutex
	funcName string
	runtime string
	containerClient cpb.ContainerClient
	concurrencyLimit int64
	concurrencyCounter int64
	timeout int64
}

// GetFuncName
func (m *Meta) GetFuncName() string {
	return m.funcName
}

func (m *Meta) GetRuntime() string {
	return m.runtime
}

// NewMeta returns a container handler which maintains a rpc connection with the container
func NewMeta(containerHost string, funcName string, runtime string) (*Meta, error) {
	conn, err := grpc.Dial(containerHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
		return nil, err
	}

	containerClient := cpb.NewContainerClient(conn)
	return &Meta{
		mu: sync.Mutex{},
		funcName: funcName,
		runtime: runtime,
		containerClient: containerClient,
		concurrencyLimit: 5,
		concurrencyCounter: 0,
		timeout: 3,
	}, nil
}

// SetConcurrencyLimit this limit influence the invoke, pls refer to invoke
func (m *Meta) SetConcurrencyLimit(limit int64) {
	m.concurrencyLimit = limit
}

// InvokeFunc exec the func in container
func (m *Meta) InvokeFunc(ctx context.Context, funcName string, payload []byte) ([]byte, error) {
	defer atomic.AddInt64(&m.concurrencyCounter, -1)
	if atomic.AddInt64(&m.concurrencyCounter, 1) <= m.concurrencyLimit {
		if m.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(m.timeout) * time.Second)
			defer cancel()
		}
		res, err := m.containerClient.Invoke(ctx, &cpb.InvokeRequest{FuncName: funcName, Payload: payload})
		if err != nil {
			return nil, err
		}
		return res.GetOutput(), nil
	}
	return nil, &ecl
}

// SetEnvVariable overwrite some envs in container
func (m *Meta) SetEnvVariable(ctx context.Context, envs []string) error {
	res, err := m.containerClient.SetEnvs(ctx, &cpb.SetEnvsRequest{Env: envs})
	if err != nil {
		return err
	} else if res.GetCode() != cpb.SetEnvsResponse_OK {
		return errors.New(res.GetCode().String())
	} else {
		return nil
	}
}

// LoadFunc reset the function settings in container
func (m *Meta) LoadFunc(ctx context.Context, funcName string, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	res, err := m.containerClient.LoadCode(ctx, &cpb.LoadCodeRequest{FuncName: funcName, Url: url})
	if err != nil {
		return err
	} else if res.GetCode() != cpb.LoadCodeResponse_OK {
		return errors.New(res.GetCode().String())
	} else {
		return nil
	}
}

// SetTimeout timeout is used to limit the invokecation exec time
func (m *Meta) SetTimeout(timeout int64) {
	m.timeout = timeout
}
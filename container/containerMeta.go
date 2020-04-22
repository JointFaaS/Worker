package container

import (
	"log"
	"sync"
	"sync/atomic"

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
	containerClient cpb.ContainerClient
	concurrencyLimit int64
	concurrencyCounter int64
}

// NewMeta returns a container handler which maintains a rpc connection with the container
func NewMeta(containerHost string) (*Meta, error) {
	conn, err := grpc.Dial(containerHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
		return nil, err
	}

	containerClient := cpb.NewContainerClient(conn)
	return &Meta{
		mu: sync.Mutex{},
		funcName: "",
		containerClient: containerClient,
		concurrencyLimit: 1,
		concurrencyCounter: 0,
	}, nil
}

// SetConcurrencyLimit this limit influence the invoke, pls refer to invoke
func (c *Meta) SetConcurrencyLimit(limit int64) {
	c.concurrencyLimit = limit
}

// InvokeFunc exec the func in container
func (c *Meta) InvokeFunc(payload []byte) ([]byte, error) {
	if atomic.AddInt64(&c.concurrencyCounter, 1) <= c.concurrencyLimit {

	} else {
		return nil, &ecl
	}
	return nil, nil
}
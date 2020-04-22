package container

import (
	"log"
	"sync"

	cpb "github.com/JointFaaS/Worker/pb/container"
	"google.golang.org/grpc"
)

type ContainerMeta struct {
	mu sync.Mutex
	funcName string
	containerClient cpb.ContainerClient
	concurrencyLimit int
}

// SetConcurrencyLimit this limit influence the invoke, pls refer to invoke
func (c *ContainerMeta) SetConcurrencyLimit(limit int) {
	c.concurrencyLimit = limit
} 

// NewContainerMeta returns a container handler which maintains a rpc connection with the container
func NewContainerMeta(containerHost string) (*ContainerMeta, error) {
	conn, err := grpc.Dial(containerHost, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
		return nil, err
	}

	containerClient := cpb.NewContainerClient(conn)
	return &ContainerMeta{
		mu: sync.Mutex{},
		funcName: "",
		containerClient: containerClient,
		concurrencyLimit: 1,
	}, nil
}
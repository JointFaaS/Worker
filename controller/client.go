package controller

import (
	"context"
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

	tasks chan *task

	funcStateMap map[string]funcState
}

// Config is used to initialize controller client
// It supports adjusting the resource limits 
type Config struct {

}

// NewClient initializes a new API client
func NewClient(config *Config) (*Client, error){
	dockerClient, err := dc.NewEnvClient()
	if err != nil {
		return nil, err
	}
	c := &Client{
		dockerClient:    dockerClient,
		tasks: make(chan * task),
	}
	return c, nil
}
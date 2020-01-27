package controller

import (

	dc "github.com/docker/docker/client"
)
// Client is the API client that performs all operations
// against a Worker.
type Client struct {
	dockerClient *dc.Client
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
	}
	return c, nil
}
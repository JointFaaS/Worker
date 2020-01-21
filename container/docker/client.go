package docker

import (
	"context"
	"sync"
	"syscall"
	"os"
	"path"
	
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

var cli *client.Client
var once sync.Once

// GetDockerClient returns a singleton obj of docker.Client
func GetDockerClient() *client.Client{
    once.Do(func(){
		var err error
		cli, err = client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}
    })
    return cli
}

// Alloc starts a specific runtime container
func Alloc(ctx context.Context, name string, image string, sequence string) (container.ContainerCreateCreatedBody, error) {
	cc := GetDockerClient()

	os.Mkdir(path.Join("/tmp", name), 0777)
	os.Mkdir(path.Join("/tmp", name, sequence), 0777)
	syscall.Mkfifo(path.Join("/tmp", name, sequence, "down") , 0666)
	syscall.Mkfifo(path.Join("/tmp", name, sequence, "up") , 0666)

	body, err := cc.ContainerCreate(ctx, 
		&container.Config{
			Image: image,
		},  
		&container.HostConfig{
			Mounts: []mount.Mount{
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", name, "up"),
					Target: "/up",
				},
				mount.Mount{
					Type: mount.TypeNamedPipe,
					Source: path.Join("/tmp", name, "down"),
					Target: "/down",
				},
			},
			NetworkMode: "none",
		},
		nil, name)
	return body, err
}

// GetNamedPipeOfEnv returns the up and down pipes for a running container
func GetNamedPipeOfEnv(name string, sequence string) (*os.File, *os.File, error){
	up, _ := os.OpenFile(path.Join("/tmp", name, sequence, "up"), os.O_RDWR, os.ModeNamedPipe)
	down, _ := os.OpenFile(path.Join("/tmp", name, sequence, "down"), os.O_RDWR, os.ModeNamedPipe)
	return up, down, nil
}
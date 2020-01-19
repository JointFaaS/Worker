package docker

import (
	"sync"
	docker "github.com/fsouza/go-dockerclient"
)

var dockerClient *docker.Client
var once sync.Once

// GetDockerClient returns a singleton obj of docker.Client
func GetDockerClient() *docker.Client {
    once.Do(func(){
		var err error
		dockerClient, err = docker.NewClientFromEnv()
		if err != nil {
			panic(err)
		}
    })
    return dockerClient
}

// Alloc starts a specific runtime container
func Alloc(name string) (*docker.Container, error) {
	dc := GetDockerClient()
	return dc.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		HostConfig: &docker.HostConfig{
			VolumesFrom: []string{""},
			Memory: 0,
		},
	})
}

func Info() {

}
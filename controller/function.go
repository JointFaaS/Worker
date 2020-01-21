package controller

import (
	"context"
	"github.com/JointFaas/Worker/container/docker"
	"sync/atomic"
)

// Function defines the prop of a function
type Function struct {
	name string
	image string
}

type invocation struct {
	name string
	image string
	args *map[string]interface{}
	res chan string
}

type envCreateReady struct {
	envID string
	job *invocation
	winner bool
}

var functions map[string]*Function
var funcEnvs map[string][]string
var funcCallMetrics map[string]*int32
var jobs chan *invocation
// GetFunc converts a id to function struct
func GetFunc(funcName string) *Function {
	return functions[funcName]
}

// Invoke pass a function request to backend
func Invoke(ctx context.Context, name string, args *map[string]interface{}, res chan string)  {
	jobs <- &invocation{name: name, args: args, res: res}
}

func dispatch(envName string, job *invocation) {

}

func alloc(job *invocation, res chan *envCreateReady)  {
	newValue := atomic.AddInt32(funcCallMetrics[job.name], 1)
	if newValue % 100 == 1 {
		c, err := docker.Alloc(context.Background(), job.name, job.image, "0")
		if err != nil {
	
		}
		ecr := &envCreateReady{
			envID: c.ID,
			job: job,
			winner: true,
		}
		res <- ecr
		atomic.AddInt32(funcCallMetrics[job.name], -1)
	} else {
		// TODO
	}

}

func work() {
	createEnvCh := make(chan *envCreateReady)
	for {
		select {
		case <- jobs:
			job := <- jobs
			availableEnvs := funcEnvs[job.name]
			if availableEnvs == nil {
				funcEnvs[job.name] = make([]string, 0)
				availableEnvs = funcEnvs[job.name]
			}
	
			if len(availableEnvs) != 0 {
				go dispatch(availableEnvs[0], job)
			} else {
				if funcCallMetrics[job.name] == nil {
					funcCallMetrics[job.name] = new(int32)
				}
				go alloc(job, createEnvCh)
			}
		case <- createEnvCh:
			ecr := <- createEnvCh
			go dispatch(ecr.envID, ecr.job)
			if ecr.winner {
				_ = append(funcEnvs[ecr.job.name], ecr.envID)
			}
		}
	}
}
// Init should be called before any other func in this file
func Init() {
	functions = make(map[string]*Function)
	funcEnvs = make(map[string][]string)
	funcCallMetrics = make(map[string]*int32)
	go work()
}
package controller

import (
	"context"
)

// Function defines the prop of a function
type Function struct {
	name string
	image string
}

type invocation struct {
	name *string
	args *map[string]interface{}
	res *chan *string
}

var functions map[string]*Function
var jobs chan *invocation
// GetFunc converts a id to function struct
func GetFunc(funcName *string) *Function {
	return functions[*funcName]
}

// Invoke pass a function request to backend
func Invoke(ctx context.Context, name *string, args *map[string]interface{}, res *chan *string)  {
	jobs <- &invocation{name: name, args: args, res: res}
}

func work() {
	for {
		job := <- jobs
		_ = job
		// TODO
	}
}
// Init should be called before any other func in this file
func Init() {
	functions = make(map[string]*Function)
	go work()
}
package controller

import (
	"github.com/JointFaas/Worker/container/docker"
)
// Invoke pass a function request to backend
func Invoke()  {
	docker.Alloc("")
}

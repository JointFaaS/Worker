package main

import (
    "fmt"
	"log"
	"net/http"
	"github.com/JointFaas/Worker/controller"
)

// CallHandler is the essential interface which
// responses manager's call requests.
func CallHandler(w http.ResponseWriter, r *http.Request) {
	controller.Invoke()
    fmt.Fprintln(w, "hello world")
}

func main() {
    http.HandleFunc("/call", CallHandler)
    log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
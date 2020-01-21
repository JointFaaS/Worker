package main

import (
    "fmt"
	"context"
	"log"
	"net/http"
	"github.com/JointFaas/Worker/controller"
)

// CallHandler is the essential interface which
// responses manager's call requests.
func CallHandler(w http.ResponseWriter, r *http.Request) {
	resCh := make(chan *string)
	name := r.URL.String()
	go controller.Invoke(context.Background(), &name, nil, &resCh)
	res := <- resCh
    fmt.Fprintln(w, res)
}

func main() {
    http.HandleFunc("/call", CallHandler)
    log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
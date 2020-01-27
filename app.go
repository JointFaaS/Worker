package main

import (
    "fmt"
	"context"
	"log"
	"net/http"
	"encoding/json"
	"github.com/JointFaas/Worker/controller"
)

type callRequestBody struct {
	funcName string
	args string
}
// CallHandler is the essential interface which
// responses manager's call requests.
func CallHandler(w http.ResponseWriter, r *http.Request) {
	var req callRequestBody
	json.NewDecoder(r.Body).Decode(&req)
	resCh := make(chan string)
	controller.Invoke(context.Background(), req.funcName, req.args, resCh)
	res := <- resCh
    fmt.Fprintln(w, res)
}

func main() {
    http.HandleFunc("/call", CallHandler)
    log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
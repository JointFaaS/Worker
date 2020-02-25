package main

import (
    "fmt"
	"os"
	"time"
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

func logInit() {
	log.SetPrefix("TRACE: ")
    log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

func main() {
	logInit()
	client, err := controller.NewClient(&controller.Config{
		SocketPath: os.Getenv("WORKER_SOCKET_PATH"),
	})
	if err != nil {
		panic(err)
	}
	callHandler := func (w http.ResponseWriter, r *http.Request) {
		var req callRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return
		}
		resCh := make(chan []byte)
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 300)
		
		client.Invoke(ctx, req.funcName, req.args, resCh)
		select {
		case res := <- resCh:
			fmt.Fprintln(w, res)
		case msg := <- ctx.Done():
			fmt.Fprintln(w, msg)
		}
	}
	http.HandleFunc("/call", callHandler)
	log.Print("start listening")
    log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
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

type CallRequestBody struct {
	FuncName string
	Args string
}

type InitRequestBody struct {
	FuncName string
	Image string
	CodeURI string
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
		var req CallRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resCh := make(chan *controller.Response)
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 300)
		
		client.Invoke(ctx, req.FuncName, req.Args, resCh)
		select {
		case res := <- resCh:
			if res.Err != nil {
				http.Error(w, res.Err.Error(), http.StatusBadRequest)
			}
			fmt.Fprintln(w, res)
		case msg := <- ctx.Done():
			fmt.Fprintln(w, msg)
		}
	}
	initHandler := func(w http.ResponseWriter, r *http.Request) {
		var req InitRequestBody
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resCh := make(chan *controller.Response)
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 300)
		
		client.Init(ctx, req.FuncName, req.Image, req.CodeURI, resCh)
		select {
		case res := <- resCh:
			if res.Err != nil {
				http.Error(w, res.Err.Error(), http.StatusBadRequest)
			}
			fmt.Fprintln(w, res)
		case msg := <- ctx.Done():
			fmt.Fprintln(w, msg)
		}
	}

	http.HandleFunc("/call", callHandler)
	http.HandleFunc("/init", initHandler)
	log.Print("start listening")
    log.Fatal(http.ListenAndServe("0.0.0.0:8000", nil))
}
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"path"
	"time"

	"gopkg.in/yaml.v2"
	"github.com/JointFaas/Worker/controller"
)

type initRequestBody struct {
	FuncName string `json:"funcName"`
	Image string	`json:"image"`
	CodeURI string	`json:"codeURI"`
}

func logInit() {
	log.SetPrefix("TRACE: ")
    log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

type config struct {
	WorkerID string `yaml:"workerID"`
	WorkerSocketPath string `yaml:"workerSocketPath"`
	ListenPort string `yaml:"listenPort"`
	ManagerAddress string `yaml:"managerAddress"`
	ContainerEnvVariables []string `yaml:"containerEnvVariables"`
}

type registrationBody struct {
	WorkerPort string `json:"workerPort"`
	WorkerID string `json:"workerID"`
}

func registerMeToManager(managerAddr string, body registrationBody) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	go func() {
		for	{
			resp, err := http.Post("http://" + managerAddr + "/register", "application/json;charset=UTF-8", bytes.NewReader(jsonBody))
			if err != nil {
				log.Print("register fail: ", err.Error())
			} else if resp.StatusCode != http.StatusOK {
				log.Print("register fail:", resp.Body)
			} else {
				log.Print("register successful:", resp.Body)
			}
			time.Sleep(time.Second * 10)
		}
	}()
}


func setHandler(client *controller.Client) {
	callHandler := func (w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Not support method", http.StatusBadRequest)
			return
		}
		r.ParseForm()
		funcName := r.FormValue("funcName")
		args, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Fail to read Payload", http.StatusBadRequest)
			return
		}
		resCh := make(chan *controller.Response)
		ctx, _ := context.WithTimeout(context.Background(), time.Second * 300)
		
		client.Invoke(ctx, funcName, args, resCh)
		select {
		case res := <- resCh:
			if res.Err != nil {
				http.Error(w, res.Err.Error(), http.StatusBadRequest)
			}
			w.Write(*res.Body)
		case msg := <- ctx.Done():
			fmt.Fprintln(w, msg)
		}
	}

	initHandler := func(w http.ResponseWriter, r *http.Request) {
		var req initRequestBody
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
}


func main() {
	logInit()
	var cfg config
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	cfgFile, err := ioutil.ReadFile(path.Join(home, "/.jfWorker/config.yml"))
	if err != nil {
		panic(err)
	}

	err = yaml.UnmarshalStrict(cfgFile, &cfg)
	if err != nil {
		panic(err)
	}
	client, err := controller.NewClient(&controller.Config{
		SocketPath: cfg.WorkerSocketPath,
		ContainerEnvVariables: cfg.ContainerEnvVariables,
	})
	if err != nil {
		panic(err)
	}
	setHandler(client)
	go log.Fatal(http.ListenAndServe("0.0.0.0:" + cfg.ListenPort, nil))
	registerMeToManager(cfg.ManagerAddress, registrationBody{WorkerID: cfg.WorkerID, WorkerPort: cfg.ListenPort})

	log.Print("start listening")
}
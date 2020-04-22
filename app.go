package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/JointFaas/Worker/controller"
	"gopkg.in/yaml.v2"
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

type workerRegistrationBody struct {
	WorkerPort string `json:"workerPort"`
	WorkerID string `json:"workerID"`
}

type workerRegistrationResponseBody struct {
	Region          string `json:"region"`
	JointfaasEnv    string `json:"jointfaasEnv"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	CenterStorage   string `json:"centerStorage"`
}

func registerMeToManager(managerAddr string, body workerRegistrationBody) (*workerRegistrationResponseBody, error) {
	time.Sleep(time.Second * 5) // wait for http server initializing
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("http://" + managerAddr + "/register", "application/json;charset=UTF-8", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Unavailable Manager")
	} 
	var res workerRegistrationResponseBody
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
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
	go registerMeToManager(cfg.ManagerAddress, workerRegistrationBody{WorkerID: cfg.WorkerID, WorkerPort: cfg.ListenPort})
	log.Fatal(http.ListenAndServe("0.0.0.0:" + cfg.ListenPort, nil))
	log.Print("start listening")
}
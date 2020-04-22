package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/JointFaaS/Worker/controller"
	wpb "github.com/JointFaaS/Worker/pb/worker"
	"google.golang.org/grpc"
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
	client, err := controller.NewClient(controller.Config{
		ContainerEnvVariables: cfg.ContainerEnvVariables,
	})
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", cfg.ListenPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	wpb.RegisterWorkerServer(s, client)
	log.Println("rpc server start")
	s.Serve(lis)

	go registerMeToManager(cfg.ManagerAddress, workerRegistrationBody{WorkerID: cfg.WorkerID, WorkerPort: cfg.ListenPort})
	log.Fatal(http.ListenAndServe("0.0.0.0:" + cfg.ListenPort, nil))
}
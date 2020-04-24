package main

import (
	"context"
	"fmt"
	"log"
	"os"

	wpb "github.com/JointFaaS/Worker/pb/worker"
	"google.golang.org/grpc"
	"github.com/spf13/cobra"
)

var (
	addr string
	funcName string
	image string
	codeURI string
	runtime string
	payload string
	memorySize int64
)

var rootCmd = &cobra.Command{
	Use:   "worker-tester",
	Short: "simple tester",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("can not connect with server %v", err)
			return
		}
		rpcClient := wpb.NewWorkerClient(conn)
		initRes, err := rpcClient.InitFunction(context.TODO(), &wpb.InitFunctionRequest{
			FuncName: funcName,
			Image: image,
			Runtime: runtime,
			CodeURI: codeURI,
			Timeout: 3,
			MemorySize: memorySize,
		})
		if err != nil {
			panic(err)
		}
		fmt.Print(initRes.GetMsg())
		if initRes.GetCode() != wpb.InitFunctionResponse_OK {
			return
		}

		invokeRes, err := rpcClient.Invoke(context.TODO(), &wpb.InvokeRequest{
			Name: funcName,
			Payload: []byte(payload),
		})
		fmt.Print(string(invokeRes.GetOutput()))
	},
}

func rootInit() {
	rootCmd.Flags().StringVarP(&addr, "addr", "a", "localhost:8001", "Tested Worker Addr")
	rootCmd.Flags().StringVarP(&funcName, "funcName", "f", "test", "Tested Func")
	rootCmd.Flags().StringVarP(&image, "image", "i", "jointfaas-java8", "Tested Image")
	rootCmd.Flags().StringVarP(&codeURI, "codeURI", "u", "uri", "Source Code URI")
	rootCmd.Flags().StringVarP(&runtime, "runtime", "r", "java8", "Tested Runtime")
	rootCmd.Flags().StringVarP(&payload, "payload", "p", "{}", "Tested Payload")
	rootCmd.Flags().Int64VarP(&memorySize, "memorySize", "m", 128, "The limitation of function memory(MB)")
}

func main() {
	rootInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

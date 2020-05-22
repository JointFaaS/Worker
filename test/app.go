package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	wpb "github.com/JointFaaS/Worker/pb/worker"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	addr       string
	funcName   string
	image      string
	codeURI    string
	runtime    string
	payload    string
	memorySize int64
)

var simpleCmd = &cobra.Command{
	Use:   "simple tester",
	Short: "simple",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("can not connect with server %v", err)
			return
		}
		rpcClient := wpb.NewWorkerClient(conn)
		initRes, err := rpcClient.InitFunction(context.TODO(), &wpb.InitFunctionRequest{
			FuncName:   funcName,
			Image:      image,
			Runtime:    runtime,
			CodeURI:    codeURI,
			Timeout:    3,
			MemorySize: memorySize,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(initRes.GetMsg())
		if initRes.GetCode() != wpb.InitFunctionResponse_OK {
			return
		}

		invokeRes, err := rpcClient.Invoke(context.TODO(), &wpb.InvokeRequest{
			Name:    funcName,
			Payload: []byte(payload),
		})
		fmt.Println(invokeRes.GetCode().String())
		fmt.Println(string(invokeRes.GetOutput()))
	},
}

var perfCmd = &cobra.Command{
	Use:   "perf tester",
	Short: "perf",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("can not connect with server %v", err)
			return
		}
		rpcClient := wpb.NewWorkerClient(conn)
		initRes, err := rpcClient.InitFunction(context.TODO(), &wpb.InitFunctionRequest{
			FuncName:   funcName,
			Image:      image,
			Runtime:    runtime,
			CodeURI:    codeURI,
			Timeout:    3,
			MemorySize: memorySize,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(initRes.GetMsg())
		if initRes.GetCode() != wpb.InitFunctionResponse_OK {
			return
		}
		timeSlice := make([]time.Duration, 20, 20)
		for i := 0; i < 20; i++ {
			start := time.Now()
			_, err = rpcClient.Invoke(context.TODO(), &wpb.InvokeRequest{
				Name:    funcName,
				Payload: []byte(payload),
			})
			cost := time.Since(start)
			timeSlice[i] = cost
			if err != nil {
				panic(err)
			}
		}

		for i := 0; i < 20; i++ {
			fmt.Println(int64(timeSlice[i]))
		}
	},
}

var rootCmd = &cobra.Command{
	Use:   "worker-tester",
	Short: "tester",
}

func rootInit() {
	rootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", "localhost:8001", "Tested Worker Addr")
	rootCmd.PersistentFlags().StringVarP(&funcName, "funcName", "f", "test", "Tested Func")
	rootCmd.PersistentFlags().StringVarP(&image, "image", "i", "jointfaas-java8", "Tested Image")
	rootCmd.PersistentFlags().StringVarP(&codeURI, "codeURI", "u", "uri", "Source Code URI")
	rootCmd.PersistentFlags().StringVarP(&runtime, "runtime", "r", "java8", "Tested Runtime")
	rootCmd.PersistentFlags().StringVarP(&payload, "payload", "p", "{}", "Tested Payload")
	rootCmd.PersistentFlags().Int64VarP(&memorySize, "memorySize", "m", 128, "The limitation of function memory(MB)")

	rootCmd.AddCommand(simpleCmd, perfCmd)
}

func main() {
	rootInit()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

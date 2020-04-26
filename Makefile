.PHONY: proto worker tester clean

worker:
	go build -o build/worker
	
proto:
	protoc -I proto proto/container/container.proto --go_out=plugins=grpc:pb/container
	protoc -I proto proto/worker/worker.proto --go_out=plugins=grpc:pb/worker

tester:
	go build -o build/tester test/app.go

clean:
	rm -rf build/*
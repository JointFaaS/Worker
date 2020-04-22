.PHONY: proto worker clean

worker:
	go build -o build/worker
	
proto:
	protoc -I proto proto/container.proto --go_out=plugins=grpc:pb/container
	protoc -I proto proto/worker.proto --go_out=plugins=grpc:pb/worker

clean:
	rm -rf build/*
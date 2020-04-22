.PHONY: proto worker clean

worker:
	go build -o build/worker
	
proto:
	protoc -I proto proto/container.proto --go_out=plugins=grpc:grpc
	protoc -I proto proto/worker.proto --go_out=plugins=grpc:grpc

clean:
	rm -rf build/*
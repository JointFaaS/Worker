.PHONY: proto worker clean

worker:
	go build -o build/worker
	
proto:
	protoc -I proto proto/container/container.proto --go_out=plugins=grpc:pb/
	protoc -I proto proto/worker/worker.proto --go_out=plugins=grpc:pb/

clean:
	rm -rf build/*

all:

gen:
	protoc -I example/proto example/proto/*.proto --go_out=plugins=grpc:example/rpc

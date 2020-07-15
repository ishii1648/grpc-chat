SERVER_BINARY=chat-server
CLIENT_BINARY=chat-client

go-proto:
	-rm -rf proto/chat.pb.go
	protoc -I proto/ proto/chat.proto --go_out=plugins=grpc:proto

build:
	go build -o $(SERVER_BINARY) ./cmd/server
	go build -o $(CLIENT_BINARY) ./cmd/client

run-server: build
	./$(SERVER_BINARY)

run-client: build
	./$(CLIENT_BINARY) -n ishii
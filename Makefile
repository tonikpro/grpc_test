APP=grpc_test
GO=/usr/local/go/bin/go 

all: clean test build

clean:
	rm -rf $(APP)

build:
	$(GO) build -o $(APP) main.go

migration:
	migrate create -ext sql -dir repository/testing/migrations -seq $(name)

test:
	$(GO) test ./...

proto:
	protoc -I transport/grpc/pb/ --go_out=plugins=grpc:transport/grpc/pb transport/grpc/pb/a2billing.proto
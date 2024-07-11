# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOPKG=$(shell go list ./...)
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

# Binary names for our commands
LIBP2P_NODE_BINARY=libp2p-node

all: test build
build: libp2p-node
test: 
		$(GOTEST) -v $(GOPKG)

clean: 
		$(GOCLEAN)
		rm -f $(GOBIN)/*

# Compile the monitor blocks use case command
libp2p-node:
		$(GOBUILD) -o $(GOBIN)/$(LIBP2P_NODE_BINARY) cmd/oracle/main.go

# Generate mocks
generate-mocks:
		mockgen -source=internal/app/domain/repository.go -destination=internal/app/domain/mocks/repository_mock.go -package=mocks

run_libp2p_node:
		$(GOBIN)/libp2p-node

tidy:
		go mod tidy

download:
		go mod download

verify:
		go mod verify

lint:
		golangci-lint run

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=main

all: verify build vet staticcheck test run_staticcheck

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

deps:
	$(GOGET) -v ./...

verify:
	go mod verify

vet:
	go vet ./...

run_staticcheck:
	staticcheck ./...

.PHONY: build test verify vet staticcheck run_staticcheck all

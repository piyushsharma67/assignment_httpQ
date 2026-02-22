APP_NAME=httpq
GO=go
PORT=24744

.PHONY: build run test race clean cert

## Generate self signed TLS certificate
cert: 
	openssl req -x509 -newkey rsa:4096 \
	-keyout server.key \
	-out server.crt \
	-days 365 \
	-nodes \
	-subj "/CN=localhost" 

## Build binary
build: 
	$(GO) build -o $(APP_NAME)

## Run the server(expected cert to exists)
run:build
	./$(APP_NAME)

## Run test
test:
	$(GO) test -v ./...

## Run test with race condition
race:
	$(GO) test -race -v ./...



## Remove binary and cert
clean:
	rm -f $(APP_NAME) server.key server.crt

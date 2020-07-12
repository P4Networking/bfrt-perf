.PHONY: build

all: tofino

tofino:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bfrt_test_tofino ./bin/main.go




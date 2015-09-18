.PHONY: lint build local

all: build

lint:
	gofmt -w $$(pwd)

build: lint
	GOARCH="amd64" GOOS="linux" go build

local:
	go build

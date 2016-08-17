IP=172.16.181.16
AMD64=GOOS=linux GOARCH=amd64
OUT=bin/watch_changes
GOBUILD=go build -o $(OUT) watch_changes.go

all: build-amd64

deps:
	go get -d .

build: deps
	$(GOBUILD)

build-amd64: deps
	env $(AMD64) $(GOBUILD)

install: build-amd64
	scp $(OUT) vagrant@$(IP):

.PHONY: all deps install


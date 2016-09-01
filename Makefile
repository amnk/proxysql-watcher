IP=172.16.181.16
AMD64=GOOS=linux GOARCH=amd64
OUT=bin/watch_changes
GOBUILD=go build -o $(OUT) watch_changes.go
DOCKER_NAME=amnk/proxysql-watcher
DOCKER_TAG=latest

all: push-docker

deps:
	go get -d .

build: deps
	$(GOBUILD)

build-amd64: deps
	env $(AMD64) $(GOBUILD)

install: build-amd64
	scp $(OUT) vagrant@$(IP):

build-docker: build-amd64
	docker build -t $(DOCKER_NAME):$(DOCKER_TAG) .

push-docker: build-docker
	docker push $(DOCKER_NAME):$(DOCKER_TAG)

.PHONY: all deps install build-docker push-docker


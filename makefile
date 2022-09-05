BINDIR=bin
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
GOBUILD=CGO_ENABLED=0 ${GO_BUILD_ENV} go build

EVENTOPS_BUILD_PATH ?= ${PROJ_PATH}/cmd
EOCTL_BUILD_PATH ?= ${PROJ_PATH}/tools/eoctl
CLIENT_BUILD_PATH ?= ${PROJ_PATH}/tools/dialerclient

GOPROXY ?= https://goproxy.cn/
GOPRIVATE ?= ""
GO_BUILD_ENV=GOPROXY=${GOPROXY} GOPRIVATE=${GOPRIVATE}

all: eventops-linux-amd64 eventops-darwin-amd64 eventops-windows-amd64 eocli-darwin-amd64 eocli-linux-amd64 eocli-windows-amd64 client-linux-amd64

eventops-darwin-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$@ $(EVENTOPS_BUILD_PATH)

eventops-linux-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$@ $(EVENTOPS_BUILD_PATH)

eventops-windows-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$@.exe $(EVENTOPS_BUILD_PATH)



eocli-darwin-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$@ $(EOCTL_BUILD_PATH)

eocli-linux-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$@ $(EOCTL_BUILD_PATH)

eocli-windows-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$@.exe $(EOCTL_BUILD_PATH)


client-linux-amd64:
	@cd ${PROJ_PATH} && GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$@ $(CLIENT_BUILD_PATH)

clean:
	@rm $(BINDIR)/*

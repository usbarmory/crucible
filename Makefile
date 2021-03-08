SHELL := /bin/bash
GO ?= go
BUILD_USER := $(shell whoami)
BUILD_HOST := $(shell hostname)
BUILD_DATE := $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD := ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV := $(shell git rev-parse --short HEAD 2> /dev/null)
PKG = "github.com/f-secure-foundry/crucible"

.PHONY: test crucible habtool

all: test crucible habtool

# requires the following dependencies in $GOPATH
#   honnef.co/go/tools/cmd/staticcheck
#   github.com/kisielk/errcheck
check:
	@${GO} vet ./...
	@${GOPATH}/bin/staticcheck ./...
	@${GOPATH}/bin/errcheck ./...

test:
	@cd fusemap && ${GO} test -cover
	@cd hab && ${GO} test -cover
	@cd otp && ${GO} test -cover *linux*.go

crucible:
	${GO} build -v \
	  -trimpath -ldflags "-s -w -X 'main.Revision=${REV}' -X 'main.Build=${BUILD}'" \
	  cmd/crucible/crucible.go
	@echo -e "compiled crucible ${REV} (${BUILD})"

habtool:
	${GO} build -v \
	  -trimpath -ldflags "-s -w -X 'main.Revision=${REV}' -X 'main.Build=${BUILD}'" \
	  cmd/habtool/habtool.go
	@echo -e "compiled habtool ${REV} (${BUILD})"

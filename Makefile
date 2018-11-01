SHELL := /bin/bash
GO ?= go
BUILD_GOPATH = ${GOPATH}:$(CURDIR)
BUILD_USER := $(shell whoami)
BUILD_HOST := $(shell hostname)
BUILD_DATE := $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD := ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV := $(shell git rev-parse --short HEAD 2> /dev/null)

.PHONY: crucible test

all: test crucible

# requires the following dependencies in $GOPATH
#   https://github.com/dominikh/go-tools/tree/master/cmd/megacheck
#   https://github.com/kisielk/errcheck
check: all
	@GOPATH="${BUILD_GOPATH}" ${GO} vet crucible
	@GOPATH="${BUILD_GOPATH}" ${GOPATH}/bin/megacheck crucible
	@GOPATH="${BUILD_GOPATH}" ${GOPATH}/bin/megacheck crucible.go
	@cd src/crucible && ${GOPATH}/bin/errcheck

test:
	@cd src/crucible && ${GO} test -cover

crucible:
	@GOPATH="${BUILD_GOPATH}" ${GO} build -v \
		-gcflags=-trimpath=${CURDIR} -asmflags=-trimpath=${CURDIR} \
		-ldflags "-s -w -X 'main.Revision=${REV}' -X 'main.Build=${BUILD}'" \
		-o crucible crucible.go
	@echo -e "compiled crucible ${REV} (${BUILD})"

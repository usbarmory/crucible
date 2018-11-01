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
#   honnef.co/go/tools/cmd/megacheck
#   github.com/kisielk/errcheck
check:
	@${GO} vet lib/*.go
	@${GOPATH}/bin/megacheck lib/*.go
	@${GOPATH}/bin/errcheck lib/*.go

test:
	@cd lib && ${GO} test -cover

# To allow a `go get` friendly repository that also supports a local make
# command we need to pass through a temporary file to change the remote import
# path to a local one for the `make` build.
crucible:
	$(eval TMP := $(shell mktemp ${CURDIR}/crucible-XXXXXX.go))
	@cp crucible.go ${TMP}
	@sed -i -e 's/github.com\/inversepath\/crucible\/lib/lib/' ${TMP}
	@GOPATH="${BUILD_GOPATH}" ${GO} build -v \
		-gcflags=-trimpath=${CURDIR} -asmflags=-trimpath=${CURDIR} \
		-ldflags "-s -w -X 'main.Revision=${REV}' -X 'main.Build=${BUILD}'" \
		-o crucible ${TMP}
	@rm ${TMP}
	@echo -e "compiled crucible ${REV} (${BUILD})"

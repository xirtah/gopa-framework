SHELL=/bin/bash

# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif

#Add the GOPATH bin directory to the path file
export PATH := $(PATH):$(GOPATH)/bin

GO        := GO15VENDOREXPERIMENT="1" go
GOTEST   := GOPATH=$(GOPATH) CGO_ENABLED=1  $(GO) test -ldflags -s

default: test

test:
	$(GOTEST) -timeout 60s ./...
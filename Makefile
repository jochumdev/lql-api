# This how we want to name the binary output
BINARY := lql-api
SHELL := /bin/bash

TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifeq ($(VERSION),)
	VERSION := $(COMMIT)-$(DATE)
else
ifneq ($(COMMIT), $(TAG_COMMIT))
	VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
endif
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags '-w -s -X github.com/jochumdev/lql-api/version.Version=${VERSION}'
LDFLAGS_STATIC=-ldflags '-extldflags "-static" -w -s -X github.com/jochumdev/lql-api/version.Version=${VERSION}'

build:
	go build -o ${BINARY} -a ${LDFLAGS}


# Builds the project
build_static:
	CGO_ENABLED=0 GOOS=linux go build -o ${BINARY} -a ${LDFLAGS_STATIC}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

.PHONY: debian
debian:
	GOPROXY= dpkg-buildpackage -b -rfakeroot -us -uc

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install
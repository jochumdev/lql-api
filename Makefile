# This how we want to name the binary output
BINARY=lql_api


TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
	VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
ifeq ($(VERSION), "")
	VERSION := $(COMMIT)-$(DATA)
endif
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X github.com/webmeisterei/lql_api/version.Version=${VERSION}"

# Builds the project
build:
	go build ${LDFLAGS} -o ${BINARY}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS_f1}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install
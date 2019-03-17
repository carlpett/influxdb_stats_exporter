CROSSBUILD_OS = linux windows darwin
CROSSBUILD_ARCH = 386 amd64

VERSION  = $(shell git describe --always --tags --dirty=-dirty)
REVISION = $(shell git rev-parse --short=8 HEAD)
BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)

BUILDUSER ?= $(USER)
BUILDHOST ?= $(HOSTNAME)
LDFLAGS    = -X github.com/prometheus/common/version.Version=${VERSION} \
             -X github.com/prometheus/common/version.Revision=${REVISION} \
             -X github.com/prometheus/common/version.Branch=${BRANCH} \
             -X github.com/prometheus/common/version.BuildUser=$(BUILDUSER)@$(BUILDHOST) \
             -X github.com/prometheus/common/version.BuildDate=$(shell date +%Y-%m-%dT%T%z)

all: build test

build:
	@echo ">> building"
	@go build -ldflags "$(LDFLAGS)"

crossbuild: $(GOPATH)/bin/gox
	@echo ">> cross-building"
	@gox -arch="$(CROSSBUILD_ARCH)" -os="$(CROSSBUILD_OS)" -ldflags="$(LDFLAGS)" -output="binaries/influx_stats_exporter_{{.OS}}_{{.Arch}}"

test:
	@echo ">> testing"
	@go test -v -cover

release: bin/github-release
	@echo ">> uploading release ${VERSION}"
	@for bin in binaries/*; do \
		./bin/github-release upload -t ${VERSION} -n $$(basename $${bin}) -f $${bin}; \
	done

docker:
	@echo ">> building docker image"
	@docker build -t carlpett/influxdb_stats_exporter .

$(GOPATH)/bin/gox:
	# Need to disable modules for this to not pollute go.mod
	@GO111MODULE=off go get -u github.com/mitchellh/gox

bin/github-release:
	@mkdir -p bin
	@curl -sL 'https://github.com/aktau/github-release/releases/download/v0.6.2/linux-amd64-github-release.tar.bz2' | tar xjf - --strip-components 3 -C bin

.PHONY: all build crossbuild test release

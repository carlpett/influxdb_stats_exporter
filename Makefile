PROMU   := $(GOPATH)/bin/promu
pkgs     = $(shell go list ./...)

all: format build

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format:
	@echo ">> formatting code"
	@go fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@go vet $(pkgs)

test:
	@echo ">> testing code"
	@go test $(pkgs)

build: $(PROMU)
	@echo ">> building binaries"
	@$(PROMU) build

tarball: build $(PROMU)
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix ./tarballs

promu: $(PROMU)
$(PROMU):
	@GOOS=$(shell uname -s | tr A-Z a-z) \
	GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
	go get -u github.com/prometheus/promu


.PHONY: all style format build vet tarball $(PROMU)
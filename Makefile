PLUGIN_NAME := push-artifactory

GO ?= go
BINDIR := $(CURDIR)/bin
LDFLAGS   := -w -s
TESTS     := ./...
TESTFLAGS :=

HAS_DEP := $(shell command -v dep;)
HAS_GOX := $(shell command -v gox;)

TARGETS ?= darwin/amd64 linux/amd64
BIN_NAME := helm-push-artifactory
DIST_DIRS = find * -type d -maxdepth 0 -exec

VERSION ?= canary

SHELL=/bin/bash

.PHONY: build
build: vet
build: fmt
build: 
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINDIR)/${BIN_NAME} cmd/push/main.go 

vet:
	@go vet ./...

fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	@go get -u github.com/golang/dep/cmd/dep
endif
ifndef HAS_GOX
	@go get -u github.com/mitchellh/gox
endif
	@dep ensure -v -vendor-only

.PHONY: dist
dist: clean build-cross
dist:
	( \
		cd _dist && \
		$(DIST_DIRS) tar -zcf ${BIN_NAME}-${VERSION}-{}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r ${BIN_NAME}-${VERSION}-{}.zip {} \; \
	)

# usage: make clean build-cross dist VERSION=v0.2-alpha
.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross: 
	CGO_ENABLED=0 gox -parallel=2 -output="_dist/{{.OS}}-{{.Arch}}/${BIN_NAME}" -osarch='$(TARGETS)' -ldflags '$(LDFLAGS)' github.com/belitre/helm-push-artifactory-plugin/cmd/push

.PHONY: clean
clean:
	@rm -rf $(BINDIR) ./_dist

.PHONY: test
test: build
	@./scripts/test.sh

.PHONY: covhtml
covhtml: test
	@go tool cover -html=.cover/cover.out

.PHONY: install
install: build
	HELM_PUSH_PLUGIN_NO_INSTALL_HOOK=1 helm plugin install $(shell pwd)

.PHONY: remove
remove:
	helm plugin remove $(PLUGIN_NAME)

include versioning.mk
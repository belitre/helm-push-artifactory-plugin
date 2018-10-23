# makefile based on the one from helm: https://github.com/kubernetes/helm/blob/master/versioning.mk
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif

BINARY_VERSION ?= ${GIT_TAG}

# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X github.com/belitre/helm-push-artifactory-plugin/pkg/version.Version=${BINARY_VERSION}
endif

LDFLAGS += -X github.com/belitre/helm-push-artifactory-plugin/pkg/version.GitCommit=${GIT_COMMIT}

info:
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"

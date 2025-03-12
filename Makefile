CONFIG_PKG = github.com/wyt-labs/wyt-core/internal/pkg/config
APP_NAME = wyt-core
APP_START_DIR = cmd/core

GO_BIN = go
ifneq (${GO},)
	GO_BIN = ${GO}
endif

BUILD_TIME = $(shell date +%F-%Z/%T)
COMMIT_ID = $(shell git rev-parse HEAD)

ifeq ($(version),)
	# not specify version: make install
	APP_VERSION = $(shell git describe --abbrev=0 --tag)
	ifeq ($(APP_VERSION),)
		APP_VERSION = dev
	endif
else
	# specify version: make install version=v0.6.1-dev
	APP_VERSION = $(version)
endif

LDFLAGS = -X "${CONFIG_PKG}.Version=${APP_VERSION}"
LDFLAGS += -X "${CONFIG_PKG}.BuildTime=${BUILD_TIME}"
LDFLAGS += -X "${CONFIG_PKG}.CommitID=${COMMIT_ID}"

ifeq ($(target),)
	COMPILE_TARGET=
else
    PARAMS=$(subst -, ,$(target))
    ifeq ($(words $(PARAMS)),2)
    	OS=$(word 1, $(PARAMS))
    	ARCH=$(word 2, $(PARAMS))
    	COMPILE_TARGET=CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH)
    else
        $(error error param: '$(target)'! example: 'target=darwin-amd64')
    endif
endif

TEST_PKGS := $(shell go list ./...)

.PHONY: init install build dev-build lint fmt test precommit compile-network-pb compile-grpc-pb

# Init subModule
init:
	${GO_BIN} mod download
	${GO_BIN} install github.com/fsgo/go_fmt/cmd/gorgeous@latest
	${GO_BIN} install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

build:
	${GO_BIN} env -w CGO_LDFLAGS=""
	cd ${APP_START_DIR}  && $(COMPILE_TARGET) ${GO_BIN} build -ldflags '-s -w $(LDFLAGS)' -o ${APP_NAME}-${APP_VERSION}

dev-build:
	${GO_BIN} env -w CGO_LDFLAGS=""
	cd ${APP_START_DIR}  && $(COMPILE_TARGET) ${GO_BIN} build -gcflags="all=-N -l" -o ${APP_NAME}-${APP_VERSION}

# Check and print out style mistakes
lint:
	golangci-lint run --timeout=5m -v

# Formats go source code
fmt:
	gorgeous -local github.com/wyt -mi

# Test unit tests of source code
unittest:
	${GO_BIN} test -short -coverprofile coverage.txt -covermode=atomic ${TEST_PKGS}

package:build
	cd ../../
	cp ./${APP_START_DIR}/${APP_NAME}-${APP_VERSION} ./deploy/tools/bin/${APP_NAME}
	tar czvf ./${APP_NAME}-${APP_VERSION}.tar.gz -C ./deploy/ .

dev-package:dev-build
	cd ../../
	rm -f ./deploy/tools/bin/${APP_NAME}
	cp ./${APP_START_DIR}/${APP_NAME}-${APP_VERSION} ./deploy/tools/bin/${APP_NAME}

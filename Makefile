VERSION ?= $(shell git tag -l --sort=v:refname | tail -1)
GIT_COMMIT := $(shell git describe --match=NeVeRmAtCh --always --abbrev=40)
BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%SZ")
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

GOOS := $(shell go env GOHOSTOS)
GOARCH := $(shell go env GOHOSTARCH)
TARGET := sshvpn-${GOOS}-${GOARCH}
OS_ARCH := ${GOOS}/${GOARCH}

BASE := github.com/wencaiwulue/tlstunnel
FOLDER := ${BASE}/cmd/sshvpn
BUILD_DIR := ./build
OUTPUT_DIR := ./bin

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=--ldflags "\
 -X ${FOLDER}/cmds.Version=${VERSION} \
 -X ${FOLDER}/cmds.GitCommit=${GIT_COMMIT} \
 -X ${FOLDER}/cmds.BuildTime=${BUILD_TIME} \
 -X ${FOLDER}/cmds.Branch=${BRANCH} \
 -X ${FOLDER}/cmds.OsArch=${OS_ARCH} \
"

GO111MODULE=on
GOPROXY=https://goproxy.cn,direct

.PHONY: all
all: sshvpn-all

.PHONY: sshvpn-all
sshvpn-all: sshvpn-darwin-amd64 sshvpn-darwin-arm64 \
sshvpn-windows-amd64 sshvpn-windows-386 sshvpn-windows-arm64 \
sshvpn-linux-amd64 sshvpn-linux-386 sshvpn-linux-arm64

.PHONY: sshvpn
sshvpn:
	make $(TARGET)

# ---------darwin-----------
.PHONY: sshvpn-darwin-amd64
sshvpn-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn ${FOLDER}
	chmod +x $(OUTPUT_DIR)/sshvpn
.PHONY: sshvpn-darwin-arm64
sshvpn-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn ${FOLDER}
	chmod +x $(OUTPUT_DIR)/sshvpn
# ---------darwin-----------

# ---------windows-----------
.PHONY: sshvpn-windows-amd64
sshvpn-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn.exe ${FOLDER}
.PHONY: sshvpn-windows-arm64
sshvpn-windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn.exe ${FOLDER}
.PHONY: sshvpn-windows-386
sshvpn-windows-386:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn.exe ${FOLDER}
# ---------windows-----------

# ---------linux-----------
.PHONY: sshvpn-linux-amd64
sshvpn-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn ${FOLDER}
	chmod +x $(OUTPUT_DIR)/sshvpn
.PHONY: sshvpn-linux-arm64
sshvpn-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn ${FOLDER}
	chmod +x $(OUTPUT_DIR)/sshvpn
.PHONY: sshvpn-linux-386
sshvpn-linux-386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build ${LDFLAGS} -o $(OUTPUT_DIR)/sshvpn ${FOLDER}
	chmod +x $(OUTPUT_DIR)/sshvpn
# ---------linux-----------
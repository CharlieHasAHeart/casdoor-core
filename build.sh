#!/bin/bash
set -euo pipefail

# Keep a fallback chain for transient proxy failures. The Dockerfile provides
# the same value, while this default also makes direct local execution robust.
: "${GOPROXY:=https://goproxy.cn|https://proxy.golang.org|direct}"
export GOPROXY
echo "Using GOPROXY=${GOPROXY}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server_linux_amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o server_linux_arm64 .

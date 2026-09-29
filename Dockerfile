# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM registry.cloudbility.com.cn/charmirror/golang@sha256:0c12ba349422ad18498f5f4279086eedac0d30c65fa0b3b37bb860d5f4d58054 AS build
WORKDIR /go/src/casdoor
ENV GOPROXY=https://goproxy.cn|https://proxy.golang.org|direct

# Copy only go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN --mount=type=cache,id=casdoor-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=casdoor-go-build,target=/root/.cache/go-build,sharing=locked \
    set -eu; \
    for attempt in 1 2 3 4 5; do \
      if go mod download -x; then \
        exit 0; \
      fi; \
      echo "go mod download failed; retry ${attempt}/5" >&2; \
      if [ "${attempt}" -lt 5 ]; then sleep 5; fi; \
    done; \
    exit 1

# Copy source files
COPY . .

RUN go test -v -run TestGetVersionInfo ./util/system_test.go ./util/system.go ./util/variable.go
RUN --mount=type=cache,id=casdoor-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=casdoor-go-build,target=/root/.cache/go-build,sharing=locked \
    ./build.sh

FROM debian:latest AS allinone
LABEL MAINTAINER="https://casdoor.org/"
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

WORKDIR /
RUN apt update
RUN apt install -y ca-certificates lsof && update-ca-certificates

WORKDIR /
COPY --from=build /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=build /go/src/casdoor/swagger ./swagger
COPY --from=build /go/src/casdoor/docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=build /go/src/casdoor/conf/app.conf ./conf/app.conf

ENTRYPOINT ["/bin/bash"]
CMD ["/docker-entrypoint.sh"]


FROM alpine:latest AS standard
LABEL MAINTAINER="https://casdoor.org/"
ARG USER=casdoor
ARG TARGETOS
ARG TARGETARCH
ENV BUILDX_ARCH="${TARGETOS:-linux}_${TARGETARCH:-amd64}"

RUN sed -i 's/https/http/' /etc/apk/repositories
RUN apk add --update sudo
RUN apk add tzdata
RUN apk add curl
RUN apk add ca-certificates && update-ca-certificates

RUN adduser -D $USER -u 1000 \
    && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
    && chmod 0440 /etc/sudoers.d/$USER \
    && mkdir logs \
    && chown -R $USER:$USER logs

USER 1000
WORKDIR /
COPY --from=build --chown=$USER:$USER /go/src/casdoor/server_${BUILDX_ARCH} ./server
COPY --from=build --chown=$USER:$USER /go/src/casdoor/swagger ./swagger
COPY --from=build --chown=$USER:$USER /go/src/casdoor/conf/app.conf ./conf/app.conf

EXPOSE 8000
ENTRYPOINT ["/server"]

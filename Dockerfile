 # Build the manager binary
FROM golang:1.13 as builder

# 增加压缩二进制压缩工具
RUN apt-get -y update && apt-get -y install upx

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer


# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# 增加backup
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="goproxy.cn"



RUN go mod download && \
    go build -a -o manager main.go && \
    go build -a -o backup cmd/backup/main.go && \
    upx manager backup
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as manager
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

# backup
FROM gcr.io/distroless/static:nonroot as backup
WORKDIR /
COPY --from=builder /workspace/backup .
USER nonroot:nonroot

ENTRYPOINT ["/backup"]
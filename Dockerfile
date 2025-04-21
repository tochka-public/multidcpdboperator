# Build the manager binary
FROM golang:1.24 as builder

ENV GO111MODULE=on

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY webhook/ webhook/
COPY metrics/ metrics/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM debian:12.10
WORKDIR /
COPY --from=builder /workspace/manager .
USER nobody
#USER 65532:65532

ENTRYPOINT ["/manager"]

FROM golang:1.26-trixie
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN apt-get -qq update \
    && apt-get -yqq install upx-ucl
RUN mkdir -p /dcr-mcp
WORKDIR /dcr-mcp
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go cli.go ./
ADD internal internal
RUN go build \
    -a \
    -ldflags "-s -w -extldflags '-static'" \
    -installsuffix cgo \
    -tags netgo \
    -o /bin/app \
    .
RUN upx -q -9 /bin/app

FROM gcr.io/distroless/static
COPY --from=0 /bin/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]

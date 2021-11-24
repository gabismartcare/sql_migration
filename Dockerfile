FROM gabismartcare/golang-proto:17.0-3.17.3-stretch as builder
LABEL stage=builder

WORKDIR /src
ENV GOPATH /go
COPY . /src/

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o /build/main /src/main.go

FROM gabismartcare/base-image:1.0.0
COPY --from=builder --chown=gabi /build/* /go/

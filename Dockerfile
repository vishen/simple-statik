FROM golang:1.12.1-alpine3.9 as builder
WORKDIR /go/src/github.com/vishen/simple-statik
COPY main.go main.go
RUN CGO_ENABLED=0 go build -tags netgo -installsuffix netgo

RUN apk add --no-cache git && go get -u github.com/hashicorp/go-getter/cmd/go-getter

FROM alpine:3.9
WORKDIR /app
RUN apk add --no-cache ca-certificates
ADD example.config example.config
COPY --from=builder /go/src/github.com/vishen/simple-statik/simple-statik simple-statik
COPY --from=builder /go/bin/go-getter go-getter
ENTRYPOINT ["/app/simple-statik"]

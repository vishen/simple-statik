FROM golang:1.12.1-alpine3.9 as builder
WORKDIR /go/src/github.com/vishen/simple-statik
COPY main.go main.go
RUN CGO_ENABLED=0 go build -tags netgo -installsuffix netgo

FROM scratch
WORKDIR /app
ADD example.config example.config
COPY --from=builder /go/src/github.com/vishen/simple-statik/simple-statik simple-statik
ENTRYPOINT ["/app/simple-statik"]

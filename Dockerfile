ARG GOVERSION=latest
FROM golang:$GOVERSION AS builder

WORKDIR /src
COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o . -ldflags="-s -w -X main.release=$(git rev-parse HEAD)" .

FROM alpine:3

RUN apk add --no-cache ca-certificates tzdata
COPY ./docker-entrypoint.sh /
COPY --from=builder /src/sentlog /bin/sentlog
ENTRYPOINT ["/docker-entrypoint.sh"]
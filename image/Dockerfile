FROM alpine:3.10

RUN apk add --no-cache ca-certificates=20191127-r2

COPY infra-server /infra-server

COPY static /etc/infra/static

ENTRYPOINT ["/infra-server"]

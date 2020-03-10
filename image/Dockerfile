FROM alpine:3.10

RUN apk add --no-cache ca-certificates=20190108-r0

COPY infra-server /infra-server

COPY static /etc/infra/static

ENTRYPOINT ["/infra-server"]

FROM alpine:3.10

RUN apk add --no-cache ca-certificates=20190108-r0

COPY static /static

COPY infra-server /infra-server

ENTRYPOINT ["/infra-server"]

FROM alpine:3.9

RUN apk add --no-cache ca-certificates=20190108-r0

COPY infra-server /infra-server

ENTRYPOINT ["/infra-server"]

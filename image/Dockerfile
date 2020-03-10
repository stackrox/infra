FROM alpine:3.10

RUN apk add --no-cache ca-certificates=20190108-r0

COPY static /etc/infra/

COPY infra-server /infra-server

COPY infractl-darwin-amd64 infractl-linux-amd64 /etc/infra/static/downloads/

ENTRYPOINT ["/infra-server"]

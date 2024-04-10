FROM golang:1.22.2 as golang-builder

WORKDIR /go/src/github.com/stackrox/infra

COPY . .

RUN make server cli

FROM node:20.11.0 as ui-builder

COPY ui ui

RUN --mount=type=secret,id=npmrc,target=/root/.npmrc make -C ui all

FROM alpine:3.19.1 as app

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infra-server-linux-amd64 /infra-server

COPY --from=ui-builder /ui/build /etc/infra/static

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infractl-* /etc/infra/static/downloads/

ENTRYPOINT ["/infra-server"]

FROM registry.access.redhat.com/ubi8/go-toolset:1.19 as golang-builder

WORKDIR /go/src/github.com/stackrox/infra

USER root

COPY . .

RUN git config --global --add safe.directory /go/src/github.com/stackrox/infra && \
    make server cli

# FROM node:16.20.2 as ui-builder

# COPY ui ui

# RUN --mount=type=secret,id=npmrc,target=/root/.npmrc make -C ui all

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest as app

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infra-server-linux-amd64 /infra-server

# COPY --from=ui-builder /ui/build /etc/infra/static

USER 4000:4000

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infractl-* /etc/infra/static/downloads/

ENTRYPOINT ["/infra-server"]

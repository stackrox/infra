# Stage 1: Build Go backend
FROM golang:1.24.7 as golang-builder

WORKDIR /go/src/github.com/stackrox/infra

# Copy dependency files first (changes less frequently)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (changes more frequently)
COPY . .

# Build
RUN make server cli

# Stage 2: Build UI
FROM node:24.8.0 as ui-builder

WORKDIR /ui

# Copy package files first (changes less frequently)
COPY ui/package.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile

# Copy UI source code (changes more frequently)
COPY ui ./

# Build
RUN make -C /ui all

# Stage 3: Final image
FROM alpine:3.22.2 as app

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infra-server-linux-amd64 /infra-server

COPY --from=ui-builder /ui/build /etc/infra/static

COPY --from=golang-builder /go/src/github.com/stackrox/infra/bin/infractl-* /etc/infra/static/downloads/

ENTRYPOINT ["/infra-server"]

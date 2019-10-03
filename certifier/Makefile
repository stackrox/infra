.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o certifier -ldflags='-s -w' *.go
	docker build -t certifier:latest .
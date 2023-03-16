.PHONY: build image

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o credtest main.go

image: build
	docker buildx build --platform linux/amd64 --tag jesseh/credtest:latest --no-cache --load --progress plain .

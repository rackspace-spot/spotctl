NAME=spotctl

all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o dist/$(NAME)-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o dist/$(NAME)-linux-arm64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o dist/$(NAME)-darwin-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o dist/$(NAME)-darwin-arm64

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o dist/$(NAME)-windows-amd64.exe

clean:
	rm -rf dist

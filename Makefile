run:
	go run cmd/glsync/main.go
build:
	go build ./...
install:
	go install ./...
test:
	go test -v ./...
build-linux-amd64:
	GOOS=linux
	GOARCH=amd64
	go build -o glsync-linux-amd64 cmd/glsync/main.go
build-windows-amd64:
	GOOS=windows
	GOARCH=amd64
	go build -o glsync-windows-amd64.exe cmd/glsync/main.go
build-darwin-amd64:
	GOOS=darwin
	GOARCH=amd64
	go build -o glsync-darwin-amd64 cmd/glsync/main.go
build-linux-arm64:
	GOOS=linux
	GOARCH=arm64
	go build -o glsync-linux-arm64 cmd/glsync/main.go
build-darwin-arm64:
	GOOS=darwin
	GOARCH=arm64
	go build -o glsync-darwin-arm64 cmd/glsync/main.go
run:
	go run cmd/glsync/main.go
build:
	go build .
install:
	go install .
test:
	go test -v ./...
build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o glsync-linux-amd64 main.go
build-windows-amd64:
	env GOOS=windows GOARCH=amd64 go build -o glsync-windows-amd64.exe main.go
build-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -o glsync-darwin-amd64 main.go
build-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -o glsync-linux-arm64 main.go
build-darwin-arm64:
	env GOOS=darwin GOARCH=arm64 go build -o glsync-darwin-arm64 main.go
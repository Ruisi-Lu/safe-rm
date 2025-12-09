.PHONY: build test clean install

# Build binary
build:
	go build -o rm ./cmd/rm

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o rm-linux-amd64 ./cmd/rm
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o rm-linux-arm64 ./cmd/rm

# Run tests
test:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -v -race ./...

# Clean build artifacts
clean:
	rm -f rm rm-linux-* rm-darwin-*

# Install to ~/.local/bin
install: build
	mkdir -p ~/.local/bin
	cp rm ~/.local/bin/rm
	@echo "Installed to ~/.local/bin/rm"
	@echo "Make sure ~/.local/bin is in your PATH"

# Install system-wide (requires sudo)
install-system: build
	sudo cp rm /usr/local/bin/rm
	@echo "Installed to /usr/local/bin/rm"

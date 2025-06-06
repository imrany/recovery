all: build

build: build-macos build-linux build-windows

dev: 
	@echo 'starting dev server'
	CompileDaemon -build="go build -o ./recovery main.go" -command="./recovery"

run:
	@echo 'running recovery...'
	sudo ./recovery

clean:
	@echo 'cleaning up...'
	rm -rf bin
	rm -f recovery
	

build-macos:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o bin/app-macos

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/app-linux

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o bin/app-windows.exe

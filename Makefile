.PHONY: build install uninstall clean

BINARY := ghosted

# Default install uses Go's standard bin directory
install:
	go install .
	@echo "Installed $(BINARY) to your Go bin directory"
	@echo "Make sure \$$GOPATH/bin or \$$HOME/go/bin is in your PATH"

build:
	go build -o $(BINARY)

clean:
	rm -f $(BINARY)

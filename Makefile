GO ?= go
GOCACHE ?= /tmp/dev-doctor-go-build
BINARY ?= dist/devdoctor
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: build install clean

build:
	mkdir -p "$(dir $(BINARY))"
	GOCACHE="$(GOCACHE)" $(GO) build -trimpath -o "$(BINARY)" ./cmd/devdoctor

install: build
	install -d "$(INSTALL_DIR)"
	install -m 0755 "$(BINARY)" "$(INSTALL_DIR)/devdoctor"

clean:
	rm -rf dist

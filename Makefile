VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)
APP_NAME = yiwoDraftViewer
BUILD_DIR = build/bin
ZIP_NAME = $(APP_NAME)-macos-$(VERSION).zip

.PHONY: dev test build package clean

dev:
	wails dev

test:
	go test ./...

build:
	wails build -m -ldflags "$(LDFLAGS)"

package: build
	cd $(BUILD_DIR) && zip -r $(ZIP_NAME) $(APP_NAME).app
	@echo "Built: $(BUILD_DIR)/$(ZIP_NAME)"

clean:
	rm -rf build/bin

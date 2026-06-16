VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "0.1.0-dev")
APP_NAME = yiwoDraftViewer
BUILD_DIR = build/bin
ZIP_NAME = $(APP_NAME)-macos-$(VERSION).zip

.PHONY: dev test build package clean

dev:
	wails dev

test:
	go test ./...

build:
	wails build -m

package: build
	cd $(BUILD_DIR) && zip -r $(ZIP_NAME) $(APP_NAME).app
	@echo "Built: $(BUILD_DIR)/$(ZIP_NAME)"

clean:
	rm -rf build/bin

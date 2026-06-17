TAG_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed -E 's/-[0-9]+-g[0-9a-f]+$$//' | grep -v '^$$' || echo "dev")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.TagVersion=$(TAG_VERSION) -X main.CommitID=$(COMMIT_ID)
APP_NAME = yiwoDraftViewer
BUILD_DIR = build/bin
ZIP_NAME = $(APP_NAME)-macos-$(TAG_VERSION)-$(COMMIT_ID).zip

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

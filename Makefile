TAG_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed -E 's/-[0-9]+-g[0-9a-f]+$$//' | grep -v '^$$' || echo "dev")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.TagVersion=$(TAG_VERSION) -X main.CommitID=$(COMMIT_ID)
APP_NAME = yiwoDraftViewer
BUILD_DIR = build/bin
STAGING_DIR = build/staging
ZIP_NAME = $(APP_NAME)-macos-$(TAG_VERSION)-$(COMMIT_ID).zip
TEST_BUNDLE_DIR = /tmp/yiwo-test/yiwo-draft-viewer.app

.PHONY: dev test build package clean staging test-staging clean-staging

dev:
	wails dev

test:
	go test ./...

build:
	wails build -m -ldflags "$(LDFLAGS)"

staging: clean-staging
	@mkdir -p $(STAGING_DIR)
	wails build -m -ldflags "$(LDFLAGS)"
	mv build/bin/yiwo-draft-viewer.app $(STAGING_DIR)/
	@echo "Staging bundle: $(STAGING_DIR)/yiwo-draft-viewer.app"

test-staging: staging
	@bash scripts/test-staging.sh

package: build
	cd $(BUILD_DIR) && zip -r $(ZIP_NAME) $(APP_NAME).app
	@echo "Built: $(BUILD_DIR)/$(ZIP_NAME)"

clean:
	rm -rf build/bin

clean-staging:
	@rm -rf $(STAGING_DIR)
	@rm -rf $(TEST_BUNDLE_DIR)
	@echo "Staging artifacts removed"

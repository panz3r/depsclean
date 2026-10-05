PREFIX ?= /usr/local
VERSION ?= $(shell git describe --tags --dirty --always | sed -e 's/^v//')
GOFLAGS ?= -buildvcs=false
IS_SNAPSHOT = $(if $(findstring -, $(VERSION)),true,false)
MAJOR_VERSION = $(word 1, $(subst ., ,$(VERSION)))
MINOR_VERSION = $(word 2, $(subst ., ,$(VERSION)))
PATCH_VERSION = $(word 3, $(subst ., ,$(word 1,$(subst -, , $(VERSION)))))
NEW_VERSION ?= $(MAJOR_VERSION).$(MINOR_VERSION).$(shell echo $$(( $(PATCH_VERSION) + 1)) )
GOVULNCHECK_PACKAGE ?= golang.org/x/vuln/cmd/govulncheck@v1

fix = false
ifeq (true,$(fix))
	FIX = --fix
endif

DEPSCLEAN ?= go run ./cmd/depsclean

.PHONY: build
build:
	GOFLAGS="$(GOFLAGS)" go build -ldflags "-X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o dist/local/depsclean ./cmd/depsclean

.PHONY: build-all
build-all:
	mkdir -p builds
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_linux_amd64 ./cmd/depsclean
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_linux_arm64 ./cmd/depsclean
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_macos_arm64 ./cmd/depsclean
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_macos_intel ./cmd/depsclean
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_windows_amd64.exe ./cmd/depsclean
	GOFLAGS="$(GOFLAGS)" CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w -X github.com/panz3r/depsclean/internal/update.Version=$(VERSION)" -o builds/depsclean_windows_arm64.exe ./cmd/depsclean

.PHONY: release-build
release-build: clean
	$(MAKE) build-all

.PHONY: format
format:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: lint-go
lint-go:
	golangci-lint run $(FIX)

.PHONY: lint
lint: lint-go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: install
install: build
	@cp dist/local/depsclean $(PREFIX)/bin/depsclean
	@chmod 755 $(PREFIX)/bin/depsclean

.PHONY: deps-tools
deps-tools: ## install tool dependencies
	"$(shell go env GOROOT)/bin/go" install $(GOVULNCHECK_PACKAGE)

.PHONY: security-check
security-check: deps-tools
	GOEXPERIMENT= "$(shell go env GOROOT)/bin/go" run $(GOVULNCHECK_PACKAGE) -show color ./...

.PHONY: pr-checks
pr-checks: tidy format lint test security-check build clean

.PHONY: clean
clean:
	rm -rf dist builds

.PHONY: upgrade
upgrade:
	go get -u ./...
	go mod tidy

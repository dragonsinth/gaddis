dev_build_version=$(shell git describe --tags --always --dirty)
short_version := $(subst v,,$(word 1,$(subst -, ,$(dev_build_version))))
git_sha := $(shell git rev-parse HEAD)
go_mod := $(shell go list .)
LDFLAGS := '-X "main.version=dev build $(dev_build_version)" -X "$(go_mod)/asm.GitSha=$(git_sha)" -X "$(go_mod)/asm.GoMod=$(go_mod)"'
GOFLAGS := -ldflags $(LDFLAGS) -trimpath
export GOWORK=off

# Disable CGO for improved compatibility across distros
export CGO_ENABLED=0

# TODO: run golint and errcheck, but only to catch *new* violations and
# decide whether to change code or not (e.g. we need to be able to whitelist
# violations already in the code). They can be useful to catch errors, but
# they are just too noisy to be a requirement for a CI -- we don't even *want*
# to fix some of the things they consider to be violations.
.PHONY: ci
ci: deps checkgofmt vet staticcheck ineffassign predeclared test

.PHONY: deps
deps:
	go get -v -t ./...
	go mod tidy

.PHONY: updatedeps
updatedeps:
	go get -d -v -t -u -f ./...
	go mod tidy

.PHONY: install
install:
	go install $(GOFLAGS) ./...
	@code --uninstall-extension dragonsinth.gaddis-vscode > /dev/null 2>&1 && echo "uninstalled previous extension" || echo "no problem"
	@code --install-extension vscode-gaddis/gaddis-vscode.vsix || echo "warning: failed to install into vscode"

.PHONY: release
release:
	@go install github.com/goreleaser/goreleaser@v1.21.0
	goreleaser release --clean

.PHONY: snapshot
snapshot:
	@go install github.com/goreleaser/goreleaser@v1.21.0
	goreleaser build --snapshot --clean

.PHONY: checkgofmt
checkgofmt:
	gofmt -s -l .
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		git diff; \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@v0.5.1
	staticcheck ./...

.PHONY: ineffassign
ineffassign:
	@go install github.com/gordonklaus/ineffassign@7953dde2c7bf
	ineffassign .

.PHONY: predeclared
predeclared:
	@go install github.com/nishanths/predeclared@245576f9a85c96ea16c750df3887f1d827f01e9c
	predeclared -ignore append ./...

# Intentionally omitted from CI, but target here for ad-hoc reports.
.PHONY: golint
golint:
	@go install golang.org/x/lint/golint@v0.0.0-20210508222113-6edffad5e616
	golint -min_confidence 0.9 -set_exit_status ./...

# Intentionally omitted from CI, but target here for ad-hoc reports.
.PHONY: errcheck
errcheck:
	@go install github.com/kisielk/errcheck@v1.2.0
	errcheck ./...

.PHONY: test
test:
	# The race detector requires CGO: https://github.com/golang/go/issues/6508
	CGO_ENABLED=1 go test -race ./...

.PHONY: plugin
plugin:
	cd vscode-gaddis && npm install
	cd vscode-gaddis && npm run compile
	mkdir -p vscode-gaddis/bin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-darwin-arm64 ./cmd/gaddis
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-darwin-amd64 ./cmd/gaddis
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-linux-arm64 ./cmd/gaddis
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-linux-amd64 ./cmd/gaddis
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-windows-arm64.exe ./cmd/gaddis
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o vscode-gaddis/bin/gaddis-windows-amd64.exe ./cmd/gaddis
	@rm -f vscode-gaddis/gaddis-vscode-*.vsix
	cd vscode-gaddis && npx vsce package -o gaddis-vscode.vsix $(short_version)

#### General ####

# All target for when make is run on its own.
.PHONY: all
all: style

#### Linting ####

# Format code. Unformatted code will fail in CI.
.PHONY: style
style:
ifdef GITHUB_ACTIONS
	goimports -l .
else
	goimports -l -w .
endif

#### Building ####

# Build binary.
.PHONY: build
build:
	@mkdir -p dist
	go build -trimpath -ldflags='-s -w' -o dist/aws-auth main.go
	@echo 'Binary build! Run it like so:'
	@echo '  $$ ./dist/aws-auth'

#### Release ####

# Build and publish binaries as Github release artifacts.
.PHONY: release
release:
ifdef GITHUB_ACTIONS
	goreleaser release
else
	goreleaser --rm-dist --skip-publish --skip-validate --snapshot
endif

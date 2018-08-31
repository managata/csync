#
#
#

export GOPATH=/z/go

TARGET := csync
BINARY := $(TARGET)

VER := $(shell git describe)
OS := $(shell go env | grep GOOS | sed -e 's/.*="\(.*\)"/\1/')
ARCH := $(shell go env | grep GOARCH | sed -e 's/.*="\(.*\)"/\1/')
RELEASE := ${OS}-${ARCH}-$(VER)

DEBUG := -tags debug

VERSION := -X main.version='$(VER)'
STRIP := -s -w
EXT := -Wl,--allow-multiple-definition

BUILD_LDFLAGS := -ldflags "$(VERSION) $(STRIP) -extldflags '$(EXT)'"
STATIC_LDFLAGS := -ldflags "$(VERSION) $(STRIP) -extldflags '-static $(EXT)'"
DEBUG_LDFLAGS := -ldflags "$(VERSION) -extldflags '$(EXT)'"

GO := go
export CC=gcc
export LD=ld


.PHONY: build
build:
	$(GO) build -v $(BUILD_LDFLAGS) -o $(BINARY)

.PHONY: static
static:
	$(GO) build -v $(STATIC_LDFLAGS) -o $(BINARY)

.PHONY: debug
debug:
	$(GO) build -v $(DEBUG) $(DEBUG_LDFLAGS) -o $(BINARY)

.PHONY: deps
deps:
	$(GO) get -v "golang.org/x/sync/errgroup"

.PHONY: dist
dist:
	make static
	- rm -rf $(TARGET)-$(RELEASE)
	mkdir $(TARGET)-$(RELEASE)
	cp -a $(BINARY) $(TARGET)-$(RELEASE)
	cp -a LICENSE $(TARGET)-$(RELEASE)
	gtar cJ --owner=managata --group=csync -f $(TARGET)-$(RELEASE).tar.xz $(TARGET)-$(RELEASE)
	rm -rf $(TARGET)-$(RELEASE)

.PHONY: clean
clean:
	rm -f $(TARGET)
	rm -f $(BINARY)
	rm -f $(TARGET).exe
	rm -rf $(TARGET)-*

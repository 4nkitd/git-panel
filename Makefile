BINARY  := git-panel
MODULE  := github.com/4nkitd/git-panel
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install clean vet

build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) ./cmd/zedgit/

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/zedgit/

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

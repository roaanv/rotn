BINARY  := rotn
BUILD   := bin/$(BINARY)
BINDIR  ?= $(HOME)/.local/bin

.PHONY: all build install uninstall test clean

all: build

build:
	go build -o $(BUILD) ./cmd/rotn

install: build
	install -d $(BINDIR)
	install -m 0755 $(BUILD) $(BINDIR)/$(BINARY)

uninstall:
	rm -f $(BINDIR)/$(BINARY)

test:
	go test ./...

clean:
	rm -rf bin

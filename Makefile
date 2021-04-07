LASTTAG := $(shell git describe --abbrev=0 --tags 2>/dev/null)

GOCMD=go
GOBUILD=$(GOCMD) build
LDFLAGS=-X github.com/markkuit/telegram-bot-archiver/internal/commons.Version=$(LASTTAG)
BINDIR=$(CURDIR)/bin
BINNAME=telegram-bot-archiver

default: build run
build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -v -o $(BINDIR)/$(BINNAME) cmd/telegram-bot-archiver/telegram-bot-archiver.go
run:
	$(BINDIR)/$(BINNAME) $(ARGS)

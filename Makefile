BINARY := cardsgen
PKG := ./cmd/cardsgen

GO ?= go
GOFLAGS ?=
LDFLAGS ?= -s -w

.PHONY: all build run test vet fmt tidy clean install

all: build

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

test:
	$(GO) test $(GOFLAGS) ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(PKG)

clean:
	rm -f $(BINARY)
	$(GO) clean

.PHONY: ui build test run clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Compila a interface (precisa de Node 20.19+ ou 22+)
ui:
	cd web && npm ci && npm run build

# Gera o binário iz com a interface e os cenários embutidos
build: ui
	go build -ldflags "-X main.version=$(VERSION)" -o iz ./cmd/iz

test:
	go test ./...

run: build
	./iz ui

clean:
	rm -rf iz web/dist/assets web/dist/index.html web/dist/favicon.svg

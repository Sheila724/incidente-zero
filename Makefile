.PHONY: ui build run clean

# Compila a interface (precisa de Node 20.19+ ou 22+)
ui:
	cd web && npm ci && npm run build

# Gera o binário iz com a interface embutida
build: ui
	go build -o iz ./cmd/iz

run: build
	./iz ui

clean:
	rm -rf iz web/dist/assets web/dist/index.html web/dist/favicon.svg

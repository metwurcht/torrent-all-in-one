.PHONY: all cli gui clean install release help

# Variables
BINARY_NAME=torrent-aio
GUI_BINARY_NAME=torrent-aio-gui
BUILD_DIR=build/bin
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: cli gui ## Compile CLI et GUI

cli: ## Compile le CLI
	@echo "🔨 Compilation du CLI..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/torrent-aio
	@echo "✅ CLI compilé : $(BUILD_DIR)/$(BINARY_NAME)"

gui: ## Compile le GUI avec Wails
	@echo "🔨 Compilation du GUI..."
	@command -v wails >/dev/null 2>&1 || { echo "❌ Wails n'est pas installé. Installez-le avec: go install github.com/wailsapp/wails/v2/cmd/wails@latest"; exit 1; }
	wails build -clean
	@echo "✅ GUI compilé : $(BUILD_DIR)/$(GUI_BINARY_NAME)"

dev-gui: ## Lance le GUI en mode développement
	wails dev

clean: ## Nettoie les fichiers de build
	@echo "🧹 Nettoyage..."
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	@echo "✅ Nettoyage terminé"

install: cli ## Installe le CLI dans /usr/local/bin
	@echo "📦 Installation..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installé dans /usr/local/bin/$(BINARY_NAME)"

release: ## Crée une release multi-plateforme
	@echo "🚀 Création des releases..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/torrent-aio
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/torrent-aio
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/torrent-aio
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/torrent-aio
	@echo "✅ Releases créées dans $(BUILD_DIR)/"

test: ## Lance les tests
	go test -v ./...

fmt: ## Formate le code
	go fmt ./...

lint: ## Lance le linter
	golangci-lint run

run: cli ## Compile et exécute le CLI
	./$(BUILD_DIR)/$(BINARY_NAME)

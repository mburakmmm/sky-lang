# SKY Programming Language - Justfile
# Just ile hızlı geliştirme komutları

# Varsayılan komut: yardım göster
default:
    @just --list

# Tüm araçları derle
build:
    @echo "🔨 Building SKY toolchain..."
    @make build

# Sadece sky'ı derle ve çalıştır
run FILE:
    @echo "🚀 Running {{FILE}}..."
    @make build-sky
    @./bin/sky run {{FILE}}

# Hızlı test
test:
    @echo "🧪 Running tests..."
    @go test -v ./...

# Race detector ile test
test-race:
    @echo "🏁 Running tests with race detector..."
    @go test -race ./...

# Belirli bir paketi test et
test-pkg PKG:
    @echo "🧪 Testing package {{PKG}}..."
    @go test -v ./{{PKG}}/...

# Lint kontrolü
lint:
    @echo "🔍 Running linters..."
    @make lint

# Kod formatla
fmt:
    @echo "✨ Formatting code..."
    @make fmt

# Lexer çıktısını göster
dump-tokens FILE:
    @echo "📝 Dumping tokens for {{FILE}}..."
    @./bin/sky dump --tokens {{FILE}}

# AST çıktısını göster
dump-ast FILE:
    @echo "🌳 Dumping AST for {{FILE}}..."
    @./bin/sky dump --ast {{FILE}}

# Semantik kontrol
check FILE:
    @echo "✅ Checking {{FILE}}..."
    @./bin/sky check {{FILE}}

# Tüm örnekleri çalıştır
run-examples:
    @echo "📚 Running all examples..."
    @for f in examples/**/*.sky; do \
        echo "Running $f..."; \
        ./bin/sky run "$f"; \
    done

# Temizlik
clean:
    @echo "🧹 Cleaning..."
    @make clean

# Dev ortamı hazırla
setup:
    @echo "⚙️  Setting up development environment..."
    @go mod download
    @go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "✅ Setup complete!"

# Coverage raporu
coverage:
    @echo "📊 Generating coverage report..."
    @make test-coverage
    @open coverage.html

# REPL başlat
repl:
    @echo "💬 Starting SKY REPL..."
    @./bin/sky repl

# Yeni örnek oluştur
new-example NAME:
    @echo "📄 Creating new example: {{NAME}}"
    @mkdir -p examples/{{NAME}}
    @echo "# {{NAME}} example\nfunction main\n  print(\"Hello from {{NAME}}\")\nend" > examples/{{NAME}}/main.sky

# Git hook'ları kur
install-hooks:
    @echo "🪝 Installing git hooks..."
    @echo '#!/bin/sh\nmake lint && make test' > .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "✅ Git hooks installed!"

# Benchmark çalıştır
bench:
    @echo "⚡ Running benchmarks..."
    @go test -bench=. -benchmem ./...

# Profiling
profile PKG:
    @echo "📈 Profiling {{PKG}}..."
    @go test -cpuprofile=cpu.prof -memprofile=mem.prof ./{{PKG}}
    @echo "Profile files created: cpu.prof, mem.prof"

# Wing komutları
wing-init:
    @echo "📦 Initializing wing project..."
    @./bin/wing init

wing-install PKG:
    @echo "📦 Installing package {{PKG}}..."
    @./bin/wing install {{PKG}}

# LSP server başlat
lsp:
    @echo "🔌 Starting LSP server..."
    @./bin/skyls

# Debugger başlat
debug FILE:
    @echo "🐛 Starting debugger for {{FILE}}..."
    @./bin/skydbg {{FILE}}

# CI pipeline lokal olarak çalıştır
ci:
    @echo "🔄 Running CI pipeline..."
    @make ci

# Release build
release VERSION:
    @echo "🚀 Building release {{VERSION}}..."
    @./scripts/release.sh {{VERSION}}

# Dokümantasyon oluştur
docs:
    @echo "📚 Generating documentation..."
    @go doc -all ./... > docs/api.txt
    @echo "✅ Documentation generated!"

# Tüm testler ve kontroller
all: lint test e2e
    @echo "✅ All checks passed!"


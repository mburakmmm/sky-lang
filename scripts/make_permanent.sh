#!/bin/bash

# 🚀 SKY Language Support - Kalıcı Kurulum Script
# Bu script extension'ı sistem genelinde kalıcı hale getirir

set -e

echo "🔧 SKY Language Support - Kalıcı Kurulum Başlıyor..."

# VS Code extension dizinini bul
VSCODE_EXT_DIR=""
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    VSCODE_EXT_DIR="$HOME/.vscode/extensions"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    VSCODE_EXT_DIR="$HOME/.vscode/extensions"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    # Windows (Git Bash/Cygwin)
    VSCODE_EXT_DIR="$HOME/.vscode/extensions"
else
    echo "❌ Desteklenmeyen işletim sistemi: $OSTYPE"
    exit 1
fi

# Extension kaynak dizini
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.vscode/extensions/sky-language"

# Hedef dizin
TARGET_DIR="$VSCODE_EXT_DIR/sky-language-0.1.0"

echo "📁 Kaynak: $SOURCE_DIR"
echo "📁 Hedef: $TARGET_DIR"

# VS Code extension dizinini oluştur
mkdir -p "$VSCODE_EXT_DIR"

# Eski extension'ı sil (varsa)
if [[ -d "$TARGET_DIR" ]]; then
    echo "🗑️  Eski extension kaldırılıyor..."
    rm -rf "$TARGET_DIR"
fi

# Yeni extension'ı kopyala
echo "📦 Extension kopyalanıyor..."
cp -r "$SOURCE_DIR" "$TARGET_DIR"

# İzinleri düzenle
chmod -R 755 "$TARGET_DIR"

echo "✅ Extension kalıcı olarak kuruldu!"
echo ""
echo "🎯 Şimdi yapmanız gerekenler:"
echo "1. VS Code'u tamamen kapatın (Cmd+Q)"
echo "2. VS Code'u yeniden açın"
echo "3. Herhangi bir .sky dosyası açın"
echo "4. Sağ alt köşede 'SKY' yazmalı"
echo ""
echo "📝 Extension artık kalıcı! VS Code'u her açtığınızda çalışacak."
echo "🔗 Konum: $TARGET_DIR"

# VS Code'un extension'ı tanıması için bir flag dosyası oluştur
echo "$(date)" > "$TARGET_DIR/.installed"

echo ""
echo "🎉 Kurulum tamamlandı!"

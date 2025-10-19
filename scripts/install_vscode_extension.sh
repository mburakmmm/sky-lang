#!/bin/bash
# Install SKY VS Code extension

echo "🎨 Installing SKY VS Code Extension..."
echo ""

# Check if VS Code is installed
if ! command -v code &> /dev/null; then
    echo "❌ VS Code 'code' command not found!"
    echo "Please install VS Code command line tools:"
    echo "  1. Open VS Code"
    echo "  2. Cmd+Shift+P"
    echo "  3. Type: 'Shell Command: Install code command in PATH'"
    exit 1
fi

# Get VS Code extensions directory
VSCODE_EXT_DIR="$HOME/.vscode/extensions"

# Create extensions directory if it doesn't exist
mkdir -p "$VSCODE_EXT_DIR"

# Copy extension
EXTENSION_NAME="sky-language-0.1.0"
SOURCE=".vscode/extensions/sky-language"
DEST="$VSCODE_EXT_DIR/$EXTENSION_NAME"

echo "📦 Copying extension to: $DEST"
rm -rf "$DEST"
cp -r "$SOURCE" "$DEST"

echo "✅ Extension installed!"
echo ""
echo "Next steps:"
echo "  1. Restart VS Code"
echo "  2. Open any .sky file"
echo "  3. Enjoy proper SKY syntax highlighting!"
echo ""
echo "Snippets available:"
echo "  - fn    → Function"
echo "  - afn   → Async function"
echo "  - class → Class"
echo "  - if    → If statement"
echo "  - for   → For loop"
echo "  - enum  → Enum"
echo "  - match → Match expression"
echo ""
echo "🎉 All done!"


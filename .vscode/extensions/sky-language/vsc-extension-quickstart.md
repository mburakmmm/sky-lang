# SKY Language Extension - Quick Start

## Development

This extension is located in `.vscode/extensions/sky-language/`

### Testing the Extension

1. Open VS Code in the sky-lang project
2. The extension is automatically loaded
3. Open any `.sky` file to test syntax highlighting

### Files

- `package.json` - Extension manifest
- `language-configuration.json` - Language behavior
- `syntaxes/sky.tmLanguage.json` - TextMate grammar for syntax highlighting
- `snippets/sky.json` - Code snippets
- `README.md` - Extension documentation
- `CHANGELOG.md` - Version history

### Publishing (Future)

```bash
# Install vsce
npm install -g vsce

# Package extension
vsce package

# Publish to marketplace
vsce publish
```

## More Information

- [VS Code Extension API](https://code.visualstudio.com/api)
- [Syntax Highlight Guide](https://code.visualstudio.com/api/language-extensions/syntax-highlight-guide)
- [Language Configuration Guide](https://code.visualstudio.com/api/language-extensions/language-configuration-guide)

**Enjoy!**


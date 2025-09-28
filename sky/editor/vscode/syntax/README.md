# Sky VSCode Extension

VSCode extension for Sky programming language providing syntax highlighting, snippets, and language support.

## Features

- **Syntax Highlighting**: Full syntax highlighting for Sky code
- **Code Snippets**: Pre-built snippets for common Sky patterns
- **Language Support**: Language configuration for Sky files
- **Indentation**: Proper indentation rules for Sky's Python-like syntax
- **Bracket Matching**: Auto-closing and surrounding pairs
- **Word Patterns**: Proper identifier recognition with Unicode support

## Installation

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/sky-lang/sky.git
   cd sky/sky/editor/vscode/syntax
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Build the extension:
   ```bash
   npm run build
   ```

4. Install in VSCode:
   ```bash
   code --install-extension sky-syntax-0.1.0.vsix
   ```

### Manual Installation

1. Copy the extension files to your VSCode extensions directory
2. Restart VSCode

## Usage

### Syntax Highlighting

The extension automatically highlights Sky files with the `.sky` extension:

```sky
# Comments are highlighted
int sayı = 42
string mesaj = "Merhaba Sky"
function selam(isim: string)
  return "Merhaba " + isim
```

### Code Snippets

Use the following snippets to quickly write Sky code:

- `var` - Variable declaration
- `int` - Integer variable
- `float` - Float variable
- `bool` - Boolean variable
- `string` - String variable
- `list` - List variable
- `map` - Map variable
- `function` - Function definition
- `async` - Async function
- `coop` - Coroutine function
- `if` - If statement
- `for` - For loop
- `while` - While loop
- `print` - Print statement
- `import` - Import statement
- `pyimport` - Python import
- `jseval` - JavaScript evaluation
- `sleep` - Async sleep
- `httpget` - HTTP GET request
- `resume` - Coroutine resume
- `done` - Coroutine done check

### Language Configuration

The extension provides:

- **Comments**: Line comments with `#`
- **Brackets**: Auto-closing pairs for `{}`, `[]`, `()`, `""`, `''`
- **Indentation**: Proper indentation rules for Sky's Python-like syntax
- **Word Patterns**: Unicode identifier support (Turkish characters)
- **Folding**: Code folding with `#region` and `#endregion`

## Configuration

### Settings

The extension respects VSCode's editor settings:

- `editor.tabSize` - Tab size for indentation
- `editor.insertSpaces` - Use spaces instead of tabs
- `editor.detectIndentation` - Auto-detect indentation
- `editor.wordWrap` - Word wrapping
- `editor.fontSize` - Font size
- `editor.fontFamily` - Font family

### Customization

You can customize the syntax highlighting by:

1. Opening VSCode settings
2. Searching for "workbench.colorCustomizations"
3. Adding custom colors for Sky tokens

Example:
```json
{
  "workbench.colorCustomizations": {
    "textMateRules": [
      {
        "scope": "keyword.control.sky",
        "settings": {
          "foreground": "#C678DD"
        }
      },
      {
        "scope": "storage.type.sky",
        "settings": {
          "foreground": "#E06C75"
        }
      }
    ]
  }
}
```

## Token Types

The extension provides the following token types:

### Keywords
- `keyword.control.sky` - Control flow keywords (`if`, `for`, `while`, etc.)
- `keyword.other.sky` - Other keywords (`var`, `function`, `async`, etc.)

### Types
- `storage.type.sky` - Type keywords (`int`, `float`, `bool`, etc.)

### Functions
- `entity.name.function.sky` - Function names
- `keyword.other.function.sky` - Function keywords

### Identifiers
- `variable.other.sky` - Variable names

### Operators
- `keyword.operator.arithmetic.sky` - Arithmetic operators
- `keyword.operator.comparison.sky` - Comparison operators
- `keyword.operator.logical.sky` - Logical operators
- `keyword.operator.assignment.sky` - Assignment operators
- `keyword.operator.unary.sky` - Unary operators

### Punctuation
- `punctuation.separator.comma.sky` - Commas
- `punctuation.separator.colon.sky` - Colons
- `punctuation.parenthesis.sky` - Parentheses
- `punctuation.bracket.sky` - Brackets
- `punctuation.brace.sky` - Braces

### Literals
- `constant.numeric.integer.sky` - Integer literals
- `constant.numeric.float.sky` - Float literals
- `string.quoted.double.sky` - String literals
- `string.quoted.single.sky` - Single-quoted strings

### Comments
- `comment.line.number-sign.sky` - Line comments
- `keyword.other.todo.sky` - TODO comments

## Development

### Building

```bash
npm install
npm run build
```

### Testing

```bash
npm test
```

### Linting

```bash
npm run lint
```

## Contributing

Contributions are welcome! Please see the [Contributing Guide](../../../CONTRIBUTING.md) for details.

### Adding New Snippets

1. Edit `snippets.json`
2. Add your snippet with a descriptive name
3. Test the snippet in VSCode
4. Submit a pull request

### Improving Syntax Highlighting

1. Edit `sky.tmLanguage.json`
2. Add or modify token patterns
3. Test the highlighting
4. Submit a pull request

## Issues

If you encounter any issues with the extension:

1. Check the [GitHub Issues](https://github.com/sky-lang/sky/issues)
2. Create a new issue with:
   - VSCode version
   - Extension version
   - Steps to reproduce
   - Expected behavior
   - Actual behavior

## License

This extension is licensed under the MIT License. See [LICENSE](../../../LICENSE) for details.

## Changelog

### 0.1.0
- Initial release
- Basic syntax highlighting
- Code snippets
- Language configuration
- Unicode identifier support
- Indentation rules

---

**Sky VSCode Extension** - Syntax highlighting and language support for Sky programming language.
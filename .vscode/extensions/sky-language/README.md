# SKY Language Support for VS Code

Official Visual Studio Code extension for the SKY Programming Language.

## Features

### ✨ Syntax Highlighting

Full syntax highlighting for SKY language with proper color coding for:

- **Keywords**: `function`, `class`, `enum`, `async`, `await`, `match`, `unsafe`, `coop`, `yield`
- **Control Flow**: `if`, `else`, `elif`, `for`, `while`, `break`, `continue`, `return`
- **Types**: `int`, `float`, `string`, `bool`, `any`
- **Constants**: `true`, `false`, `nil`
- **Operators**: `+`, `-`, `*`, `/`, `==`, `!=`, `=>`, etc.
- **Builtin Functions**: `print`, `len`, `range`, `int`, `str`, `float`, `bool`, `type`, etc.
- **Stdlib Functions**: `fs_read_text`, `crypto_sha256`, `json_encode`, etc.
- **Comments**: `#` line comments

### 📝 Code Snippets

Smart code snippets for common patterns:

| Prefix | Description |
|--------|-------------|
| `fn` | Function definition |
| `afn` | Async function |
| `class` | Class definition |
| `if` | If statement |
| `for` | For loop |
| `enum` | Enum definition |
| `match` | Match expression |

### 🎯 Smart Features

- **Auto-closing pairs**: `()`, `[]`, `{}`, `""`
- **Auto-indentation**: Smart indent after `function`, `if`, `class`, etc.
- **Comment toggling**: `Cmd+/` (Mac) or `Ctrl+/` (Win/Linux)
- **Bracket matching**: Highlights matching brackets

## Installation

### From Marketplace (Coming Soon)
```
Search for "SKY Language" in VS Code Extensions
```

### Manual Installation

1. Copy extension to VS Code extensions folder:
   ```bash
   cp -r .vscode/extensions/sky-language ~/.vscode/extensions/sky-language-0.1.0
   ```

2. Restart VS Code

3. Open any `.sky` file - syntax highlighting will be active!

Or use the installer script:
```bash
./scripts/install_vscode_extension.sh
```

## Usage

### Running SKY Code

**Method 1: Terminal**
```bash
sky run yourfile.sky
```

**Method 2: Custom Keybinding**

Add to `keybindings.json`:
```json
{
  "key": "cmd+shift+r",
  "command": "workbench.action.terminal.sendSequence",
  "args": { "text": "sky run ${file}\n" },
  "when": "editorLangId == sky"
}
```

### Using Snippets

1. Type snippet prefix (e.g., `fn`)
2. Press **Tab**
3. Fill in the placeholders

Example:
```
fn<Tab> → 

function name(params)
  [cursor here]
end
```

## Language Features

### Supported Syntax

```sky
# Variables
let x = 10
const PI = 3.14

# Functions
function add(a, b)
  return a + b
end

# Async
async function fetch()
  let data = await getData()
  return data
end

# Classes
class Person
  function init(name)
    self.name = name
  end
end

# Enums & Pattern Matching
enum Result
  Ok(int)
  Error(string)
end

match value
  Ok(x) => print(x)
  Error(e) => print(e)
end

# Unsafe blocks
unsafe
  let ptr = 0xFF00
end
```

## Recommended Themes

SKY syntax looks best with:

**Dark:**
- One Dark Pro ⭐
- Monokai Pro
- Dracula
- GitHub Dark

**Light:**
- One Light
- GitHub Light

## Known Issues

- Language Server (skyls) integration coming soon
- Debugger integration coming soon
- Auto-formatting coming soon

## Release Notes

### 0.1.0

Initial release:
- Full syntax highlighting
- Code snippets
- Auto-indentation
- Comment support

## Feedback

- **GitHub**: https://github.com/mburakmmm/sky-lang
- **Issues**: Report bugs or request features

## License

MIT License

---

**Enjoy coding with SKY! 🚀**


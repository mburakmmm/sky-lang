# 🎨 VS Code Setup for SKY Language

## 🚀 Quick Setup (Automatic)

```bash
cd /path/to/sky-lang
./scripts/setup_vscode.sh
```

Then restart VS Code!

---

## 📝 Manual Setup

### Option 1: Use SKY Extension (Recommended)

1. Copy the extension folder:
```bash
cp -r .vscode/extensions/sky-language ~/.vscode/extensions/
```

2. Restart VS Code

3. Open any `.sky` file - you'll see **proper SKY syntax highlighting!**

### Option 2: Simple Python-like Highlighting

Add to `.vscode/settings.json`:
```json
{
  "files.associations": {
    "*.sky": "python"
  }
}
```

---

## 🎨 What You Get

### ✅ Syntax Highlighting

**SKY-specific highlighting for:**
- Keywords: `function`, `class`, `enum`, `async`, `await`, `match`, `unsafe`
- Control flow: `if`, `else`, `for`, `while`, `break`, `continue`
- Types: `int`, `float`, `string`, `bool`
- Constants: `true`, `false`, `nil`
- Operators: `+`, `-`, `*`, `/`, `==`, `=>`, etc.
- Comments: `#`
- Builtin functions: `print`, `len`, `range`, `int`, `str`
- Stdlib functions: `fs_read_text`, `crypto_sha256`, etc.

### ✅ Code Snippets

Type and press **Tab**:

| Prefix | Expands to |
|--------|------------|
| `fn` | Function template |
| `afn` | Async function |
| `class` | Class template |
| `if` | If statement |
| `for` | For loop |
| `enum` | Enum definition |
| `match` | Match expression |

### ✅ Smart Features

- **Auto-closing**: Automatically closes `()`, `[]`, `{}`, `""`
- **Auto-indent**: Smart indentation after `function`, `if`, etc.
- **End detection**: Auto-dedent on `end`
- **Comment toggle**: Cmd/Ctrl+/ for `#` comments

---

## 🎯 Color Theme Recommendations

SKY syntax looks best with these themes:

**Dark Themes:**
- **One Dark Pro** (recommended)
- **Monokai Pro**
- **Dracula**
- **GitHub Dark**

**Light Themes:**
- **One Light**
- **GitHub Light**

---

## 🔧 Advanced Configuration

### Custom Keybindings

Add to `keybindings.json`:
```json
{
  "key": "cmd+shift+r",
  "command": "workbench.action.terminal.sendSequence",
  "args": {
    "text": "sky run ${file}\n"
  },
  "when": "editorLangId == sky"
}
```

Now **Cmd+Shift+R** runs the current SKY file!

### Task Runner

Create `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Run SKY",
      "type": "shell",
      "command": "sky",
      "args": ["run", "${file}"],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "presentation": {
        "reveal": "always",
        "panel": "new"
      }
    },
    {
      "label": "Check SKY",
      "type": "shell",
      "command": "sky",
      "args": ["check", "${file}"],
      "problemMatcher": []
    }
  ]
}
```

Now **Cmd+Shift+B** runs your SKY code!

---

## 🐛 Troubleshooting

### Syntax highlighting not working?

1. Restart VS Code completely
2. Check file association:
   ```
   Cmd+Shift+P → "Change Language Mode" → Select "SKY"
   ```
3. Verify extension is loaded:
   ```
   Cmd+Shift+P → "Developer: Show Running Extensions"
   Look for "sky-language"
   ```

### Snippets not appearing?

1. Type the prefix exactly: `fn` (not `Fn` or `FN`)
2. Press **Tab** (not Enter)
3. Check snippet settings: `Editor: Snippet Suggestions` should be "inline" or "top"

---

## 📖 More Information

- **Full extension**: `.vscode/extensions/sky-language/`
- **Grammar file**: `syntaxes/sky.tmLanguage.json`
- **Snippets**: `snippets/sky.json`

---

## 🎉 You're All Set!

Open any `.sky` file and enjoy:
- ✅ Proper SKY syntax highlighting
- ✅ Smart code snippets
- ✅ Auto-indentation
- ✅ Comment toggling

**Happy coding with SKY! 🚀**


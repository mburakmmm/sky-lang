# 📦 SKY Installation Guide

## 🚀 Quick Install (Recommended)

### macOS / Linux

```bash
# 1. Clone repository
git clone https://github.com/mburakmmm/sky-lang.git
cd sky-lang

# 2. Build
make build

# 3. Add to PATH
echo 'export PATH="$PATH:'$(pwd)'/bin"' >> ~/.zshrc
source ~/.zshrc

# 4. Verify
sky --version
wing --version
```

---

## 🛠️ Prerequisites

### Install Go (1.22+)

**macOS:**
```bash
brew install go
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install golang-1.22
```

**Check version:**
```bash
go version  # Should be 1.22 or higher
```

### Install LLVM (15+)

**macOS:**
```bash
brew install llvm@15
```

**Linux:**
```bash
sudo apt install llvm-15 llvm-15-dev libllvm15
```

---

## 📥 Installation Methods

### Method 1: System-Wide Install (Recommended)

```bash
# Build
make build

# Install to /usr/local/bin
sudo make install

# Now you can use 'sky' from anywhere
cd ~
sky --version
```

### Method 2: User Install (No sudo needed)

```bash
# Build
make build

# Create user bin directory
mkdir -p ~/bin

# Copy binaries
cp bin/* ~/bin/

# Add to PATH
echo 'export PATH="$PATH:$HOME/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Method 3: Project-Local (Development)

```bash
# Build
make build

# Use with full path
./bin/sky run examples/01_hello/hello.sky

# Or add alias
echo 'alias sky="'$(pwd)'/bin/sky"' >> ~/.zshrc
echo 'alias wing="'$(pwd)'/bin/wing"' >> ~/.zshrc
source ~/.zshrc
```

---

## 🧪 Verify Installation

```bash
# Check if sky is in PATH
which sky

# Check version
sky --version

# Run a test
sky run examples/01_hello/hello.sky

# Check wing
wing --version
```

Expected output:
```
sky version 0.1.0
wing version 0.1.0
Hello, SKY!
```

---

## 🎨 VS Code Setup

### 1. Install Recommended Extensions

Open VS Code in the `sky-lang` directory:
```bash
cd sky-lang
code .
```

### 2. Apply Workspace Settings

The repository includes:
- `.vscode/settings.json` - Syntax highlighting
- `.vscode/sky.code-snippets` - Code snippets

These are automatically applied when you open the project!

### 3. Test Syntax Highlighting

Open any `.sky` file - you should see Python-like syntax highlighting.

### 4. Use Snippets

Type these prefixes and press Tab:
- `fn` → Function
- `afn` → Async function
- `class` → Class
- `if` → If statement
- `for` → For loop
- `enum` → Enum
- `match` → Match expression

---

## 🔧 Troubleshooting

### "sky: command not found"

Check PATH:
```bash
echo $PATH | grep sky
```

If not there:
```bash
export PATH="$PATH:/path/to/sky-lang/bin"
```

### "LLVM not found" during build

**macOS:**
```bash
brew install llvm@15
export PATH="/opt/homebrew/opt/llvm@15/bin:$PATH"
export LDFLAGS="-L/opt/homebrew/opt/llvm@15/lib"
export CPPFLAGS="-I/opt/homebrew/opt/llvm@15/include"
```

**Linux:**
```bash
sudo apt install llvm-15-dev
```

### Permission denied

```bash
chmod +x bin/*
```

---

## 📚 Next Steps

1. **Read the docs:**
   - [API Reference (EN)](docs/API_REFERENCE_EN.md)
   - [API Reference (TR)](docs/API_REFERENCE_TR.md)

2. **Try examples:**
   ```bash
   sky run examples/01_hello/hello.sky
   sky run examples/08_stdlib/complete_demo.sky
   ```

3. **Create your first project:**
   ```bash
   wing init
   ```

4. **Join the community:**
   - GitHub: https://github.com/mburakmmm/sky-lang
   - Issues: Report bugs
   - Discussions: Ask questions

---

## 🚀 You're Ready!

```bash
# Create a new file
echo 'function main
  print("I installed SKY!")
end' > test.sky

# Run it
sky run test.sky
```

**Happy coding with SKY! 🌌**


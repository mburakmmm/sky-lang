# Wing Package Manager

Wing is the official package manager for the SKY programming language. It provides dependency management, package publishing, and registry integration using GitHub as the backend.

## Features

- **Package Management**: Install, update, and remove dependencies
- **GitHub Registry**: Uses GitHub releases as package registry
- **TOML Configuration**: Uses `sky.project.toml` for project configuration
- **Archive Creation**: Creates tar.gz packages for distribution
- **Build Integration**: Integrates with SKY compiler
- **Script Execution**: Run custom build and test scripts

## Installation

```bash
# Build from source
make build

# Install to system
make install
```

## Usage

### Initialize a Project

```bash
wing init my-project
```

This creates:
- `sky.project.toml` - Project manifest
- `src/main.sky` - Main source file
- `.gitignore` - Git ignore file

### Add Dependencies

```bash
# Add a dependency
wing add http --version 1.0.0

# Add dev dependency
wing add json --dev

# Add with latest version
wing add crypto
```

### Build Project

```bash
# Build project
wing build

# Build as package (creates tar.gz)
# Set target = "package" in sky.project.toml
wing build
```

### Publish Package

```bash
# Set GitHub token
export GITHUB_TOKEN=your_github_token

# Publish to GitHub registry
wing publish
```

### Other Commands

```bash
# List dependencies
wing list

# Search packages
wing search http

# Run scripts
wing run build
wing run test

# Clean build artifacts
wing clean
```

## Project Configuration

The `sky.project.toml` file defines your project:

```toml
[package]
name = "my-project"
version = "0.1.0"
description = "A SKY project"
authors = ["Developer"]
license = "MIT"
repository = "https://github.com/user/repo"
homepage = "https://example.com"
keywords = ["sky", "example"]

[dependencies]
http = "1.0.0"
json = "2.1.0"

[dev-dependencies]
test = "1.0.0"

[scripts]
build = "sky build"
test = "sky test"
run = "sky run src/main.sky"

[build]
target = "native"  # or "package"
optimization = "debug"
output-dir = "dist"
source-dir = "src"
```

## GitHub Registry

Wing uses GitHub as the package registry:

1. **Package Storage**: Packages are stored as GitHub releases
2. **Index Management**: `packages.json` tracks all packages
3. **Automated Updates**: GitHub Actions updates the index on new releases
4. **Download URLs**: Packages are downloaded from GitHub releases

### Registry Structure

```
wing-packages/
├── packages.json          # Package index
├── packages/              # Package manifests
│   ├── http/
│   │   └── sky.project.toml
│   └── json/
│       └── sky.project.toml
└── .github/
    └── workflows/
        └── update-registry.yml
```

## Development

### Building

```bash
# Build all binaries
make build

# Build specific binary
make build-sky
make build-wing

# Run tests
make test

# Lint code
make lint
```

### Adding New Features

1. Implement in `internal/wing/`
2. Add CLI commands in `cmd/wing/`
3. Update tests
4. Update documentation

## License

MIT License - see LICENSE file for details.

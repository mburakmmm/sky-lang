# Contributing to Sky

Thank you for your interest in contributing to Sky! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to [conduct@sky-lang.org](mailto:conduct@sky-lang.org).

## Getting Started

### Prerequisites

- Rust 1.79+
- Python 3.8+ (for Python bridge)
- Node.js (for JavaScript bridge)
- Git

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/sky.git
   cd sky
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/sky-lang/sky.git
   ```

## Development Setup

### Build from Source

```bash
git clone https://github.com/sky-lang/sky.git
cd sky
cargo build --release
```

### Run Tests

```bash
cargo test
```

### Run Examples

```bash
cargo run --release -- run sky/examples/hello.sky
```

### Development Tools

Install useful development tools:

```bash
# Format code
cargo install rustfmt

# Lint code
cargo install clippy

# Security audit
cargo install cargo-audit

# Coverage
cargo install cargo-tarpaulin
```

## Contributing Guidelines

### Areas for Contribution

- **Language Features**: New syntax, operators, built-in functions
- **Performance**: VM optimizations, GC improvements
- **Bridges**: Python/JS bridge enhancements
- **Tooling**: CLI improvements, better error messages
- **Documentation**: Examples, tutorials, specifications
- **Testing**: Unit tests, integration tests, benchmarks
- **Bug Fixes**: Bug reports and fixes

### Code Style

- Follow Rust conventions
- Use `cargo fmt` for formatting
- Use `cargo clippy` for linting
- Write comprehensive tests
- Document public APIs

### Commit Messages

Use conventional commits format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

Examples:
```
feat(parser): add support for list comprehensions
fix(vm): resolve memory leak in coroutine handling
docs(grammar): update coroutine documentation
```

## Pull Request Process

### Before Submitting

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes**:
   - Write code following the style guidelines
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**:
   ```bash
   cargo test
   cargo clippy
   cargo fmt --all -- --check
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat(parser): add support for list comprehensions"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/amazing-feature
   ```

### Pull Request Guidelines

- **Title**: Use conventional commit format
- **Description**: Clearly describe what the PR does
- **Tests**: Ensure all tests pass
- **Documentation**: Update docs if needed
- **Breaking Changes**: Clearly mark breaking changes

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

## Issue Reporting

### Bug Reports

Use the bug report template:

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Run command '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment**
- OS: [e.g. Ubuntu 20.04]
- Sky version: [e.g. 0.1.0]
- Rust version: [e.g. 1.79.0]

**Additional context**
Any other context about the problem.
```

### Feature Requests

Use the feature request template:

```markdown
**Is your feature request related to a problem?**
A clear description of what the problem is.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Alternative solutions or features you've considered.

**Additional context**
Any other context about the feature request.
```

## Development Workflow

### Branch Strategy

- `main`: Stable, production-ready code
- `develop`: Integration branch for features
- `feature/*`: Feature branches
- `bugfix/*`: Bug fix branches
- `hotfix/*`: Critical bug fixes

### Workflow

1. Create feature branch from `develop`
2. Make changes and commit
3. Push to your fork
4. Create pull request to `develop`
5. After review and merge, create release PR to `main`

## Testing

### Unit Tests

Write unit tests for new functionality:

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_feature() {
        // Test implementation
        assert_eq!(expected, actual);
    }
}
```

### Integration Tests

Add integration tests in `tests/` directory:

```rust
use sky::compiler::lexer::lex;

#[test]
fn test_lexer_integration() {
    let result = lex("int x = 42");
    assert!(result.is_ok());
}
```

### Golden Tests

For parser and formatter, use golden tests:

```rust
#[test]
fn test_parser_golden() {
    let input = include_str!("test.sky");
    let expected = include_str!("test.sky.expected");
    
    let result = parse(input);
    assert_eq!(result, expected);
}
```

### Performance Tests

Add benchmarks for performance-critical code:

```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn benchmark_lexer(c: &mut Criterion) {
    c.bench_function("lexer", |b| {
        b.iter(|| lex(black_box("int x = 42")))
    });
}

criterion_group!(benches, benchmark_lexer);
criterion_main!(benches);
```

## Documentation

### Code Documentation

Document public APIs:

```rust
/// Lexes the input string into tokens.
/// 
/// # Arguments
/// * `input` - The source code to lex
/// 
/// # Returns
/// * `Result<Vec<Token>, Diagnostic>` - Tokens or error
/// 
/// # Examples
/// ```
/// let tokens = lex("int x = 42")?;
/// ```
pub fn lex(input: &str) -> Result<Vec<Token>, Diagnostic> {
    // Implementation
}
```

### User Documentation

Update user-facing documentation:

- README.md
- Grammar specification
- Example programs
- Tutorial content

### API Documentation

Generate and review API docs:

```bash
cargo doc --open
```

## Release Process

### Version Bumping

Follow semantic versioning:

- **Major**: Breaking changes
- **Minor**: New features (backward compatible)
- **Patch**: Bug fixes (backward compatible)

### Release Checklist

- [ ] Update CHANGELOG.md
- [ ] Update version in Cargo.toml
- [ ] Run full test suite
- [ ] Update documentation
- [ ] Create release tag
- [ ] Build release artifacts
- [ ] Publish to GitHub Releases

### Release Commands

```bash
# Update version
cargo set-version 0.2.0

# Create release tag
git tag -a v0.2.0 -m "Release version 0.2.0"
git push origin v0.2.0

# Create GitHub release
gh release create v0.2.0 --title "Sky v0.2.0" --notes "Release notes"
```

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Discord**: Real-time chat and community support
- **Email**: [dev@sky-lang.org](mailto:dev@sky-lang.org)

### Resources

- [Sky Documentation](https://docs.sky-lang.org)
- [Rust Book](https://doc.rust-lang.org/book/)
- [Programming Language Implementation](https://craftinginterpreters.com/)

## Recognition

Contributors will be recognized in:

- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to Sky! 🚀

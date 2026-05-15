# Contributing to depsclean

Thanks for your interest in contributing to depsclean.

## Development setup

1. Fork the repository
2. Clone your fork: `git clone git@github.com:yourusername/depsclean.git`
3. Create a branch: `git checkout -b my-feature-branch`

## Building and testing

```bash
# Run the test suite
go test ./...

# Build the project
go build ./cmd/depsclean

# Format touched files
gofmt -w cmd/depsclean/main.go internal/
```

## Pull requests

1. Push your changes to your fork
2. Open a pull request against `main`
3. Keep the description focused on the user-visible change
4. Make sure tests pass locally
5. Update documentation when behavior changes

## Code style

- Use standard Go formatting with `gofmt`
- Prefer clear, small changes over large refactors
- Add tests for bug fixes and behavior changes

## Reporting bugs

Please include:

- Steps to reproduce
- Expected behavior
- Actual behavior
- OS, Go version, and how the binary was installed

## Feature requests

Feature requests are welcome. Include:

- The problem you are trying to solve
- The behavior you would like
- Any alternatives you considered

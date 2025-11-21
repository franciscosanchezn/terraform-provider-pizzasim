# Contributing to terraform-provider-pizzasim

Thank you for contributing! This is a learning project demonstrating Terraform provider development best practices.

## Development Setup

### Prerequisites

- Go >= 1.24, Terraform >= 1.5
- golangci-lint: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

### Quick Start

```bash
# Clone and setup
git clone https://github.com/YOUR_USERNAME/terraform-provider-pizzasim.git
cd terraform-provider-pizzasim
go mod download

# Build and install
go build -v .
go install

# Configure local development (optional)
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "franciscosanchezn/pizzasim" = "$HOME/go/bin"
  }
  direct {}
}
EOF
```

## Code Style

- **Format**: `gofmt -s -w .`
- **Lint**: `golangci-lint run ./...`
- **Document** all exported types/functions with godoc comments
- **Handle errors** with descriptive messages:

```go
resp.Diagnostics.AddError(
    "Unable to Create Pizza",
    fmt.Sprintf("Error: %s", err.Error()),
)
```

## Testing

```bash
# Unit tests
go test ./...

# Acceptance tests
export PIZZASIM_ENDPOINT="http://localhost:8080"
export PIZZASIM_CLIENT_ID="test-client"
export PIZZASIM_CLIENT_SECRET="test-secret"
export TF_ACC=1
go test -v ./internal/provider/
```

**Requirements**: All features need tests, use table-driven patterns, follow `TestAcc<Resource>_<scenario>` naming.

## Submitting Changes

### Branch Names
- `feat/feature-name`, `fix/bug-name`, `docs/topic`, `chore/task`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

Closes #<issue>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Process

1. Update fork: `git fetch upstream && git rebase upstream/main`
2. Create branch: `git checkout -b feat/my-feature`
3. Make changes, run tests and linter
4. Commit: `git commit -m "feat: description"`
5. Push and open PR with clear title and issue references

### PR Checklist

- [ ] Code formatted (`gofmt`)
- [ ] Tests pass locally
- [ ] New tests added
- [ ] Documentation updated
- [ ] No linting errors
- [ ] Commit messages follow convention
- [ ] CHANGELOG.md updated (significant changes)

## Code of Conduct

- Be respectful, collaborative, and constructive
- Focus on what's best for the project
- Keep discussions professional
- No harassment, discrimination, or trolling

## Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Plugin Development Tutorials](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework)
- [Effective Go](https://golang.org/doc/effective_go)

## Questions?

- General: [Discussions](https://github.com/franciscosanchezn/terraform-provider-pizzasim/discussions)
- Bugs/Features: [Issues](https://github.com/franciscosanchezn/terraform-provider-pizzasim/issues)

---

Thank you for contributing! 🍕

# Terraform Provider for PizzaSim

A Terraform provider for managing pizza resources in the PizzaSim API. This provider demonstrates OAuth2 authentication, CRUD operations, and importing resources using the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## What is PizzaSim?

PizzaSim is a simulated pizza management API that allows you to create, manage, and bake virtual pizzas. This Terraform provider enables infrastructure-as-code management of your pizza portfolio with features like:

- **Pizza Resource Management**: Create, update, and delete pizzas with various ingredients
- **OAuth2 Authentication**: Secure client credentials flow authentication
- **Import Support**: Import existing pizzas into Terraform state
- **Validation**: Built-in validation for pizza attributes (name length, ingredient count, price range)
- **Easter Egg**: Special handling for controversial ingredients 🍍

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- PizzaSim API credentials (client ID and secret)

## Using the Provider

### Installation

The provider will be available from the Terraform Registry. Add it to your `required_providers` block:

```terraform
terraform {
  required_providers {
    pizzasim = {
      source  = "franciscosanchezn/pizzasim"
      version = "~> 0.1"
    }
  }
}
```

### Configuration

Configure the provider with your PizzaSim API credentials:

```terraform
provider "pizzasim" {
  endpoint      = "https://pizza-api.example.com"
  client_id     = var.pizzasim_client_id
  client_secret = var.pizzasim_client_secret
}
```

Or use environment variables:

```bash
export PIZZASIM_ENDPOINT="https://pizza-api.example.com"
export PIZZASIM_CLIENT_ID="your-client-id"
export PIZZASIM_CLIENT_SECRET="your-client-secret"
```

### Example Usage

```terraform
# Create a classic Margherita pizza
resource "pizzasim_pizza" "margherita" {
  name        = "Margherita"
  ingredients = ["tomato sauce", "mozzarella", "basil"]
  price       = 12.99
  description = "Classic Italian pizza"
}

# Create a pepperoni pizza with default description
resource "pizzasim_pizza" "pepperoni" {
  name        = "Pepperoni"
  ingredients = ["tomato sauce", "mozzarella", "pepperoni"]
  price       = 14.99
}

# Output the last baked timestamp
output "margherita_last_baked" {
  value = pizzasim_pizza.margherita.last_baked
}
```

For more examples, see the [examples](./examples) directory.

## Building The Provider

1. Clone the repository:

```shell
git clone https://github.com/franciscosanchezn/terraform-provider-pizzasim.git
cd terraform-provider-pizzasim
```

2. Build the provider using the Go `install` command:

```shell
go install
```

This will build the provider and place it in your `$GOPATH/bin` directory.

## Developing the Provider

### Prerequisites

- [Go](http://www.golang.org) >= 1.24
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- Make (optional, for convenience commands)

### Local Development Setup

1. Build the provider:

```shell
go build -o terraform-provider-pizzasim
```

2. Create a local override configuration for Terraform to use your local build:

```shell
# Create the override file
cat > ~/.terraformrc << EOF
provider_installation {
  dev_overrides {
    "franciscosanchezn/pizzasim" = "$HOME/go/bin"
  }
  direct {}
}
EOF
```

3. Install the provider to your local Go bin:

```shell
go install
```

### Generating Documentation

Documentation is generated from the provider schema and examples:

```shell
cd tools
go generate
```

Or using make:

```shell
make generate
```

### Running Tests

#### Unit Tests

Run unit tests without acceptance tests:

```shell
go test ./... -v
```

#### Acceptance Tests

Acceptance tests create real resources against a PizzaSim API instance. Set up your environment:

```shell
export PIZZASIM_ENDPOINT="https://test-api.example.com"
export PIZZASIM_CLIENT_ID="test-client-id"
export PIZZASIM_CLIENT_SECRET="test-client-secret"
export TF_ACC=1
```

Run acceptance tests:

```shell
make testacc
```

Or run specific tests:

```shell
go test ./internal/provider -v -run TestAccPizzaResource
```

~> **Warning**: Acceptance tests create real resources and may incur costs or rate limits.

### Project Structure

```
.
├── internal/provider/      # Provider implementation
│   ├── provider.go        # Provider configuration and setup
│   ├── pizza_resource.go  # Pizza resource implementation
│   └── *_test.go          # Test files
├── examples/              # Example Terraform configurations
│   ├── provider/          # Provider configuration examples
│   ├── resources/         # Resource examples
│   └── data-sources/      # Data source examples
├── docs/                  # Generated documentation
│   ├── index.md          # Provider documentation
│   └── resources/        # Resource documentation
├── tools/                 # Code generation tools
└── go.mod                # Go module definition
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Generate documentation: `make generate`
6. Commit your changes: `git commit -am 'Add new feature'`
7. Push to the branch: `git push origin feature/my-feature`
8. Create a Pull Request

### Code Style

- Follow standard Go formatting (use `gofmt`)
- Add comments for exported functions and types
- Include unit tests for new functionality
- Update documentation when adding features

## Testing Guide

See [TESTING.md](./TESTING.md) for detailed testing instructions and troubleshooting.

## Releases

This project follows [Semantic Versioning](https://semver.org/): MAJOR.MINOR.PATCH

**Creating a Release** (automated via GitHub Actions):

1. Update `CHANGELOG.md`:
   ```markdown
   ## [0.2.0] - 2025-11-18
   ### Added
   - New feature
   ### Fixed
   - Bug fix
   ```

2. Create and push tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

GitHub Actions will automatically build multi-platform binaries, create checksums, sign with GPG, and publish the release.

**Prerequisites**: Configure GitHub secrets `GPG_PRIVATE_KEY` and `PASSPHRASE` (see `.tasks/manual-next-steps.md`)

**Local Testing**:
```bash
goreleaser check              # Validate config
goreleaser build --snapshot   # Test build
```

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](./LICENSE) file for details.

## Acknowledgments

Built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

## Support

For issues, questions, or contributions:

- Open an issue on [GitHub](https://github.com/franciscosanchezn/terraform-provider-pizzasim/issues)
- Review the [documentation](./docs)
- Check existing [examples](./examples)

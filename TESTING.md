# Testing Guide

Instructions for running tests in the terraform-provider-pizzasim project.

## Prerequisites

- Go 1.21+ (see `go.mod`)
- Terraform CLI 1.0+
- Make
- PizzaSim API (for acceptance tests only)

## Environment Variables

### Required for Acceptance Tests

| Variable | Description | Example |
|----------|-------------|---------|
| `TF_ACC` | Enable acceptance tests | `1` |
| `PIZZASIM_CLIENT_ID` | OAuth2 Client ID | `your-client-id` |
| `PIZZASIM_CLIENT_SECRET` | OAuth2 Client Secret | `your-client-secret` |
| `PIZZASIM_ENDPOINT` | API endpoint | `http://localhost:8080` |

### Optional

| Variable | Description |
|----------|-------------|
| `TF_LOG` | Terraform log level (`TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`) |

**Setup:**
```bash
# Linux/macOS
export TF_ACC=1
export PIZZASIM_CLIENT_ID="your-client-id"
export PIZZASIM_CLIENT_SECRET="your-client-secret"
export PIZZASIM_ENDPOINT="http://localhost:8080"

# Windows PowerShell
$env:TF_ACC = "1"
$env:PIZZASIM_CLIENT_ID = "your-client-id"
```

## Running Tests

**Unit tests only (no API needed):**
```bash
make test                    # All tests (skips acceptance if TF_ACC not set)
go test -v ./...            # Direct go test
```

**Acceptance tests (requires API):**
```bash
make testacc                # All acceptance tests
TF_ACC=1 go test -v ./internal/provider -run TestAcc
```

**Specific tests:**
```bash
go test -v ./internal/provider -run TestPizzaValidation
TF_ACC=1 go test -v ./internal/provider -run TestAccPizzaResource_basic
```

**With debug logging:**
```bash
TF_ACC=1 TF_LOG=TRACE go test -v ./internal/provider -run TestAcc
```

**Coverage:**
```bash
go test -v -cover ./...
```

## Test Types

**Acceptance Tests** (`TestAcc*`): End-to-end tests with real API. Requires `TF_ACC=1`, credentials, and running API.

**Unit Tests**: Validation logic without API calls. No special requirements.

**Edge Case Tests**: Mock server tests for error conditions (404, empty responses, concurrent operations).

**Provider Tests**: Provider configuration and error handling tests.

## Local API Setup

**Docker:**
```bash
docker run -d -p 8080:8080 \
  -e CLIENT_ID=test-client \
  -e CLIENT_SECRET=test-secret \
  pizzasim-api:latest

export PIZZASIM_ENDPOINT="http://localhost:8080"
export PIZZASIM_CLIENT_ID="test-client"
export PIZZASIM_CLIENT_SECRET="test-secret"
```

**API Endpoints:**
- `POST /api/v1/oauth/token` - OAuth2 token
- `POST /api/v1/pizzas` - Create
- `GET /api/v1/public/pizzas/{id}` - Read
- `PUT /api/v1/pizzas/{id}` - Update
- `DELETE /api/v1/pizzas/{id}` - Delete

## Troubleshooting

**Tests skipped:** Set `TF_ACC=1` and credentials.

**OAuth errors:** Verify credentials are correct and API is running at the endpoint.

**Certificate errors:** Add provider block with `insecure_skip_verify = true` for local development with self-signed certs.

**Timeouts:** Increase timeout: `go test -v -timeout 30m ./internal/provider`

**Connection refused:** Verify API is running: `curl http://localhost:8080/health`

**Pineapple errors:** Expected! The provider rejects pineapple as an easter egg.

**Debug logging:**
```bash
TF_ACC=1 TF_LOG=TRACE go test -v ./internal/provider -run TestAcc
```

**Test API connection:**
```bash
curl -X POST http://localhost:8080/api/v1/oauth/token \
  -d "grant_type=client_credentials" \
  -d "client_id=${PIZZASIM_CLIENT_ID}" \
  -d "client_secret=${PIZZASIM_CLIENT_SECRET}"
```

## CI/CD Example

```yaml
# .github/workflows/test.yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version-file: 'go.mod'
      - run: go test -v -cover ./...
      - name: Acceptance tests
        if: github.event_name == 'push'
        env:
          TF_ACC: 1
          PIZZASIM_CLIENT_ID: ${{ secrets.PIZZASIM_CLIENT_ID }}
          PIZZASIM_CLIENT_SECRET: ${{ secrets.PIZZASIM_CLIENT_SECRET }}
          PIZZASIM_ENDPOINT: ${{ secrets.PIZZASIM_ENDPOINT }}
        run: make testacc
```

## Resources

- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Plugin Framework Docs](https://developer.hashicorp.com/terraform/plugin/framework)

## 0.1.0 (Unreleased)

FEATURES:

* **New Resource**: `pizzasim_pizza` - Manage pizza resources with full CRUD operations
* **OAuth2 Authentication**: Client credentials flow for secure API authentication
* **Import Support**: Import existing pizzas into Terraform state using pizza name
* **Comprehensive Validation**: Built-in validation for pizza attributes including:
  - Name length (3-50 characters)
  - Ingredients count (1-20 items)
  - Price range ($0.01-$100.00)
* **Environment Variable Support**: Configure provider using `PIZZASIM_ENDPOINT`, `PIZZASIM_CLIENT_ID`, and `PIZZASIM_CLIENT_SECRET`
* **TLS Configuration**: Optional `insecure_skip_verify` for development environments
* **Auto-computed Attributes**: `last_baked` timestamp automatically updated on create/update
* **Easter Egg**: Special validation for controversial pizza ingredients 🍍

IMPROVEMENTS:

* Provider built on Terraform Plugin Framework for modern development experience
* Comprehensive documentation with examples
* Full test coverage including acceptance tests
* Retry logic with exponential backoff for API requests

NOTES:

* Initial release of the PizzaSim Terraform provider
* Requires PizzaSim API v1.0 or later
* Minimum Terraform version: 1.0

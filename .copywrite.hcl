schema_version = 1

project {
  license        = "MPL-2.0"
  copyright_year = 2025

  # Paths to exclude from license header checks
  header_ignore = [
    # Task tracking and documentation
    ".tasks/**",
    "META.d/**/*.yaml",
    
    # Examples used within documentation
    "examples/**",
    
    # Test data and fixtures
    "**/*_test.go",
    "testdata/**",
    
    # GitHub issue template configuration
    ".github/ISSUE_TEMPLATE/*.yml",
    
    # Tool configuration files
    ".golangci.yml",
    ".goreleaser.yml",
    ".copywrite.hcl",
    
    # Terraform files in test directories
    "tf-pizza-test/**",
    
    # Third-party or reference code
    "terraform-provider-scaffolding-framework/**",
    "tools/**",
    
    # Documentation
    "docs/**",
    "*.md",
  ]
}

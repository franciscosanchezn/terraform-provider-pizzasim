// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Ensure PizzaSimProvider satisfies various provider interfaces.
var _ provider.Provider = &PizzaSimProvider{}

// PizzaSimProvider defines the provider implementation.
type PizzaSimProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version  string
	client   *http.Client
	Endpoint string
}

// PizzaSimProviderModel describes the provider data model.
type PizzaSimProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	ClientId           types.String `tfsdk:"client_id"`
	ClientSecret       types.String `tfsdk:"client_secret"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func (p *PizzaSimProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pizzasim"
	resp.Version = p.version
}

func (p *PizzaSimProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "PizzaSim provider for managing pizza simulation resources.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The PizzaSim API endpoint URL. Can also be set via the PIZZASIM_ENDPOINT environment variable.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID for OAuth2 authentication. Can also be set via the PIZZASIM_CLIENT_ID environment variable.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret for OAuth2 authentication. Can also be set via the PIZZASIM_CLIENT_SECRET environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "If true, skips TLS certificate verification. Defaults to false.",
				Optional:            true,
			},
		},
	}
}

func (p *PizzaSimProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PizzaSimProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Support environment variable fallbacks
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("PIZZASIM_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = "https://pizza-api.local"
	}

	clientID := data.ClientId.ValueString()
	if clientID == "" {
		clientID = os.Getenv("PIZZASIM_CLIENT_ID")
	}

	clientSecret := data.ClientSecret.ValueString()
	if clientSecret == "" {
		clientSecret = os.Getenv("PIZZASIM_CLIENT_SECRET")
	}

	insecureSkipVerify := false
	if !data.InsecureSkipVerify.IsNull() {
		insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
	}

	tflog.Debug(ctx, "Configuring PizzaSim Provider", map[string]any{
		"endpoint":          endpoint,
		"has_client_id":     clientID != "",
		"has_client_secret": clientSecret != "",
	})

	// Validate required credentials
	if clientID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Client ID",
			"The provider requires a Client ID. Set it in the provider configuration or via the PIZZASIM_CLIENT_ID environment variable.",
		)
	}
	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Client Secret",
			"The provider requires a Client Secret. Set it in the provider configuration or via the PIZZASIM_CLIENT_SECRET environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	config := &Config{
		Endpoint:           endpoint,
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		InsecureSkipVerify: insecureSkipVerify,
	}

	httpClient, err := config.Client(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create PizzaSim Client",
			fmt.Sprintf("An error was encountered creating the PizzaSim client: %s", err.Error()),
		)
		return
	}

	// FIXED: Reuse the provider instance instead of creating a new one
	p.client = httpClient
	p.Endpoint = endpoint

	// Make client available to resources/data sources:
	resp.ResourceData = p
	resp.DataSourceData = p
}

// DataSources defines the data sources implemented in the provider.
func (p *PizzaSimProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *PizzaSimProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPizzaResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PizzaSimProvider{
			version: version,
		}
	}
}

type Config struct {
	Endpoint           string
	ClientID           string
	ClientSecret       string
	InsecureSkipVerify bool
}

func (c *Config) Client(ctx context.Context) (*http.Client, error) {
	if c.InsecureSkipVerify {
		tflog.Warn(ctx, "InsecureSkipVerify is enabled; TLS certificate verification will be skipped")
	}

	// HTTP transport with optional TLS config
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.InsecureSkipVerify,
		},
	}

	// OAuth2 client credentials configuration
	oauthConfig := clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     c.Endpoint + "/api/v1/oauth/token",
		// Scopes are optional - your API doesn't require them
		// Scopes: []string{},
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: httpTransport,
	})

	// Create token source (handles auto-refresh!)
	tokenSource := oauthConfig.TokenSource(ctx)

	// Verify we can get a token (checks credentials early)
	_, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to obtain OAuth2 token: %w", err)
	}

	// Create OAuth2 client (automatically adds Authorization: Bearer header)
	oauthClient := oauth2.NewClient(ctx, tokenSource)

	// Wrap with retryable HTTP client for resilience
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = oauthClient // Use oauth2 client underneath
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 5 * time.Second
	retryClient.RetryWaitMax = 15 * time.Second

	// Return standard http.Client
	return retryClient.StandardClient(), nil
}

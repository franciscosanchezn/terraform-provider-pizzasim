// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PizzaResource{}
var _ resource.ResourceWithImportState = &PizzaResource{}
var _ resource.ResourceWithConfigure = &PizzaResource{}

func NewPizzaResource() resource.Resource {
	return &PizzaResource{}
}

// PizzaResource defines the resource implementation.
type PizzaResource struct {
	client   *http.Client
	Endpoint string
}

// PizzaResourceModel describes the resource data model.
type PizzaResourceModel struct {
	Name        types.String  `tfsdk:"name"`
	Ingredients types.Set     `tfsdk:"ingredients"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	LastBaked   types.String  `tfsdk:"last_baked"`
	Id          types.String  `tfsdk:"id"`
}

func (r *PizzaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pizza"
}

func (r *PizzaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pizza resource for managing pizza simulation pizzas.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the pizza.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(50),
				},
			},
			"ingredients": schema.SetAttribute{
				MarkdownDescription: "List of ingredients on the pizza.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(20), // A pizza can have at most 20 ingredients
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the pizza.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Created by your Infra"),
			},
			"price": schema.Float64Attribute{
				MarkdownDescription: "Price of the pizza.",
				Required:            true,
				Validators: []validator.Float64{
					float64validator.AtLeast(0.01),  // Minimum price: 1 cent´
					float64validator.AtMost(100.00), // Maximum price: $100.00, more is too expensive!
				},
			},
			"last_baked": schema.StringAttribute{
				MarkdownDescription: "Timestamp of when the pizza was last baked.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PizzaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*PizzaSimProvider)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *PizzaSimProvider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = provider.client
	r.client.Timeout = 30 * time.Second // Set a timeout of 30 seconds for all requests
	r.Endpoint = provider.Endpoint
}

func (r *PizzaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PizzaResourceModel

	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// parse ingredients from Set to []string
	ingredients := func() []string {
		var ingredients []string
		for _, v := range data.Ingredients.Elements() {
			ingredients = append(ingredients, v.(types.String).ValueString())
		}
		return ingredients
	}()

	// Build the request body
	requestBodyData := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"ingredients": ingredients,
		"description": data.Description.ValueString(),
		"price":       data.Price.ValueFloat64(),
	}

	// easter egg if ingredients contains pineapple, run an error "Angry Italian will not cook this pizza!
	for _, ingredient := range ingredients {
		if strings.ToLower(ingredient) == "pineapple" {
			resp.Diagnostics.AddError("Invalid Ingredient", "Angry Italian will not cook this pizza with pineapple!")
			return
		}
	}

	requestBodyBytes, err := json.Marshal(requestBodyData)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("marshaling request body", data.Name.ValueString(), "", err)...)
		return
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/pizzas", r.Endpoint), bytes.NewReader(requestBodyBytes))
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("creating HTTP request", data.Name.ValueString(), "", err)...)
		return
	}

	// Add Content-Type header
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("creating pizza", data.Name.ValueString(), "", err)...)
		return
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := json.Marshal(httpResp.Body)
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("creating pizza", data.Name.ValueString(), "", httpResp.StatusCode, string(bodyBytes))...)
		return
	}

	// Check empty response body
	if httpResp.ContentLength == 0 {
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("creating pizza", data.Name.ValueString(), "", httpResp.StatusCode, "empty response body")...)
		return
	}

	// Parse response body
	var apiResponse struct {
		Id          int      `json:"id"`
		Name        string   `json:"name"`
		Ingredients []string `json:"ingredients"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		CreatedBy   int      `json:"created_by"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	}

	err = json.NewDecoder(httpResp.Body).Decode(&apiResponse)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("decoding response body", data.Name.ValueString(), "", err)...)
		return
	}
	// For the purposes of this example code, hardcoding a response value to
	data.Id = types.StringValue(fmt.Sprintf("%d", apiResponse.Id))
	data.LastBaked = types.StringValue(apiResponse.UpdatedAt)

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PizzaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PizzaResourceModel

	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/api/v1/public/pizzas/%s", r.Endpoint, data.Id.ValueString())
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("creating HTTP request", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}

	// Execute the request
	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("reading pizza", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}
	defer httpResp.Body.Close()

	// Check response status, if 404 then the resource has been deleted outside of Terraform
	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	} else if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := json.Marshal(httpResp.Body)
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("reading pizza", data.Name.ValueString(), data.Id.ValueString(), httpResp.StatusCode, string(bodyBytes))...)
		return
	}

	// Check empty response body
	if httpResp.ContentLength == 0 {
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("reading pizza", data.Name.ValueString(), data.Id.ValueString(), httpResp.StatusCode, "empty response body")...)
		return
	}

	// Parse response body
	var apiResponse struct {
		Id          int      `json:"id"`
		Name        string   `json:"name"`
		Ingredients []string `json:"ingredients"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		CreatedBy   int      `json:"created_by"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	}

	err = json.NewDecoder(httpResp.Body).Decode(&apiResponse)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("decoding response body", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}

	// Map API response to resource data model
	data.Name = types.StringValue(apiResponse.Name)
	data.Description = types.StringValue(apiResponse.Description)
	data.Price = types.Float64Value(apiResponse.Price)
	data.LastBaked = types.StringValue(apiResponse.UpdatedAt)

	// Convert ingredients slice to Set
	var ingredientValues []types.String
	for _, ingredient := range apiResponse.Ingredients {
		ingredientValues = append(ingredientValues, types.StringValue(ingredient))
	}

	var errDiags diag.Diagnostics
	data.Ingredients, errDiags = types.SetValueFrom(ctx, types.StringType, ingredientValues)
	if errDiags.HasError() {
		resp.Diagnostics.Append(errDiags...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PizzaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PizzaResourceModel

	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// parse ingredients from Set to []string
	ingredients := func() []string {
		var ingredients []string
		for _, v := range data.Ingredients.Elements() {
			ingredients = append(ingredients, v.(types.String).ValueString())
		}
		return ingredients
	}()

	// Build the request body
	requestBodyData := map[string]interface{}{
		"name":        data.Name.ValueString(),
		"description": data.Description.ValueString(),
		"ingredients": ingredients,
		"price":       data.Price.ValueFloat64(),
	}

	// easter egg if ingredients contains pineapple, run an error "Angry Italian will not cook this pizza!
	for _, ingredient := range ingredients {
		if strings.ToLower(ingredient) == "pineapple" {
			resp.Diagnostics.AddError("Invalid Ingredient", "Angry Italian will not cook this pizza with pineapple!")
			return
		}
	}

	requestBodyBytes, err := json.Marshal(requestBodyData)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("marshaling request body", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}
	// Create HTTP request
	requestURL := fmt.Sprintf("%s/api/v1/pizzas/%s", r.Endpoint, data.Id.ValueString())
	httpReq, err := http.NewRequest("PUT", requestURL, bytes.NewReader(requestBodyBytes))
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("creating HTTP request", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}

	// Add Content-Type header
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the request
	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("updating pizza", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := json.Marshal(httpResp.Body)
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("updating pizza", data.Name.ValueString(), data.Id.ValueString(), httpResp.StatusCode, string(bodyBytes))...)
		return
	}

	// Check empty response body
	if httpResp.ContentLength == 0 {
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("updating pizza", data.Name.ValueString(), data.Id.ValueString(), httpResp.StatusCode, "empty response body")...)
		return
	}

	// Parse response body
	var apiResponse struct {
		Id          int      `json:"id"`
		Name        string   `json:"name"`
		Ingredients []string `json:"ingredients"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		CreatedBy   int      `json:"created_by"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	}

	err = json.NewDecoder(httpResp.Body).Decode(&apiResponse)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("decoding response body", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}

	data.LastBaked = types.StringValue(apiResponse.UpdatedAt)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PizzaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PizzaResourceModel

	// Prevent panic if the provider has not been configured.
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured HTTP Client",
			"Expected configured HTTP client. Please report this issue to the provider developers.",
		)
		return
	}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create HTTP request
	requestURL := fmt.Sprintf("%s/api/v1/pizzas/%s", r.Endpoint, data.Id.ValueString())
	httpReq, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("creating HTTP request", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}

	// Execute the request
	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.Append(clientErrorDiagnosticsWithContext("deleting pizza", data.Name.ValueString(), data.Id.ValueString(), err)...)
		return
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		bodyBytes, _ := json.Marshal(httpResp.Body)
		resp.Diagnostics.Append(apiErrorDiagnosticsWithContext("deleting pizza", data.Name.ValueString(), data.Id.ValueString(), httpResp.StatusCode, string(bodyBytes))...)
		return
	}
}

func (r *PizzaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func clientErrorDiagnosticsWithContext(operation string, pizzaName string, pizzaID string, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	contextInfo := buildPizzaContext(pizzaName, pizzaID)
	diags.AddError(
		"Client Error",
		fmt.Sprintf("An unexpected error occurred during %s for pizza %s: %s", operation, contextInfo, err.Error()),
	)
	return diags
}

func apiErrorDiagnosticsWithContext(operation string, pizzaName string, pizzaID string, statusCode int, responseBody string) diag.Diagnostics {
	var diags diag.Diagnostics
	contextInfo := buildPizzaContext(pizzaName, pizzaID)
	diags.AddError(
		"API Error",
		fmt.Sprintf("API request failed during %s for pizza %s with status code %d. Response body: %s", operation, contextInfo, statusCode, responseBody),
	)
	return diags
}

func buildPizzaContext(pizzaName string, pizzaID string) string {
	if pizzaName != "" && pizzaID != "" {
		return fmt.Sprintf("'%s' (ID: %s)", pizzaName, pizzaID)
	} else if pizzaName != "" {
		return fmt.Sprintf("'%s'", pizzaName)
	} else if pizzaID != "" {
		return fmt.Sprintf("(ID: %s)", pizzaID)
	}
	return "(unknown)"
}

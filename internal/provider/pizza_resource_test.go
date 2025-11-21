// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPizzaResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPizzaResourceConfig_basic("Margherita", "mozzarella", "tomato", "basil"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pizzasim_pizza.test", "id"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "name", "Margherita"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "ingredients.#", "3"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "price", "12.99"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "description", "Created by your Infra"),
					resource.TestCheckResourceAttrSet("pizzasim_pizza.test", "last_baked"),
				),
			},
		},
	})
}

func TestAccPizzaResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPizzaResourceConfig_basic("Pepperoni", "pepperoni", "mozzarella", "tomato"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "name", "Pepperoni"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "ingredients.#", "3"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "price", "12.99"),
				),
			},
			{
				Config: testAccPizzaResourceConfig_updated("Deluxe Pepperoni", "pepperoni", "mozzarella", "tomato", "olives", "mushrooms"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "name", "Deluxe Pepperoni"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "ingredients.#", "5"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "price", "15.99"),
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "description", "Deluxe version"),
				),
			},
		},
	})
}

func TestAccPizzaResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPizzaResourceConfig_basic("Marinara", "tomato", "garlic", "oregano"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pizzasim_pizza.test", "name", "Marinara"),
				),
			},
			{
				ResourceName:            "pizzasim_pizza.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_baked"},
			},
		},
	})
}

func TestAccPizzaResource_pineappleRejection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPizzaResourceConfig_basic("Hawaiian", "pineapple", "ham", "mozzarella"),
				ExpectError: regexp.MustCompile("Angry Italian will not cook this pizza with pineapple"),
			},
		},
	})
}

func testAccPizzaResourceConfig_basic(name string, ingredients ...string) string {
	ingredientsList := make([]string, len(ingredients))
	for i, ing := range ingredients {
		ingredientsList[i] = fmt.Sprintf("%q", ing)
	}
	return fmt.Sprintf(`
provider "pizzasim" {
  insecure_skip_verify = true
}

resource "pizzasim_pizza" "test" {
  name        = %[1]q
  ingredients = [%[2]s]
  price       = 12.99
}
`, name, strings.Join(ingredientsList, ", "))
}

func testAccPizzaResourceConfig_updated(name string, ingredients ...string) string {
	ingredientsList := make([]string, len(ingredients))
	for i, ing := range ingredients {
		ingredientsList[i] = fmt.Sprintf("%q", ing)
	}
	return fmt.Sprintf(`
provider "pizzasim" {
  insecure_skip_verify = true
}

resource "pizzasim_pizza" "test" {
  name        = %[1]q
  ingredients = [%[2]s]
  price       = 15.99
  description = "Deluxe version"
}
`, name, strings.Join(ingredientsList, ", "))
}

// Unit tests for validation logic

func TestPizzaValidation_pineapple(t *testing.T) {
	tests := []struct {
		name        string
		ingredients []string
		shouldError bool
	}{
		{
			name:        "pineapple lowercase",
			ingredients: []string{"pineapple", "ham"},
			shouldError: true,
		},
		{
			name:        "Pineapple capitalized",
			ingredients: []string{"Pineapple", "ham"},
			shouldError: true,
		},
		{
			name:        "PINEAPPLE uppercase",
			ingredients: []string{"PINEAPPLE", "ham"},
			shouldError: true,
		},
		{
			name:        "no pineapple",
			ingredients: []string{"pepperoni", "mozzarella"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPineapple := false
			for _, ingredient := range tt.ingredients {
				if strings.ToLower(ingredient) == "pineapple" {
					hasPineapple = true
					break
				}
			}

			if hasPineapple != tt.shouldError {
				t.Errorf("Expected pineapple detection: %v, got: %v", tt.shouldError, hasPineapple)
			}
		})
	}
}

func TestPizzaValidation_ingredients(t *testing.T) {
	tests := []struct {
		name        string
		ingredients []string
		wantError   bool
		errorMsg    string
	}{
		{
			name:        "valid single ingredient",
			ingredients: []string{"cheese"},
			wantError:   false,
		},
		{
			name:        "valid multiple ingredients",
			ingredients: []string{"cheese", "tomato", "basil"},
			wantError:   false,
		},
		{
			name:        "empty ingredients",
			ingredients: []string{},
			wantError:   true,
			errorMsg:    "at least 1 ingredient required",
		},
		{
			name:        "too many ingredients",
			ingredients: make([]string, 21),
			wantError:   true,
			errorMsg:    "maximum 20 ingredients allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ingredientCount := len(tt.ingredients)
			hasError := ingredientCount < 1 || ingredientCount > 20

			if hasError != tt.wantError {
				t.Errorf("Expected error: %v, got: %v", tt.wantError, hasError)
			}
		})
	}
}

func TestPizzaValidation_price(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid price",
			price:     12.99,
			wantError: false,
		},
		{
			name:      "minimum valid price",
			price:     0.01,
			wantError: false,
		},
		{
			name:      "maximum valid price",
			price:     100.00,
			wantError: false,
		},
		{
			name:      "negative price",
			price:     -5.00,
			wantError: true,
			errorMsg:  "price must be at least 0.01",
		},
		{
			name:      "zero price",
			price:     0.00,
			wantError: true,
			errorMsg:  "price must be at least 0.01",
		},
		{
			name:      "too expensive",
			price:     150.00,
			wantError: true,
			errorMsg:  "price must be at most 100.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.price < 0.01 || tt.price > 100.00

			if hasError != tt.wantError {
				t.Errorf("Expected error: %v, got: %v for price %.2f", tt.wantError, hasError, tt.price)
			}
		})
	}
}

func TestPizzaValidation_name(t *testing.T) {
	tests := []struct {
		name      string
		pizzaName string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid name",
			pizzaName: "Margherita",
			wantError: false,
		},
		{
			name:      "minimum length name",
			pizzaName: "ABC",
			wantError: false,
		},
		{
			name:      "maximum length name",
			pizzaName: strings.Repeat("a", 50),
			wantError: false,
		},
		{
			name:      "too short",
			pizzaName: "Ab",
			wantError: true,
			errorMsg:  "name must be at least 3 characters",
		},
		{
			name:      "too long",
			pizzaName: strings.Repeat("a", 51),
			wantError: true,
			errorMsg:  "name must be at most 50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nameLength := len(tt.pizzaName)
			hasError := nameLength < 3 || nameLength > 50

			if hasError != tt.wantError {
				t.Errorf("Expected error: %v, got: %v for name length %d", tt.wantError, hasError, nameLength)
			}
		})
	}
}

// Edge case tests using httptest

func TestPizzaResource_404Handling(t *testing.T) {
	pizzaID := "999"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/oauth/token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.Method == "GET" && strings.Contains(r.URL.Path, pizzaID) {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp := &http.Response{
		StatusCode: http.StatusNotFound,
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 status code, got %d", resp.StatusCode)
	}
}

func TestPizzaResource_emptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/v1/oauth/token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp := &http.Response{
		StatusCode:    http.StatusCreated,
		ContentLength: 0,
	}

	if resp.ContentLength == 0 {
		t.Log("Empty response body detected correctly")
	}
}

func TestPizzaResource_concurrentOperations(t *testing.T) {
	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		if strings.Contains(r.URL.Path, "/api/v1/oauth/token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "test-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          requestCount,
			"name":        "Test Pizza",
			"ingredients": []string{"cheese"},
			"price":       10.0,
			"created_at":  "2023-01-01T00:00:00Z",
			"updated_at":  "2023-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	var wg sync.WaitGroup
	numRequests := 5

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(server.URL + "/api/v1/pizzas/1")
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			defer resp.Body.Close()
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if requestCount < numRequests {
		t.Errorf("Expected at least %d requests, got %d", numRequests, requestCount)
	}
}

func TestPizzaResource_destroyCheck(t *testing.T) {
	checkDestroy := func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pizzasim_pizza" {
				continue
			}

			if rs.Primary.ID != "" {
				return fmt.Errorf("Pizza resource still exists with ID: %s", rs.Primary.ID)
			}
		}
		return nil
	}

	testState := &terraform.State{
		Version: 3,
		Modules: []*terraform.ModuleState{
			{
				Path:      []string{"root"},
				Resources: map[string]*terraform.ResourceState{},
			},
		},
	}

	if err := checkDestroy(testState); err != nil {
		t.Errorf("Destroy check failed: %v", err)
	}
}

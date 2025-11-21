# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    pizzasim = {
      source = "hashicorp.com/franciscosanchezn/pizzasim"
    }
  }
}

provider "pizzasim" {
  endpoint      = "https://pizza-api.local"
  client_id     = var.pizzasim_client_id
  client_secret = var.pizzasim_client_secret
  insecure_skip_verify = true
}

# Once you add resources, you can test them here
# resource "pizzasim_pizza" "example" {
#   name = "Margherita"
# }

resource "pizzasim_pizza" "example" {
  name        = "Pepperoni Special"
  description = "A classic pepperoni pizza"
  price       = 15.99
  ingredients = ["Tomato Sauce", "Mozzarella", "Pepperoni", "Special"]
}

# Copyright (c) HashiCorp, Inc.

# Basic pizza resource
resource "pizzasim_pizza" "margherita" {
  name        = "Margherita"
  ingredients = ["tomato sauce", "mozzarella", "basil"]
  price       = 12.99
  description = "Classic Italian pizza"
}

# Pizza with custom description
resource "pizzasim_pizza" "pepperoni" {
  name        = "Pepperoni"
  ingredients = ["tomato sauce", "mozzarella", "pepperoni"]
  price       = 14.99
}

# Gourmet pizza with many ingredients
resource "pizzasim_pizza" "quattro_formaggi" {
  name = "Quattro Formaggi"
  ingredients = [
    "tomato sauce",
    "mozzarella",
    "gorgonzola",
    "parmesan",
    "ricotta"
  ]
  price       = 16.99
  description = "Four cheese pizza"
}

package model

import (
	"slices"
	"time"
)

// Category groups places by the kind of food they are ranked for.
type Category string

const (
	CategoryPizza     Category = "pizza"
	CategoryBurger    Category = "burger"
	CategorySushi     Category = "sushi"
	CategoryTacos     Category = "tacos"
	CategoryCoffee    Category = "coffee"
	CategoryDessert   Category = "dessert"
	CategoryBreakfast Category = "breakfast"
	CategoryOther     Category = "other"
)

// Categories is the closed list the UI offers when creating a place.
var Categories = []Category{
	CategoryPizza,
	CategoryBurger,
	CategorySushi,
	CategoryTacos,
	CategoryCoffee,
	CategoryDessert,
	CategoryBreakfast,
	CategoryOther,
}

// Valid reports whether the category is one the API accepts.
func (c Category) Valid() bool {
	return slices.Contains(Categories, c)
}

// Place is a restaurant pinned on the map.
type Place struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  Category  `json:"category"`
	Address   string    `json:"address"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Ratings   []Rating  `json:"ratings"`
	CreatedAt time.Time `json:"createdAt"`
}

// Score averages every rating left on the place. Zero means nobody rated it yet.
func (p Place) Score() float64 {
	if len(p.Ratings) == 0 {
		return 0
	}

	total := 0.0
	for _, rating := range p.Ratings {
		total += rating.Score
	}

	return total / float64(len(p.Ratings))
}

// Rating is one member of the duo scoring a place from 1 to 5.
type Rating struct {
	ID        string    `json:"id"`
	PlaceID   string    `json:"placeId"`
	Author    string    `json:"author"`
	Score     float64   `json:"score"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
}

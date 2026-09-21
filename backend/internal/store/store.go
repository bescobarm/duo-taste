package store

import (
	"context"
	"errors"

	"github.com/bescobarm/duo-taste/backend/internal/model"
)

// ErrNotFound is returned when a place does not exist.
var ErrNotFound = errors.New("place not found")

// PlaceStore is the persistence boundary. The in-memory implementation backs
// local development; a Postgres/PostGIS one will slot in behind it later.
type PlaceStore interface {
	List(ctx context.Context) ([]model.Place, error)
	Get(ctx context.Context, id string) (model.Place, error)
	Create(ctx context.Context, place model.Place) (model.Place, error)
	AddRating(ctx context.Context, placeID string, rating model.Rating) (model.Place, error)
}

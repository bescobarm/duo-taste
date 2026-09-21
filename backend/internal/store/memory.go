package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"

	"github.com/bescobarm/duo-taste/backend/internal/model"
)

// Memory keeps places in process memory. Everything is lost on restart, which
// is fine until the database lands.
type Memory struct {
	mu     sync.RWMutex
	places map[string]model.Place
}

// NewMemory builds an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{places: make(map[string]model.Place)}
}

// List returns every place, newest first.
func (m *Memory) List(ctx context.Context) ([]model.Place, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	places := make([]model.Place, 0, len(m.places))
	for _, place := range m.places {
		places = append(places, place)
	}

	sort.Slice(places, func(i, j int) bool {
		return places[i].CreatedAt.After(places[j].CreatedAt)
	})

	return places, nil
}

// Get returns one place by id.
func (m *Memory) Get(ctx context.Context, id string) (model.Place, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	place, ok := m.places[id]
	if !ok {
		return model.Place{}, ErrNotFound
	}

	return place, nil
}

// Create stores a new place, assigning its id and creation time.
func (m *Memory) Create(ctx context.Context, place model.Place) (model.Place, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	place.ID = NewID()
	place.CreatedAt = time.Now().UTC()
	if place.Ratings == nil {
		place.Ratings = []model.Rating{}
	}

	m.places[place.ID] = place

	return place, nil
}

// AddRating appends a rating to an existing place and returns the updated place.
func (m *Memory) AddRating(ctx context.Context, placeID string, rating model.Rating) (model.Place, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	place, ok := m.places[placeID]
	if !ok {
		return model.Place{}, ErrNotFound
	}

	rating.ID = NewID()
	rating.PlaceID = placeID
	rating.CreatedAt = time.Now().UTC()

	place.Ratings = append(place.Ratings, rating)
	m.places[placeID] = place

	return place, nil
}

// NewID returns a random 128-bit hex identifier.
func NewID() string {
	buf := make([]byte, 16)

	_, err := rand.Read(buf)
	if err != nil {
		// crypto/rand never fails on the platforms we target.
		panic(err)
	}

	return hex.EncodeToString(buf)
}

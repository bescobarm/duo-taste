package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/bescobarm/duo-taste/backend/internal/model"
	"github.com/bescobarm/duo-taste/backend/internal/store"
)

type placeHandler struct {
	places store.PlaceStore
}

type createPlaceRequest struct {
	Name     string         `json:"name"`
	Category model.Category `json:"category"`
	Address  string         `json:"address"`
	Lat      float64        `json:"lat"`
	Lng      float64        `json:"lng"`
}

type createRatingRequest struct {
	Author string  `json:"author"`
	Score  float64 `json:"score"`
	Notes  string  `json:"notes"`
}

// placeResponse is a place plus the average score, so the map can color pins
// without recomputing it on the client.
type placeResponse struct {
	model.Place
	Score float64 `json:"score"`
}

func newPlaceResponse(place model.Place) placeResponse {
	return placeResponse{Place: place, Score: place.Score()}
}

func (h placeHandler) list(w http.ResponseWriter, r *http.Request) {
	places, err := h.places.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list places")
		return
	}

	category := model.Category(r.URL.Query().Get("category"))

	response := make([]placeResponse, 0, len(places))
	for _, place := range places {
		if category != "" && place.Category != category {
			continue
		}

		response = append(response, newPlaceResponse(place))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h placeHandler) get(w http.ResponseWriter, r *http.Request) {
	place, err := h.places.Get(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "place not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read place")
		return
	}

	writeJSON(w, http.StatusOK, newPlaceResponse(place))
}

func (h placeHandler) create(w http.ResponseWriter, r *http.Request) {
	var body createPlaceRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if !body.Category.Valid() {
		writeError(w, http.StatusBadRequest, "unknown category")
		return
	}
	if body.Lat < -90 || body.Lat > 90 {
		writeError(w, http.StatusBadRequest, "lat must be between -90 and 90")
		return
	}
	if body.Lng < -180 || body.Lng > 180 {
		writeError(w, http.StatusBadRequest, "lng must be between -180 and 180")
		return
	}

	place, err := h.places.Create(r.Context(), model.Place{
		Name:     body.Name,
		Category: body.Category,
		Address:  strings.TrimSpace(body.Address),
		Lat:      body.Lat,
		Lng:      body.Lng,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create place")
		return
	}

	writeJSON(w, http.StatusCreated, newPlaceResponse(place))
}

func (h placeHandler) addRating(w http.ResponseWriter, r *http.Request) {
	var body createRatingRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	body.Author = strings.TrimSpace(body.Author)
	if body.Author == "" {
		writeError(w, http.StatusBadRequest, "author is required")
		return
	}
	if body.Score < 1 || body.Score > 5 {
		writeError(w, http.StatusBadRequest, "score must be between 1 and 5")
		return
	}

	place, err := h.places.AddRating(r.Context(), r.PathValue("id"), model.Rating{
		Author: body.Author,
		Score:  body.Score,
		Notes:  strings.TrimSpace(body.Notes),
	})
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "place not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not add rating")
		return
	}

	writeJSON(w, http.StatusCreated, newPlaceResponse(place))
}

func listCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.Categories)
}

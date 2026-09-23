package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bescobarm/duo-taste/backend/internal/model"
	"github.com/bescobarm/duo-taste/backend/internal/store"
)

var (
	errInvalidJSON = errors.New("invalid JSON body")

	errInvalidPlace    = errors.New("invalid place")
	errMissingName     = errors.New("name is required")
	errUnknownCategory = errors.New("unknown category")
	errInvalidLat      = errors.New("lat must be between -90 and 90")
	errInvalidLng      = errors.New("lng must be between -180 and 180")
	errMissingRating   = errors.New("rating is required")

	errInvalidRating = errors.New("invalid rating")
	errMissingAuthor = errors.New("author is required")
	errInvalidScore  = errors.New("score must be between 1 and 5")
)

// errorResponse is the JSON body of every failed request.
type errorResponse struct {
	Error string `json:"error"`
}

type placeHandler struct {
	places store.PlaceStore
	logger *slog.Logger
}

type createPlaceRequest struct {
	Name     string         `json:"name"`
	Category model.Category `json:"category"`
	Address  string         `json:"address"`
	Lat      float64        `json:"lat"`
	Lng      float64        `json:"lng"`
	Rating   *Rating        `json:"rating"`
}

// Rating is a single score for a place, with an author and optional notes.
type Rating struct {
	Author string  `json:"author"`
	Score  float64 `json:"score"`
	Notes  string  `json:"notes,omitempty"`
}

// toPlace validates the request and returns the place it describes, first
// rating included. Every error wraps errInvalidPlace or errInvalidRating.
func (body createPlaceRequest) toPlace() (model.Place, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return model.Place{}, fmt.Errorf("%w: %w", errInvalidPlace, errMissingName)
	}

	if !body.Category.Valid() {
		return model.Place{}, fmt.Errorf("%w: %w", errInvalidPlace, errUnknownCategory)
	}

	if body.Lat < -90 || body.Lat > 90 {
		return model.Place{}, fmt.Errorf("%w: %w", errInvalidPlace, errInvalidLat)
	}

	if body.Lng < -180 || body.Lng > 180 {
		return model.Place{}, fmt.Errorf("%w: %w", errInvalidPlace, errInvalidLng)
	}

	if body.Rating == nil {
		return model.Place{}, fmt.Errorf("%w: %w", errInvalidPlace, errMissingRating)
	}

	rating, err := body.Rating.toRating()
	if err != nil {
		return model.Place{}, err
	}

	return model.Place{
		Name:     name,
		Category: body.Category,
		Address:  strings.TrimSpace(body.Address),
		Lat:      body.Lat,
		Lng:      body.Lng,
		Ratings:  []model.Rating{rating},
	}, nil
}

// toRating validates the request and returns the rating it describes. Scores
// are any real number from 1 to 5, such as 4.82. Every error wraps
// errInvalidRating.
func (body Rating) toRating() (model.Rating, error) {
	author := strings.TrimSpace(body.Author)
	if author == "" {
		return model.Rating{}, fmt.Errorf("%w: %w", errInvalidRating, errMissingAuthor)
	}

	if body.Score < 1 || body.Score > 5 {
		return model.Rating{}, fmt.Errorf("%w: %w", errInvalidRating, errInvalidScore)
	}

	return model.Rating{Author: author, Score: body.Score, Notes: strings.TrimSpace(body.Notes)}, nil
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
		h.logger.Error("list places", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not list places"})
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
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	case err != nil:
		h.logger.Error("get place", "id", r.PathValue("id"), "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not read place"})
		return
	}

	writeJSON(w, http.StatusOK, newPlaceResponse(place))
}

func (h placeHandler) create(w http.ResponseWriter, r *http.Request) {
	var body createPlaceRequest

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errInvalidJSON.Error()})
		return
	}

	// The message carries both the group and the reason, such as
	// "invalid rating: author is required".
	place, err := body.toPlace()
	switch {
	case errors.Is(err, errInvalidPlace), errors.Is(err, errInvalidRating):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	case err != nil:
		h.logger.Error("validate place", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not create place"})
		return
	}

	place, err = h.places.Create(r.Context(), place)
	if err != nil {
		h.logger.Error("create place", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not create place"})
		return
	}

	writeJSON(w, http.StatusCreated, newPlaceResponse(place))
}

func (h placeHandler) addRating(w http.ResponseWriter, r *http.Request) {
	var body Rating

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: errInvalidJSON.Error()})
		return
	}

	rating, err := body.toRating()
	switch {
	case errors.Is(err, errInvalidRating):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	case err != nil:
		h.logger.Error("validate rating", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not add rating"})
		return
	}

	place, err := h.places.AddRating(r.Context(), r.PathValue("id"), rating)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
		return
	case err != nil:
		h.logger.Error("add rating", "placeId", r.PathValue("id"), "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "could not add rating"})
		return
	}

	writeJSON(w, http.StatusCreated, newPlaceResponse(place))
}

func listCategories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, model.Categories)
}

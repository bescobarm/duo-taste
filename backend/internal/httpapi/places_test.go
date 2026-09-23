package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bescobarm/duo-taste/backend/internal/config"
	"github.com/bescobarm/duo-taste/backend/internal/store"
)

func validPlaceRequest() createPlaceRequest {
	return createPlaceRequest{
		Name:     "Pizzeria",
		Category: "pizza",
		Lat:      4.6,
		Lng:      -74,
		Rating:   &Rating{Author: "Ana", Score: 4.5},
	}
}

// send runs one request through a router backed by an empty store and returns
// the status code and the decoded JSON body.
func send(t *testing.T, method, path, body string) (int, map[string]any) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(store.NewMemory(), config.Config{}, logger)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, path, strings.NewReader(body)))

	var payload map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	require.NoError(t, err)

	return recorder.Code, payload
}

func TestToRatingAcceptsRealScore(t *testing.T) {
	c := require.New(t)

	rating, err := Rating{Author: " Ana ", Score: 4.82, Notes: " crisp "}.toRating()

	c.NoError(err)
	c.Equal("Ana", rating.Author)
	c.Equal(4.82, rating.Score)
	c.Equal("crisp", rating.Notes)
}

func TestToRatingRejectsBlankAuthor(t *testing.T) {
	c := require.New(t)

	_, err := Rating{Author: "  ", Score: 4}.toRating()

	c.ErrorIs(err, errMissingAuthor)
	c.ErrorIs(err, errInvalidRating)
}

func TestToRatingRejectsScoreBelowOne(t *testing.T) {
	c := require.New(t)

	_, err := Rating{Author: "Ana", Score: 0.99}.toRating()

	c.ErrorIs(err, errInvalidScore)
	c.ErrorIs(err, errInvalidRating)
}

func TestToRatingRejectsScoreAboveFive(t *testing.T) {
	c := require.New(t)

	_, err := Rating{Author: "Ana", Score: 5.01}.toRating()

	c.ErrorIs(err, errInvalidScore)
	c.ErrorIs(err, errInvalidRating)
}

func TestToPlaceAttachesFirstRating(t *testing.T) {
	c := require.New(t)

	place, err := validPlaceRequest().toPlace()

	c.NoError(err)
	c.Equal("Pizzeria", place.Name)
	c.Len(place.Ratings, 1)
	c.Equal("Ana", place.Ratings[0].Author)
	c.Equal(4.5, place.Ratings[0].Score)
}

func TestToPlaceRejectsBlankName(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Name = " "

	_, err := body.toPlace()

	c.ErrorIs(err, errMissingName)
	c.ErrorIs(err, errInvalidPlace)
}

func TestToPlaceRejectsUnknownCategory(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Category = "soup"

	_, err := body.toPlace()

	c.ErrorIs(err, errUnknownCategory)
	c.ErrorIs(err, errInvalidPlace)
}

func TestToPlaceRejectsLatOutOfRange(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Lat = 91

	_, err := body.toPlace()

	c.ErrorIs(err, errInvalidLat)
	c.ErrorIs(err, errInvalidPlace)
}

func TestToPlaceRejectsLngOutOfRange(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Lng = -181

	_, err := body.toPlace()

	c.ErrorIs(err, errInvalidLng)
	c.ErrorIs(err, errInvalidPlace)
}

func TestToPlaceRejectsMissingRating(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Rating = nil

	_, err := body.toPlace()

	c.ErrorIs(err, errMissingRating)
	c.ErrorIs(err, errInvalidPlace)
}

func TestToPlaceRejectsInvalidFirstRating(t *testing.T) {
	c := require.New(t)
	body := validPlaceRequest()
	body.Rating = &Rating{Author: "Ana", Score: 6}

	_, err := body.toPlace()

	c.ErrorIs(err, errInvalidScore)
	c.ErrorIs(err, errInvalidRating)
}

func TestCreatePlaceReturnsCreated(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places",
		`{"name":"A","category":"pizza","lat":4.6,"lng":-74,"rating":{"author":"Ana","score":4.82}}`)

	c.Equal(http.StatusCreated, status)
	c.Equal(4.82, payload["score"])
	c.Len(payload["ratings"], 1)
}

func TestCreatePlaceRejectsMalformedJSON(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places", "{")

	c.Equal(http.StatusBadRequest, status)
	c.Equal("invalid JSON body", payload["error"])
}

func TestCreatePlaceRejectsMissingRating(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places",
		`{"name":"A","category":"pizza","lat":4.6,"lng":-74}`)

	c.Equal(http.StatusBadRequest, status)
	c.Equal("invalid place: rating is required", payload["error"])
}

func TestCreatePlaceRejectsInvalidScore(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places",
		`{"name":"A","category":"pizza","lat":4.6,"lng":-74,"rating":{"author":"Ana","score":5.2}}`)

	c.Equal(http.StatusBadRequest, status)
	c.Equal("invalid rating: score must be between 1 and 5", payload["error"])
}

func TestGetPlaceReturnsNotFound(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodGet, "/api/places/nope", "")

	c.Equal(http.StatusNotFound, status)
	c.Equal("place not found", payload["error"])
}

func TestAddRatingRejectsMalformedJSON(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places/nope/ratings", "{")

	c.Equal(http.StatusBadRequest, status)
	c.Equal("invalid JSON body", payload["error"])
}

func TestAddRatingRejectsBlankAuthor(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places/nope/ratings", `{"author":" ","score":4}`)

	c.Equal(http.StatusBadRequest, status)
	c.Equal("invalid rating: author is required", payload["error"])
}

func TestAddRatingReturnsNotFoundForUnknownPlace(t *testing.T) {
	c := require.New(t)

	status, payload := send(t, http.MethodPost, "/api/places/nope/ratings", `{"author":"Ana","score":4}`)

	c.Equal(http.StatusNotFound, status)
	c.Equal("place not found", payload["error"])
}

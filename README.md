# DuoTaste

A close app to rank restaurants and food by category and leave a pin in a map.

## Stack

| Layer | Choice | Why |
| --- | --- | --- |
| API | Go 1.23, standard library only | `net/http` routing patterns cover the whole API; no framework needed yet |
| Web | Vue 3 + TypeScript + Vite | |
| Map | Leaflet + OpenStreetMap tiles | free, no API key, no billing account |
| Design | DuoTaste design system v2 | tokens mirrored into `src/tokens.css` from the Claude Design canvas |
| Storage | in-memory | swap for Postgres/PostGIS behind `store.PlaceStore` |

## Layout

```
backend/
  cmd/api/            entry point, graceful shutdown
  internal/config/    env-based configuration
  internal/model/     Place, Rating, Category
  internal/store/     PlaceStore interface + in-memory implementation
  internal/httpapi/   router, handlers, CORS and request logging
frontend/
  src/api.ts          typed fetch client
  src/types.ts        shared DTOs
  src/tokens.css      design tokens: colour (light + dark), type, spacing, radii
  src/categories.ts   category -> icon, and the score bands
  src/components/     IconSprite, AppIcon, BrandMark, MapView
  src/App.vue         sidebar: ranking, add place, place detail
```

## Design

The UI follows the DuoTaste canvas at
https://claude.ai/artifact/GXynMhVW6iZxS3bHDa4Dmd. Its design system is
mirrored into `src/tokens.css`; every colour in the app resolves to a token
there, so a palette change is an edit to that one file. Light and dark values
both come from the system and switch on `prefers-color-scheme`.

It covers two breakpoints. The desktop artboards are 1280 wide; the phone
artboards are 390 x 844. Below 900px the sidebar stops being a column and
becomes a sheet over the map: the ranking peeks at 300px and expands from the
grabber, a panel (add place, place detail) takes the whole sheet, the filter
chips scroll in one row instead of wrapping, and the toolbar's Add place button
becomes a floating action button. Controls grow to 48-52px for touch, and under
380px the category grid drops from four columns to three.

Two rules from the canvas that shape the code:

- **Score bands.** 4.2 and up is teal, 3.2 to 4.1 amber, below 3.2 rust,
  unrated is a plain outline. The numeral is printed on every pin, pill and
  row, so colour is never the only signal (`bandFor` in `src/categories.ts`).
- **No emoji in product surfaces.** Categories are stroked 24px icons from the
  sprite in `IconSprite.vue`.

## Running it

Two processes. The Vite dev server proxies `/api` and `/health` to the Go API,
so the browser only ever talks to `localhost:5173`.

```
make api   # http://localhost:8080
make web   # http://localhost:5173
```

## API

| Method | Path | Body / notes |
| --- | --- | --- |
| GET | `/health` | |
| GET | `/api/categories` | the closed category list |
| GET | `/api/places` | optional `?category=pizza` |
| POST | `/api/places` | `{name, category, address, lat, lng}` |
| GET | `/api/places/{id}` | |
| POST | `/api/places/{id}/ratings` | `{author, score (1-5), notes}` |

Places come back with a computed `score`, the average of their ratings (`0` when unrated).

## Configuration

| Variable | Default |
| --- | --- |
| `DUOTASTE_ADDR` | `:8080` |
| `DUOTASTE_ALLOW_ORIGIN` | `http://localhost:5173` |

## Not built yet

Persistence, auth / accounts for the duo, photos, search and geocoding of addresses.

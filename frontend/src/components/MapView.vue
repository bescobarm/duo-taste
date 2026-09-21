<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { bandFor, iconFor } from '../categories'
import type { Place } from '../types'

const props = defineProps<{
  places: Place[]
  draft: { lat: number; lng: number } | null
  selectedId: string | null
}>()

const emit = defineEmits<{
  (event: 'pick', coords: { lat: number; lng: number }): void
  (event: 'select', place: Place): void
}>()

const container = ref<HTMLDivElement | null>(null)
let map: L.Map | null = null
let markers: L.LayerGroup | null = null
let draftMarker: L.Marker | null = null

// A pin is the score pill from the canvas: category icon plus the numeral, on
// a fill from the score band. The numeral is always printed, so the band color
// is never the only signal.
function pinIcon(place: Place, selected: boolean) {
  const band = bandFor(place.score)
  const label = place.score ? place.score.toFixed(1) : '—'

  return L.divIcon({
    className: 'pin',
    html: `<span class="pin-body pin-${band}${selected ? ' is-selected' : ''}">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <use href="#${iconFor(place.category)}"></use>
      </svg><span class="pin-score">${label}</span></span>`,
    iconSize: [64, 36],
    iconAnchor: [32, 18],
  })
}

let framed = false

// The map opens on Chapinero, as the design frames it. Once places load, fit
// the view to them once, so pins elsewhere are not left off-screen.
function frameToPlaces() {
  if (!map || framed || !props.places.length) return

  const bounds = L.latLngBounds(props.places.map((place) => [place.lat, place.lng]))
  map.fitBounds(bounds, { padding: [80, 80], maxZoom: 16 })
  framed = true
}

function renderPlaces() {
  if (!map || !markers) return

  markers.clearLayers()

  for (const place of props.places) {
    const marker = L.marker([place.lat, place.lng], {
      icon: pinIcon(place, place.id === props.selectedId),
      title: place.name,
      alt: place.score
        ? `${place.name}, rated ${place.score.toFixed(1)} out of 5`
        : `${place.name}, not rated yet`,
    })

    marker.on('click', () => emit('select', place))
    markers.addLayer(marker)
  }

  frameToPlaces()
}

function renderDraft() {
  if (!map) return

  if (draftMarker) {
    map.removeLayer(draftMarker)
    draftMarker = null
  }

  if (!props.draft) return

  draftMarker = L.marker([props.draft.lat, props.draft.lng], {
    icon: L.divIcon({
      className: 'pin',
      html: `<span class="pin-body pin-draft">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <use href="#ic-pin"></use>
        </svg><span class="pin-score">New</span></span>`,
      iconSize: [72, 36],
      iconAnchor: [36, 18],
    }),
  })
  draftMarker.addTo(map)
}

onMounted(() => {
  if (!container.value) return

  map = L.map(container.value, { zoomControl: false }).setView([4.6655, -74.0578], 15)
  L.control.zoom({ position: 'topright' }).addTo(map)

  // OpenStreetMap raster tiles: free and keyless.
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  markers = L.layerGroup().addTo(map)
  map.on('click', (event: L.LeafletMouseEvent) => {
    emit('pick', { lat: event.latlng.lat, lng: event.latlng.lng })
  })

  renderPlaces()
})

onUnmounted(() => {
  map?.remove()
  map = null
})

// Lets the toolbar's Add place button drop a draft pin where the user is looking.
defineExpose({
  center: () => {
    const point = map?.getCenter()

    return point ? { lat: point.lat, lng: point.lng } : null
  },
})

watch(() => props.places, renderPlaces, { deep: true })
watch(() => props.selectedId, renderPlaces)
watch(() => props.draft, renderDraft, { deep: true })
</script>

<template>
  <div ref="container" class="map"></div>
</template>

<style>
.map {
  width: 100%;
  height: 100%;
  background: var(--map-canvas);
}

.pin .pin-body {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  height: 36px;
  padding: 0 var(--space-2);
  border: 2px solid var(--surface-100);
  border-radius: var(--radius-pill);
  box-shadow: 0 1px 4px rgb(11 43 48 / 25%);
}

.pin .pin-score {
  font-size: 14px;
  font-weight: 700;
}

.pin .is-selected {
  outline: 2px solid var(--focus-ring);
  outline-offset: 2px;
}

.pin .pin-high {
  color: var(--on-teal);
  background: var(--score-high);
}

.pin .pin-mid {
  color: var(--on-amber);
  background: var(--score-mid);
}

.pin .pin-low {
  color: var(--on-rust);
  background: var(--score-low);
}

.pin .pin-none {
  color: var(--ink-900);
  background: var(--surface-100);
  border-color: var(--border-200);
}

.pin .pin-draft {
  color: var(--on-cyan);
  background: var(--pin-saved);
}

/* Leaflet's own controls, restyled to the design system. */
.leaflet-touch .leaflet-bar,
.leaflet-bar {
  border: 1px solid var(--border-200);
  border-radius: var(--radius-md);
  box-shadow: none;
}

.leaflet-bar a,
.leaflet-bar a:hover {
  width: 44px;
  height: 44px;
  font-size: 20px;
  line-height: 44px;
  color: var(--ink-900);
  background: var(--surface-100);
  border-bottom-color: var(--surface-300);
}

.leaflet-bar a:first-child {
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}

.leaflet-bar a:last-child {
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}

.leaflet-bar a:hover {
  background: var(--surface-200);
}

.leaflet-control-attribution {
  padding: 2px var(--space-2);
  font-size: 13px;
  line-height: 18px;
  color: var(--ink-600);
  background: var(--surface-100);
  border-radius: var(--radius-sm);
}

.leaflet-control-attribution a {
  color: var(--teal-700);
}

.leaflet-top.leaflet-right {
  top: 76px;
}
</style>

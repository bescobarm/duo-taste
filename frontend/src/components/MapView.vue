<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import L from 'leaflet'
import { bandFor, iconFor, type ScoreBand } from '../categories'
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

// Pins are HTML strings, so their utilities live here where Tailwind can scan them.
const pinBody = 'inline-flex h-9 items-center gap-1 rounded-pill border-2 px-2 shadow-pin'
// The numeral keeps the line height it inherits from .leaflet-container.
const pinScore = 'text-body-sm leading-[inherit] font-bold'

// An unrated pin is a white pill, so it takes the border color for its edge.
const pinFill: Record<ScoreBand, string> = {
  high: 'border-surface-100 bg-score-high text-on-teal',
  mid: 'border-surface-100 bg-score-mid text-on-amber',
  low: 'border-surface-100 bg-score-low text-on-rust',
  none: 'border-border-200 bg-surface-100 text-ink-900',
}

const pinSelected = 'outline-2 outline-offset-2 outline-focus-ring'

// Leaflet marks an ancestor with .leaflet-dragging while the pin is dragged.
const pinDraft =
  'border-surface-100 bg-pin-saved text-on-cyan cursor-grab in-[.leaflet-dragging]:cursor-grabbing'

// A pin is the score pill from the canvas: category icon plus the numeral, on
// a fill from the score band. The numeral is always printed, so the band color
// is never the only signal.
function pinIcon(place: Place, selected: boolean) {
  const band = bandFor(place.score)
  const label = place.score ? place.score.toFixed(2) : '—'

  return L.divIcon({
    // An empty class keeps Leaflet's default white .leaflet-div-icon box off.
    className: '',
    html: `<span class="${pinBody} ${pinFill[band]}${selected ? ` ${pinSelected}` : ''}">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
        stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <use href="#${iconFor(place.category)}"></use>
      </svg><span class="${pinScore}">${label}</span></span>`,
    iconSize: [76, 36],
    iconAnchor: [38, 18],
  })
}

let framed = false
let locating = false
let userMarker: L.CircleMarker | null = null

// The map opens on Chapinero, as the design frames it. Once places load, fit
// the view to them once, so pins elsewhere are not left off-screen. While the
// browser is still asking for the user's location, hold off: their position
// wins when it arrives.
function frameToPlaces() {
  if (!map || framed || locating || !props.places.length) return

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
        ? `${place.name}, rated ${place.score.toFixed(2)} out of 5`
        : `${place.name}, not rated yet`,
    })

    marker.on('click', () => emit('select', place))
    markers.addLayer(marker)
  }

  frameToPlaces()
}

// The draft pin can be dragged: dropping it reports the new point like a map
// click does, which also refreshes the address in the form.
function renderDraft() {
  if (!map) return

  if (!props.draft) {
    if (draftMarker) map.removeLayer(draftMarker)
    draftMarker = null
    return
  }

  if (draftMarker) {
    draftMarker.setLatLng([props.draft.lat, props.draft.lng])
    return
  }

  draftMarker = L.marker([props.draft.lat, props.draft.lng], {
    draggable: true,
    autoPan: true,
    title: 'New place: drag to move',
    icon: L.divIcon({
      className: '',
      html: `<span class="${pinBody} ${pinDraft}">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <use href="#ic-pin"></use>
        </svg><span class="${pinScore}">New</span></span>`,
      iconSize: [72, 36],
      iconAnchor: [36, 18],
    }),
  })
  draftMarker.on('dragend', () => {
    const point = draftMarker?.getLatLng()
    if (point) emit('pick', { lat: point.lat, lng: point.lng })
  })
  draftMarker.addTo(map)
}

function showUser(lat: number, lng: number, accuracy: number) {
  if (!map) return

  if (!userMarker) {
    userMarker = L.circleMarker([lat, lng], {
      className: 'fill-teal-500 stroke-surface-100',
      radius: 8,
      weight: 3,
      fillOpacity: 1,
      interactive: false,
    }).addTo(map)
  } else {
    userMarker.setLatLng([lat, lng])
  }

  userMarker.bindTooltip(`You are here (±${Math.round(accuracy)} m)`)
}

// Asks the browser for the user's position. The first time it runs the browser
// shows its permission prompt; a denial or a timeout leaves the map where it is.
function locate(onDone?: (found: boolean) => void) {
  if (!('geolocation' in navigator)) {
    onDone?.(false)
    return
  }

  navigator.geolocation.getCurrentPosition(
    ({ coords }) => {
      showUser(coords.latitude, coords.longitude, coords.accuracy)
      map?.setView([coords.latitude, coords.longitude], 15)
      onDone?.(true)
    },
    () => onDone?.(false),
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 60000 },
  )
}

const LocateControl = L.Control.extend({
  onAdd() {
    const bar = L.DomUtil.create('div', 'leaflet-bar leaflet-control')
    // flex! because leaflet.css sets display: block on .leaflet-bar a, unlayered.
    const button = L.DomUtil.create('a', 'flex! items-center justify-center', bar)
    button.href = '#'
    button.role = 'button'
    button.title = 'Show my location'
    button.setAttribute('aria-label', 'Show my location')
    button.innerHTML = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor"
      stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <use href="#ic-locate"></use></svg>`

    L.DomEvent.disableClickPropagation(bar)
    L.DomEvent.on(button, 'click', (event) => {
      L.DomEvent.preventDefault(event)
      locate()
    })

    return bar
  },
})

onMounted(() => {
  if (!container.value) return

  map = L.map(container.value, { zoomControl: false }).setView([4.6655, -74.0578], 15)
  L.control.zoom({ position: 'topright' }).addTo(map)
  new LocateControl({ position: 'topright' }).addTo(map)

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

  // Open on the user's position when they allow it; otherwise fall back to
  // framing the places around the default view.
  locating = true
  locate((found) => {
    locating = false
    if (found) framed = true
    else frameToPlaces()
  })
})

onUnmounted(() => {
  map?.remove()
  map = null
  userMarker = null
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
  <div ref="container" class="h-full w-full"></div>
</template>

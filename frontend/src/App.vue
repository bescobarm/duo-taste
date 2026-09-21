<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppIcon from './components/AppIcon.vue'
import BrandMark from './components/BrandMark.vue'
import IconSprite from './components/IconSprite.vue'
import MapView from './components/MapView.vue'
import { api } from './api'
import { bandFor, iconFor, labelFor } from './categories'
import type { Category, Place } from './types'

const mapView = ref<InstanceType<typeof MapView> | null>(null)
const places = ref<Place[]>([])
const categories = ref<Category[]>([])
const filter = ref<Category | ''>('')
const search = ref('')
const draft = ref<{ lat: number; lng: number } | null>(null)
const selected = ref<Place | null>(null)
const error = ref('')
const sheetExpanded = ref(false)

const form = ref({ name: '', category: 'other' as Category, address: '' })
const rating = ref({ author: '', score: 4.5, notes: '' })

const visible = computed(() => {
  const term = search.value.trim().toLowerCase()

  return places.value.filter((place) => {
    const matchesCategory = !filter.value || place.category === filter.value
    const matchesTerm =
      !term ||
      place.name.toLowerCase().includes(term) ||
      place.address.toLowerCase().includes(term)

    return matchesCategory && matchesTerm
  })
})

const ranked = computed(() =>
  [...visible.value].sort((a, b) => b.score - a.score || a.name.localeCompare(b.name)),
)

// The header counts come from the data, not from a fixed pair of tasters.
const tasters = computed(() => {
  const names = new Set<string>()
  for (const place of places.value) {
    for (const entry of place.ratings) names.add(entry.author)
  }

  return [...names]
})

const countFor = (category: Category) =>
  places.value.filter((place) => place.category === category).length

const rankOf = (place: Place) => ranked.value.findIndex((entry) => entry.id === place.id) + 1

// Each taster's own average for the selected place, for the meter rows.
const tasterScores = computed(() => {
  if (!selected.value) return []

  const totals = new Map<string, { sum: number; count: number }>()
  for (const entry of selected.value.ratings) {
    const current = totals.get(entry.author) ?? { sum: 0, count: 0 }
    totals.set(entry.author, { sum: current.sum + entry.score, count: current.count + 1 })
  }

  return [...totals].map(([author, { sum, count }]) => ({ author, score: sum / count }))
})

const tierLabel = (score: number) => {
  const band = bandFor(score)
  if (band === 'high') return 'Top tier'
  if (band === 'mid') return 'Worth a visit'
  if (band === 'low') return 'Below par'

  return 'Not rated'
}

const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? '' : 's'}`

const initial = (name: string) => name.trim().charAt(0).toUpperCase() || '?'
const scoreText = (score: number) => (score ? score.toFixed(1) : '—')

async function load() {
  try {
    ;[places.value, categories.value] = await Promise.all([api.places(), api.categories()])
  } catch (err) {
    error.value = (err as Error).message
  }
}

function pick(coords: { lat: number; lng: number }) {
  draft.value = coords
  selected.value = null
}

// On a phone the sidebar is a sheet: the ranking peeks, a panel takes it all.
const sheetOpen = computed(() => sheetExpanded.value || !!draft.value || !!selected.value)

function addAtCenter() {
  const center = mapView.value?.center()
  if (center) pick(center)
}

function openDetail(place: Place) {
  selected.value = place
  draft.value = null
}

function closePanels() {
  draft.value = null
  selected.value = null
  sheetExpanded.value = false
}

async function createPlace() {
  if (!draft.value) return

  try {
    const place = await api.createPlace({ ...form.value, ...draft.value })
    places.value = [place, ...places.value]
    draft.value = null
    form.value = { name: '', category: 'other', address: '' }
    error.value = ''
  } catch (err) {
    error.value = (err as Error).message
  }
}

async function ratePlace() {
  if (!selected.value) return

  try {
    const updated = await api.ratePlace(selected.value.id, { ...rating.value })
    places.value = places.value.map((place) => (place.id === updated.id ? updated : place))
    selected.value = updated
    rating.value = { author: rating.value.author, score: 4.5, notes: '' }
    error.value = ''
  } catch (err) {
    error.value = (err as Error).message
  }
}

onMounted(load)
</script>

<template>
  <IconSprite />

  <div class="layout">
    <header class="brand">
        <BrandMark />
        <div class="brand-text">
          <div class="t-heading-m">DuoTaste</div>
          <div class="t-caption muted">
            {{ plural(places.length, 'place') }} · {{ plural(tasters.length, 'taster') }}
          </div>
        </div>
        <div class="tasters">
          <span v-for="name in tasters.slice(0, 2)" :key="name" class="avatar" :title="name">
            {{ initial(name) }}
          </span>
        </div>
    </header>

    <div class="map-toolbar">
      <label for="search" class="sr-only">Search places</label>
      <div class="search">
        <AppIcon name="ic-search" />
        <input id="search" v-model="search" type="text" placeholder="Search places" />
      </div>
      <span style="flex-grow: 1"></span>
      <button type="button" class="btn" @click="addAtCenter">
        <AppIcon name="ic-plus" :stroke="1.8" />
        Add place
      </button>
    </div>

    <div v-if="!draft && !selected" class="filters">
      <span class="t-caption muted">Filter by category</span>
      <div class="chips">
        <button type="button" class="chip is-bare" :aria-pressed="filter === ''" @click="filter = ''">
          All {{ places.length }}
        </button>
        <button
          v-for="category in categories.filter((entry) => countFor(entry) > 0)"
          :key="category"
          type="button"
          class="chip"
          :aria-pressed="filter === category"
          @click="filter = category"
        >
          <AppIcon :name="iconFor(category)" :size="16" />
          {{ labelFor(category) }}
        </button>
      </div>
    </div>

    <aside class="sidebar" :class="{ 'is-expanded': sheetOpen }">
      <button
        type="button"
        class="sheet-grabber"
        :aria-expanded="sheetOpen"
        aria-label="Expand or collapse the list"
        @click="draft || selected ? closePanels() : (sheetExpanded = !sheetExpanded)"
      ></button>

      <p v-if="error" class="error">{{ error }}</p>

      <!-- Add a place ------------------------------------------------------>
      <section v-if="draft" class="panel">
        <div class="panel-head">
          <h2>New place</h2>
          <button
            type="button"
            class="btn btn-icon"
            aria-label="Cancel adding a place"
            @click="closePanels()"
          >
            <AppIcon name="ic-close" :size="18" :stroke="1.6" />
          </button>
        </div>

        <div class="coords">
          <AppIcon name="ic-pin" :size="18" />
          <code>{{ draft.lat.toFixed(5) }}, {{ draft.lng.toFixed(5) }}</code>
          <span class="t-caption muted">Click the map to adjust</span>
        </div>

        <form @submit.prevent="createPlace">
          <div class="field">
            <label for="place-name">Name</label>
            <input id="place-name" v-model="form.name" type="text" required />
          </div>

          <div class="field" style="margin-top: var(--space-3)">
            <span id="category-label">Category</span>
            <div class="category-grid" role="group" aria-labelledby="category-label">
              <button
                v-for="category in categories"
                :key="category"
                type="button"
                class="category-cell"
                :aria-pressed="form.category === category"
                @click="form.category = category"
              >
                <AppIcon :name="iconFor(category)" />
                {{ labelFor(category) }}
              </button>
            </div>
          </div>

          <div class="field" style="margin-top: var(--space-3)">
            <label for="place-address">Address (optional)</label>
            <input id="place-address" v-model="form.address" type="text" />
          </div>

          <div class="btn-row" style="margin-top: var(--space-3)">
            <button type="submit" class="btn btn-grow">Save place</button>
            <button type="button" class="btn btn-quiet" @click="closePanels()">Cancel</button>
          </div>
        </form>
      </section>

      <!-- Place detail ----------------------------------------------------->
      <template v-else-if="selected">
        <div class="detail-bar">
          <button type="button" class="btn btn-quiet btn-sm" @click="closePanels()">
            <AppIcon name="ic-back" :size="18" :stroke="1.6" />
            Ranking
          </button>
        </div>

        <div class="detail">
          <div class="field">
            <h2 class="t-heading-m" style="margin: 0">{{ selected.name }}</h2>
            <div style="display: flex; align-items: center; gap: var(--space-2)">
              <span class="chip-static">
                <AppIcon :name="iconFor(selected.category)" :size="16" />
                {{ labelFor(selected.category) }}
              </span>
              <span class="t-body-sm muted">{{ selected.address || 'No address' }}</span>
            </div>
          </div>

          <div class="score-hero">
            <span class="t-display">{{ scoreText(selected.score) }}</span>
            <span class="muted" style="padding-bottom: 6px">/ 5</span>
            <span style="flex-grow: 1"></span>
            <div style="text-align: right">
              <span class="tier" :class="`score-${bandFor(selected.score)}`">
                {{ tierLabel(selected.score) }}
              </span>
              <span class="t-caption muted" style="display: block; margin-top: var(--space-1)">
                #{{ rankOf(selected) }} of {{ ranked.length }}
              </span>
            </div>
          </div>

          <div v-if="tasterScores.length" class="taster-card">
            <div v-for="taster in tasterScores" :key="taster.author" class="taster">
              <span class="avatar avatar-sm">{{ initial(taster.author) }}</span>
              <span class="taster-name">{{ taster.author }}</span>
              <span class="meter">
                <span :style="{ width: `${(taster.score / 5) * 100}%` }"></span>
              </span>
              <span class="taster-score">{{ taster.score.toFixed(1) }}</span>
            </div>
          </div>

          <form class="field" @submit.prevent="ratePlace">
            <span class="t-section">ADD A RATING</span>

            <div class="field" style="margin-top: var(--space-2)">
              <label for="author">Taster</label>
              <input id="author" v-model="rating.author" type="text" required />
            </div>

            <div style="display: flex; align-items: center; gap: var(--space-2); margin-top: var(--space-3)">
              <label for="score" class="t-caption muted">Score</label>
              <span style="flex-grow: 1"></span>
              <span class="score" :class="`score-${bandFor(rating.score)}`">
                {{ rating.score.toFixed(1) }}
              </span>
            </div>
            <input
              id="score"
              v-model.number="rating.score"
              class="slider"
              type="range"
              min="1"
              max="5"
              step="0.5"
            />
            <div class="slider-scale">
              <span>1</span><span>2</span><span>3</span><span>4</span><span>5</span>
            </div>

            <label for="notes" class="sr-only">Notes</label>
            <input
              id="notes"
              v-model="rating.notes"
              type="text"
              placeholder="Notes (optional)"
              style="margin-top: var(--space-2)"
            />

            <button type="submit" class="btn" style="margin-top: var(--space-2)">Add rating</button>
          </form>

          <ul v-if="selected.ratings.length" class="ratings">
            <li class="t-section">{{ plural(selected.ratings.length, 'RATING') }}</li>
            <li v-for="entry in selected.ratings" :key="entry.id">
              <div class="rating-head">
                <span style="font-weight: 600">{{ entry.author }}</span>
                <span style="font-weight: 700">{{ entry.score.toFixed(1) }}</span>
                <span style="flex-grow: 1"></span>
                <span class="t-caption muted">
                  {{ new Date(entry.createdAt).toLocaleDateString(undefined, { day: 'numeric', month: 'short' }) }}
                </span>
              </div>
              <p v-if="entry.notes" class="rating-notes">{{ entry.notes }}</p>
            </li>
          </ul>
        </div>
      </template>

      <!-- Ranking ---------------------------------------------------------->
      <template v-else>
        <div class="list-head">
          <span class="t-section" style="flex-grow: 1">RANKING</span>
          <span class="t-caption muted">{{ plural(ranked.length, 'place') }}</span>
        </div>

        <div class="ranking">
          <p v-if="!ranked.length" class="hint">
            No places yet. Click anywhere on the map to add one.
          </p>
          <button
            v-for="(place, index) in ranked"
            :key="place.id"
            type="button"
            class="row"
            @click="openDetail(place)"
          >
            <span class="row-rank">{{ index + 1 }}</span>
            <span class="row-icon"><AppIcon :name="iconFor(place.category)" /></span>
            <span class="row-body">
              <span class="row-name">{{ place.name }}</span>
              <span class="row-meta">
                {{ place.address || 'No address' }} · {{ plural(place.ratings.length, 'rating') }}
              </span>
            </span>
            <span class="score" :class="`score-${bandFor(place.score)}`">
              {{ scoreText(place.score) }}
            </span>
          </button>
        </div>
      </template>
    </aside>

    <main class="map-area">
      <MapView
        ref="mapView"
        :places="visible"
        :draft="draft"
        :selected-id="selected?.id ?? null"
        @pick="pick"
        @select="openDetail"
      />

      <p v-if="!draft" class="map-note">
        <AppIcon name="ic-pin" />
        Click anywhere on the map to add a place.
      </p>

      <button v-if="!sheetOpen" type="button" class="fab" aria-label="Add place" @click="addAtCenter">
        <AppIcon name="ic-plus" :size="24" :stroke="1.8" />
      </button>
    </main>
  </div>
</template>

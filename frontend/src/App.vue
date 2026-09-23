<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppIcon from './components/AppIcon.vue'
import BrandMark from './components/BrandMark.vue'
import IconSprite from './components/IconSprite.vue'
import MapView from './components/MapView.vue'
import ScoreInput from './components/ScoreInput.vue'
import { api, reverseGeocode } from './api'
import { bandFor, iconFor, labelFor } from './categories'
import { bandClass, btn, chip, chipBare, scorePill, textInput } from './ui'
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
// The new place's first rating: a place is never saved unrated.
const firstRating = ref({ author: '', score: 4.5, notes: '' })
const lookingUp = ref(false)

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
const scoreText = (score: number) => (score ? score.toFixed(2) : '—')

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

// Every time the new pin moves, fill the address from where it now sits. The
// lookup waits for the pin to settle and drops any answer a newer move replaced.
let lookupTimer: ReturnType<typeof setTimeout> | undefined
let lookup: AbortController | null = null

watch(draft, (point) => {
  clearTimeout(lookupTimer)
  lookup?.abort()
  lookingUp.value = false
  if (!point) return

  lookingUp.value = true
  lookupTimer = setTimeout(async () => {
    const controller = new AbortController()
    lookup = controller

    try {
      const address = await reverseGeocode(point.lat, point.lng, controller.signal)
      if (address) form.value.address = address
    } catch {
      // No address for this point, or the lookup failed: the field stays editable.
    } finally {
      if (lookup === controller) lookingUp.value = false
    }
  }, 500)
})

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
    const place = await api.createPlace({
      ...form.value,
      ...draft.value,
      rating: { ...firstRating.value },
    })
    places.value = [place, ...places.value]
    draft.value = null
    form.value = { name: '', category: 'other', address: '' }
    firstRating.value = { author: firstRating.value.author, score: 4.5, notes: '' }
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

  <div
    class="relative grid h-full grid-cols-[var(--sidebar-w)_1fr] grid-rows-[auto_auto_1fr] [--sidebar-w:360px] phone:grid-cols-1 phone:grid-rows-[auto_auto_auto_1fr] phone:*:min-w-0"
  >
    <header
      class="col-1 row-1 flex items-center gap-2 border-r border-b border-r-border-200 border-b-surface-300 p-4 phone:border-r-0"
    >
        <BrandMark />
        <div class="grow">
          <div class="text-heading-m">DuoTaste</div>
          <div class="text-caption text-ink-600">
            {{ plural(places.length, 'place') }} · {{ plural(tasters.length, 'taster') }}
          </div>
        </div>
        <div class="flex items-center">
          <span
            v-for="name in tasters.slice(0, 2)"
            :key="name"
            class="inline-flex size-7 items-center justify-center rounded-pill border-2 border-surface-100 bg-surface-300 text-[13px] font-semibold text-ink-900 not-first:-ml-2 not-first:bg-pin-saved not-first:text-on-cyan"
            :title="name"
          >
            {{ initial(name) }}
          </span>
        </div>
    </header>

    <div
      class="absolute top-4 right-4 left-[calc(var(--sidebar-w)+var(--space-3))] z-500 flex items-center gap-4 phone:static phone:col-1 phone:row-2 phone:px-4 phone:pt-4 phone:pb-2"
    >
      <label for="search" class="sr-only">Search places</label>
      <div
        class="flex h-11 w-105 max-w-[55%] shrink-0 items-center gap-2 rounded-md border border-border-200 bg-surface-100 px-4 text-ink-600 phone:w-full phone:max-w-none"
      >
        <AppIcon name="ic-search" />
        <input
          id="search"
          v-model="search"
          type="text"
          placeholder="Search places"
          class="min-w-0 grow border-0 bg-transparent p-0 text-body text-ink-900 outline-none"
        />
      </div>
      <span class="grow"></span>
      <button type="button" :class="[btn.primary, 'phone:hidden']" @click="addAtCenter">
        <AppIcon name="ic-plus" :stroke="1.8" />
        Add place
      </button>
    </div>

    <div
      v-if="!draft && !selected"
      class="col-1 row-2 flex flex-col gap-2 border-r border-b border-r-border-200 border-b-surface-300 p-4 phone:row-3 phone:border-r-0 phone:border-b-0 phone:pt-0"
    >
      <span class="text-caption text-ink-600">Filter by category</span>
      <div
        class="flex flex-wrap gap-2 phone:min-w-0 phone:flex-nowrap phone:overflow-x-auto phone:[scrollbar-width:none] phone:[&::-webkit-scrollbar]:hidden"
      >
        <button type="button" :class="chipBare" :aria-pressed="filter === ''" @click="filter = ''">
          All {{ places.length }}
        </button>
        <button
          v-for="category in categories.filter((entry) => countFor(entry) > 0)"
          :key="category"
          type="button"
          :class="chip"
          :aria-pressed="filter === category"
          @click="filter = category"
        >
          <AppIcon :name="iconFor(category)" :size="16" />
          {{ labelFor(category) }}
        </button>
      </div>
    </div>

    <aside
      class="col-1 row-3 flex min-h-0 flex-col border-r border-border-200 bg-surface-100 phone:absolute phone:inset-x-0 phone:bottom-0 phone:z-1100 phone:col-auto phone:row-auto phone:rounded-t-lg phone:border phone:border-b-0 phone:transition-[height] phone:duration-180 phone:ease-[ease]"
      :class="sheetOpen ? 'phone:h-[calc(100%-64px)]' : 'phone:h-75'"
    >
      <button
        type="button"
        class="hidden h-6 w-full shrink-0 cursor-pointer items-center justify-center border-0 bg-transparent p-0 before:h-1 before:w-9 before:rounded-pill before:bg-surface-300 phone:flex"
        :aria-expanded="sheetOpen"
        aria-label="Expand or collapse the list"
        @click="draft || selected ? closePanels() : (sheetExpanded = !sheetExpanded)"
      ></button>

      <p v-if="error" class="mx-4 rounded-sm bg-rust-600 px-3 py-2 text-body-sm text-on-rust">{{ error }}</p>

      <section
        v-if="draft"
        class="m-4 flex flex-col gap-4 rounded-lg border border-border-200 bg-surface-200 p-4 phone:mt-0"
      >
        <div class="flex items-center gap-2">
          <h2 class="grow text-heading-m">New place</h2>
          <button
            type="button"
            :class="btn.icon"
            aria-label="Cancel adding a place"
            @click="closePanels()"
          >
            <AppIcon name="ic-close" :size="18" :stroke="1.6" />
          </button>
        </div>

        <form @submit.prevent="createPlace">
          <div class="flex flex-col gap-1">
            <label for="place-address" class="text-caption text-ink-600">Address</label>
            <input
              id="place-address"
              v-model="form.address"
              type="text"
              :class="textInput"
              :placeholder="lookingUp ? 'Looking up address…' : 'Street address'"
            />
            <span class="text-caption text-ink-600">
              {{ lookingUp ? 'Looking up address…' : 'Drag the New pin or click the map to move it' }}
            </span>
          </div>

          <div class="mt-4 flex flex-col gap-1">
            <label for="place-name" class="text-caption text-ink-600">Name</label>
            <input id="place-name" v-model="form.name" type="text" :class="textInput" required />
          </div>

          <div class="mt-4 flex flex-col gap-1">
            <span id="category-label" class="text-caption text-ink-600">Category</span>
            <div class="grid grid-cols-4 gap-2 narrow:grid-cols-3" role="group" aria-labelledby="category-label">
              <button
                v-for="category in categories"
                :key="category"
                type="button"
                class="flex h-15 cursor-pointer flex-col items-center justify-center gap-1 rounded-md border border-border-200 bg-surface-100 text-[13px] leading-[normal] font-medium text-ink-900 phone:h-16 aria-pressed:border-teal-500 aria-pressed:bg-teal-500 aria-pressed:font-semibold aria-pressed:text-on-teal"
                :aria-pressed="form.category === category"
                @click="form.category = category"
              >
                <AppIcon :name="iconFor(category)" />
                {{ labelFor(category) }}
              </button>
            </div>
          </div>

          <div class="mt-4 flex flex-col gap-1">
            <span class="text-caption text-ink-600">YOUR RATING</span>

            <div class="mt-2 flex flex-col gap-1">
              <label for="first-author" class="text-caption text-ink-600">Taster</label>
              <input id="first-author" v-model="firstRating.author" type="text" :class="textInput" required />
            </div>

            <ScoreInput id="first-score" v-model="firstRating.score" class="mt-4" />

            <label for="first-notes" class="sr-only">Notes</label>
            <input
              id="first-notes"
              v-model="firstRating.notes"
              type="text"
              placeholder="Notes (optional)"
              :class="[textInput, 'mt-2']"
            />
          </div>

          <div class="mt-4 flex gap-2">
            <button type="submit" :class="[btn.primary, 'grow']">Save place</button>
            <button type="button" :class="btn.quiet" @click="closePanels()">Cancel</button>
          </div>
        </form>
      </section>

      <template v-else-if="selected">
        <div class="flex items-center gap-2 border-b border-surface-300 px-4 py-2">
          <button type="button" :class="btn.quietSmall" @click="closePanels()">
            <AppIcon name="ic-back" :size="18" :stroke="1.6" />
            Ranking
          </button>
        </div>

        <div class="flex flex-col gap-4 overflow-y-auto p-4">
          <div class="flex flex-col gap-1">
            <h2 class="text-heading-m">{{ selected.name }}</h2>
            <div class="flex items-center gap-2">
              <span
                class="inline-flex h-7 items-center gap-1 rounded-pill bg-surface-200 pr-2.5 pl-2 text-caption text-ink-900"
              >
                <AppIcon :name="iconFor(selected.category)" :size="16" />
                {{ labelFor(selected.category) }}
              </span>
              <span class="text-body-sm text-ink-600">{{ selected.address || 'No address' }}</span>
            </div>
          </div>

          <div class="flex items-end gap-2">
            <span class="text-display">{{ scoreText(selected.score) }}</span>
            <span class="pb-1.5 text-ink-600">/ 5</span>
            <span class="grow"></span>
            <div class="text-right">
              <span
                class="inline-flex h-6 items-center rounded-sm px-2 text-caption font-bold"
                :class="bandClass(bandFor(selected.score))"
              >
                {{ tierLabel(selected.score) }}
              </span>
              <span class="mt-1 block text-caption text-ink-600">
                #{{ rankOf(selected) }} of {{ ranked.length }}
              </span>
            </div>
          </div>

          <div
            v-if="tasterScores.length"
            class="flex flex-col gap-2 rounded-md border border-border-200 bg-surface-200 p-4"
          >
            <div v-for="taster in tasterScores" :key="taster.author" class="flex items-center gap-2">
              <span
                class="inline-flex size-6 items-center justify-center rounded-pill bg-surface-300 text-[13px] font-semibold text-ink-900"
              >
                {{ initial(taster.author) }}
              </span>
              <span class="w-14.5 truncate text-body-sm">{{ taster.author }}</span>
              <span class="h-1.5 grow overflow-hidden rounded-pill bg-surface-300">
                <span
                  class="block h-1.5 rounded-pill bg-teal-500"
                  :style="{ width: `${(taster.score / 5) * 100}%` }"
                ></span>
              </span>
              <span class="w-10 text-right text-body-sm font-bold">{{ taster.score.toFixed(2) }}</span>
            </div>
          </div>

          <form class="flex flex-col gap-1" @submit.prevent="ratePlace">
            <span class="text-caption text-ink-600">ADD A RATING</span>

            <div class="mt-2 flex flex-col gap-1">
              <label for="author" class="text-caption text-ink-600">Taster</label>
              <input id="author" v-model="rating.author" type="text" :class="textInput" required />
            </div>

            <ScoreInput id="score" v-model="rating.score" class="mt-4" />

            <label for="notes" class="sr-only">Notes</label>
            <input
              id="notes"
              v-model="rating.notes"
              type="text"
              placeholder="Notes (optional)"
              :class="[textInput, 'mt-2']"
            />

            <button type="submit" :class="[btn.primary, 'mt-2']">Add rating</button>
          </form>

          <ul v-if="selected.ratings.length" class="flex flex-col gap-2 border-t border-surface-300 pt-2 pl-10">
            <li class="text-caption tracking-section text-ink-600">
              {{ plural(selected.ratings.length, 'RATING') }}
            </li>
            <li v-for="entry in selected.ratings" :key="entry.id">
              <div class="flex items-center gap-2 text-body-sm">
                <span class="font-semibold">{{ entry.author }}</span>
                <span class="font-bold">{{ entry.score.toFixed(2) }}</span>
                <span class="grow"></span>
                <span class="text-caption text-ink-600">
                  {{ new Date(entry.createdAt).toLocaleDateString(undefined, { day: 'numeric', month: 'short' }) }}
                </span>
              </div>
              <p v-if="entry.notes" class="text-body-sm text-ink-600">{{ entry.notes }}</p>
            </li>
          </ul>
        </div>
      </template>

      <template v-else>
        <div class="flex items-center gap-2 px-4 py-2">
          <span class="grow text-caption tracking-section text-ink-600">RANKING</span>
          <span class="text-caption text-ink-600">{{ plural(ranked.length, 'place') }}</span>
        </div>

        <div class="min-h-0 grow overflow-y-auto">
          <p v-if="!ranked.length" class="my-3.5 px-4 pb-4 text-body-sm text-ink-600">
            No places yet. Click anywhere on the map to add one.
          </p>
          <button
            v-for="(place, index) in ranked"
            :key="place.id"
            type="button"
            class="flex w-full cursor-pointer items-center gap-2 border-0 border-b border-surface-300 bg-surface-100 px-4 py-2 text-left hover:bg-surface-200"
            @click="openDetail(place)"
          >
            <span class="w-5 text-right text-[13px] font-medium text-ink-600">{{ index + 1 }}</span>
            <span class="inline-flex size-9 shrink-0 items-center justify-center rounded-md bg-surface-200 text-ink-900">
              <AppIcon :name="iconFor(place.category)" />
            </span>
            <span class="min-w-0 grow">
              <span class="block truncate text-body font-semibold text-ink-900">{{ place.name }}</span>
              <span class="block truncate text-body-sm text-ink-600">
                {{ place.address || 'No address' }} · {{ plural(place.ratings.length, 'rating') }}
              </span>
            </span>
            <span :class="[scorePill, bandClass(bandFor(place.score))]">
              {{ scoreText(place.score) }}
            </span>
          </button>
        </div>
      </template>
    </aside>

    <main class="relative col-2 row-[1/-1] min-h-0 min-w-0 phone:col-1 phone:row-4">
      <MapView
        ref="mapView"
        :places="visible"
        :draft="draft"
        :selected-id="selected?.id ?? null"
        @pick="pick"
        @select="openDetail"
      />

      <p
        v-if="!draft"
        class="absolute bottom-7.5 left-4 z-500 flex items-center gap-2 rounded-md border border-border-200 bg-surface-100 px-4 py-2 text-body-sm text-ink-900 phone:hidden"
      >
        <AppIcon name="ic-pin" class="text-teal-700" />
        Click anywhere on the map to add a place.
      </p>

      <button
        v-if="!sheetOpen"
        type="button"
        class="absolute right-4 bottom-79 z-1101 hidden size-14 cursor-pointer items-center justify-center rounded-pill border-2 border-surface-100 bg-teal-500 p-0 text-on-teal phone:inline-flex"
        aria-label="Add place"
        @click="addAtCenter"
      >
        <AppIcon name="ic-plus" :size="24" :stroke="1.8" />
      </button>
    </main>
  </div>
</template>

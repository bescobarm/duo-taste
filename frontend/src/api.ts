import type { Category, NewPlace, NewRating, Place } from './types'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })

  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(body.error ?? 'request failed')
  }

  return response.json() as Promise<T>
}

export const api = {
  categories: () => request<Category[]>('/api/categories'),
  places: (category?: Category | '') =>
    request<Place[]>(category ? `/api/places?category=${category}` : '/api/places'),
  createPlace: (place: NewPlace) =>
    request<Place>('/api/places', { method: 'POST', body: JSON.stringify(place) }),
  ratePlace: (placeId: string, rating: NewRating) =>
    request<Place>(`/api/places/${placeId}/ratings`, {
      method: 'POST',
      body: JSON.stringify(rating),
    }),
}

interface NominatimAddress {
  road?: string
  pedestrian?: string
  house_number?: string
  neighbourhood?: string
  suburb?: string
  city?: string
  town?: string
  village?: string
}

// Turns a map point into a street address with OpenStreetMap's Nominatim, which
// is free and keyless but asks for at most one request a second: callers
// debounce, and pass a signal so a newer lookup cancels the one in flight.
export async function reverseGeocode(
  lat: number,
  lng: number,
  signal?: AbortSignal,
): Promise<string> {
  const params = new URLSearchParams({
    format: 'jsonv2',
    lat: String(lat),
    lon: String(lng),
    zoom: '18',
    addressdetails: '1',
    'accept-language': navigator.language,
  })

  const response = await fetch(`https://nominatim.openstreetmap.org/reverse?${params}`, { signal })
  if (!response.ok) throw new Error('address lookup failed')

  const body = (await response.json()) as { display_name?: string; address?: NominatimAddress }
  const address = body.address ?? {}
  const street = [address.road ?? address.pedestrian, address.house_number].filter(Boolean).join(' ')
  const area = address.suburb ?? address.neighbourhood
  const city = address.city ?? address.town ?? address.village
  const short = [street, area, city].filter(Boolean).join(', ')

  return short || body.display_name || ''
}

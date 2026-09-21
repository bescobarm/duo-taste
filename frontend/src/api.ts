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

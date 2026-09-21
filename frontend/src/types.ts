export type Category =
  | 'pizza'
  | 'burger'
  | 'sushi'
  | 'tacos'
  | 'coffee'
  | 'dessert'
  | 'breakfast'
  | 'other'

export interface Rating {
  id: string
  placeId: string
  author: string
  score: number
  notes: string
  createdAt: string
}

export interface Place {
  id: string
  name: string
  category: Category
  address: string
  lat: number
  lng: number
  score: number
  ratings: Rating[]
  createdAt: string
}

export interface NewPlace {
  name: string
  category: Category
  address: string
  lat: number
  lng: number
}

export interface NewRating {
  author: string
  score: number
  notes: string
}

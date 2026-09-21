import type { Category } from './types'

// The design system rules emoji out of product surfaces: every category is a
// stroked 24px icon from the sprite in IconSprite.vue.
export const categoryIcon: Record<Category, string> = {
  pizza: 'ic-pizza',
  burger: 'ic-burger',
  sushi: 'ic-sushi',
  tacos: 'ic-tacos',
  coffee: 'ic-coffee',
  dessert: 'ic-dessert',
  breakfast: 'ic-breakfast',
  other: 'ic-other',
}

export const categoryLabel: Record<Category, string> = {
  pizza: 'Pizza',
  burger: 'Burger',
  sushi: 'Sushi',
  tacos: 'Tacos',
  coffee: 'Coffee',
  dessert: 'Dessert',
  breakfast: 'Breakfast',
  other: 'Other',
}

export function iconFor(category: Category): string {
  return categoryIcon[category] ?? categoryIcon.other
}

export function labelFor(category: Category): string {
  return categoryLabel[category] ?? categoryLabel.other
}

export type ScoreBand = 'high' | 'mid' | 'low' | 'none'

// Score bands from the canvas: 4.2 and up teal, 3.2 to 4.1 amber, below 3.2
// rust. The numeral is printed alongside every band, so color is never the
// only signal.
export function bandFor(score: number): ScoreBand {
  if (!score) return 'none'
  if (score >= 4.2) return 'high'
  if (score >= 3.2) return 'mid'

  return 'low'
}

import type { ScoreBand } from './categories'

// Class strings for the pieces that repeat across the templates. Each variant
// is spelled out in full: two utilities for the same property on one element
// resolve by stylesheet order, not by the order they are written in.

// Buttons and the score pill keep the browser's normal line height, which
// Preflight would otherwise swap for the body's 24px and nudge the label.
const btnBase = 'inline-flex cursor-pointer items-center justify-center rounded-md leading-[normal]'

export const btn = {
  primary: `${btnBase} h-11 gap-2 bg-teal-500 px-4 text-body font-semibold text-on-teal phone:h-13`,
  quiet: `${btnBase} h-11 gap-2 border border-border-200 bg-transparent px-4 text-body font-medium text-ink-900 phone:h-13`,
  quietSmall: `${btnBase} h-9 gap-2 border border-border-200 bg-transparent pr-3 pl-2 text-body-sm font-medium text-ink-900 phone:h-11`,
  // The phone height matches the other buttons in the row, so it is taller than wide.
  icon: `${btnBase} size-9 border border-border-200 bg-transparent text-teal-700 phone:h-13 phone:w-12`,
}

// Filter chips toggle through aria-pressed, which also carries their state to
// screen readers.
const chipBase =
  'inline-flex h-8 cursor-pointer items-center gap-1 rounded-pill border border-border-200 bg-surface-100 text-caption text-ink-900 phone:shrink-0 aria-pressed:border-teal-500 aria-pressed:bg-teal-500 aria-pressed:font-semibold aria-pressed:text-on-teal'

export const chip = `${chipBase} pr-3 pl-2`
export const chipBare = `${chipBase} px-3`

export const textInput =
  'h-11 rounded-sm border border-border-200 bg-surface-100 px-3 text-body text-ink-900 phone:h-12'

export const scorePill =
  'inline-flex h-7 min-w-11 items-center justify-center rounded-sm px-2 text-body-sm leading-[normal] font-bold'

const bandFill: Record<ScoreBand, string> = {
  high: 'bg-score-high text-on-teal',
  mid: 'bg-score-mid text-on-amber',
  low: 'bg-score-low text-on-rust',
  none: 'bg-surface-200 text-ink-600',
}

export function bandClass(band: ScoreBand): string {
  return bandFill[band]
}

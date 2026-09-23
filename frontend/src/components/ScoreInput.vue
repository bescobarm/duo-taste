<script setup lang="ts">
import { bandFor } from '../categories'
import { bandClass, scorePill } from '../ui'

const score = defineModel<number>({ required: true })

defineProps<{ id: string }>()
</script>

<template>
  <div class="flex flex-col">
    <div class="flex items-center gap-2">
      <label :for="id" class="text-caption text-ink-600">Score</label>
      <span class="grow"></span>
      <input
        :id="id"
        v-model.number="score"
        :class="[
          scorePill,
          bandClass(bandFor(score)),
          'w-19 border-0 text-center [appearance:textfield] [&::-webkit-inner-spin-button]:m-0 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:m-0 [&::-webkit-outer-spin-button]:appearance-none',
        ]"
        type="number"
        min="1"
        max="5"
        step="0.01"
        inputmode="decimal"
        required
      />
    </div>
    <input
      v-model.number="score"
      class="m-0 h-5 w-full accent-teal-700"
      type="range"
      min="1"
      max="5"
      step="0.01"
      aria-hidden="true"
      tabindex="-1"
    />
    <div class="flex justify-between text-[13px]/[18px] text-ink-600">
      <span>1</span><span>2</span><span>3</span><span>4</span><span>5</span>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ size?: number }>(), { size: 36 })
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 64 64" role="img" aria-label="DuoTaste">
    <defs>
      <mask id="dt-mark">
        <rect x="0" y="0" width="64" height="64" rx="18" fill="#ffffff" />
        <circle cx="60" cy="6" r="13" fill="#000000" />
      </mask>
    </defs>
    <rect x="0" y="0" width="64" height="64" rx="18" fill="#0f6f74" mask="url(#dt-mark)" />
    <path
      fill="#ffffff"
      fill-rule="evenodd"
      d="M12 20h5v28h-5zM17 20a14 14 0 0 1 0 28zM17 28a6 6 0 0 1 0 12z"
    />
    <path fill="#fab242" d="M34 20h18v6h-6v22h-6V26h-6z" />
  </svg>
</template>

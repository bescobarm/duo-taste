<script setup lang="ts">
// One hidden sprite mounted once at the app root; every icon is a <use href>
// into it. Paths are the stroked 24px set from the DuoTaste canvas.
</script>

<template>
  <svg width="0" height="0" class="absolute" aria-hidden="true" focusable="false">
    <symbol id="ic-pizza" viewBox="0 0 24 24">
      <path d="M12 3.2 4.3 19.4a17.6 17.6 0 0 0 15.4 0z" />
      <circle cx="10" cy="11.4" r="1.1" />
      <circle cx="14.1" cy="14.8" r="1.1" />
    </symbol>
    <symbol id="ic-burger" viewBox="0 0 24 24">
      <path d="M4 10.2a8 4.6 0 0 1 16 0z" />
      <path d="M4.4 13.2h15.2" />
      <path d="M4 16.2h16a3.6 3.6 0 0 1-3.6 3.6H7.6A3.6 3.6 0 0 1 4 16.2z" />
    </symbol>
    <symbol id="ic-sushi" viewBox="0 0 24 24">
      <circle cx="12" cy="12" r="8.2" />
      <circle cx="12" cy="12" r="3.4" />
    </symbol>
    <symbol id="ic-tacos" viewBox="0 0 24 24">
      <path d="M3.2 16.6a8.8 8.8 0 0 1 17.6 0z" />
      <path d="M7.8 12.6 9.3 14M12 10.8l1.3 1.6M16.4 12.8 15 14.2" />
    </symbol>
    <symbol id="ic-coffee" viewBox="0 0 24 24">
      <path d="M4.6 8.6h11.2v5.6a5.2 5.2 0 0 1-5.2 5.2h-.8a5.2 5.2 0 0 1-5.2-5.2z" />
      <path d="M15.8 10h1.6a2.6 2.6 0 0 1 0 5.2h-1.6" />
      <path d="M8 3.8v2.2M11.6 3.4v2.6" />
    </symbol>
    <symbol id="ic-dessert" viewBox="0 0 24 24">
      <path d="M8 11.6 12 20.4l4-8.8" />
      <path d="M7.4 11.6a4.6 4.6 0 0 1 9.2 0z" />
    </symbol>
    <symbol id="ic-breakfast" viewBox="0 0 24 24">
      <path
        d="M9.6 4.8c3.2-1.4 5.4 1 6.6 2.8s3.8 1.6 3.3 4.5-3.2 4.6-5.6 4.6-3.6 2.4-6.3.6S4.2 10.6 6.1 7.7a6 6 0 0 1 3.5-2.9z"
      />
      <circle cx="12.3" cy="10.9" r="2.4" />
    </symbol>
    <symbol id="ic-other" viewBox="0 0 24 24">
      <path d="M6.6 3.6v4.6a2.8 2.8 0 0 0 5.6 0V3.6" />
      <path d="M9.4 11v9.4" />
      <path d="M17.8 3.6c-1.7 1.5-2.6 3.4-2.6 5.6 0 1.7.9 2.8 2.6 3.2v8" />
    </symbol>
    <symbol id="ic-search" viewBox="0 0 24 24">
      <circle cx="10.8" cy="10.8" r="6.6" />
      <path d="m15.6 15.6 4.8 4.8" />
    </symbol>
    <symbol id="ic-plus" viewBox="0 0 24 24"><path d="M12 5v14M5 12h14" /></symbol>
    <symbol id="ic-close" viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18" /></symbol>
    <symbol id="ic-back" viewBox="0 0 24 24"><path d="M19 12H5M11 6l-6 6 6 6" /></symbol>
    <symbol id="ic-pin" viewBox="0 0 24 24">
      <path d="M12 21s6.6-6.1 6.6-10.4A6.6 6.6 0 0 0 5.4 10.6C5.4 14.9 12 21 12 21z" />
      <circle cx="12" cy="10.4" r="2.4" />
    </symbol>
    <symbol id="ic-locate" viewBox="0 0 24 24">
      <circle cx="12" cy="12" r="6.4" />
      <circle cx="12" cy="12" r="2" />
      <path d="M12 2.8v3M12 18.2v3M2.8 12h3M18.2 12h3" />
    </symbol>
  </svg>
</template>

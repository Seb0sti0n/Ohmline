<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<template>
  <header class="flex flex-col gap-1.5">
    <h1 class="text-display tracking-[-0.01em] [font-stretch:112%]">{{ title }}</h1>
    <p class="text-body text-ink-muted">Esta pantalla llega en la siguiente fase.</p>
  </header>
  <RouterLink
    to="/"
    class="w-fit text-[15px] font-semibold text-brand no-underline hover:text-brand-strong"
    >Volver al resumen</RouterLink
  >
</template>

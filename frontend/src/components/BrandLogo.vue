<script setup lang="ts">
withDefaults(defineProps<{ size?: number; textClass?: string }>(), {
  size: 30,
  textClass: 'text-[19px]',
})
</script>

<template>
  <div class="flex items-center gap-2.5">
    <svg :width="size" :height="size" viewBox="0 0 30 30" aria-hidden="true">
      <rect width="30" height="30" rx="7" fill="var(--color-brand)" />
      <path
        d="M5 19h5l3-9 4 12 3-7h5"
        fill="none"
        stroke="#FFFFFF"
        stroke-width="2.2"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </svg>
    <span class="font-bold [font-stretch:112%]" :class="textClass">Ohmline</span>
  </div>
</template>

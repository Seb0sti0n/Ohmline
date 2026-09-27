<script setup lang="ts">
defineProps<{ to: string; label: string }>()
</script>

<template>
  <RouterLink
    :to="to"
    class="inline-flex h-12 items-center gap-2.5 self-start rounded-md bg-brand px-[22px] text-[16px] leading-5 font-semibold text-white no-underline hover:bg-brand-strong"
  >
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="#FFFFFF"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <path d="M15 5l-7 7 7 7" />
    </svg>
    <span>{{ label }}</span>
  </RouterLink>
</template>

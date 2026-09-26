<script setup lang="ts">
defineProps<{ to: string; label: string }>()
</script>

<template>
  <RouterLink
    :to="to"
    class="inline-flex h-8 items-center gap-1.5 self-start text-[15px] font-semibold text-brand no-underline hover:text-brand-strong"
  >
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <path d="M15 5l-7 7 7 7" />
    </svg>
    {{ label }}
  </RouterLink>
</template>

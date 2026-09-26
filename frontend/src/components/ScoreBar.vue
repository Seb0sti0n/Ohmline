<script setup lang="ts">
withDefaults(defineProps<{ fraction: number; barClass?: string; thin?: boolean }>(), {
  barClass: 'bg-brand',
  thin: false,
})
</script>

<template>
  <span class="block rounded-[3px] bg-line" :class="thin ? 'h-[5px]' : 'h-1.5'" aria-hidden="true">
    <span
      class="block rounded-[3px]"
      :class="[barClass, thin ? 'h-[5px]' : 'h-1.5']"
      :style="{ width: `${Math.round(Math.max(0, Math.min(1, fraction)) * 100)}%` }"
    />
  </span>
</template>

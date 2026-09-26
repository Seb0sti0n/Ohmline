<script setup lang="ts">
import { computed } from 'vue'
import { sparkline } from '@/utils/sparkline'

const props = defineProps<{ values: number[]; baseline: number }>()
const shape = computed(() => sparkline(props.values, props.baseline))
</script>

<template>
  <svg width="96" height="28" viewBox="0 0 96 28" aria-hidden="true">
    <line
      x1="0"
      x2="96"
      :y1="shape.baselineY"
      :y2="shape.baselineY"
      stroke="var(--color-line)"
      stroke-dasharray="3 3"
    />
    <path
      :d="shape.path"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linejoin="round"
    />
  </svg>
</template>

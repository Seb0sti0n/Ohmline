<script setup lang="ts">
import type { AnomalyType } from '@/types/api'
import { TYPE_BADGE, TYPE_LABEL, TYPE_LABEL_SHORT } from '@/utils/labels'

defineProps<{ type: AnomalyType; short?: boolean }>()
</script>

<template>
  <span
    class="inline-flex h-6 items-center whitespace-nowrap rounded-pill px-2.5 text-caption"
    :class="TYPE_BADGE[type]"
  >
    {{ short ? TYPE_LABEL_SHORT[type] : TYPE_LABEL[type] }}
  </span>
</template>

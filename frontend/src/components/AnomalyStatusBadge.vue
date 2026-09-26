<script setup lang="ts">
import type { AnomalyStatus } from '@/types/api'
import { ANOMALY_STATUS_BADGE, ANOMALY_STATUS_LABEL } from '@/utils/labels'

defineProps<{ status: AnomalyStatus; large?: boolean }>()
</script>

<template>
  <span
    class="inline-flex items-center whitespace-nowrap font-semibold"
    :class="[
      ANOMALY_STATUS_BADGE[status],
      large ? 'h-8 rounded-2xl px-3.5 text-label' : 'h-6 rounded-pill px-2.5 text-caption',
    ]"
  >
    {{ ANOMALY_STATUS_LABEL[status] }}
  </span>
</template>

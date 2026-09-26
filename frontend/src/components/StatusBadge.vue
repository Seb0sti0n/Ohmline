<script setup lang="ts">
import type { MeterStatus } from '@/types/api'
import { STATUS_BADGE, STATUS_LABEL } from '@/utils/labels'

defineProps<{ status: MeterStatus }>()
</script>

<template>
  <span
    class="inline-flex h-6 items-center gap-1.5 whitespace-nowrap rounded-pill px-2.5 text-caption"
    :class="STATUS_BADGE[status]"
  >
    <span class="size-[7px] rounded-full bg-current" />
    {{ STATUS_LABEL[status] }}
  </span>
</template>

<script setup lang="ts">
import type { EventType } from '@/types/api'
import { EVENT_BADGE, EVENT_LABEL } from '@/utils/labels'

defineProps<{ type: EventType; plain?: boolean }>()
</script>

<template>
  <span
    class="inline-flex h-6 items-center gap-1.5 whitespace-nowrap rounded-pill px-2.5 text-caption"
    :class="plain ? 'bg-surface text-ink-muted' : EVENT_BADGE[type]"
  >
    <span class="size-[7px] rounded-full bg-current" />
    {{ EVENT_LABEL[type] }}
  </span>
</template>

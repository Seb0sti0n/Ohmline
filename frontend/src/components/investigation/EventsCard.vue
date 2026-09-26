<script setup lang="ts">
import EventTypeBadge from '@/components/EventTypeBadge.vue'
import type { EventMatch } from '@/types/api'
import { formatDataTime } from '@/utils/format'
import { eventText } from '@/utils/labels'

defineProps<{ events: EventMatch[] }>()
</script>

<template>
  <section class="flex flex-col gap-2.5 rounded-lg border border-line bg-surface px-6 py-[22px]">
    <h2 class="text-heading">Eventos relacionados</h2>
    <p v-if="!events.length" class="text-label font-normal text-ink-muted">
      No hay eventos del medidor cerca del inicio del cambio.
    </p>
    <div
      v-for="e in events"
      :key="e.timestamp + e.type"
      class="flex flex-col gap-1.5 rounded-md bg-ground px-3.5 py-3"
    >
      <div class="flex items-center justify-between gap-2">
        <span class="text-label font-semibold tabular-nums">{{ formatDataTime(e.timestamp) }}</span>
        <EventTypeBadge :type="e.type" plain />
      </div>
      <span class="text-label font-normal text-ink-muted">{{ eventText(e.description) }}</span>
      <span class="text-caption font-normal capitalize-first text-ink-muted">{{ e.reason }}.</span>
    </div>
    <p class="text-label font-normal leading-[1.45] text-ink-muted">
      Ventana revisada: 6 h antes y 2 h después del inicio del cambio.
    </p>
  </section>
</template>

<style scoped>
.capitalize-first::first-letter {
  text-transform: uppercase;
}
</style>

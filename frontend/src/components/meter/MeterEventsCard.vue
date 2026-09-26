<script setup lang="ts">
import EventTypeBadge from '@/components/EventTypeBadge.vue'
import type { MeterEvent } from '@/types/api'
import { formatDataTime } from '@/utils/format'
import { eventText } from '@/utils/labels'

defineProps<{ events: MeterEvent[] }>()
</script>

<template>
  <section class="flex flex-col gap-3 rounded-lg border border-line bg-surface px-[22px] py-[18px]">
    <h2 class="text-[16px] font-[650]">Eventos del medidor</h2>
    <p v-if="!events.length" class="text-body text-ink-muted">
      Este medidor no tiene eventos registrados.
    </p>
    <div v-for="e in events" :key="e.timestamp + e.type" class="flex items-center gap-4">
      <span class="w-[110px] shrink-0 text-body tabular-nums text-ink-muted">{{
        formatDataTime(e.timestamp)
      }}</span>
      <EventTypeBadge :type="e.type" />
      <span class="text-body">{{ eventText(e.description) }}</span>
    </div>
  </section>
</template>

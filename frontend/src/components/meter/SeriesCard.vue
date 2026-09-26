<script setup lang="ts">
import { computed } from 'vue'
import HourlySeriesChart from '@/components/HourlySeriesChart.vue'
import type { ReadingPoint } from '@/types/api'
import { summarizeSeries } from '@/utils/anomalies'

const props = defineProps<{
  title: string
  metric: 'voltage' | 'current' | 'pf'
  points: ReadingPoint[]
  window: { start: string; end: string } | null
  meterId: string
}>()

const summary = computed(() => summarizeSeries(props.points, props.metric, props.window))
</script>

<template>
  <section class="flex flex-col gap-2 rounded-lg border border-line bg-surface px-5 py-[18px]">
    <div class="flex items-baseline justify-between">
      <h3 class="text-[16px] font-[650]">{{ title }}</h3>
      <span
        class="text-label font-semibold"
        :class="summary.alert ? 'text-status-critical' : 'text-ink'"
        >{{ summary.delta }}</span
      >
    </div>
    <span class="text-caption font-normal text-ink-muted">{{ summary.sub }}</span>
    <HourlySeriesChart
      :points="points"
      :metric="metric"
      :window="window"
      :description="`${title} horario de ${meterId}`"
    />
  </section>
</template>

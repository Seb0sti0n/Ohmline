<script setup lang="ts">
import { computed } from 'vue'
import type { AnalysisRun, Anomaly, MeterSummary } from '@/types/api'
import { signed, signedPct } from '@/utils/anomalies'
import { formatDataTime, formatDayMonth, formatKwh, formatLocalTime } from '@/utils/format'
import { STATUS_LABEL, STATUS_TEXT } from '@/utils/labels'

const props = defineProps<{
  meter: MeterSummary
  anomaly: Anomaly | null
  run: AnalysisRun | null
}>()

const lastDay = computed(() => {
  const d = new Date(`${props.meter.daily_from}T00:00:00Z`)
  d.setUTCDate(d.getUTCDate() + props.meter.daily_kwh.length - 1)
  return d.toISOString().slice(0, 10)
})
const firstDays = computed(
  () =>
    `${new Date(`${props.meter.daily_from}T00:00:00Z`).getUTCDate()}–${formatDayMonth(lastDayOfBaseline.value)}`,
)
const lastDayOfBaseline = computed(() => {
  const d = new Date(`${props.meter.daily_from}T00:00:00Z`)
  d.setUTCDate(d.getUTCDate() + 6)
  return d.toISOString().slice(0, 10)
})

const extra = computed(() => {
  const e = props.anomaly?.evidence
  if (!e || e.kind !== 'consumption_shift') return { value: '—', sub: 'Sin cambio de consumo' }
  return {
    value: `${signed(e.metrics.extra_kwh, 0).replace(/^\+/, '')} kWh`,
    sub: e.window.ongoing
      ? `Desde el ${formatDataTime(e.window.start)}`
      : 'Durante la ventana anómala',
  }
})

const analysed = computed(() =>
  props.run?.status === 'COMPLETED' && props.run.finished_at
    ? formatLocalTime(props.run.finished_at)
    : null,
)
const alertVariation = computed(() => Math.abs(props.meter.variation_pct) > 10)
</script>

<template>
  <section
    aria-label="Indicadores"
    class="grid grid-cols-5 rounded-lg border border-line bg-surface"
  >
    <div class="flex flex-col gap-1.5 border-r border-line px-[22px] py-[18px]">
      <span class="text-label font-normal text-ink-muted">Consumo último día</span>
      <span class="text-[28px] font-bold tabular-nums">{{ formatKwh(meter.consumption_kwh) }}</span>
      <span class="text-caption font-normal text-ink-muted">{{ formatDayMonth(lastDay) }}</span>
    </div>
    <div class="flex flex-col gap-1.5 border-r border-line px-[22px] py-[18px]">
      <span class="text-label font-normal text-ink-muted">Baseline diario</span>
      <span class="text-[28px] font-bold tabular-nums">{{ formatKwh(meter.baseline_kwh) }}</span>
      <span class="text-caption font-normal text-ink-muted">Mediana horaria, {{ firstDays }}</span>
    </div>
    <div class="flex flex-col gap-1.5 border-r border-line px-[22px] py-[18px]">
      <span class="text-label font-normal text-ink-muted">Variación</span>
      <span
        class="text-[28px] font-bold tabular-nums"
        :class="alertVariation ? 'text-status-critical' : ''"
        >{{ signedPct(meter.variation_pct) }}</span
      >
      <span class="text-caption font-normal text-ink-muted">vs baseline</span>
    </div>
    <div class="flex flex-col gap-1.5 border-r border-line px-[22px] py-[18px]">
      <span class="text-label font-normal text-ink-muted">Energía extra</span>
      <span class="text-[28px] font-bold tabular-nums">{{ extra.value }}</span>
      <span class="text-caption font-normal text-ink-muted">{{ extra.sub }}</span>
    </div>
    <div class="flex flex-col gap-1.5 px-[22px] py-[18px]">
      <span class="text-label font-normal text-ink-muted">Estado</span>
      <span class="text-[28px] font-bold" :class="STATUS_TEXT[meter.status]">{{
        STATUS_LABEL[meter.status]
      }}</span>
      <span class="text-caption font-normal text-ink-muted">{{
        analysed ? `Último análisis ${analysed.day}, ${analysed.time}` : 'Sin analizar'
      }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DailyBaselineChart from '@/components/DailyBaselineChart.vue'
import type { Anomaly, MeterSummary } from '@/types/api'
import { formatKwh, formatPeriod } from '@/utils/format'

const props = defineProps<{ anomaly: Anomaly; meter: MeterSummary }>()

const last = computed(() => {
  const d = new Date(`${props.meter.daily_from}T00:00:00Z`)
  d.setUTCDate(d.getUTCDate() + props.meter.daily_kwh.length - 1)
  return d.toISOString().slice(0, 10)
})

const energy = computed(() => {
  const extra = props.anomaly.evidence?.metrics.extra_kwh ?? 0
  if (Math.abs(extra) < 1) return ''
  return extra > 0
    ? ` Energía extra acumulada: ${formatKwh(extra)}.`
    : ` Energía no consumida durante la ventana: ${formatKwh(-extra)}.`
})
</script>

<template>
  <section class="flex flex-col gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]">
    <h2 class="text-heading">Comparación contra baseline</h2>
    <DailyBaselineChart
      :values="meter.daily_kwh"
      :from="meter.daily_from"
      :baseline="meter.baseline_kwh"
      :description="`Consumo diario de ${meter.meter_id} frente a su baseline de ${formatKwh(meter.baseline_kwh)}`"
    />
    <p class="text-label font-normal text-ink-muted">
      Consumo diario, {{ formatPeriod(meter.daily_from, last) }}. En rojo, días que se desvían más
      de un 20% del baseline.{{ energy }}
    </p>
  </section>
</template>

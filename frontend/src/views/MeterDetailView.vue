<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import BackLink from '@/components/BackLink.vue'
import ErrorState from '@/components/ErrorState.vue'
import HourlySeriesChart from '@/components/HourlySeriesChart.vue'
import MeterEventsCard from '@/components/meter/MeterEventsCard.vue'
import MeterKpis from '@/components/meter/MeterKpis.vue'
import SeriesCard from '@/components/meter/SeriesCard.vue'
import VerdictBar from '@/components/meter/VerdictBar.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import { useMeterDetailStore } from '@/stores/meterDetail'
import { formatDataTime } from '@/utils/format'

const route = useRoute()
const detail = useMeterDetailStore()

const meterId = computed(() => String(route.params.meterId))
watch(meterId, (id) => void detail.load(id), { immediate: true })

// The window to shade: the anomaly's own window, if the latest analysis found one.
const window = computed(() => {
  const w = detail.anomaly?.evidence?.window
  return w ? { start: w.start, end: w.end } : null
})
const markerLabel = computed(() =>
  window.value ? `${formatDataTime(window.value.start)}: inicio del cambio` : '',
)
</script>

<template>
  <BackLink to="/meters" label="Medidores" />

  <section
    v-if="detail.notFound"
    class="flex flex-col items-start gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
  >
    <p class="text-[16px] font-semibold">No existe el medidor {{ meterId }}</p>
    <RouterLink
      to="/meters"
      class="inline-flex h-11 items-center rounded-md bg-brand px-4 text-[15px] font-semibold text-white no-underline hover:bg-brand-strong"
      >Ver medidores</RouterLink
    >
  </section>
  <ErrorState v-else-if="detail.error" :message="detail.error" @retry="detail.load(meterId)" />
  <p v-else-if="detail.loading || !detail.meter" class="text-body text-ink-muted">
    Cargando medidor…
  </p>

  <template v-else>
    <header class="flex flex-col gap-2">
      <div class="flex items-center gap-3.5">
        <h1 class="text-display [font-stretch:112%]">{{ detail.meter.meter_id }}</h1>
        <StatusBadge :status="detail.meter.status" />
      </div>
      <p class="text-body text-ink-muted">{{ detail.meter.name }}, {{ detail.meter.location }}</p>
    </header>

    <VerdictBar :anomaly="detail.anomaly" :run="detail.run" />
    <MeterKpis :meter="detail.meter" :anomaly="detail.anomaly" :run="detail.run" />

    <section
      class="flex flex-col gap-3 rounded-lg border border-line bg-surface px-[26px] pb-[18px] pt-[22px]"
    >
      <div class="flex items-baseline justify-between gap-4">
        <h2 class="text-heading">Consumo horario</h2>
        <div class="flex flex-wrap gap-[18px] text-caption font-normal text-ink-muted">
          <span class="inline-flex items-center gap-1.5">
            <svg width="18" height="4" aria-hidden="true">
              <line x1="0" x2="18" y1="2" y2="2" stroke="var(--color-ink)" stroke-width="2" />
            </svg>
            Consumo (kWh)
          </span>
          <span class="inline-flex items-center gap-1.5">
            <svg width="18" height="10" aria-hidden="true">
              <rect width="18" height="10" fill="var(--color-chart-band)" />
              <line
                x1="0"
                x2="18"
                y1="5"
                y2="5"
                stroke="var(--color-brand)"
                stroke-dasharray="4 3"
              />
            </svg>
            Baseline y rango esperado
          </span>
          <span v-if="window" class="inline-flex items-center gap-1.5">
            <svg width="18" height="10" aria-hidden="true">
              <rect width="18" height="10" fill="var(--color-status-critical-bg)" />
            </svg>
            Ventana anómala
          </span>
        </div>
      </div>
      <HourlySeriesChart
        :points="detail.points"
        metric="consumption"
        large
        :height="300"
        :window="window"
        :marker-label="markerLabel"
        :description="`Consumo horario de ${detail.meter.meter_id} frente a su banda de baseline`"
      />
    </section>

    <div class="grid grid-cols-3 gap-4">
      <SeriesCard
        title="Voltaje (V)"
        metric="voltage"
        :points="detail.points"
        :window="window"
        :meter-id="detail.meter.meter_id"
      />
      <SeriesCard
        title="Corriente (A)"
        metric="current"
        :points="detail.points"
        :window="window"
        :meter-id="detail.meter.meter_id"
      />
      <SeriesCard
        title="Factor de potencia"
        metric="pf"
        :points="detail.points"
        :window="window"
        :meter-id="detail.meter.meter_id"
      />
    </div>

    <MeterEventsCard :events="detail.events" />
  </template>
</template>

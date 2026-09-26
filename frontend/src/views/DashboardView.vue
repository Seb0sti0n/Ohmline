<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'
import AnalysisCard from '@/components/AnalysisCard.vue'
import DailyConsumptionChart from '@/components/DailyConsumptionChart.vue'
import ErrorState from '@/components/ErrorState.vue'
import KpiStrip from '@/components/KpiStrip.vue'
import MeterTiles from '@/components/MeterTiles.vue'
import PriorityList from '@/components/PriorityList.vue'
import { useAnalysisStore } from '@/stores/analysis'
import { useDashboardStore } from '@/stores/dashboard'
import { formatKwh, formatPeriod } from '@/utils/format'
import { buildKpis } from '@/utils/kpis'

const dashboard = useDashboardStore()
const analysis = useAnalysisStore()

onMounted(() => {
  void dashboard.load()
  void analysis.loadLatest() // also resumes following a run that is still in progress
})
// The polling loop belongs to whoever started it: leaving the screen cancels it.
onBeforeUnmount(() => analysis.stop())

const summary = computed(() => dashboard.summary)
// The live run can be newer than what the last summary knows.
const run = computed(() => analysis.run ?? summary.value?.last_analysis ?? null)
const kpis = computed(() => (summary.value ? buildKpis(summary.value, run.value) : []))
const days = computed(() => summary.value?.daily_consumption ?? [])
const baseline = computed(() => days.value[0]?.baseline_kwh ?? 0)
const subtitle = computed(() => {
  if (!summary.value || days.value.length === 0) return ''
  return `${summary.value.meters_count} medidores, ${formatPeriod(days.value[0].date, days.value[days.value.length - 1].date)}`
})
</script>

<template>
  <header class="flex items-end justify-between gap-6">
    <div class="flex flex-col gap-1.5">
      <h1 class="text-display tracking-[-0.01em] [font-stretch:112%]">Resumen operativo</h1>
      <p class="text-body text-ink-muted">{{ subtitle }}</p>
    </div>
    <button
      type="button"
      :disabled="analysis.isRunning || !summary"
      class="inline-flex h-12 cursor-pointer items-center gap-2.5 rounded-md bg-brand px-[22px] text-[16px] font-semibold text-white hover:bg-brand-strong disabled:cursor-not-allowed disabled:opacity-70"
      @click="analysis.start()"
    >
      <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="#FFFFFF"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M7 4l13 8-13 8V4z" />
      </svg>
      <span>{{ analysis.isRunning ? 'Analizando…' : 'Run AI Analysis' }}</span>
    </button>
  </header>

  <ErrorState
    v-if="dashboard.error && !summary"
    :message="dashboard.error"
    @retry="dashboard.load()"
  />
  <p v-else-if="!summary" class="text-body text-ink-muted">Cargando resumen…</p>

  <template v-else>
    <AnalysisCard :run="run" :by-type="summary.anomalies_by_type" :error="analysis.error" />

    <KpiStrip :kpis="kpis" />

    <div class="grid grid-cols-[minmax(0,1.35fr)_minmax(0,1fr)] gap-6">
      <section
        class="flex flex-col gap-3.5 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
      >
        <div class="flex items-baseline justify-between">
          <h2 class="text-heading">Consumo total diario</h2>
          <span class="inline-flex items-center gap-2 text-caption font-normal text-ink-muted">
            <svg width="22" height="4" aria-hidden="true">
              <line
                x1="0"
                x2="22"
                y1="2"
                y2="2"
                stroke="var(--color-ink)"
                stroke-dasharray="5 4"
                stroke-width="1.5"
              />
            </svg>
            Baseline {{ formatKwh(baseline) }}/día
          </span>
        </div>
        <DailyConsumptionChart :points="days" />
        <p class="text-label font-normal text-ink-muted">
          Cada barra es un día del periodo. En verde oscuro, días más de un 3% por encima del
          baseline.
        </p>
      </section>

      <section
        class="flex flex-col gap-3.5 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
      >
        <h2 class="text-heading">Qué atender primero</h2>
        <PriorityList :items="summary.top_priorities" :has-results="summary.has_results" />
      </section>
    </div>

    <section
      aria-label="Estado de los medidores"
      class="flex flex-col gap-3.5 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
    >
      <div class="flex items-baseline justify-between">
        <h2 class="text-heading">Estado de los medidores</h2>
        <RouterLink
          to="/meters"
          class="text-[15px] font-semibold text-brand no-underline hover:text-brand-strong"
          >Ver todos</RouterLink
        >
      </div>
      <MeterTiles :meters="summary.meters" :evaluated="summary.has_results" />
    </section>
  </template>
</template>

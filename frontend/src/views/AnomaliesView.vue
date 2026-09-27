<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'
import AnomaliesTable from '@/components/AnomaliesTable.vue'
import ErrorState from '@/components/ErrorState.vue'
import TypeBadge from '@/components/TypeBadge.vue'
import { useAnalysisStore } from '@/stores/analysis'
import { useAnomaliesStore } from '@/stores/anomalies'
import type { AnomalyType } from '@/types/api'
import { formatInt, formatLocalTime } from '@/utils/format'
import { TYPE_HELP } from '@/utils/labels'

const anomalies = useAnomaliesStore()
const analysis = useAnalysisStore()

onMounted(() => {
  void anomalies.load()
  void analysis.loadLatest()
})
onBeforeUnmount(() => analysis.stop())

const list = computed(() => anomalies.list)
const high = computed(() => list.value.filter((a) => a.severity === 'HIGH').length)
const discarded = computed(() => list.value.filter((a) => a.type === 'FALSE_POSITIVE').length)

const run = computed(() => analysis.run)
const hasResults = computed(() => run.value?.status === 'COMPLETED' || list.value.length > 0)
const subtitle = computed(() => {
  const r = run.value
  if (r?.status !== 'COMPLETED' || !r.finished_at || !r.summary) return 'Ordenadas por prioridad.'
  const when = formatLocalTime(r.finished_at)
  return `Ordenadas por prioridad. Análisis ${when.day === 'hoy' ? 'de hoy' : `del ${when.day}`} a las ${when.time} sobre ${formatInt(r.summary.readings_count)} lecturas.`
})

const legend = Object.keys(TYPE_HELP) as AnomalyType[]
</script>

<template>
  <header class="flex items-end justify-between gap-6">
    <div class="flex flex-col gap-1.5">
      <h1 class="text-display tracking-[-0.01em] [font-stretch:112%]">Anomalías IA</h1>
      <p class="text-body text-ink-muted">{{ subtitle }}</p>
    </div>
    <div v-if="list.length" class="flex gap-7">
      <div class="flex flex-col items-end">
        <span class="text-[28px] font-bold">{{ list.length }}</span>
        <span class="text-caption font-normal text-ink-muted">{{
          list.length === 1 ? 'detectada' : 'detectadas'
        }}</span>
      </div>
      <div class="flex flex-col items-end">
        <span class="text-[28px] font-bold" :class="high > 0 ? 'text-status-critical' : ''">{{
          high
        }}</span>
        <span class="text-caption font-normal text-ink-muted">alta prioridad</span>
      </div>
      <div class="flex flex-col items-end">
        <span class="text-[28px] font-bold">{{ discarded }}</span>
        <span class="text-caption font-normal text-ink-muted">{{
          discarded === 1 ? 'descartada' : 'descartadas'
        }}</span>
      </div>
    </div>
  </header>

  <ErrorState v-if="anomalies.error" :message="anomalies.error" @retry="anomalies.load()" />
  <p v-else-if="anomalies.loading && !list.length" class="text-body text-ink-muted">
    Cargando anomalías…
  </p>

  <section
    v-else-if="!list.length"
    class="flex flex-col items-start gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
  >
    <p class="text-[16px] font-semibold">
      {{ hasResults ? 'No hay anomalías que atender' : 'Aún no hay anomalías' }}
    </p>
    <p class="max-w-[520px] text-label font-normal text-ink-muted">
      {{
        hasResults
          ? 'El último análisis no encontró anomalías en ninguno de los medidores.'
          : 'Ejecuta el análisis IA desde el resumen para detectar anomalías y ordenarlas según lo que requiere atención primero.'
      }}
    </p>
    <RouterLink
      to="/"
      class="inline-flex h-11 items-center rounded-md bg-brand px-4 text-[15px] font-semibold text-white no-underline hover:bg-brand-strong leading-5"
    >
      Ir al resumen
    </RouterLink>
  </section>

  <template v-else>
    <section class="overflow-hidden rounded-lg border border-line bg-surface">
      <AnomaliesTable :items="list" />
    </section>

    <section aria-label="Tipos de anomalía" class="grid grid-cols-4 gap-4">
      <div
        v-for="t in legend"
        :key="t"
        class="flex flex-col gap-1.5 rounded-[12px] border border-line bg-surface px-[18px] py-4"
      >
        <TypeBadge :type="t" class="self-start" />
        <span class="text-label font-normal leading-[1.4] text-ink-muted">{{ TYPE_HELP[t] }}</span>
      </div>
    </section>
  </template>
</template>

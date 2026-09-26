<script setup lang="ts">
import { computed } from 'vue'
import type { Evidence } from '@/types/api'
import { changeRows, qualityRows } from '@/utils/anomalies'
import { formatDataTime, formatInt } from '@/utils/format'

const props = defineProps<{ evidence: Evidence }>()

const isQuality = computed(() => props.evidence.kind === 'data_quality')
const rows = computed(() => changeRows(props.evidence))
const checks = computed(() => qualityRows(props.evidence))
const since = computed(() => {
  const w = props.evidence.window
  return w.ongoing ? `Desde el ${formatDataTime(w.start).split(', ')[0]}` : 'Durante la ventana'
})
const th = 'pb-2 text-caption text-ink-muted'
const td = 'border-t border-line py-3 text-body'
</script>

<template>
  <section class="flex flex-col gap-2 rounded-lg border border-line bg-surface px-[26px] py-[22px]">
    <h2 class="mb-1 text-heading">
      {{ isQuality ? 'Lecturas inconsistentes' : 'Variables que cambiaron' }}
    </h2>

    <table v-if="!isQuality" class="w-full border-collapse">
      <thead>
        <tr>
          <th scope="col" :class="[th, 'text-left']">Variable</th>
          <th scope="col" :class="[th, 'text-right']">Antes (baseline)</th>
          <th scope="col" :class="[th, 'text-right']">{{ since }}</th>
          <th scope="col" :class="[th, 'text-right']">Cambio</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.label">
          <td :class="td">{{ r.label }}</td>
          <td :class="[td, 'text-right tabular-nums text-ink-muted']">{{ r.before }}</td>
          <td :class="[td, 'text-right font-semibold tabular-nums']">{{ r.after }}</td>
          <td
            :class="[
              td,
              'text-right font-[650] tabular-nums',
              r.significant ? 'text-status-critical' : 'text-ink',
            ]"
          >
            {{ r.change }}
          </td>
        </tr>
      </tbody>
    </table>

    <template v-else>
      <p class="text-label font-normal text-ink-muted">
        El consumo se mantiene en su baseline, pero
        {{ formatInt(evidence.quality?.flagged_hours ?? 0) }} de
        {{ formatInt(evidence.quality?.window_hours ?? 0) }} horas tienen lecturas eléctricas que no
        pasan las comprobaciones de calidad.
      </p>
      <table class="w-full border-collapse">
        <thead>
          <tr>
            <th scope="col" :class="[th, 'text-left']">Comprobación</th>
            <th scope="col" :class="[th, 'text-right']">Lecturas afectadas</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in checks" :key="c.label">
            <td :class="td">{{ c.label }}</td>
            <td :class="[td, 'text-right font-[650] tabular-nums text-status-critical']">
              {{ formatInt(c.count) }}
            </td>
          </tr>
        </tbody>
      </table>
    </template>
  </section>
</template>

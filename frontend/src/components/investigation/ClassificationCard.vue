<script setup lang="ts">
import { computed } from 'vue'
import ScoreBar from '@/components/ScoreBar.vue'
import SeverityBadge from '@/components/SeverityBadge.vue'
import TypeBadge from '@/components/TypeBadge.vue'
import type { Anomaly } from '@/types/api'
import { confidenceBase, confidenceLabel, confidenceParts, ordinal } from '@/utils/anomalies'
import { formatDecimal, formatInt } from '@/utils/format'

const props = defineProps<{ anomaly: Anomaly; rank: { rank: number; total: number } | null }>()

const parts = computed(() =>
  props.anomaly.evidence ? confidenceParts(props.anomaly.evidence, props.anomaly.type) : [],
)
const base = computed(() => (props.anomaly.evidence ? confidenceBase(props.anomaly.evidence) : 0))
</script>

<template>
  <section class="flex flex-col gap-4 rounded-lg border border-line bg-surface px-6 py-[22px]">
    <h2 class="text-heading">Clasificación</h2>
    <div class="flex flex-wrap gap-2">
      <TypeBadge :type="anomaly.type" />
      <SeverityBadge :severity="anomaly.severity" long />
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-0.5">
        <span class="text-caption font-normal text-ink-muted">Prioridad</span>
        <span class="text-kpi tabular-nums"
          >{{ formatInt(anomaly.priority_score)
          }}<span class="text-body font-medium text-ink-muted">/100</span></span
        >
        <span class="text-caption font-normal text-ink-muted">{{
          rank ? ordinal(rank.rank, rank.total) : ''
        }}</span>
      </div>
      <div class="flex flex-col gap-0.5">
        <span class="text-caption font-normal text-ink-muted">Confianza</span>
        <span class="text-kpi tabular-nums">{{ formatDecimal(anomaly.confidence, 2) }}</span>
        <span class="text-caption font-normal text-ink-muted">{{
          confidenceLabel(anomaly.confidence)
        }}</span>
      </div>
    </div>

    <div v-if="parts.length" class="flex flex-col gap-3 border-t border-line pt-3.5">
      <span class="text-label font-[650]">Por qué esta confianza</span>
      <div v-for="p in parts" :key="p.key" class="flex flex-col gap-[5px]">
        <div class="flex justify-between text-label font-normal">
          <span>{{ p.label }}</span>
          <span class="font-semibold tabular-nums">{{ formatDecimal(p.fraction, 2) }}</span>
        </div>
        <ScoreBar :fraction="p.fraction" />
      </div>
      <p class="text-caption font-normal leading-[1.4] text-ink-muted">
        La confianza parte de {{ formatDecimal(base, 2) }} y suma estos tres componentes según su
        peso, hasta un máximo de 0,99.
      </p>
    </div>
  </section>
</template>

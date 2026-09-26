<script setup lang="ts">
import type { Anomaly } from '@/types/api'
import { confidenceLabel } from '@/utils/anomalies'
import { formatDecimal, formatInt } from '@/utils/format'
import { ACTION_LABEL } from '@/utils/labels'
import AnomalyStatusBadge from './AnomalyStatusBadge.vue'
import ScoreBar from './ScoreBar.vue'
import SeverityBadge from './SeverityBadge.vue'
import TypeBadge from './TypeBadge.vue'

defineProps<{ items: Anomaly[] }>()

const head = 'h-11 border-b border-line px-4 text-left text-caption text-ink-muted'
const cell = 'border-b border-line px-4 py-[18px] align-top text-body'
</script>

<template>
  <table class="w-full border-collapse">
    <thead>
      <tr class="bg-surface-muted">
        <th scope="col" :class="head">Prioridad</th>
        <th scope="col" :class="head">Medidor</th>
        <th scope="col" :class="head">Tipo</th>
        <th scope="col" :class="head">Severidad</th>
        <th scope="col" :class="head">Confianza</th>
        <th scope="col" :class="head">Motivo</th>
        <th scope="col" :class="head">Acción</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(a, i) in items" :key="a.id">
        <td :class="cell">
          <div class="flex items-center gap-3">
            <span
              class="w-5 text-[24px] font-bold"
              :class="a.severity === 'HIGH' ? 'text-status-critical' : 'text-ink-muted'"
              >{{ i + 1 }}</span
            >
            <div class="flex w-16 flex-col gap-1.5">
              <span class="text-caption font-normal tabular-nums text-ink-muted"
                >{{ formatInt(a.priority_score) }}/100</span
              >
              <ScoreBar
                :fraction="a.priority_score / 100"
                thin
                :bar-class="a.severity === 'HIGH' ? 'bg-status-critical' : 'bg-ink-muted'"
              />
            </div>
          </div>
        </td>
        <td :class="cell">
          <RouterLink
            :to="`/meters/${a.meter_id}`"
            class="flex flex-col gap-0.5 text-ink no-underline"
          >
            <span class="font-bold">{{ a.meter_id }}</span>
            <span class="text-caption font-normal text-ink-muted">{{ a.meter_name }}</span>
          </RouterLink>
        </td>
        <td :class="cell"><TypeBadge :type="a.type" /></td>
        <td :class="cell"><SeverityBadge :severity="a.severity" /></td>
        <td :class="cell">
          <div class="flex flex-col gap-0.5">
            <span class="font-[650]">{{ confidenceLabel(a.confidence) }}</span>
            <span class="text-caption font-normal tabular-nums text-ink-muted">{{
              formatDecimal(a.confidence, 2)
            }}</span>
          </div>
        </td>
        <td :class="[cell, 'max-w-[380px] leading-[1.45]']">{{ a.reason }}</td>
        <td :class="cell">
          <div class="flex flex-col items-start gap-2">
            <RouterLink
              :to="`/anomalies/${a.id}`"
              class="inline-flex h-10 items-center justify-center whitespace-nowrap rounded-[9px] border px-3.5 text-[14px] font-semibold no-underline"
              :class="
                a.severity === 'HIGH'
                  ? 'border-brand bg-brand text-white hover:bg-brand-strong'
                  : 'border-line bg-surface text-ink hover:bg-surface-muted'
              "
            >
              {{ ACTION_LABEL[a.type] }}
            </RouterLink>
            <AnomalyStatusBadge v-if="a.status !== 'OPEN'" :status="a.status" />
          </div>
        </td>
      </tr>
    </tbody>
  </table>
</template>

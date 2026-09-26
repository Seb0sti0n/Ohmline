<script setup lang="ts">
import type { MeterSort, MeterSummary, SortOrder } from '@/types/api'
import { formatKwh, formatPct } from '@/utils/format'
import { STATUS_TEXT } from '@/utils/labels'
import Sparkline from './Sparkline.vue'
import StatusBadge from './StatusBadge.vue'
import TypeBadge from './TypeBadge.vue'

defineProps<{ rows: MeterSummary[]; sort: MeterSort; order: SortOrder }>()
defineEmits<{ sort: [key: MeterSort] }>()

const head = 'h-11 border-b border-line px-4 text-left text-caption text-ink-muted'
const arrow = (active: boolean, order: SortOrder) => (active ? (order === 'desc' ? '↓' : '↑') : '')
</script>

<template>
  <table class="w-full border-collapse">
    <thead>
      <tr class="bg-surface-muted">
        <th scope="col" :class="head">Medidor</th>
        <th scope="col" :class="head">Ubicación</th>
        <th
          scope="col"
          :class="[head, 'text-right']"
          :aria-sort="
            sort === 'consumption' ? (order === 'desc' ? 'descending' : 'ascending') : 'none'
          "
        >
          <button
            type="button"
            class="inline-flex h-11 cursor-pointer items-center gap-1 text-caption"
            :class="sort === 'consumption' ? 'text-ink' : 'text-ink-muted'"
            @click="$emit('sort', 'consumption')"
          >
            Consumo<span aria-hidden="true">{{ arrow(sort === 'consumption', order) }}</span>
          </button>
        </th>
        <th scope="col" :class="[head, 'text-right']">Baseline</th>
        <th
          scope="col"
          :class="[head, 'text-right']"
          :aria-sort="
            sort === 'variation' ? (order === 'desc' ? 'descending' : 'ascending') : 'none'
          "
        >
          <button
            type="button"
            class="inline-flex h-11 cursor-pointer items-center gap-1 text-caption"
            :class="sort === 'variation' ? 'text-ink' : 'text-ink-muted'"
            @click="$emit('sort', 'variation')"
          >
            Variación<span aria-hidden="true">{{ arrow(sort === 'variation', order) }}</span>
          </button>
        </th>
        <th scope="col" :class="head">Últimos 14 días</th>
        <th
          scope="col"
          :class="head"
          :aria-sort="
            sort === 'severity' ? (order === 'desc' ? 'descending' : 'ascending') : 'none'
          "
        >
          <button
            type="button"
            class="inline-flex h-11 cursor-pointer items-center gap-1 text-caption"
            :class="sort === 'severity' ? 'text-ink' : 'text-ink-muted'"
            @click="$emit('sort', 'severity')"
          >
            Estado<span aria-hidden="true">{{ arrow(sort === 'severity', order) }}</span>
          </button>
        </th>
        <th scope="col" :class="head">Anomalía IA</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="m in rows" :key="m.meter_id">
        <td class="h-[60px] border-b border-line px-4 text-body">
          <RouterLink
            :to="`/meters/${m.meter_id}`"
            class="flex flex-col gap-0.5 text-ink no-underline"
          >
            <span class="font-bold">{{ m.meter_id }}</span>
            <span class="text-caption font-normal text-ink-muted">{{ m.name }}</span>
          </RouterLink>
        </td>
        <td class="h-[60px] border-b border-line px-4 text-body text-ink-muted">
          {{ m.location }}
        </td>
        <td
          class="h-[60px] border-b border-line px-4 text-right text-body font-semibold tabular-nums"
        >
          {{ formatKwh(m.consumption_kwh) }}
        </td>
        <td
          class="h-[60px] border-b border-line px-4 text-right text-body tabular-nums text-ink-muted"
        >
          {{ formatKwh(m.baseline_kwh) }}
        </td>
        <td
          class="h-[60px] border-b border-line px-4 text-right text-body font-semibold tabular-nums"
          :class="Math.abs(m.variation_pct) > 10 ? 'text-status-critical' : 'text-ink'"
        >
          {{ formatPct(m.variation_pct) }}
        </td>
        <td class="h-[60px] border-b border-line px-4">
          <span :class="STATUS_TEXT[m.status]"
            ><Sparkline :values="m.daily_kwh" :baseline="m.baseline_kwh"
          /></span>
        </td>
        <td class="h-[60px] border-b border-line px-4"><StatusBadge :status="m.status" /></td>
        <td class="h-[60px] border-b border-line px-4">
          <TypeBadge v-if="m.anomaly" :type="m.anomaly.type" />
          <span v-else class="text-ink-muted">—</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

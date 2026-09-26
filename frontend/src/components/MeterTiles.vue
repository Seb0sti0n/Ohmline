<script setup lang="ts">
import type { MeterSummary } from '@/types/api'
import { STATUS_LABEL, TILE_STYLE, TILE_UNEVALUATED } from '@/utils/labels'
import { formatPct } from '@/utils/format'

defineProps<{ meters: MeterSummary[]; evaluated: boolean }>()
</script>

<template>
  <div class="grid grid-cols-6 gap-2.5">
    <RouterLink
      v-for="m in meters"
      :key="m.meter_id"
      :to="`/meters/${m.meter_id}`"
      class="flex flex-col gap-1 rounded-md px-3.5 py-3 no-underline"
      :class="evaluated ? TILE_STYLE[m.status] : TILE_UNEVALUATED"
    >
      <span class="text-[15px] font-bold">{{ m.meter_id }}</span>
      <span class="text-caption font-normal opacity-85">
        {{ evaluated ? `${STATUS_LABEL[m.status]}, ${formatPct(m.variation_pct)}` : 'Sin evaluar' }}
      </span>
    </RouterLink>
  </div>
</template>

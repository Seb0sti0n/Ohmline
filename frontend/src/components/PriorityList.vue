<script setup lang="ts">
import type { Anomaly } from '@/types/api'
import TypeBadge from './TypeBadge.vue'

defineProps<{ items: Anomaly[]; hasResults: boolean }>()
</script>

<template>
  <div v-if="!hasResults" class="flex grow flex-col items-start justify-center gap-3 py-3">
    <p class="text-[16px] font-semibold">Aún no hay prioridades</p>
    <p class="max-w-[340px] text-label font-normal text-ink-muted">
      Ejecuta el análisis IA para detectar anomalías y ordenarlas según lo que requiere atención
      primero.
    </p>
  </div>

  <p v-else-if="items.length === 0" class="py-3 text-label font-normal text-ink-muted">
    El último análisis no encontró anomalías. No hay nada que atender.
  </p>

  <ol v-else class="m-0 flex list-none flex-col p-0">
    <li v-for="(a, i) in items" :key="a.id" class="border-t border-line">
      <RouterLink
        :to="`/anomalies/${a.id}`"
        class="flex items-center gap-3.5 py-3.5 text-ink no-underline"
      >
        <span
          class="w-[22px] text-[22px] font-bold"
          :class="a.severity === 'HIGH' ? 'text-status-critical' : 'text-ink-muted'"
          >{{ i + 1 }}</span
        >
        <span class="flex min-w-0 grow flex-col gap-[3px]">
          <span class="text-[16px] font-[650]">
            {{ a.meter_id }} <span class="font-medium text-ink-muted">{{ a.meter_name }}</span>
          </span>
          <span class="line-clamp-2 text-label font-normal text-ink-muted">{{ a.reason }}</span>
        </span>
        <TypeBadge :type="a.type" short />
      </RouterLink>
    </li>
  </ol>
</template>

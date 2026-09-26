<script setup lang="ts">
import type { Kpi } from '@/utils/kpis'

defineProps<{ kpis: Kpi[] }>()
</script>

<template>
  <section
    aria-label="Indicadores"
    class="grid grid-cols-6 rounded-lg border border-line bg-surface"
  >
    <div
      v-for="kpi in kpis"
      :key="kpi.label"
      class="flex min-w-0 flex-col gap-1.5 border-r border-line px-[22px] py-[18px] last:border-r-0"
    >
      <span class="text-label font-normal text-ink-muted">{{ kpi.label }}</span>
      <span
        class="flex flex-wrap items-baseline font-bold tabular-nums"
        :class="[
          kpi.compact ? 'text-[20px] leading-9 font-[650]' : 'text-kpi',
          kpi.tone === 'critical' ? 'text-status-critical' : '',
        ]"
      >
        {{ kpi.value
        }}<span v-if="kpi.unit" class="ml-1 text-[16px] font-medium text-ink-muted">{{
          kpi.unit
        }}</span>
      </span>
      <span class="text-caption font-normal text-ink-muted">{{ kpi.sub }}</span>
    </div>
  </section>
</template>

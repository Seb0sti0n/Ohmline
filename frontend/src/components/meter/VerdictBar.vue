<script setup lang="ts">
import { computed } from 'vue'
import type { AnalysisRun, Anomaly } from '@/types/api'
import { verdict } from '@/utils/anomalies'
import { SEVERITY_LABEL, TYPE_LABEL } from '@/utils/labels'

const props = defineProps<{ anomaly: Anomaly | null; run: AnalysisRun | null }>()

const analysed = computed(() => props.run?.status === 'COMPLETED')
const text = computed(() => {
  if (props.anomaly) {
    return verdict(
      props.anomaly,
      TYPE_LABEL[props.anomaly.type],
      SEVERITY_LABEL[props.anomaly.severity].toLowerCase(),
    )
  }
  if (analysed.value) {
    return {
      strong: 'Sin anomalías.',
      rest: 'El último análisis no encontró nada que atender en este medidor.',
    }
  }
  return {
    strong: 'Aún sin análisis.',
    rest: 'Ejecuta el análisis IA desde el resumen para ver qué encuentra en este medidor.',
  }
})
</script>

<template>
  <section
    class="flex items-center justify-between gap-5 rounded-lg bg-ink px-[22px] py-4 text-white"
  >
    <div class="flex items-center gap-3.5">
      <svg
        width="22"
        height="22"
        viewBox="0 0 24 24"
        fill="none"
        stroke="var(--color-brand-on-dark)"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path
          d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5 18 18M18 6l-2.5 2.5M8.5 15.5 6 18"
        />
      </svg>
      <p class="text-[16px] leading-[1.45]">
        <strong>{{ text.strong }}</strong> {{ text.rest }}
      </p>
    </div>
    <RouterLink
      v-if="anomaly"
      :to="`/anomalies/${anomaly.id}`"
      class="inline-flex h-11 shrink-0 items-center rounded-md bg-white px-[18px] text-[15px] font-semibold text-ink no-underline hover:opacity-90 leading-5"
    >
      Ver investigación
    </RouterLink>
    <RouterLink
      v-else-if="!analysed"
      to="/"
      class="inline-flex h-11 shrink-0 items-center rounded-md bg-white px-[18px] text-[15px] font-semibold text-ink no-underline hover:opacity-90 leading-5"
    >
      Ir al resumen
    </RouterLink>
  </section>
</template>

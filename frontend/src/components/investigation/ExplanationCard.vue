<script setup lang="ts">
import type { Anomaly } from '@/types/api'

defineProps<{ anomaly: Anomaly }>()
</script>

<template>
  <section class="flex flex-col gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]">
    <div class="flex items-center justify-between gap-4">
      <h2 class="text-heading">Qué encontró la IA</h2>
      <span class="inline-flex items-center gap-1.5 text-caption font-normal text-brand-strong">
        <svg
          width="15"
          height="15"
          viewBox="0 0 24 24"
          fill="none"
          stroke="var(--color-brand)"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path
            d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5 18 18M18 6l-2.5 2.5M8.5 15.5 6 18"
          />
        </svg>
        {{
          anomaly.explanation_source === 'LLM'
            ? 'Explicación generada por IA a partir de la evidencia'
            : 'Explicación generada a partir de la evidencia (plantilla)'
        }}
      </span>
    </div>
    <p class="max-w-[720px] text-body-lg">{{ anomaly.explanation }}</p>
  </section>
</template>

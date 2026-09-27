<script setup lang="ts">
import { computed } from 'vue'
import type { AnalysisRun, AnomalyType } from '@/types/api'
import { buildSteps, resultText, statusText } from '@/utils/analysis'

const props = defineProps<{
  run: AnalysisRun | null
  byType: Partial<Record<AnomalyType, number>>
  error: string | null
}>()

const steps = computed(() => buildSteps(props.run, props.byType))
const status = computed(() => statusText(props.run))
const result = computed(() => resultText(props.run))
</script>

<template>
  <section
    aria-label="Análisis IA"
    class="flex flex-col gap-[18px] rounded-lg border border-line bg-surface px-[26px] py-[22px]"
  >
    <div class="flex items-center justify-between">
      <h2 class="text-heading">Análisis IA</h2>
      <span class="text-label font-normal text-ink-muted" aria-live="polite">{{ status }}</span>
    </div>

    <ol class="m-0 grid list-none grid-cols-7 p-0">
      <li v-for="(step, i) in steps" :key="step.label" class="flex min-w-0 flex-col gap-2 pr-2">
        <div class="flex items-center">
          <span
            class="flex size-[30px] shrink-0 items-center justify-center rounded-full text-[13px] font-bold"
            :class="{
              'bg-brand text-white': step.state === 'done',
              'border-[3px] border-brand bg-surface text-brand': step.state === 'active',
              'border-2 border-line bg-surface text-ink-muted': step.state === 'pending',
            }"
          >
            <svg
              v-if="step.state === 'done'"
              width="15"
              height="15"
              viewBox="0 0 24 24"
              fill="none"
              stroke="#FFFFFF"
              stroke-width="2.6"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="m5 12 5 5 9-10" />
            </svg>
            <span v-else>{{ step.number }}</span>
          </span>
          <span
            v-if="i < steps.length - 1"
            class="mx-2 h-0.5 grow"
            :class="step.state === 'done' ? 'bg-brand' : 'bg-line'"
          />
        </div>
        <span
          class="text-body font-semibold"
          :class="step.state === 'pending' ? 'text-ink-muted' : 'text-ink'"
        >
          {{ step.label }}
          <span class="sr-only">{{
            step.state === 'done'
              ? '(completado)'
              : step.state === 'active'
                ? '(en curso)'
                : '(pendiente)'
          }}</span>
        </span>
        <span class="min-h-[18px] text-caption font-normal text-ink-muted">{{ step.detail }}</span>
      </li>
    </ol>

    <p
      v-if="error"
      role="alert"
      class="rounded-md bg-status-critical-bg px-[18px] py-3.5 text-[16px] font-semibold text-status-critical"
    >
      {{ error }}
    </p>

    <div
      v-if="result"
      class="flex items-center justify-between gap-4 rounded-md bg-brand-soft px-[18px] py-3.5"
    >
      <p class="text-[17px] font-semibold text-brand-strong">{{ result }}</p>
      <RouterLink
        to="/anomalies"
        class="inline-flex h-11 items-center text-[15px] font-semibold text-brand no-underline hover:text-brand-strong leading-5"
      >
        Ver anomalías
      </RouterLink>
    </div>
  </section>
</template>

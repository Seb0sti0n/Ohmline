<script setup lang="ts">
import type { Anomaly, AnomalyStatus } from '@/types/api'

defineProps<{ anomaly: Anomaly; updating: boolean; error: string | null }>()
defineEmits<{ set: [status: AnomalyStatus] }>()
</script>

<template>
  <section class="flex flex-col gap-3.5 rounded-lg bg-ink px-6 py-[22px] text-white">
    <h2 class="text-heading">Acción recomendada</h2>
    <p class="text-[16px] leading-normal">{{ anomaly.recommended_action }}</p>

    <div class="flex gap-2.5">
      <button
        v-if="anomaly.status === 'OPEN'"
        type="button"
        :disabled="updating"
        class="h-11 grow cursor-pointer rounded-md bg-brand-on-dark text-[15px] font-[650] text-ink hover:opacity-90 disabled:cursor-wait disabled:opacity-60 leading-5"
        @click="$emit('set', 'ACKNOWLEDGED')"
      >
        Marcar en investigación
      </button>
      <button
        v-else-if="anomaly.status === 'ACKNOWLEDGED'"
        type="button"
        disabled
        class="h-11 grow rounded-md bg-brand-on-dark text-[15px] font-[650] text-ink opacity-60 leading-5"
      >
        En investigación
      </button>
      <button
        v-else
        type="button"
        :disabled="updating"
        class="h-11 grow cursor-pointer rounded-md bg-brand-on-dark text-[15px] font-[650] text-ink hover:opacity-90 disabled:cursor-wait disabled:opacity-60 leading-5"
        @click="$emit('set', 'OPEN')"
      >
        Reabrir
      </button>

      <button
        v-if="anomaly.status !== 'RESOLVED'"
        type="button"
        :disabled="updating"
        class="h-11 cursor-pointer rounded-md border border-ink-muted bg-transparent px-4 text-[15px] font-semibold text-white hover:bg-sidebar-active disabled:cursor-wait disabled:opacity-60 leading-5"
        @click="$emit('set', 'RESOLVED')"
      >
        Resolver
      </button>
    </div>

    <p
      v-if="error"
      role="alert"
      class="rounded-md bg-status-critical-bg px-3 py-2 text-label font-semibold text-status-critical"
    >
      {{ error }}
    </p>
  </section>
</template>

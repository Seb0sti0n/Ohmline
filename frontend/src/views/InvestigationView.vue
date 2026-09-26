<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import AnomalyStatusBadge from '@/components/AnomalyStatusBadge.vue'
import BackLink from '@/components/BackLink.vue'
import ErrorState from '@/components/ErrorState.vue'
import ActionCard from '@/components/investigation/ActionCard.vue'
import BaselineCard from '@/components/investigation/BaselineCard.vue'
import ChangedVariablesCard from '@/components/investigation/ChangedVariablesCard.vue'
import ClassificationCard from '@/components/investigation/ClassificationCard.vue'
import EventsCard from '@/components/investigation/EventsCard.vue'
import EvidenceCard from '@/components/investigation/EvidenceCard.vue'
import ExplanationCard from '@/components/investigation/ExplanationCard.vue'
import { useInvestigationStore } from '@/stores/investigation'
import { investigationTitle, windowSentence } from '@/utils/anomalies'
import { formatLocalTime } from '@/utils/format'

const route = useRoute()
const inv = useInvestigationStore()

const id = computed(() => Number(route.params.id))
// Reload when navigating from one anomaly to another.
watch(id, (value) => Number.isInteger(value) && void inv.load(value), { immediate: true })

const evidence = computed(() => inv.anomaly?.evidence)
const subtitle = computed(() => {
  const a = inv.anomaly
  if (!a || !evidence.value) return ''
  const detected = formatLocalTime(a.detected_at)
  const when = detected.day === 'hoy' ? 'hoy' : `el ${detected.day}`
  return `${inv.meter?.name ?? a.meter_name}, ${inv.meter?.location ?? ''}. Detectada ${when} a las ${detected.time}; ${windowSentence(evidence.value)}.`
})
</script>

<template>
  <BackLink to="/anomalies" label="Anomalías IA" />

  <section
    v-if="inv.notFound"
    class="flex flex-col items-start gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
  >
    <p class="text-[16px] font-semibold">No encontramos esta anomalía</p>
    <p class="max-w-[520px] text-label font-normal text-ink-muted">
      Puede haberse reemplazado por un análisis más reciente. Vuelve a la lista para ver las
      actuales.
    </p>
    <RouterLink
      to="/anomalies"
      class="inline-flex h-11 items-center rounded-md bg-brand px-4 text-[15px] font-semibold text-white no-underline hover:bg-brand-strong"
      >Ver anomalías</RouterLink
    >
  </section>
  <ErrorState v-else-if="inv.error" :message="inv.error" @retry="inv.load(id)" />
  <p
    v-else-if="inv.loading || !inv.anomaly || !inv.meter || !evidence"
    class="text-body text-ink-muted"
  >
    Cargando investigación…
  </p>

  <template v-else>
    <header class="flex items-end justify-between gap-6">
      <div class="flex flex-col gap-2">
        <h1 class="text-display [font-stretch:112%]">{{ investigationTitle(inv.anomaly) }}</h1>
        <p class="text-body text-ink-muted">
          <RouterLink
            :to="`/meters/${inv.anomaly.meter_id}`"
            class="font-semibold text-brand no-underline hover:text-brand-strong"
            >{{ inv.anomaly.meter_id }}</RouterLink
          >,
          {{ subtitle }}
        </p>
      </div>
      <AnomalyStatusBadge :status="inv.anomaly.status" large />
    </header>

    <div class="grid grid-cols-[minmax(0,1fr)_380px] items-start gap-[22px]">
      <div class="flex min-w-0 flex-col gap-[22px]">
        <ExplanationCard :anomaly="inv.anomaly" />
        <BaselineCard :anomaly="inv.anomaly" :meter="inv.meter" />
        <ChangedVariablesCard :evidence="evidence" />
      </div>
      <div class="flex min-w-0 flex-col gap-[22px]">
        <ClassificationCard :anomaly="inv.anomaly" :rank="inv.rank" />
        <ActionCard
          :anomaly="inv.anomaly"
          :updating="inv.updating"
          :error="inv.updateError"
          @set="inv.setStatus($event)"
        />
        <EventsCard :events="evidence.events" />
        <EvidenceCard :rules="evidence.rules_fired" />
      </div>
    </div>
  </template>
</template>

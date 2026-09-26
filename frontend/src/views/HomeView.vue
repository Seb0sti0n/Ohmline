<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getHealth, type Health } from '@/services/api'

const health = ref<Health | null>(null)
const failed = ref(false)

onMounted(async () => {
  try {
    health.value = await getHealth()
  } catch {
    failed.value = true
  }
})
</script>

<template>
  <main class="mx-auto max-w-xl p-s10">
    <h1 class="text-display font-bold" style="font-stretch: 112%">Astrophage</h1>
    <p class="mt-s2 text-body text-ink-muted">Plataforma de gestión energética con IA.</p>

    <section class="mt-s6 rounded-lg border border-line bg-surface p-s6">
      <h2 class="text-heading">Estado del servicio</h2>
      <p v-if="failed" class="mt-s3 text-status-critical">API no disponible</p>
      <p v-else-if="!health" class="mt-s3 text-ink-muted">Comprobando…</p>
      <p v-else class="mt-s3 text-status-ok">API y base de datos: {{ health.db }}</p>
    </section>
  </main>
</template>

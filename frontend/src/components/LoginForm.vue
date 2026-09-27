<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { errorMessage } from '@/services/api'

const auth = useAuthStore()
const router = useRouter()

// The demo credentials come prefilled (see design/DESIGN.md).
const email = ref('demo@energy.io')
const password = ref('demo123')
const submitting = ref(false)
const error = ref('')

async function submit() {
  if (submitting.value) return
  submitting.value = true
  error.value = ''
  try {
    await auth.login(email.value.trim(), password.value)
    await router.push({ name: 'dashboard' })
  } catch (e) {
    error.value =
      axios.isAxiosError(e) && e.response?.status === 401
        ? 'Correo o contraseña incorrectos'
        : errorMessage(e, 'No se pudo iniciar sesión')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="flex w-full max-w-[400px] flex-col gap-5" @submit.prevent="submit">
    <div class="flex flex-col gap-1.5">
      <h2 class="text-[28px] font-bold">Iniciar sesión</h2>
      <p class="text-body text-ink-muted">Usa la cuenta de demostración para entrar.</p>
    </div>

    <label class="flex flex-col gap-2 text-label font-semibold">
      Correo
      <input
        v-model="email"
        type="email"
        autocomplete="username"
        required
        class="h-12 w-full rounded-md border border-line bg-surface px-3.5 text-[16px] font-normal text-ink leading-5"
      />
    </label>

    <label class="flex flex-col gap-2 text-label font-semibold">
      Contraseña
      <input
        v-model="password"
        type="password"
        autocomplete="current-password"
        required
        class="h-12 w-full rounded-md border border-line bg-surface px-3.5 text-[16px] font-normal text-ink leading-5"
      />
    </label>

    <p
      v-if="error"
      role="alert"
      class="rounded-md bg-status-critical-bg px-3.5 py-2.5 text-body font-semibold text-status-critical"
    >
      {{ error }}
    </p>

    <button
      type="submit"
      :disabled="submitting"
      class="flex h-[50px] cursor-pointer items-center justify-center rounded-md bg-brand text-[16px] font-semibold text-white hover:bg-brand-strong disabled:cursor-wait disabled:opacity-70"
    >
      {{ submitting ? 'Entrando…' : 'Entrar' }}
    </button>
  </form>
</template>

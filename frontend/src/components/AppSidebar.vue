<script setup lang="ts">
import { useRouter } from 'vue-router'
import BrandLogo from './BrandLogo.vue'
import { useAuthStore } from '@/stores/auth'
import { useAnalysisStore } from '@/stores/analysis'

const auth = useAuthStore()
const analysis = useAnalysisStore()
const router = useRouter()

const items = [
  { to: '/', label: 'Resumen', icon: 'grid', exact: true },
  { to: '/meters', label: 'Medidores', icon: 'gauge', exact: false },
  { to: '/anomalies', label: 'Anomalías IA', icon: 'alert', exact: false },
]

function logout() {
  analysis.stop()
  auth.logout()
  void router.push({ name: 'login' })
}
</script>

<template>
  <aside class="sticky top-0 flex h-screen w-[232px] shrink-0 flex-col bg-ink px-4 py-6 text-white">
    <BrandLogo class="px-2 pb-7" />

    <nav aria-label="Principal" class="flex flex-col gap-1">
      <RouterLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        custom
        v-slot="{ href, navigate, isActive, isExactActive }"
      >
        <a
          :href="href"
          :aria-current="(item.exact ? isExactActive : isActive) ? 'page' : undefined"
          class="flex h-11 items-center gap-3 rounded-sm px-3.5 text-[15px] no-underline"
          :class="
            (item.exact ? isExactActive : isActive)
              ? 'bg-sidebar-active font-semibold text-white'
              : 'font-medium text-sidebar-text hover:bg-sidebar-active/60 hover:text-white'
          "
          @click="navigate"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path
              v-if="item.icon === 'grid'"
              d="M4 4h6v6H4zM14 4h6v6h-6zM4 14h6v6H4zM14 14h6v6h-6z"
            />
            <template v-else-if="item.icon === 'gauge'">
              <circle cx="12" cy="13" r="8" />
              <path d="M12 13l4-4M8 5.5 7 3.5M16 5.5l1-2" />
            </template>
            <template v-else>
              <path d="M12 3 2 20h20L12 3z" />
              <path d="M12 10v4M12 17v.5" />
            </template>
          </svg>
          <span>{{ item.label }}</span>
        </a>
      </RouterLink>
    </nav>

    <div class="grow" />

    <div class="flex items-center gap-2 border-t border-[#2E3B36] pt-4">
      <div
        class="flex size-[34px] items-center justify-center rounded-full bg-[#2E3B36] text-[13px] font-semibold"
      >
        OP
      </div>
      <div class="flex min-w-0 grow flex-col gap-0.5">
        <span class="text-[14px] font-semibold">Operador demo</span>
        <span class="truncate text-[12px] text-sidebar-text">{{ auth.email }}</span>
      </div>
      <button
        type="button"
        aria-label="Cerrar sesión"
        class="flex size-11 cursor-pointer items-center justify-center rounded-sm text-sidebar-text hover:text-white"
        @click="logout"
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M15 4h4v16h-4M10 8l-4 4 4 4M6 12h10" />
        </svg>
      </button>
    </div>
  </aside>
</template>

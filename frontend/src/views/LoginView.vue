<script setup lang="ts">
import BrandLogo from '@/components/BrandLogo.vue'
import LoginForm from '@/components/LoginForm.vue'
import LoginIllustration from '@/components/LoginIllustration.vue'
</script>

<template>
  <div class="grid min-h-screen bg-ground lg:grid-cols-[minmax(0,760px)_minmax(0,1fr)]">
    <section class="hidden flex-col justify-between bg-ink px-[70px] py-14 text-white lg:flex">
      <BrandLogo :size="36" text-class="text-[24px]" />
      <div class="flex flex-col gap-7">
        <h1
          class="max-w-[560px] text-[52px] font-bold leading-[1.05] tracking-[-0.02em] [font-stretch:118%]"
        >
          Sabe qué medidor revisar primero, y por qué.
        </h1>
        <p class="max-w-[520px] text-[18px] leading-normal text-sidebar-text">
          La IA compara cada lectura con su comportamiento esperado, descarta lo que explican los
          eventos operativos y te dice qué hacer.
        </p>
        <LoginIllustration />
      </div>
      <p class="text-label font-normal text-[#8A9993]">12 medidores monitorizados</p>
    </section>

    <section class="flex items-center justify-center px-6 py-10">
      <LoginForm />
    </section>
  </div>
</template>

<template>
  <div class="fullscreen bg-blue text-white text-center q-pa-md flex flex-center">
    <div>
      <div style="font-size: 30vh">404</div>

      <div class="text-h2" style="opacity: 0.4">{{ t('auto.oops_nothing_here') }}</div>

      <q-btn class="q-mt-xl" color="white" text-color="blue" unelevated to="/" :label="t('auto.go_home')" no-caps rounded />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
//
</script>

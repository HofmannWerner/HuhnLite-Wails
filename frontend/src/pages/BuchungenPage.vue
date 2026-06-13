<template>
  <q-page padding :class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-1'">
    <div class="row items-center justify-between q-mb-lg">
      <div class="text-h4 text-weight-bold text-primary">{{ t('auto.buchungen') }}</div>
    </div>

    <q-card flat bordered class="rounded-borders shadow-2 overflow-hidden" :class="$q.dark.isActive ? 'bg-grey-10 text-white' : 'bg-white'">
      <q-tabs
        v-model="tab"
        dense
        active-color="primary"
        indicator-color="primary"
        align="left"
        narrow-indicator
        :class="$q.dark.isActive ? 'bg-grey-10 text-grey-4' : 'bg-white text-grey-7'"
      >
        <q-tab name="leistung" :label="t('auto.leistung')" icon="receipt_long" />
        <q-tab name="verluste" :label="t('auto.verluste')" icon="remove_circle" />
        <q-tab name="tierbewegungen" :label="t('auto.tierbewegungen')" icon="swap_horiz" />
        <q-tab name="eilager" :label="t('grid.eggStorage')" icon="inventory" />
        <q-tab name="futter" :label="t('auto.futter')" icon="local_shipping" />
        <q-tab name="verkauf" :label="t('auto.verkauf')" icon="shopping_cart" />
        <q-tab name="aktionen" :label="t('form.actions')" icon="task" />
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="tab" animated class="bg-transparent">
        <q-tab-panel name="leistung" class="q-pa-none">
          <LeistungGrid />
        </q-tab-panel>

        <q-tab-panel name="verluste" class="q-pa-none">
          <VerlusteGrid />
        </q-tab-panel>

        <q-tab-panel name="tierbewegungen" class="q-pa-none">
          <TierbewegungenGrid />
        </q-tab-panel>

        <q-tab-panel name="eilager" class="q-pa-none">
          <EilagerBuchungenGrid />
        </q-tab-panel>

        <q-tab-panel name="futter" class="q-pa-none">
          <FutterGrid />
        </q-tab-panel>

        <q-tab-panel name="verkauf" class="q-pa-none">
          <VerkaufGrid />
        </q-tab-panel>

        <q-tab-panel name="aktionen" class="q-pa-none">
          <AktionenGrid />
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import LeistungGrid from 'components/LeistungGrid.vue';
import VerlusteGrid from 'components/VerlusteGrid.vue';
import TierbewegungenGrid from 'components/TierbewegungenGrid.vue';
import EilagerBuchungenGrid from 'components/EilagerBuchungenGrid.vue';
import FutterGrid from 'components/FutterGrid.vue';
import VerkaufGrid from 'components/VerkaufGrid.vue';
import AktionenGrid from 'components/AktionenGrid.vue';

const $q = useQuasar();
const route = useRoute();
const tab = ref((route.query.tab as string) || 'leistung');

watch(() => route.query.tab, (newTab) => {
  if (newTab) {
    tab.value = newTab as string;
  }
});
</script>

<style scoped>
/* Tabs styling refinement */
</style>

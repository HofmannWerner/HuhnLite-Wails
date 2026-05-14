<template>
  <q-page padding class="bg-grey-1">
    <div class="row items-center justify-between q-mb-lg">
      <div class="text-h4 text-weight-bold text-primary">Buchungen</div>
    </div>

    <q-card class="shadow-2 rounded-borders overflow-hidden">
      <q-tabs
        v-model="tab"
        dense
        class="bg-white text-grey-7"
        active-color="primary"
        indicator-color="primary"
        align="left"
        narrow-indicator
      >
        <q-tab name="leistung" label="Leistung" icon="receipt_long" />
        <q-tab name="verluste" label="Verluste" icon="remove_circle" />
        <q-tab name="tierbewegungen" label="Tierbewegungen" icon="swap_horiz" />
        <q-tab name="eilager" label="Eilager" icon="inventory" />
        <q-tab name="futter" label="Futter" icon="local_shipping" />
        <q-tab name="verkauf" label="Verkauf" icon="shopping_cart" />
        <q-tab name="aktionen" label="Aktionen" icon="task" />
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="tab" animated class="bg-grey-1">
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
import { ref, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import LeistungGrid from 'components/LeistungGrid.vue';
import VerlusteGrid from 'components/VerlusteGrid.vue';
import TierbewegungenGrid from 'components/TierbewegungenGrid.vue';
import EilagerBuchungenGrid from 'components/EilagerBuchungenGrid.vue';
import FutterGrid from 'components/FutterGrid.vue';
import VerkaufGrid from 'components/VerkaufGrid.vue';
import AktionenGrid from 'components/AktionenGrid.vue';

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

<template>
  <q-page padding :class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-1'">
    <div class="row items-center justify-between q-mb-lg">
      <div class="text-h4 text-weight-bold text-primary">{{ t('auto.einrichtungen') }}</div>
    </div>

    <q-card flat bordered class="rounded-borders shadow-2 overflow-hidden" :class="$q.dark.isActive ? 'bg-grey-10 text-white' : 'bg-white'">
      <q-tabs
        v-model="mainTab"
        dense
        active-color="primary"
        indicator-color="primary"
        align="left"
        narrow-indicator
        :class="$q.dark.isActive ? 'bg-grey-10 text-grey-4' : 'bg-white text-grey-7'"
      >
        <q-tab name="silos" :label="t('auto.silos')" icon="warehouse" />
        <q-tab name="stalle" :label="t('auto.staelle')" icon="home" />
        <q-tab name="eilager" :label="t('grid.eggStorage')" icon="inventory_2" />
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="mainTab" animated class="bg-transparent">
        <!-- Tab 1: Silos -->
        <q-tab-panel name="silos" class="q-pa-none">
          <SiloGrid />
        </q-tab-panel>

        <!-- Tab 2: Ställe -->
        <q-tab-panel name="stalle" class="q-pa-none">
          <StallGrid />
        </q-tab-panel>

        <!-- Tab 3: Eilager (mit seinen eigenen Unter-Tabs) -->
        <q-tab-panel name="eilager" class="q-pa-none">
          <div class="q-pa-md">
            <q-tabs
              v-model="eilagerTab"
              dense
              active-color="secondary"
              indicator-color="secondary"
              align="left"
              narrow-indicator
              class="q-mb-md"
              :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'"
            >
              <q-tab name="eilager" :label="t('grid.eggStorage')" icon="inventory_2" />
              <q-tab name="lagerplatz" :label="t('auto.lagerplatzverwaltung')" icon="place" />
              <q-tab name="bestandsuebersicht" :label="t('auto.bestandsuebersicht')" icon="analytics" />
            </q-tabs>

            <q-tab-panels v-model="eilagerTab" animated class="rounded-borders shadow-1" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'">
              <q-tab-panel name="eilager" class="q-pa-none">
                <EilagerGrid />
              </q-tab-panel>
              <q-tab-panel name="lagerplatz" class="q-pa-none">
                <LagerplatzVerwaltung />
              </q-tab-panel>
              <q-tab-panel name="bestandsuebersicht" class="q-pa-none">
                <BestandsUebersicht />
              </q-tab-panel>
            </q-tab-panels>
          </div>
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import SiloGrid from 'components/SiloGrid.vue';
import StallGrid from 'components/StallGrid.vue';
import EilagerGrid from 'components/EilagerGrid.vue';
import LagerplatzVerwaltung from 'components/LagerplatzVerwaltung.vue';
import BestandsUebersicht from 'components/BestandsUebersicht.vue';

const $q = useQuasar();
const mainTab = ref('silos');
const eilagerTab = ref('eilager');
</script>

<style scoped>
/* Optional: Feinabstimmung für das Design */
</style>

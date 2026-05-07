<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">Herden & Rassen</div>
    </div>

    <q-card flat bordered class="rounded-borders shadow-2 overflow-hidden"
            :class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-1'">
      <q-tabs
        v-model="tab"
        dense
        class="text-grey bg-white shadow-1"
        active-color="primary"
        indicator-color="primary"
        align="left"
        narrow-indicator
        :class="$q.dark.isActive ? 'bg-grey-10 text-grey-4' : 'bg-white text-grey-7'"
      >
        <q-tab name="herden" label="Herden-Stamm" icon="pets"/>
        <q-tab name="rassen" label="Rassen" icon="category"/>
      </q-tabs>

      <q-separator/>

      <q-tab-panels v-model="tab" animated class="bg-transparent">
        <q-tab-panel name="herden" class="q-pa-none">
          <HerdeGrid ref="herdeGrid"/>
        </q-tab-panel>

        <q-tab-panel name="rassen" class="q-pa-none">
          <RasseGrid @updated="refreshHerden"/>
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import {ref} from 'vue';
import {useQuasar} from 'quasar';
import HerdeGrid from '../components/HerdeGrid.vue';
import RasseGrid from '../components/RasseGrid.vue';

const $q = useQuasar();
const tab = ref('herden');
const herdeGrid = ref<any>(null);

function refreshHerden() {
  if (herdeGrid.value && herdeGrid.value.loadData) {
    herdeGrid.value.loadData();
  }
}
</script>

<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">{{ t('auto.direkt_update_erzwungen') }}</div>
    </div>

    <q-card flat bordered class="rounded-borders shadow-10" style="max-width: 600px; margin: auto;">
      <q-card-section class="bg-primary text-white row items-center q-pa-md">
        <q-icon name="speed" size="md" class="q-mr-md" />
        <div class="text-h6">{{ t('auto.daten_verschiebung_v3') }}</div>
      </q-card-section>

      <q-card-section class="q-pa-xl">
        <div class="text-h6 q-mb-md">{{ t('auto.wieviele_tage_verschieben') }}</div>

        <q-input
          v-model="days"
          filled
          bg-color="white"
          outlined
          :placeholder="t('auto.zahl_hier_eingeben')"
          class="q-mb-lg"
          input-style="font-size: 2.5rem; text-align: center; color: #1976D2;"
        />

        <q-btn
          color="negative"
          icon="rocket_launch"
          label="JETZT AKTUALISIEREN"
          class="full-width"
          size="xl"
          rounded
          unelevated
          :loading="loading"
          @click="startAction"
        />

        <q-btn
          :label="t('form.cancel')"
          color="grey-7"
          flat
          rounded
          class="full-width q-mt-md"
          size="lg"
          to="/"
        />

        <div class="q-mt-lg text-center text-grey-8">
          {{ t('auto.db_stand') }} <strong>{{ latestDate || 'Lade...' }}</strong> {{ t('auto.vorschlag') }} <strong>{{ suggested }}</strong>
        </div>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, onMounted } from 'vue';
import { useQuasar, date as qDate } from 'quasar';
import { api } from '../boot/api';

const $q = useQuasar();
const days = ref('666'); // Testwert: Wenn man 666 sieht, ist es diese Datei!
const latestDate = ref('');
const suggested = ref(0);
const loading = ref(false);

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

async function init() {
  try {
    const res = await api.get('/api/admin/latest-date');
    if (res.data && res.data.max_date) {
      latestDate.value = extractString(res.data.max_date).split(' ')[0] || '';
      const last = new Date(latestDate.value);
      const today = new Date();
      last.setHours(0,0,0,0);
      today.setHours(0,0,0,0);
      const diff = qDate.getDateDiff(today, last, 'days');
      suggested.value = diff;
      days.value = String(diff);
    }
  } catch (err) {
    console.error('Init-Fehler:', err);
  }
}

function startAction() {
  const num = parseInt(days.value);
  if (isNaN(num)) return;

  $q.dialog({
    title: 'Finaler Check',
    message: `Echt jetzt? Verschieben um ${num} Tage?`,
    ok: { label: 'JA, STARTEN', color: 'negative', rounded: true },
    cancel: { label: 'Abbrechen', flat: true }
  }).onOk(() => {
    run(num);
  });
}

async function run(num: number) {
  loading.value = true;
  try {
    await api.post('/api/admin/shift-test-dates', {DAYS: num});
    $q.notify('Erfolg!');
    setTimeout(() => location.reload(), 1500);
  } catch (e: any) {
    const errorMsg = e.response?.data?.error || e.message;
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Verschieben: ' + errorMsg,
      position: 'top',
      timeout: 10000
    });
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  init();
});
</script>

<style scoped>
.rounded-borders { border-radius: 20px; }
</style>

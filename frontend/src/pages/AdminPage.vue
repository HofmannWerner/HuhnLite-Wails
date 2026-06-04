<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">{{ t('auto.system_wartung_direkt_modus') }}</div>
    </div>

    <div class="row q-col-gutter-lg">
      <div class="col-12 col-md-8">
        <q-card flat bordered class="rounded-borders shadow-5">
          <q-card-section class="bg-primary text-white row items-center q-pa-md">
            <q-icon name="manage_accounts" size="md" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">{{ t('auto.datumswerte_verschieben') }}</div>
          </q-card-section>

          <q-card-section class="q-pa-xl">
            <div class="row q-col-gutter-xl items-center">

              <!-- Eingabe links -->
              <div class="col-12 col-sm-6">
                <div class="text-subtitle1 text-weight-bold q-mb-sm">{{ t('auto.tage_verschieben') }}</div>
                <q-input
                  v-model="daysInput"
                  filled
                  bg-color="white"
                  outlined
                  :placeholder="t('auto.zahl_eingeben')"
                  input-style="font-size: 2rem; text-align: center; font-weight: bold;"
                  :hint="t('auto.positive_zahl_zukunft_negative_zahl_verg')"
                />

                <div class="q-mt-md text-grey-7">
                  {{ t('auto.aktuellstes_datum_in_db') }} <br>
                  <strong class="text-primary text-h6">{{ latestDate || 'Lade...' }}</strong>
                </div>
              </div>

              <!-- Buttons rechts -->
              <div class="col-12 col-sm-6 column q-gutter-y-md">
                <q-btn
                  color="negative"
                  icon="send"
                  label="JETZT AKTUALISIEREN"
                  class="full-width"
                  size="xl"
                  rounded
                  unelevated
                  padding="lg"
                  :loading="processing"
                  @click="confirmAndRun"
                />
                <q-btn
                  :label="t('form.cancel')"
                  color="grey-7"
                  outline
                  rounded
                  class="full-width"
                  size="lg"
                  to="/"
                />
                <div class="text-center q-mt-md text-orange-9 text-weight-bold">
                  Vorschlag: {{ suggestedDiff }} Tage
                </div>
              </div>

            </div>

            <div class="q-mt-xl bg-red-1 text-red-9 q-pa-md rounded-borders border-red shadow-1">
               <strong>{{ t('auto.warnung') }}</strong> {{ t('auto.diese_aktion_aendert_alle_datumsfelder_i') }}
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, onMounted } from 'vue';
import { useQuasar, date as qDate } from 'quasar';
import { api } from '../boot/api';

const $q = useQuasar();

const latestDate = ref('');
const daysInput = ref('999'); // Offensichtlicher Testwert für Cache-Check
const suggestedDiff = ref(0);
const processing = ref(false);

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

async function loadInitData() {
  console.log('AdminPage: Lade-Vorgang gestartet');
  try {
    const res = await api.get('/api/admin/latest-date');
    if (res.data && res.data.max_date) {
      latestDate.value = extractString(res.data.max_date).split(' ')[0] || '';

      const last = new Date(latestDate.value);
      const today = new Date();
      last.setHours(0,0,0,0);
      today.setHours(0,0,0,0);

      const diff = qDate.getDateDiff(today, last, 'days');
      suggestedDiff.value = diff;
      daysInput.value = String(diff); // Wert als String setzen
      console.log('Berechnete Differenz:', diff);
    }
  } catch (err) {
    console.error('Ladefehler Init:', err);
  }
}

function confirmAndRun() {
  const val = parseInt(String(daysInput.value));
  if (isNaN(val)) {
    $q.notify({ message: 'Bitte eine gültige Zahl eingeben!', color: 'warning' });
    return;
  }

  $q.dialog({
    title: 'Sicherheitsbestätigung',
    message: `Sollen alle Daten wirklich um <b>${val} Tage</b> verschieben werden?`,
    html: true,
    ok: { label: 'Ja, jetzt ausführen', color: 'negative', rounded: true },
    cancel: { label: 'Abbrechen', flat: true },
    persistent: true
  }).onOk(() => {
    runUpdate(val);
  });
}

async function runUpdate(days: number) {
  processing.value = true;
  try {
    await api.post('/api/admin/shift-test-dates', {
      DAYS: days
    });
    $q.notify({ type: 'positive', message: 'Verschiebung erfolgreich!' });
    setTimeout(() => location.reload(), 2000);
  } catch (err: any) {
    const msg = err.response?.data?.error || err.message;
    $q.notify({ type: 'negative', message: 'Fehler: ' + msg });
  } finally {
    processing.value = false;
  }
}

onMounted(() => {
  loadInitData();
});
</script>

<style scoped>
.rounded-borders {
  border-radius: 20px;
}
.border-red { border: 1px solid #f44336; }
</style>

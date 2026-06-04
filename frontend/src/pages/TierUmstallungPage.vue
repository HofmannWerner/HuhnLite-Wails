<template>
  <q-page padding>
    <div class="row items-center justify-between q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">{{ t('auto.tier_umstallung') }}</div>
    </div>

    <div class="row items-center q-gutter-md q-mb-lg">
      <div style="min-width: 200px">
        <q-select
          v-model="timeframe"
          :options="timeframeOptions"
          :label="t('auto.zeitraum_tage_zurueck')"
          filled
          dense
          emit-value
          map-options
        >
          <template v-slot:prepend>
            <q-icon name="history" />
          </template>
        </q-select>
      </div>
      <div class="text-caption text-grey-7">
        {{ timeframe === 0 ? t('auto.shows_all_bookings') : t('auto.shows_bookings_days', { days: timeframe }) }}
      </div>
    </div>

    <div class="row q-col-gutter-lg items-center">
      <!-- Links: Abgebende Herde -->
      <div class="col-12 col-md-5">
        <q-card flat bordered class="rounded-borders shadow-2">
          <q-card-section class="bg-red-7 text-white row items-center q-pa-sm">
            <q-icon name="outbox" size="sm" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">{{ t('auto.abgebende_herde_geben') }}</div>
          </q-card-section>
          
          <q-table
            :rows="filteredBookings"
            :columns="columns"
            row-key="id"
            flat
            dense
            :pagination="{ rowsPerPage: 10 }"
            selection="single"
            v-model:selected="selectedLeft"
            :loading="loading"
            class="full-width"
            style="height: 500px"
          />
        </q-card>
      </div>

      <!-- Mitte: Action Button -->
      <div class="col-12 col-md-2 text-center">
        <q-btn
          round
          size="32px"
          :color="isActionActive ? 'positive' : 'grey-5'"
          :disable="!isActionActive"
          icon="east"
          @click="openAmountDialog"
          class="shadow-5"
        >
          <q-tooltip v-if="isActionActive">{{ t('auto.tiere_umbuchen') }}</q-tooltip>
        </q-btn>
      </div>

      <!-- Rechts: Empfangende Herde -->
      <div class="col-12 col-md-5">
        <q-card flat bordered class="rounded-borders shadow-2">
          <q-card-section class="bg-green-7 text-white row items-center q-pa-sm">
            <q-icon name="inbox" size="sm" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">{{ t('auto.empfangende_herde_nehmen') }}</div>
          </q-card-section>
          
          <q-table
            :rows="filteredRightBookings"
            :columns="columns"
            row-key="id"
            flat
            dense
            :pagination="{ rowsPerPage: 10 }"
            selection="single"
            v-model:selected="selectedRight"
            :loading="loading"
            :no-data-label="selectedLeft.length === 0 ? t('auto.select_herd_left') : t('auto.no_matching_herds_same_date')"
            class="full-width"
            style="height: 500px"
          />
        </q-card>
      </div>
    </div>

    <!-- Menge Dialog -->
    <q-dialog v-model="amountDialog" persistent>
      <q-card style="min-width: 350px">
        <q-card-section class="bg-primary text-white">
          <div class="text-h6">{{ t('auto.umstallungsmenge') }}</div>
        </q-card-section>

        <q-card-section class="q-pt-md">
          <div class="q-mb-md text-weight-medium">
            {{ t('auto.tiere_von_herde') }} <strong>{{ herdeLinks?.herdennummer }}</strong> 
            {{ t('auto.nach_herde') }} <strong>{{ herdeRechts?.herdennummer }}</strong> {{ t('auto.verschieben') }}
          </div>
          <q-input
            v-model.number="amount"
            type="number"
            :label="t('auto.anzahl_stueck')"
            filled
            autofocus
            :rules="[
              val => !!val || t('message.required'),
              val => val > 0 || t('auto.must_be_greater_than_zero'),
              val => val <= (herdeLinks?.tierbestand || 0) || t('auto.insufficient_stock')
            ]"
          />
          <q-select
            v-model="idTexte"
            :options="grundOptions"
            option-value="id"
            option-label="betreff"
            emit-value
            map-options
            :label="t('auto.grund_fuer_umstallung')"
            filled
            class="q-mt-md"
            :rules="[val => !!val || t('auto.select_reason')]"
          />
          <div class="text-caption text-grey-7 q-mt-xs">
            {{ t('auto.available_stock') }}: {{ herdeLinks?.tierbestand || 0 }} {{ t('auto.animals') }}
          </div>
        </q-card-section>

        <q-card-actions align="right" class="text-primary q-pa-md">
          <q-btn flat :label="t('form.cancel')" v-close-popup />
          <q-btn :label="t('auto.bestaetigen')" color="primary" @click="confirmUmbuchung" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted, watch } from 'vue';
import { api } from '../boot/api';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const $q = useQuasar();
const loading = ref(false);
const bookings = ref<any[]>([]);
const selectedLeft = ref<any[]>([]);
const selectedRight = ref<any[]>([]);
const amountDialog = ref(false);
const amount = ref(0);
const idTexte = ref<number | null>(null);
const grundOptions = ref<any[]>([]);

const timeframe = ref(7);
const timeframeOptions = computed(() => [
  { label: t('auto.last_7_days'), value: 7 },
  { label: t('auto.last_30_days'), value: 30 },
  { label: t('auto.last_90_days'), value: 90 },
  { label: t('auto.all_bookings'), value: 0 },
]);

const herdeLinks = computed(() => selectedLeft.value[0] || null);
const herdeRechts = computed(() => selectedRight.value[0] || null);

const columns = computed(() => [
  { name: 'HERDENNUMMER', align: 'left', label: t('grid.herdNumber'), field: 'herdennummer', sortable: true },
  { name: 'BEZEICHNUNG', align: 'left', label: t('grid.designation'), field: 'herden_bezeichnung', sortable: true },
  { name: 'DATUM', align: 'left', label: t('auto.datum'), field: 'buchungsdatum', sortable: true },
  { name: 'BESTAND', align: 'right', label: t('grid.stockAnimals'), field: 'tierbestand', sortable: true },
]);

const filteredBookings = computed(() => {
  if (timeframe.value === 0) return bookings.value;
  
  const cutoff = new Date();
  cutoff.setDate(cutoff.getDate() - timeframe.value);
  cutoff.setHours(0, 0, 0, 0);

  return bookings.value.filter(b => {
    const d = new Date(b.buchungsdatum);
    return d >= cutoff;
  });
});

const filteredRightBookings = computed(() => {
  if (selectedLeft.value.length === 0) return [];
  const left = selectedLeft.value[0];
  return bookings.value.filter(b => 
    b.buchungsdatum === left.buchungsdatum && 
    b.id !== left.id
  );
});

watch(selectedLeft, () => {
  selectedRight.value = [];
});

const isActionActive = computed(() => {
  if (selectedLeft.value.length === 0 || selectedRight.value.length === 0) return false;
  return selectedLeft.value[0].buchungsdatum === selectedRight.value[0].buchungsdatum;
});

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/buchungen/detailed');
    bookings.value = res.data || [];
    
    const resTexte = await api.get('/api/texte/typ/T');
    grundOptions.value = (resTexte.data || []).map((t: any) => ({
      id: t.ID || t.id,
      betreff: t.BETREFF || t.betreff
    }));
  } catch (err) {
    console.error('Error loading data:', err);
  } finally {
    loading.value = false;
  }
}

function openAmountDialog() {
  amount.value = 0;
  amountDialog.value = true;
}

async function confirmUmbuchung() {
  if (amount.value <= 0 || amount.value > (herdeLinks.value?.tierbestand || 0)) {
    $q.notify({ type: 'warning', message: t('auto.invalid_amount') });
    return;
  }

  try {
    await api.post('/api/tierbewegungen/umbuchung', {
      ID_BUCHUNG_VON: herdeLinks.value.id,
      ID_HERDEN_VON: herdeLinks.value.id_herden,
      ID_BUCHUNG_NACH: herdeRechts.value.id,
      ID_HERDEN_NACH: herdeRechts.value.id_herden,
      MENGE: amount.value,
      DATUM: herdeLinks.value.buchungsdatum,
      ID_TEXTE: idTexte.value
    });

    $q.notify({ type: 'positive', message: t('auto.relocation_success') });
    amountDialog.value = false;
    selectedLeft.value = [];
    selectedRight.value = [];
    loadData();
  } catch (err: any) {
    const msg = err.response?.data?.error || err.message;
    $q.notify({ type: 'negative', message: t('message.error') + ': ' + msg });
  }
}

onMounted(loadData);
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
</style>

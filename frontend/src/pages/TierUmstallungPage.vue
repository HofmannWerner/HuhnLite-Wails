<template>
  <q-page padding>
    <div class="row items-center justify-between q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">Tier-Umstallung</div>
    </div>

    <div class="row items-center q-gutter-md q-mb-lg">
      <div style="min-width: 200px">
        <q-select
          v-model="timeframe"
          :options="timeframeOptions"
          label="Zeitraum (Tage zurück)"
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
        Zeigt Buchungen der letzten {{ timeframe === 0 ? 'alle' : timeframe }} Tage an.
      </div>
    </div>

    <div class="row q-col-gutter-lg items-center">
      <!-- Links: Abgebende Herde -->
      <div class="col-12 col-md-5">
        <q-card flat bordered class="rounded-borders shadow-2">
          <q-card-section class="bg-red-7 text-white row items-center q-pa-sm">
            <q-icon name="outbox" size="sm" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">Abgebende Herde (Geben)</div>
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
          <q-tooltip v-if="isActionActive">Tiere umbuchen</q-tooltip>
        </q-btn>
      </div>

      <!-- Rechts: Empfangende Herde -->
      <div class="col-12 col-md-5">
        <q-card flat bordered class="rounded-borders shadow-2">
          <q-card-section class="bg-green-7 text-white row items-center q-pa-sm">
            <q-icon name="inbox" size="sm" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">Empfangende Herde (Nehmen)</div>
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
            :no-data-label="selectedLeft.length === 0 ? 'Bitte zuerst links eine Herde wählen' : 'Keine passenden Herden zum selben Datum gefunden'"
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
          <div class="text-h6">Umstallungsmenge</div>
        </q-card-section>

        <q-card-section class="q-pt-md">
          <div class="q-mb-md text-weight-medium">
            Tiere von Herde <strong>{{ herdeLinks?.herdennummer }}</strong> 
            nach Herde <strong>{{ herdeRechts?.herdennummer }}</strong> verschieben.
          </div>
          <q-input
            v-model.number="amount"
            type="number"
            label="Anzahl (Stück) *"
            filled
            autofocus
            :rules="[
              val => !!val || 'Erforderlich',
              val => val > 0 || 'Muss größer 0 sein',
              val => val <= (herdeLinks?.tierbestand || 0) || 'Bestand nicht ausreichend!'
            ]"
          />
          <q-select
            v-model="idTexte"
            :options="grundOptions"
            option-value="id"
            option-label="betreff"
            emit-value
            map-options
            label="Grund für Umstallung *"
            filled
            class="q-mt-md"
            :rules="[val => !!val || 'Bitte wählen Sie einen Grund']"
          />
          <div class="text-caption text-grey-7 q-mt-xs">
            Verfügbarer Bestand: {{ herdeLinks?.tierbestand || 0 }} Tiere
          </div>
        </q-card-section>

        <q-card-actions align="right" class="text-primary q-pa-md">
          <q-btn flat label="ABBRUCH" v-close-popup />
          <q-btn label="Bestätigen" color="primary" @click="confirmUmbuchung" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { api } from '../boot/api';
import { useQuasar } from 'quasar';

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
const timeframeOptions = [
  { label: 'Letzte 7 Tage', value: 7 },
  { label: 'Letzte 30 Tage', value: 30 },
  { label: 'Letzte 90 Tage', value: 90 },
  { label: 'Alle Buchungen', value: 0 },
];

const herdeLinks = computed(() => selectedLeft.value[0] || null);
const herdeRechts = computed(() => selectedRight.value[0] || null);

const columns = [
  { name: 'HERDENNUMMER', align: 'left', label: 'Herde-Nr', field: 'herdennummer', sortable: true },
  { name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: 'herden_bezeichnung', sortable: true },
  { name: 'DATUM', align: 'left', label: 'Datum', field: 'buchungsdatum', sortable: true },
  { name: 'BESTAND', align: 'right', label: 'Bestand', field: 'tierbestand', sortable: true },
];

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
  // Filtern nach gleichem Datum und andere Herde
  // Wir nutzen hier alle bookings (unabhängig vom timeframe-Filter der linken Seite), 
  // um sicherzustellen dass wir auch Ziele finden falls der Zeitraum links anders gefiltert wäre.
  return bookings.value.filter(b => 
    b.buchungsdatum === left.buchungsdatum && 
    b.id !== left.id
  );
});

// Automatisch rechts löschen wenn links gewechselt wird
watch(selectedLeft, () => {
  selectedRight.value = [];
});

const isActionActive = computed(() => {
  if (selectedLeft.value.length === 0 || selectedRight.value.length === 0) return false;
  // Datum prüfen
  return selectedLeft.value[0].buchungsdatum === selectedRight.value[0].buchungsdatum;
});

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/buchungen/detailed');
    bookings.value = res.data || [];
    
    // Gründe laden (Typ 'T' für Umstallung/Abgang)
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
    $q.notify({ type: 'warning', message: 'Ungültige Menge' });
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

    $q.notify({ type: 'positive', message: 'Umstallung erfolgreich durchgeführt' });
    amountDialog.value = false;
    selectedLeft.value = [];
    selectedRight.value = [];
    loadData();
  } catch (err: any) {
    const msg = err.response?.data?.error || err.message;
    $q.notify({ type: 'negative', message: 'Fehler: ' + msg });
  }
}

onMounted(loadData);
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
</style>

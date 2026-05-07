<template>
  <div class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="row q-gutter-md">
        <q-btn color="primary" icon="add" label="Neue Bewegung" @click="openCreate" :disable="!filterHerde" rounded unelevated />
      </div>
      <div class="text-h6 text-primary">Tierbewegungen</div>
    </div>

    <div class="row q-col-gutter-md q-mb-md items-center">
      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterHerde"
          :options="herdeOptions"
          option-value="id"
          option-label="bezeichnung"
          emit-value
          map-options
          clearable
          label="Herde auswählen"
          filled
          stack-label
        >
          <template v-slot:prepend>
            <q-icon name="pets" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-4 col-md-2">
        <q-select
          v-model="timeframe"
          :options="timeframeOptions"
          label="Zeitraum"
          filled
          dense
          emit-value
          map-options
          @update:model-value="onTimeframeUpdate"
        >
          <template v-slot:prepend>
            <q-icon name="history" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-4 col-md-2">
        <q-input filled v-model="filterDateRangeText" label="Manuelles Datum" stack-label readonly dense>
          <template v-slot:prepend>
            <q-icon name="event" class="cursor-pointer">
              <q-popup-proxy cover transition-show="scale" transition-hide="scale">
                <q-date v-model="filterDateRange" range @update:model-value="onDateRangeUpdate">
                  <div class="row items-center justify-end">
                    <q-btn v-close-popup label="Schließen" color="primary" flat />
                  </div>
                </q-date>
              </q-popup-proxy>
            </q-icon>
          </template>
        </q-input>
      </div>

      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterText"
          :options="textOptions"
          option-value="id"
          option-label="betreff"
          emit-value
          map-options
          clearable
          label="Grund"
          filled
          stack-label
          dense
        >
          <template v-slot:prepend>
            <q-icon name="comment" />
          </template>
        </q-select>
      </div>
    </div>

    <!-- Table -->
    <!-- Table View (Single or Dual) -->
    <div v-if="isUmstallung">
      <!-- Dual Grid Layout for Umstallung -->
      <div class="row q-col-gutter-lg items-center">
        <!-- Links: Abgebende Herde -->
        <div class="col-12 col-md-5">
          <q-card flat bordered class="rounded-borders shadow-2">
            <q-card-section class="bg-red-7 text-white row items-center q-pa-sm">
              <q-icon name="outbox" size="sm" class="q-mr-md" />
              <div class="text-subtitle1 text-weight-bold">Abgebende Herde (Geben)</div>
            </q-card-section>
            
            <q-table
              :rows="filteredRows"
              :columns="bookingColumns"
              row-key="id"
              flat
              dense
              :pagination="{ rowsPerPage: 10 }"
              selection="single"
              v-model:selected="selectedLeft"
              :loading="loading"
              class="full-width resizable-table"
              style="height: 500px"
              :bg-color="$q.dark.isActive ? 'grey-10' : 'white'"
            >
              <template v-slot:header-cell="props">
                <q-th :props="props" 
                      class="resizable-column" 
                      :style="{ width: (columnWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
                  <div class="ellipsis">{{ props.col.label }}</div>
                  <div class="resizer" 
                       :class="{ 'is-resizing': isResizing === props.col.name }"
                       @pointerdown.stop.prevent.capture="startResize($event, props.col.name)">
                  </div>
                </q-th>
              </template>
            </q-table>
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
              <div class="text-subtitle1 text-weight-bold">Empfangende Herde (Nehmen)</div>
            </q-card-section>
            
            <q-table
              :rows="filteredRightBookings"
              :columns="bookingColumns"
              row-key="id"
              flat
              dense
              :pagination="{ rowsPerPage: 10 }"
              selection="single"
              v-model:selected="selectedRight"
              :loading="loading"
              class="full-width resizable-table"
              style="height: 500px"
              :bg-color="$q.dark.isActive ? 'grey-10' : 'white'"
            >
              <template v-slot:header-cell="props">
                <q-th :props="props" 
                      class="resizable-column" 
                      :style="{ width: (columnWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
                  <div class="ellipsis">{{ props.col.label }}</div>
                  <div class="resizer" 
                       :class="{ 'is-resizing': isResizing === props.col.name }"
                       @pointerdown.stop.prevent.capture="startResize($event, props.col.name)">
                  </div>
                </q-th>
              </template>
            </q-table>
          </q-card>
        </div>
      </div>
    </div>

    <div v-else>
      <!-- Standard Single Grid View (Leistungsbuchungen) -->
      <q-table
        flat
        bordered
        :rows="filteredRows"
        :columns="bookingColumns"
        row-key="ID"
        :loading="loading"
        :pagination="pagination"
        class="shadow-2 rounded-borders overflow-hidden resizable-table"
        :bg-color="$q.dark.isActive ? 'grey-10' : 'white'"
      >
        <template v-slot:header-cell="props">
          <q-th :props="props" 
                class="resizable-column" 
                :style="{ width: (columnWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
            <div class="ellipsis">{{ props.col.label }}</div>
            <div class="resizer" 
                 :class="{ 'is-resizing': isResizing === props.col.name }"
                 @pointerdown.stop.prevent.capture="startResize($event, props.col.name)">
            </div>
          </q-th>
        </template>

        <template v-slot:body-cell-actions="props">
          <q-td :props="props" class="q-gutter-x-sm">
            <q-btn flat round color="primary" icon="add" size="sm" @click="openCreateForBooking(props.row)">
              <q-tooltip>Bewegung hinzufügen</q-tooltip>
            </q-btn>
          </q-td>
        </template>
      </q-table>
    </div>

    <!-- Amount Dialog for Umstallung -->
    <q-dialog v-model="amountDialog" persistent>
      <q-card style="min-width: 350px; border-radius: 16px;">
        <q-card-section class="bg-primary text-white q-pa-md row items-center">
          <q-icon name="swap_horiz" size="sm" class="q-mr-sm" />
          <div class="text-h6 text-weight-bold">Tier-Umstallung</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup color="white" />
        </q-card-section>

        <q-card-section class="q-pa-lg">
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
          <div class="text-caption text-grey-7 q-mt-xs">
            Verfügbarer Bestand: {{ herdeLinks?.tierbestand || 0 }} Tiere
          </div>

          <div class="row justify-end q-gutter-sm q-mt-lg">
            <q-btn label="Abbrechen" color="negative" flat v-close-popup />
            <q-btn label="Durchführen" color="positive" unelevated @click="confirmUmbuchung" />
          </div>
        </q-card-section>
      </q-card>
    </q-dialog>

    <TierbewegungDialog 
      v-model="showDialog" 
      :is-editing="isEditing" 
      :edit-id="editId" 
      :initial-herde-id="initialHerdeId"
      :initial-typ="initialTyp"
      @saved="loadData" 
    />
  </div>
</template>

<script setup lang="ts">
const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  return Number(val) || 0;
};

import { ref, onMounted, computed, watch } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

/* eslint-disable @typescript-eslint/no-explicit-any */

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Tierbewegungen');

const loading = ref(false);
const rows = ref<Record<string, any>[]>([]);
const herdeOptions = ref<any[]>([]);
const filterHerde = ref<number | null>(null);
const filterText = ref<number | null>(null);
const filterDateRange = ref<{ from: string; to: string } | null>(null);

// For Umstallung (Dual Grid)
const selectedLeft = ref<any[]>([]);
const selectedRight = ref<any[]>([]);
const amountDialog = ref(false);
const amount = ref(0);

// For Dialog
const initialHerdeId = ref<number | null>(null);
const initialTyp = ref<string>('');

const isUmstallung = computed(() => {
  if (!filterText.value) return false;
  const selected = textOptions.value.find(t => t.id === filterText.value);
  return selected?.kz === 'U';
});

const bookingColumns = [
  { name: 'actions', label: 'Aktion', field: 'actions', align: 'center', style: 'width: 80px' },
  { name: 'HERDENNUMMER', align: 'left', label: 'Herde-Nr', field: (row: any) => row.herdennummer || row.HERDENNUMMER, sortable: true },
  { name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: (row: any) => row.herden_bezeichnung || row.HERDEN_BEZEICHNUNG, sortable: true },
  { name: 'DATUM', align: 'left', label: 'Datum', field: (row: any) => extractString(row.buchungsdatum || row.BUCHUNGSDATUM), sortable: true },
  { name: 'BESTAND', align: 'right', label: 'Bestand', field: (row: any) => extractInt(row.tierbestand || row.TIERBESTAND), sortable: true },
];

const herdeLinks = computed(() => selectedLeft.value[0] || null);
const herdeRechts = computed(() => selectedRight.value[0] || null);

const isActionActive = computed(() => {
  if (selectedLeft.value.length === 0 || selectedRight.value.length === 0) return false;
  return extractString(selectedLeft.value[0].buchungsdatum) === extractString(selectedRight.value[0].buchungsdatum);
});

const filteredRightBookings = computed(() => {
  // Wenn links noch nichts ausgewählt ist, zeigen wir einfach 
  // die gleichen gefilterten Daten wie links an (Initialzustand)
  if (selectedLeft.value.length === 0) return filteredRows.value;

  const left = selectedLeft.value[0];
  const leftDatumRaw = extractString(left.buchungsdatum || left.BUCHUNGSDATUM);
  const leftDatum = leftDatumRaw.substring(0, 10);
  
  // Sobald links ausgewählt ist, zeigen wir alle Herden an,
  // die am EXAKT gleichen Tag eine Buchung haben.
  return rows.value.filter(b => {
    const rowDatumRaw = extractString(b.buchungsdatum || b.BUCHUNGSDATUM);
    const rowDatum = rowDatumRaw.substring(0, 10);
    return rowDatum === leftDatum && 
           extractInt(b.id || b.ID) !== extractInt(left.id || left.ID);
  });
});
const filterDateRangeText = computed(() => {
  if (!filterDateRange.value) return '';
  const from = filterDateRange.value.from.split('/').reverse().join('.');
  const to = filterDateRange.value.to.split('/').reverse().join('.');
  return `${from} - ${to}`;
});

function onDateRangeUpdate() {
  // Logic already handled by computed and watch
}

const textTypOptions = ref<any[]>([]);
const textOptions = ref<any[]>([]);

const timeframe = ref(7);
const timeframeOptions = [
  { label: 'Letzte 7 Tage', value: 7 },
  { label: 'Letzte 30 Tage', value: 30 },
  { label: 'Letzte 90 Tage', value: 90 },
  { label: 'Alle Buchungen', value: 0 },
];

function onTimeframeUpdate(val: number) {
  if (val === 0) {
    filterDateRange.value = null;
    return;
  }
  const end = new Date();
  const start = new Date();
  start.setDate(end.getDate() - val);
  
  filterDateRange.value = {
    from: start.toISOString().split('T')[0].replace(/-/g, '/'),
    to: end.toISOString().split('T')[0].replace(/-/g, '/')
  };
}

onMounted(async () => {
  initWidths(bookingColumns);
  onTimeframeUpdate(timeframe.value);
  try {
    await Promise.all([loadData(), fetchHerden(), fetchRelevantTexts()]);
  } catch (err) {
    console.error('Initial load failed:', err);
  }
});

// Automatisch rechts löschen wenn links gewechselt wird
watch(selectedLeft, () => {
  selectedRight.value = [];
});

const filteredRows = computed(() => {
  let list = rows.value;

  if (filterHerde.value !== null && filterHerde.value !== undefined) {
    const herdeId = Number(filterHerde.value);
    // Find herdennummer for this ID
    const herde = herdeOptions.value.find(h => h.id === herdeId);
    if (herde) {
      list = list.filter(row => extractInt(row.herdennummer || row.HERDENNUMMER) === herde.herdennummer);
    }
  }


  // Datum Filter
  if (filterDateRange.value) {
    const from = new Date(filterDateRange.value.from.replace(/\//g, '-'));
    const to = new Date(filterDateRange.value.to.replace(/\//g, '-'));
    from.setHours(0, 0, 0, 0);
    to.setHours(23, 59, 59, 999);
    
    list = list.filter(row => {
      const dateStr = extractString(row.buchungsdatum || row.BUCHUNGSDATUM);
      if (!dateStr || dateStr === '0001-01-01') return false;
      const rowDate = new Date(dateStr);
      return rowDate >= from && rowDate <= to;
    });
  }

  return list;
});

// bookingColumns moved up

function getTypLabel(val: string) {
  if (val === 'V') return 'Verkauf';
  if (val === 'T') return 'Abgang';
  if (val === 'U') return 'Umbuchung';
  if (val === 'Z') return 'Zugang';
  return val;
}

function getTypColor(val: string) {
  if (val === 'V') return 'positive';
  if (val === 'T') return 'negative';
  if (val === 'U') return 'blue';
  return 'grey';
}

interface Pagination {
  rowsPerPage: number;
  sortBy: string | null;
  descending: boolean;
  page: number;
}
const pagination = ref<Pagination>({ 
  rowsPerPage: 15, 
  sortBy: 'BEWEGUNGSDATUM', 
  descending: true,
  page: 1
});

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

// Duplicate onMounted removed

async function fetchRelevantTexts() {
  try {
    // ONLY type 'T' (Transfer/Umbuchung)
    const res = await api.get('/api/texte');
    const all = res.data || [];
    textOptions.value = all
      .filter((t: any) => (t.TEXT_TYP_KZ || t.text_typ_kz) === 'T')
      .map((t: any) => ({
        id: t.ID || t.id,
        kz: t.KZ || t.kz,
        betreff: t.BETREFF || t.betreff || `Grund ${t.ID}`
      }));
  } catch (err) {
    console.error('Error fetching texts:', err);
  }
}


async function loadData() {
  loading.value = true;
  try {
    // ONLY Performance Bookings
    const res = await api.get('/api/buchungen/detailed');
    rows.value = res.data || [];
  } catch (err) {
    console.error('Error loading movements:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Leistungsbuchungen' });
  } finally {
    loading.value = false;
  }
}

async function fetchHerden() {
  const res = await api.get('/api/herden/lookup');
  herdeOptions.value = (res.data || []).map((h: any) => {
    const id = h.id || h.ID;
    const bez = h.bezeichnung || h.BEZEICHNUNG;
    const num = h.herdennummer || h.HERDENNUMMER;
    return {
      id: id,
      bezeichnung: bez ? bez : (num ? `Herde ${num}` : `Herde ${id}`),
      herdennummer: num
    };
  });
}

function openCreate() {
  if (!filterHerde.value) {
    $q.notify({ type: 'warning', message: 'Bitte wählen Sie zuerst eine Herde aus' });
    return;
  }
  isEditing.value = false;
  editId.value = null;
  initialHerdeId.value = filterHerde.value;
  initialTyp.value = 'Z'; // Default to Zugang as per requirements
  showDialog.value = true;
}

function openCreateForBooking(row: any) {
  isEditing.value = false;
  editId.value = null;
  initialHerdeId.value = extractInt(row.id_herden || row.ID_HERDEN);
  initialTyp.value = 'Z';
  showDialog.value = true;
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
      ID_BUCHUNG_VON: extractInt(herdeLinks.value.id || herdeLinks.value.ID),
      ID_HERDEN_VON: extractInt(herdeLinks.value.id_herden || herdeLinks.value.ID_HERDEN),
      ID_BUCHUNG_NACH: extractInt(herdeRechts.value.id || herdeRechts.value.ID),
      ID_HERDEN_NACH: extractInt(herdeRechts.value.id_herden || herdeRechts.value.ID_HERDEN),
      MENGE: amount.value,
      DATUM: extractString(herdeLinks.value.buchungsdatum || herdeLinks.value.BUCHUNGSDATUM),
      ID_TEXTE: filterText.value // The selected Umstallung reason
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
</script>

<template>
  <div class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="text-h6 text-primary">Verkaufs-Statistik</div>
    </div>

    <!-- Table -->
    <q-table separator="cell"
      :rows="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      v-model:pagination="pagination"
      class="q-mb-lg cursor-pointer shadow-2 rounded-borders overflow-hidden resizable-table"
      @row-dblclick="(evt, row) => onEdit(row)"
    >
      <!-- Resizable Header Cells -->
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
        <q-td :props="props" auto-width>
          <div class="row no-wrap q-gutter-x-xs justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm">
              <q-tooltip>Bearbeiten</q-tooltip>
            </q-btn>
          </div>
        </q-td>
      </template>

      <template v-slot:body-cell-bio="props">
        <q-td :props="props" align="center">
          <q-icon :name="props.value ? 'check_circle' : 'cancel'" :color="props.value ? 'positive' : 'grey-5'" size="sm" />
        </q-td>
      </template>

      <template v-slot:body-cell-verbucht="props">
        <q-td :props="props" align="center">
          <q-chip :color="props.value ? 'positive' : 'warning'" text-color="white" dense size="sm">
            {{ props.value ? 'Ja' : 'Nein' }}
          </q-chip>
        </q-td>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 400px; max-width: 700px;" :class="$q.dark.isActive ? 'bg-grey-10 text-white' : ''">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6">{{ isEditing ? 'Verkauf bearbeiten' : 'Neuer Verkauf' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated />
        </q-card-section>

        <q-card-section>
          <div v-if="form.ID_EILAGERBUCHUNG" class="q-mb-md">
            <q-banner dense class="bg-blue-1 text-blue-9 rounded-borders">
              Dieser Verkauf ist mit einer Eilager-Umbuchung (ID {{ form.ID_EILAGERBUCHUNG }}) verknüpft. Mengen sind schreibgeschützt.
            </q-banner>
          </div>
          
          <q-form @submit="onSubmit" class="q-gutter-md">
            <div class="row q-col-gutter-md">
              <div class="col-12 col-md-6">
                <q-input v-model="form.BUCHUNGSDATUM" type="date" label="Buchungsdatum *" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => !!val || 'Datum ist ein Pflichtfeld']" :readonly="form.VERBUCHT" />
              </div>
              <div class="col-12 col-md-4">
                <q-input v-model.number="form.ID_EILAGERBUCHUNG" type="number" label="ID Eilagerbuchung" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>
              <div class="col-12 col-md-4">
                <q-input v-model.number="form.ID_BUCHUNG" type="number" label="ID Leistung" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>

              <!-- Mengen (READONLY) -->
              <div class="col-6 col-md-3">
                <q-input v-model.number="form.MENGESMALL" type="number" label="Menge S" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>
              <div class="col-6 col-md-3">
                <q-input v-model.number="form.MENGEMEDIUM" type="number" label="Menge M" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>
              <div class="col-6 col-md-3">
                <q-input v-model.number="form.MENGELARGE" type="number" label="Menge L" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>
              <div class="col-6 col-md-3">
                <q-input v-model.number="form.MENGEXL" type="number" label="Menge XL" filled stack-label readonly :bg-color="$q.dark.isActive ? 'grey-8' : 'grey-3'" />
              </div>

              <!-- Preise (EDITABLE) -->
              <div class="col-6 col-md-3">
                <q-input :model-value="formatCurrency(form.PREISSMALL)" @change="val => { form.PREISSMALL = parseCurrency(val); calcGesamt(); }" type="text" label="Preis S" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" input-class="text-right" />
              </div>
              <div class="col-6 col-md-3">
                <q-input :model-value="formatCurrency(form.PREISMEDIUM)" @change="val => { form.PREISMEDIUM = parseCurrency(val); calcGesamt(); }" type="text" label="Preis M" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" input-class="text-right" />
              </div>
              <div class="col-6 col-md-3">
                <q-input :model-value="formatCurrency(form.PREISLARGE)" @change="val => { form.PREISLARGE = parseCurrency(val); calcGesamt(); }" type="text" label="Preis L" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" input-class="text-right" />
              </div>
              <div class="col-6 col-md-3">
                <q-input :model-value="formatCurrency(form.PREISXL)" @change="val => { form.PREISXL = parseCurrency(val); calcGesamt(); }" type="text" label="Preis XL" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" input-class="text-right" />
              </div>

              <div class="col-12">
                <div class="row items-center q-gutter-sm">
                  <q-input :model-value="formatCurrency(form.GESAMTPREIS)" @change="val => { form.GESAMTPREIS = parseCurrency(val); onGesamtpreisManualChange(form.GESAMTPREIS); }" type="text" label="Gesamtpreis (€)" filled stack-label class="col" :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" input-class="text-right text-bold" />
                  <q-btn flat round icon="calculate" color="primary" @click="calcGesamt" :disable="form.VERBUCHT">
                    <q-tooltip>Neu berechnen</q-tooltip>
                  </q-btn>
                </div>
              </div>

              <div class="col-6">
                <q-input v-model="form.CHARGE" label="Charge" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :disable="isEditing" :readonly="form.VERBUCHT" />
              </div>
              <div class="col-6">
                <q-input v-model.number="form.RABATTPROZENT" type="number" step="0.01" label="Rabatt (%)" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :readonly="form.VERBUCHT" @update:model-value="updateCalculatedPrices" />
              </div>
              <div class="col-6">
                <q-toggle v-model="form.BIO" label="Bio-Ware" color="positive" :disable="form.VERBUCHT" @update:model-value="updateCalculatedPrices" />
              </div>
              <div class="col-6">
                <q-toggle v-model="form.VERBUCHT" label="Bereits verbucht" color="orange" />
              </div>
            </div>

            <div class="q-mt-md">
              <q-btn :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated />
              <q-btn label="Abbrechen" color="negative" class="q-ml-sm" @click="closeDialog" rounded unelevated />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import type { QTableProps } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

function formatCurrency(val: any) {
  if (val === null || val === undefined) return '0,00';
  const num = typeof val === 'string' ? Number(val.replace(',', '.')) : Number(val);
  if (isNaN(num)) return '0,00';
  return num.toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function parseCurrency(val: any) {
  if (!val) return 0;
  const str = String(val).replace(/\./g, '').replace(',', '.');
  const num = Number(str);
  return isNaN(num) ? 0 : num;
}

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Verkauf');

const loading = ref(false);
const showDialog = ref(false);
const isEditing = ref(false);
const firmenparameter = ref<any>(null);
const basePrices = reactive({ SMALL: 0, MEDIUM: 0, LARGE: 0, XL: 0 });

const pagination = ref({
  sortBy: 'buchungsdatum',
  descending: true,
  page: 1,
  rowsPerPage: 20
});

const form = reactive({
  ID: 0,
  ID_EILAGERBUCHUNG: 0,
  ID_BUCHUNG: 0,
  BUCHUNGSDATUM: new Date().toISOString().split('T')[0],
  MENGESMALL: 0,
  MENGEMEDIUM: 0,
  MENGELARGE: 0,
  MENGEXL: 0,
  PREISSMALL: 0,
  PREISMEDIUM: 0,
  PREISLARGE: 0,
  PREISXL: 0,
  GESAMTPREIS: 0,
  BIO: false,
  VERBUCHT: false,
  CHARGE: '',
  RABATTPROZENT: 0
});

const rows = ref([]);

const columns: QTableProps['columns'] = [
  { name: 'actions', label: 'Aktionen', field: 'actions', align: 'center', style: 'width: 80px' },
  { name: 'buchungsdatum', label: 'Datum', field: 'buchungsdatum', align: 'left', sortable: true },
  { name: 'mengesmall', label: 'S', field: 'mengesmall', align: 'right' },
  { name: 'mengemedium', label: 'M', field: 'mengemedium', align: 'right' },
  { name: 'mengelarge', label: 'L', field: 'mengelarge', align: 'right' },
  { name: 'mengexl', label: 'XL', field: 'mengexl', align: 'right' },
  { name: 'gesamtpreis', label: 'Gesamt (€)', field: 'gesamtpreis', align: 'right', format: (val: any) => Number(val).toLocaleString('de-DE', { minimumFractionDigits: 2 }) },
  { name: 'charge', label: 'Charge', field: 'charge', align: 'left' },
  { name: 'rabattprozent', label: 'Rabatt (%)', field: 'rabattprozent', align: 'right' },
  { name: 'bio', label: 'Bio', field: 'bio', align: 'center', format: (val: any) => !!val },
  { name: 'verbucht', label: 'Verb.', field: 'verbucht', align: 'center', format: (val: any) => !!val }
];

async function loadData() {
  loading.value = true;
  try {
    const [resV, resF, resP] = await Promise.all([
      api.get('/api/verkauf'),
      api.get('/api/firmenparameter/get-or-create/-1'),
      api.get('/api/eierpreise')
    ]);
    
    rows.value = (resV.data || []).map((r: any) => ({
      ...r,
      bio: r.bio === 1 || r.bio === true || r.BIO === 1 || r.BIO === true,
      verbucht: r.verbucht === 1 || r.verbucht === true || r.VERBUCHT === 1 || r.VERBUCHT === true
    }));

    firmenparameter.value = resF.data;
    
    // Default base prices from eierpreise
    if (resP.data) {
      resP.data.forEach((p: any) => {
        const val = Number(p.preis_von || p.PREIS_VON) / 100.0;
        if (p.eierklasse === 'S' || p.EIERKLASSE === 'S') basePrices.SMALL = val;
        if (p.eierklasse === 'M' || p.EIERKLASSE === 'M') basePrices.MEDIUM = val;
        if (p.eierklasse === 'L' || p.EIERKLASSE === 'L') basePrices.LARGE = val;
        if (p.eierklasse === 'XL' || p.EIERKLASSE === 'XL') basePrices.XL = val;
      });
    }
  } catch (err: any) {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Daten: ' + (err.response?.data?.error || err.message) });
  } finally {
    loading.value = false;
  }
}

function updateCalculatedPrices() {
  const isBioActive = form.BIO && firmenparameter.value?.bio === 1;
  const surcharge = isBioActive ? (Number(firmenparameter.value.bioaufschlag) || 0) : 0;
  const multiplier = 1 - (Number(form.RABATTPROZENT) / 100);

  form.PREISSMALL = Number(((basePrices.SMALL + surcharge) * multiplier).toFixed(4));
  form.PREISMEDIUM = Number(((basePrices.MEDIUM + surcharge) * multiplier).toFixed(4));
  form.PREISLARGE = Number(((basePrices.LARGE + surcharge) * multiplier).toFixed(4));
  form.PREISXL = Number(((basePrices.XL + surcharge) * multiplier).toFixed(4));
  
  calcGesamt();
}

function onGesamtpreisManualChange(newVal: number) {
  if (form.VERBUCHT) return;
  
  const isBioActive = form.BIO && firmenparameter.value?.bio === 1;
  const surcharge = isBioActive ? (Number(firmenparameter.value.bioaufschlag) || 0) : 0;
  
  // Berechne theoretischen Gesamtpreis ohne Rabatt (aber mit Bio-Aufschlag falls aktiv)
  const fullPriceTotal = 
    (form.MENGESMALL * (basePrices.SMALL + surcharge)) +
    (form.MENGEMEDIUM * (basePrices.MEDIUM + surcharge)) +
    (form.MENGELARGE * (basePrices.LARGE + surcharge)) +
    (form.MENGEXL * (basePrices.XL + surcharge));

  if (fullPriceTotal > 0) {
    const diff = fullPriceTotal - newVal;
    form.RABATTPROZENT = Number(((diff / fullPriceTotal) * 100).toFixed(2));
    updateCalculatedPrices();
  }
}

function calcGesamt() {
  form.GESAMTPREIS = Number((
    (form.MENGESMALL * form.PREISSMALL) +
    (form.MENGEMEDIUM * form.PREISMEDIUM) +
    (form.MENGELARGE * form.PREISLARGE) +
    (form.MENGEXL * form.PREISXL)
  ).toFixed(2));
}

function openCreate() {
  isEditing.value = false;
  Object.assign(form, {
    ID: 0,
    ID_EILAGERBUCHUNG: 0,
    ID_BUCHUNG: 0,
    BUCHUNGSDATUM: new Date().toISOString().split('T')[0],
    MENGESMALL: 0,
    MENGEMEDIUM: 0,
    MENGELARGE: 0,
    MENGEXL: 0,
    PREISSMALL: 0,
    PREISMEDIUM: 0,
    PREISLARGE: 0,
    PREISXL: 0,
    GESAMTPREIS: 0,
    BIO: false,
    VERBUCHT: false,
    CHARGE: '',
    RABATTPROZENT: 0
  });
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  const rabatt = row.rabattprozent || row.RABATTPROZENT || 0;
  const isBio = !!(row.bio || row.BIO);
  const surcharge = (isBio && firmenparameter.value?.bio === 1) ? (Number(firmenparameter.value.bioaufschlag) || 0) : 0;
  const multiplier = 1 - (rabatt / 100);

  // Rekonstruiere basePrices falls möglich (falls multiplier > 0)
  if (multiplier > 0) {
    basePrices.SMALL = ((row.preissmall || row.PREISSMALL || 0) / multiplier) - surcharge;
    basePrices.MEDIUM = ((row.preismedium || row.PREISMEDIUM || 0) / multiplier) - surcharge;
    basePrices.LARGE = ((row.preislarge || row.PREISLARGE || 0) / multiplier) - surcharge;
    basePrices.XL = ((row.preisxl || row.PREISXL || 0) / multiplier) - surcharge;
  }

  Object.assign(form, {
    ID: row.id || row.ID,
    ID_EILAGERBUCHUNG: row.id_eilagerbuchung || row.ID_EILAGERBUCHUNG,
    ID_BUCHUNG: row.id_buchung || row.ID_BUCHUNG,
    BUCHUNGSDATUM: row.buchungsdatum || row.BUCHUNGSDATUM,
    MENGESMALL: row.mengesmall || row.MENGESMALL,
    MENGEMEDIUM: row.mengemedium || row.MENGEMEDIUM,
    MENGELARGE: row.mengelarge || row.MENGELARGE,
    MENGEXL: row.mengexl || row.MENGEXL,
    PREISSMALL: row.preissmall || row.PREISSMALL,
    PREISMEDIUM: row.preismedium || row.PREISMEDIUM,
    PREISLARGE: row.preislarge || row.PREISLARGE,
    PREISXL: row.preisxl || row.PREISXL,
    GESAMTPREIS: row.gesamtpreis || row.GESAMTPREIS,
    BIO: isBio,
    VERBUCHT: !!(row.verbucht || row.VERBUCHT),
    CHARGE: row.charge || row.CHARGE || '',
    RABATTPROZENT: rabatt
  });
  showDialog.value = true;
}

async function onSubmit() {
  try {
    const payload = {
      ...form,
      BIO: form.BIO,
      VERBUCHT: form.VERBUCHT
    };

    if (isEditing.value) {
      await api.put(`/api/verkauf/${form.ID}`, payload);
      $q.notify({ type: 'positive', message: 'Verkauf aktualisiert' });
    } else {
      await api.post('/api/verkauf', payload);
      $q.notify({ type: 'positive', message: 'Verkauf gespeichert' });
    }
    showDialog.value = false;
    loadData();
  } catch (err: any) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern: ' + (err.response?.data?.error || err.message) });
  }
}

async function onDelete(row: any) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diesen Verkauf wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/verkauf/${row.id}`);
      $q.notify({ type: 'positive', message: 'Verkauf gelöscht' });
      loadData();
    } catch (err: any) {
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen: ' + (err.response?.data?.error || err.message) });
    }
  });
}

function closeDialog() {
  showDialog.value = false;
}

function onDialogShow() {
  // Focus helper if needed
}

onMounted(() => {
  initWidths(columns);
  loadData();
});
</script>

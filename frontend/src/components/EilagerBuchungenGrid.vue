<template>
  <div class="q-pa-md">
    <q-card flat bordered class="q-mb-md bg-grey-1">
      <q-card-section class="row items-center q-gutter-sm">
        <div class="text-h6 text-primary">Eilager Buchungen</div>
        <q-space />
        <q-select
          v-model="selectedLagerId"
          :options="filteredLagerOptions"
          option-value="id"
          option-label="dropdownLabel"
          emit-value
          map-options
          label="Lager wählen (Alle, wenn leer)"
          filled
          dense
          clearable
          style="min-width: 250px"
          @update:model-value="loadBuchungen"
        />
        <q-select
          v-model="kzFilter"
          :options="lagerTypOptions"
          option-value="kz"
          option-label="betreff"
          emit-value
          map-options
          label="Typ Filter"
          filled
          dense
          clearable
          style="min-width: 150px"
          @update:model-value="onKzFilterChange"
        />
        <q-btn color="primary" icon="add" label="Neu" @click="openCreate" rounded unelevated />
        <span class="text-caption q-ml-sm text-grey">Debug KZ: {{ currentViewKz }}</span>
      </q-card-section>
    </q-card>

    <q-table
      separator="cell"
      :rows="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      v-model:pagination="pagination"
      class="cursor-pointer shadow-2 resizable-table"
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
            <q-btn v-if="currentViewKz === 'E'" dense round icon="shopping_cart" color="orange" @click="onSell(props.row)" unelevated size="sm">
              <q-tooltip>Verkauf / Umbuchung erstellen</q-tooltip>
            </q-btn>
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm">
              <q-tooltip>Löschen</q-tooltip>
            </q-btn>
          </div>
        </q-td>
      </template>
      <template v-slot:body-cell-verkauf="props">
        <q-td :props="props" align="center">
          <q-icon :name="props.value ? 'shopping_cart' : 'remove'" :color="props.value ? 'primary' : 'grey-4'" size="xs" />
        </q-td>
      </template>
    </q-table>

    <!-- Dialog -->
    <q-dialog v-model="showDialog" persistent>
      <q-card style="min-width: 600px; max-width: 900px;">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6">{{ isEditing ? 'Buchung bearbeiten' : 'Neue Buchung' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated />
        </q-card-section>
        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-md">
            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-input v-model="form.BUCHUNGSDATUM" type="date" label="Buchungsdatum *" filled stack-label :rules="[val => !!val || 'Pflichtfeld']" />
              </div>
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="form.BUCHUNGSTYP"
                  :options="editBuchungsTypOptions"
                  option-value="kz"
                  option-label="betreff"
                  emit-value
                  map-options
                  label="Buchungstyp *"
                  filled stack-label
                  @update:model-value="onTypeChange"
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="form.ID_EILAGER"
                  :options="lagerOptions"
                  option-value="id"
                  option-label="dropdownLabel"
                  emit-value
                  map-options
                  label="Lager *"
                  filled stack-label
                  :readonly="isMove"
                  @update:model-value="updateStockReference"
                />
              </div>
              <div v-if="['U', 'V'].includes(form.BUCHUNGSTYP)" class="col-12 col-sm-6">
                <q-select
                  v-model="form.ID_FREMDESLAGER"
                  :options="targetLagerOptions"
                  option-value="id"
                  option-label="dropdownLabel"
                  emit-value
                  map-options
                  label="Ziel-Lager (für Umbuchung)"
                  filled stack-label
                  clearable
                  @update:model-value="onTargetLagerChange"
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="form.ID_LAGERPLATZ"
                  :options="filteredLagerplaetze"
                  option-value="id"
                  option-label="bezeichnung"
                  emit-value
                  map-options
                  label="Lagerplatz"
                  filled stack-label
                  clearable
                />
              </div>
              <div class="col-12 col-sm-12">
                <q-select
                  v-model="form.ID_BUCHUNG"
                  :options="buchungOptions"
                  option-value="id"
                  option-label="dropdownLabel"
                  emit-value
                  map-options
                  label="Bezug zur Legeleistung (Optional)"
                  filled stack-label
                  clearable
                  :readonly="isMove"
                  @update:model-value="onBuchungChange"
                  hint="Wähle eine Leistung aus für Restmengen-Prüfung und Lager-Vorbelegung"
                />
              </div>
            </div>


            <!-- Row 1: Stock Reference -->
            <div class="row q-col-gutter-sm items-center q-mb-sm">
              <div class="col-12 col-md-2 text-weight-bold text-primary">Bestand Lager:</div>
              <div class="col-4 col-md" v-for="s in eggFields" :key="'stock-'+s">
                <q-input :model-value="availableQuantities[s as keyof typeof availableQuantities].toString()" :label="s" dense filled stack-label readonly />
              </div>
            </div>

            <q-separator class="q-my-md" />

            <!-- Row 2: Mengen Inputs -->
            <div class="row q-col-gutter-sm items-center">
              <div class="col-12 col-md-2 text-weight-bold text-secondary">Mengen:</div>
              <div class="col-4 col-md" v-for="s in eggFields" :key="'input-'+s">
                 <q-input v-model.number="form[s as keyof typeof form]" type="number" :label="s" dense filled stack-label
                          :bg-color="$q.dark.isActive ? 'grey-10' : 'blue-1'" 
                          :error="!!form.ID_BUCHUNG && isFieldNegative(s)"
                          hide-bottom-space/>
              </div>
              <div class="col-4 col-md"><q-input v-model.number="form.SCHMUTZ" type="number" label="Schmutz" dense filled stack-label /></div>
              <div class="col-4 col-md"><q-input v-model.number="form.KNICKEIER" type="number" label="Knick" dense filled stack-label /></div>
              <div class="col-4 col-md"><q-input v-model.number="form.BRUCHEIER" type="number" label="Bruch" dense filled stack-label /></div>
            </div>

            <div class="row q-col-gutter-sm q-mt-sm">
              <div class="col-12 col-sm-6"><q-input v-model="form.CHARGE" label="Charge" filled stack-label :disable="isEditing" /></div>
              <div class="col-12 col-sm-6 flex items-center">
                <q-checkbox v-model="form.VERKAUF" label="Automatisch Verkauf anlegen" color="primary" />
              </div>
            </div>

            <div class="q-mt-md row justify-end">
              <q-btn label="Abbrechen" color="negative" outline class="q-mr-sm" @click="closeDialog" rounded unelevated />
              <q-btn :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated
                     :disable="!!form.ID_BUCHUNG && hasValidationError"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useQuasar, date } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('EilagerBuchungen');

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const currentViewKz = computed(() => {
  if (selectedLagerId.value) {
    const l = lagerOptions.value.find(o => extractInt(o.id || o.ID) === selectedLagerId.value);
    return l ? extractString(l.kz || l.KZ) : '';
  }
  return kzFilter.value || '';
});

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  if (typeof val === 'object' && 'Float64' in val) return Number(val.Float64) || 0;
  return Number(val) || 0;
};

const eggFields = ['JUMBOS', 'XL', 'LARGE', 'MEDIUM', 'SMALL', 'VOLLEIKG'] as const;

const loading = ref(false);
const rows = ref<Record<string, unknown>[]>([]);
const lagerOptions = ref<LagerOption[]>([]);
const lagerTypOptions = ref<{ kz: string; betreff: string }[]>([]);
const allBuchungsTypOptions = ref<{ kz: string; betreff: string }[]>([]);
const editBuchungsTypOptions = ref<{ kz: string; betreff: string }[]>([]);
const lagerplaetze = ref<Lagerplatz[]>([]);
const selectedLagerId = ref<number | null>(null);
const kzFilter = ref<string | null>(null);

const filteredLagerOptions = computed(() => {
  if (!kzFilter.value) return lagerOptions.value;
  return lagerOptions.value.filter(l => extractString(l.kz || l.KZ) === kzFilter.value);
});

const destinationLagerId = computed(() => {
  if (['U', 'V'].includes(form.BUCHUNGSTYP)) {
    return form.ID_FREMDESLAGER;
  }
  return form.ID_EILAGER;
});

const filteredLagerplaetze = computed(() => {
  const targetId = destinationLagerId.value;
  if (!targetId) return [];
  return lagerplaetze.value.filter(lp => lp.id_eilager === targetId);
});

const targetLagerOptions = computed(() => {
  return lagerOptions.value.filter(l => l.id !== form.ID_EILAGER);
});

const pagination = ref({ rowsPerPage: 15 });
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const isMove = ref(false);

const availableQuantities = reactive({JUMBOS: 0, XL: 0, LARGE: 0, MEDIUM: 0, SMALL: 0, VOLLEIKG: 0});

const form = reactive({
  ID_EILAGER: null as number | null,
  ID_BUCHUNG: null as number | null,
  ID_FREMDESLAGER: null as number | null,
  ID_FREMDEBUCHUNG: null as number | null,
  ID_LAGERPLATZ: null as number | null,
  BUCHUNGSDATUM: '',
  BUCHUNGSTYP: 'E',
  JUMBOS: 0, XL: 0, LARGE: 0, MEDIUM: 0, SMALL: 0, VOLLEIKG: 0,
  SCHMUTZ: 0, KNICKEIER: 0, BRUCHEIER: 0,
  CHARGE: '',
  VERKAUF: false
});

function isFieldNegative(s: string) {
  const available = (availableQuantities as any)[s] || 0;
  const input = (form as any)[s] || 0;
  return (available - Number(input)) < -0.001;
}

function getDiff(s: string) {
  const available = (availableQuantities as any)[s] || 0;
  const input = (form as any)[s] || 0;
  const diff = available - Number(input);
  return diff.toFixed(s === 'VOLLEIKG' ? 2 : 0);
}

const hasValidationError = computed(() => {
  return eggFields.some(s => isFieldNegative(s));
});

const validationErrorMessage = computed(() => {
  const over = eggFields.filter(s => isFieldNegative(s));
  return `Eingabe übersteigt Bestand bei: ${over.join(', ')}`;
});

const displayStock = availableQuantities;

const validationLabel = computed(() => {
  if (isMove.value) return 'Prüfung gegen gewählte Zeile (Ursprung)';
  return 'Restmengen aus Leistung';
});

const buchungOptions = ref<any[]>([]);

async function loadBuchungen() {
  if (!selectedLagerId.value && !kzFilter.value) {
    rows.value = [];
    return;
  }
  loading.value = true;
  try {
    const url = selectedLagerId.value ? `/api/eilagerbuchungen/lager/${selectedLagerId.value}` : `/api/eilagerbuchungen/kz/${kzFilter.value}`;
    const res = await api.get(url);
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Buchungen' });
  } finally {
    loading.value = false;
  }
}

async function fetchPerformanceOptions() {
  try {
    const res = await api.get('/api/buchung');
    buchungOptions.value = (res.data || []).map((b: any) => {
      const hNum = extractInt(b.HERDEN_NUMMER_REL || b.herden_nummer_rel) || extractInt(b.ID_HERDEN || b.id_herden);
      const hBez = extractString(b.HERDEN_BEZEICHNUNG_REL || b.herden_bezeichnung_rel);
      const hLabel = hBez ? `${hNum} (${hBez})` : `${hNum}`;
      return {
        id: b.ID || b.id,
        dropdownLabel: `${date.formatDate(b.BUCHUNGSDATUM || b.buchungsdatum, 'DD.MM.YYYY')} - Herde ${hLabel}`,
        data: b
      };
    });
  } catch (err) {
    console.error('Fehler beim Laden der Leistungen', err);
  }
}

function updateStockReference() {
  // Nur relevant für Stall -> Lager (Initial)
  if (isMove.value) return;
  // logic handled in onBuchungChange
}

async function fetchSourceUsedStock(buchungId: number, lagerId: number) {
  try {
    const res = await api.get(`/api/eilagerbuchungen/sum-by-source/${buchungId}/${lagerId}`);
    if (res.data && isMove.value) {
      // In this mode, we ALREADY set availableQuantities from the row.
      // We don't want to subtract anything else.
    }
  } catch (err) {
    console.error('Error fetching source used stock:', err);
  }
}

async function onBuchungChange(val: number | null) {
  if (!val || isMove.value) return;

  try {
    const res = await api.get(`/api/buchung/${val}`);
    if (res.data) {
      availableQuantities.JUMBOS = extractInt(res.data.kl6 || res.data.KL6);
      availableQuantities.XL = extractInt(res.data.xl || res.data.XL);
      availableQuantities.LARGE = extractInt(res.data.large || res.data.LARGE);
      availableQuantities.MEDIUM = extractInt(res.data.medium || res.data.MEDIUM);
      availableQuantities.SMALL = extractInt(res.data.small || res.data.SMALL);
      availableQuantities.VOLLEIKG = extractInt(res.data.vollei || res.data.VOLLEI);
      
      // Used Stock abziehen
      const resUsed = await api.get(`/api/eilagerbuchungen/sum-by-buchung/${val}`);
      if (resUsed.data) {
        availableQuantities.JUMBOS -= extractInt(resUsed.data.jumbos);
        availableQuantities.XL -= extractInt(resUsed.data.xl);
        availableQuantities.LARGE -= extractInt(resUsed.data.large);
        availableQuantities.MEDIUM -= extractInt(resUsed.data.medium);
        availableQuantities.SMALL -= extractInt(resUsed.data.small);
        availableQuantities.VOLLEIKG -= extractInt(resUsed.data.volleikg);
      }
    }
  } catch (err) {
    console.error('Fehler beim Laden der Buchungsdaten:', err);
  }
}

function resetForm() {
  form.ID_EILAGER = null;
  form.ID_BUCHUNG = null;
  form.ID_FREMDESLAGER = null;
  form.ID_FREMDEBUCHUNG = null;
  form.BUCHUNGSDATUM = '';
  form.BUCHUNGSTYP = 'A';
  form.JUMBOS = 0; form.XL = 0; form.LARGE = 0; form.MEDIUM = 0; form.SMALL = 0; form.VOLLEIKG = 0;
  form.SCHMUTZ = 0; form.KNICKEIER = 0; form.BRUCHEIER = 0;
  form.CHARGE = '';
  form.VERKAUF = false;
  eggFields.forEach(f => availableQuantities[f] = 0);
}

function openCreate() {
  isMove.value = false;
  resetForm();
  isEditing.value = false;
  editId.value = null;
  form.ID_EILAGER = selectedLagerId.value;
  form.BUCHUNGSDATUM = date.formatDate(new Date(), 'YYYY-MM-DD');
  void fetchPerformanceOptions();
  showDialog.value = true;
}

function onEdit(row: any) {
  try {
    const buchungstyp = extractString(row.BUCHUNGSTYP || row.buchungstyp);
    isMove.value = !!extractInt(row.ID_FREMDESLAGER || row.id_fremdeslager) || ['U', 'V'].includes(buchungstyp);
    isEditing.value = true;
    editId.value = extractInt(row.ID || row.id);
    
    if (['U', 'V'].includes(buchungstyp)) {
      form.ID_EILAGER = extractInt(row.ID_FREMDESLAGER || row.id_fremdeslager);
      form.ID_FREMDESLAGER = extractInt(row.ID_EILAGER || row.id_eilager);
    } else {
      form.ID_EILAGER = extractInt(row.ID_EILAGER || row.id_eilager);
      form.ID_FREMDESLAGER = extractInt(row.ID_FREMDESLAGER || row.id_fremdeslager);
    }
    
    form.ID_BUCHUNG = extractInt(row.ID_BUCHUNG || row.id_buchung);
    form.BUCHUNGSDATUM = parseTimestampToDate(row.BUCHUNGSDATUM || row.buchungsdatum);
    form.BUCHUNGSTYP = extractString(row.BUCHUNGSTYP || row.buchungstyp);
    
    form.JUMBOS = extractInt(row.JUMBOS || row.jumbos);
    form.XL = extractInt(row.XL || row.xl);
    form.LARGE = extractInt(row.LARGE || row.large);
    form.MEDIUM = extractInt(row.MEDIUM || row.medium);
    form.SMALL = extractInt(row.SMALL || row.small);
    form.VOLLEIKG = extractInt(row.VOLLEIKG || row.volleikg);
    form.SCHMUTZ = extractInt(row.SCHMUTZ || row.schmutz);
    form.KNICKEIER = extractInt(row.KNICKEIER || row.knickeier);
    form.BRUCHEIER = extractInt(row.BRUCHEIER || row.brucheier);
    form.ID_LAGERPLATZ = extractInt(row.ID_LAGERPLATZ || row.id_lagerplatz);
    form.CHARGE = extractString(row.CHARGE || row.charge);
    form.VERKAUF = !!extractInt(row.VERKAUF || row.verkauf);
    form.ID_FREMDEBUCHUNG = extractInt(row.ID_FREMDEBUCHUNG || row.id_fremdebuchung);

    if (isMove.value) {
      availableQuantities.JUMBOS = form.JUMBOS;
      availableQuantities.XL = form.XL;
      availableQuantities.LARGE = form.LARGE;
      availableQuantities.MEDIUM = form.MEDIUM;
      availableQuantities.SMALL = form.SMALL;
      availableQuantities.VOLLEIKG = form.VOLLEIKG;
    }

    void fetchPerformanceOptions().then(() => {
      if (form.ID_BUCHUNG && !isMove.value) void onBuchungChange(form.ID_BUCHUNG);
    });
    showDialog.value = true;
  } catch (err) {
    console.error('Error in onEdit:', err);
  }
}

function onSell(row: any) {
  try {
    isMove.value = true;
    resetForm();
    isEditing.value = false;
    editId.value = null;

    form.ID_EILAGER = extractInt(row.ID_EILAGER || row.id_eilager);
    form.ID_BUCHUNG = extractInt(row.ID_BUCHUNG || row.id_buchung);
    form.BUCHUNGSDATUM = date.formatDate(Date.now(), 'YYYY-MM-DD');
    form.BUCHUNGSTYP = 'V';
    form.CHARGE = extractString(row.CHARGE || row.charge);
    form.ID_FREMDEBUCHUNG = extractInt(row.ID || row.id);
    form.VERKAUF = true;

    // Automatisch das Verkaufslager (kz = 'V') als Ziel auswählen, falls vorhanden
    const verkaufslager = lagerOptions.value.find((l: any) => l.kz === 'V' || l.KZ === 'V');
    if (verkaufslager) {
      form.ID_FREMDESLAGER = extractInt(verkaufslager.id || verkaufslager.ID);
    } else {
      form.ID_FREMDESLAGER = null;
    }

    availableQuantities.JUMBOS = extractInt(row.JUMBOS || row.jumbos);
    availableQuantities.XL = extractInt(row.XL || row.xl);
    availableQuantities.LARGE = extractInt(row.LARGE || row.large);
    availableQuantities.MEDIUM = extractInt(row.MEDIUM || row.medium);
    availableQuantities.SMALL = extractInt(row.SMALL || row.small);
    availableQuantities.VOLLEIKG = extractInt(row.VOLLEIKG || row.volleikg);

    form.JUMBOS = availableQuantities.JUMBOS;
    form.XL = availableQuantities.XL;
    form.LARGE = availableQuantities.LARGE;
    form.MEDIUM = availableQuantities.MEDIUM;
    form.SMALL = availableQuantities.SMALL;
    form.VOLLEIKG = availableQuantities.VOLLEIKG;

    void fetchPerformanceOptions();
    showDialog.value = true;
  } catch (err) {
    console.error('Error in onSell:', err);
  }
}

async function onSubmit() {
  if (hasValidationError.value) {
    $q.notify({ type: 'negative', message: 'Mengen prüfen!' });
    return;
  }
  try {
    const data = { ...form };
    
    // Swap the source and target so the DB record formally belongs to the target warehouse
    const isTransfer = ['U', 'V'].includes(data.BUCHUNGSTYP);
    if (isTransfer) {
      data.ID_EILAGER = form.ID_FREMDESLAGER;
      data.ID_FREMDESLAGER = form.ID_EILAGER;
    }
    
    if (isEditing.value && editId.value) {
      await api.put(`/api/eilagerbuchungen/${editId.value}`, data);
      $q.notify({ type: 'positive', message: 'Aktualisiert' });
    } else {
      await api.post('/api/eilagerbuchungen', data);
      $q.notify({ type: 'positive', message: 'Gespeichert' });
    }
    showDialog.value = false;
    void loadBuchungen();
  } catch (err: any) {
    $q.notify({ type: 'negative', message: err.response?.data?.error || 'Fehler' });
  }
}

function onDelete(row: any) {
  $q.dialog({ title: 'Löschen', message: 'Wirklich?', cancel: true }).onOk(async () => {
    try {
      await api.delete(`/api/eilagerbuchungen/${row.ID || row.id}`);
      void loadBuchungen();
    } catch {
      $q.notify({ type: 'negative', message: 'Fehler' });
    }
  });
}

function onTargetLagerChange() { updateStockReference(); }
function onTypeChange() { }
function onKzFilterChange() { selectedLagerId.value = null; void loadBuchungen(); }
function closeDialog() { showDialog.value = false; }

function formatDate(val: any) {
  if (!val) return '';
  const d = (val && val.String) ? val.String : val;
  return date.formatDate(d, 'DD.MM.YYYY');
}

function parseTimestampToDate(val: any) {
  if (!val) return '';
  const d = (val && val.String) ? val.String : val;
  return date.formatDate(d, 'YYYY-MM-DD');
}

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  { name: 'lager', align: 'left', label: 'Lager', field: (row: any) => {
    const id = extractInt(row.id_eilager || row.ID_EILAGER);
    const l = lagerOptions.value.find(o => o.id === id);
    return l ? l.dropdownLabel : id;
  }, sortable: true },
  { name: 'datum', align: 'left', label: 'Datum', field: (row: any) => formatDate(row.BUCHUNGSDATUM || row.buchungsdatum), sortable: true },
  { name: 'typ', align: 'left', label: 'Typ', field: (row: any) => {
    const kz = extractString(row.BUCHUNGSTYP || row.buchungstyp);
    const opt = allBuchungsTypOptions.value.find(o => o.kz === kz);
    return opt ? opt.betreff : kz;
  }, sortable: true },
  { name: 'jumbos', align: 'right', label: 'Jumbos', field: (row: any) => extractInt(row.JUMBOS || row.jumbos), format: (v: number) => v.toLocaleString('de-DE') },
  { name: 'xl', align: 'right', label: 'XL', field: (row: any) => extractInt(row.XL || row.xl), format: (v: number) => v.toLocaleString('de-DE') },
  { name: 'large', align: 'right', label: 'L', field: (row: any) => extractInt(row.LARGE || row.large), format: (v: number) => v.toLocaleString('de-DE') },
  { name: 'medium', align: 'right', label: 'M', field: (row: any) => extractInt(row.MEDIUM || row.medium), format: (v: number) => v.toLocaleString('de-DE') },
  { name: 'small', align: 'right', label: 'S', field: (row: any) => extractInt(row.SMALL || row.small), format: (v: number) => v.toLocaleString('de-DE') },
  { name: 'charge', align: 'left', label: 'Charge', field: (row: any) => extractString(row.CHARGE || row.charge) },
  { name: 'verkauf', align: 'center', label: 'VK', field: (row: any) => !!extractInt(row.VERKAUF || row.verkauf) }
];

interface LagerOption { id: number; lagernummer: any; bezeichnung: any; kz: any; dropdownLabel: string; jumbos?: any; xl?: any; large?: any; medium?: any; small?: any; volleikg?: any; }
interface Lagerplatz { id: number; id_eilager: number; bezeichnung: string; }
interface BuchungOption { id: number; dropdownLabel: string; data: any; }

async function loadLagerOptions() {
  try {
    const [resL, resP] = await Promise.all([api.get('/api/eilager'), api.get('/api/lagerplatz')]);
    lagerOptions.value = (resL.data || []).map((l: any) => ({ 
      ...l, 
      id: extractInt(l.id || l.ID),
      lagernummer: extractString(l.lagernummer || l.LAGERNUMMER),
      dropdownLabel: `${extractString(l.lagernummer || l.LAGERNUMMER) || '?'} - ${extractString(l.bezeichnung || l.BEZEICHNUNG) || 'Lager'}` 
    }));
    lagerplaetze.value = (resP.data || []).map((lp: any) => ({ 
      id: extractInt(lp.id || lp.ID), 
      id_eilager: extractInt(lp.id_eilager || lp.ID_EILAGER), 
      bezeichnung: extractString(lp.bezeichnung || lp.BEZEICHNUNG) 
    }));
    if (selectedLagerId.value === null && kzFilter.value) {
      void loadBuchungen();
    }
  } catch (err) { console.error(err); }
}

async function loadTypeOptions() {
  try {
    const [resL, resE] = await Promise.all([api.get('/api/texte/typ/L'), api.get('/api/texte/typ/E')]);
    lagerTypOptions.value = (resL.data || []).map((t: any) => ({ kz: extractString(t.kz || t.KZ), betreff: extractString(t.betreff || t.BETREFF) }));
    
    const alleTypen = (resE.data || []).map((t: any) => ({ kz: extractString(t.kz || t.KZ), betreff: extractString(t.betreff || t.BETREFF) }));
    allBuchungsTypOptions.value = alleTypen;
    editBuchungsTypOptions.value = alleTypen.filter((t: any) => ['A', 'U', 'R', 'V'].includes(t.kz));
    
    if (!kzFilter.value) {
      kzFilter.value = 'E';
    }
  } catch (err) { console.error(err); }
}

onMounted(async () => { 
  initWidths(columns);
  await loadTypeOptions();
  await loadLagerOptions(); 
  if (selectedLagerId.value === null && kzFilter.value) {
    void loadBuchungen();
  }
});
</script>

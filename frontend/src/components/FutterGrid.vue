<template>
  <div class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="row q-gutter-md">
        <q-btn color="primary" icon="add" label="Neue Futterbuchung" @click="openCreate" rounded unelevated />
        <q-btn v-if="futterinventurActive" color="secondary" icon="assignment" label="Futterinventur" @click="openInventur" rounded unelevated />
      </div>
      <div class="text-h6 text-primary">Futter-Buchung</div>
    </div>

    <div class="row q-col-gutter-md q-mb-md items-center">
      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterSilo"
          :options="siloOptions"
          option-value="id"
          option-label="label"
          emit-value
          map-options
          clearable
          label="Silo filtern"
          filled
          stack-label
          dense
        >
          <template v-slot:prepend>
            <q-icon name="filter_list" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterText"
          :options="textOptions"
          option-value="id"
          option-label="bezeichnung"
          emit-value
          map-options
          clearable
          label="Grund / Sorte"
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
    <q-table
      :rows="filteredRows"
      :columns="columns"
      row-key="ID"
      :loading="loading"
      :pagination="pagination"
      @update:pagination="(val) => { pagination = val }"
      class="huhnlite-grid-standard resizable-table q-mb-lg shadow-2 cursor-pointer"
      :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
      :dark="$q.dark.isActive"
      table-header-class="text-weight-bold"
      separator="cell"
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
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm" />
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm" />
          </div>
        </q-td>
      </template>

      <!-- Custom cell for numbers -->
      <template v-slot:body-cell="props">
        <q-td :props="props">
          <template v-if="typeof props.value === 'number'">
            {{ props.value.toLocaleString('de-DE', { minimumFractionDigits: (props.col.name === 'SILONUMMER' ? 0 : 2), maximumFractionDigits: 2 }) }}
          </template>
          <template v-else>
            {{ props.value }}
          </template>
        </q-td>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent full-width @show="onDialogShow">
      <q-card style="max-width: 1000px; margin: auto; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('auto.edit_futter') : t('auto.new_futter') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            
            <!-- Globaler Form Header (nur bei Neuanlage) -->
            <FormHeader v-if="!isEditing" v-model="form.fullTimestamp" />

            <!-- Messdaten Zeile (Datum, Silo, Herde) -->
            <div class="row q-col-gutter-md items-start">
               <div class="col-12 col-sm-4">
                <q-input 
                  :model-value="form.fullTimestamp ? form.fullTimestamp.split(' ')[0] : ''" 
                  type="date" 
                  label="Lieferdatum" 
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :dark="$q.dark.isActive"
                  readonly 
                  hide-bottom-space 
                />
              </div>
              <div class="col-12 col-sm-4">
                <q-select
                   v-model="form.ID_SILO"
                   :options="siloOptions"
                   option-value="id"
                   option-label="label"
                  emit-value
                  map-options
                  label="Silo *"
                  filled
                  stack-label
                  hide-bottom-space
                  :rules="[val => !!val || 'Erforderlich']"
                  @update:model-value="onSiloChange"
                />
              </div>
            </div>

            <q-separator />

            <!-- Lieferant & Menge -->
            <div class="row q-col-gutter-md">
              <div class="col-12 col-md-4">
                <q-select
                   v-model="form.ID_PERSON"
                   :options="personOptions"
                   option-value="ID"
                   option-label="LABEL"
                  emit-value
                  map-options
                  label="Lieferant"
                  filled
                  stack-label
                />
              </div>
              <div class="col-12 col-md-4">
                <q-select
                  v-model="form.ID_FUTTERSORTEN"
                  :options="dialogTextOptions"
                  option-value="id"
                  option-label="bezeichnung"
                  emit-value
                  map-options
                  label="Grund / Sorte"
                  filled
                  stack-label
                  :rules="[val => !!val || 'Erforderlich']"
                />
              </div>
              <div class="col-12 col-sm-6 col-md-2">
                <q-input 
                  v-model="displayForm.LIEFERMENGE" 
                  label="Menge (kg)" 
                  filled 
                  stack-label 
                  @update:model-value="val => form.LIEFERMENGE = parseNumberLocalized(String(val))"
                  @blur="displayForm.LIEFERMENGE = formatNumberLocalized(form.LIEFERMENGE, 2)"
                />
              </div>
              <div class="col-12 col-sm-6 col-md-2">
                <q-input 
                  v-model="displayForm.PREISDT" 
                  label="Preis / dt" 
                  filled 
                  stack-label 
                  @update:model-value="val => form.PREISDT = parseNumberLocalized(String(val))"
                  @blur="displayForm.PREISDT = formatNumberLocalized(form.PREISDT, 2)"
                />
              </div>
            </div>

            <q-separator />

            <!-- Preise & MwSt -->
            <div class="row q-col-gutter-md items-center">
              <div class="col-12 col-sm-6 col-md-3">
                <q-input 
                  v-model="displayForm.NETTO" 
                  label="Netto (€)" 
                  filled 
                  stack-label 
                  @update:model-value="val => form.NETTO = parseNumberLocalized(String(val))"
                  @blur="displayForm.NETTO = formatNumberLocalized(form.NETTO, 2)"
                />
              </div>
              <div class="col-12 col-sm-6 col-md-3">
                 <q-select
                   v-model="form.MWSTKZ"
                   :options="mwstOptions"
                   option-value="MWSTKZ"
                   option-label="LABEL"
                  emit-value
                  map-options
                  label="MwSt"
                  filled
                  stack-label
                  @update:model-value="onMwstChange"
                />
              </div>
              <div class="col-12 col-sm-6 col-md-3">
                <q-input 
                  v-model="displayForm.BRUTTO" 
                  label="Brutto (€)" 
                  filled 
                  stack-label 
                  @update:model-value="val => form.BRUTTO = parseNumberLocalized(String(val))"
                  @blur="displayForm.BRUTTO = formatNumberLocalized(form.BRUTTO, 2)"
                />
              </div>
               <div class="col-12 col-sm-6 col-md-3">
                <q-input 
                  v-model="displayForm.RABATTPROZ" 
                  label="Rabatt (%)" 
                  filled 
                  stack-label 
                  @update:model-value="val => form.RABATTPROZ = parseNumberLocalized(String(val))"
                  @blur="displayForm.RABATTPROZ = formatNumberLocalized(form.RABATTPROZ, 2)"
                />
              </div>
            </div>

            <div class="row justify-end q-mt-xl q-gutter-x-sm">
              <q-btn ref="cancelBtn" :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog" />
              <q-btn ref="saveBtn" :label="isEditing ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Futterinventur Dialog -->
    <q-dialog v-model="showInventurDialog" persistent>
      <q-card style="min-width: 400px; max-width: 500px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-secondary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">Futterinventur</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup unelevated color="white" flat />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmitInventur" class="q-gutter-md">
            <q-select
              v-model="inventurForm.ID_SILO"
              :options="siloOptions"
              option-value="id"
              option-label="label"
              emit-value
              map-options
              label="Silo *"
              filled
              stack-label
              :rules="[val => !!val || 'Erforderlich']"
            />
            
            <q-input
              v-model="inventurForm.INVENTURDATUMNEU"
              type="date"
              label="Inventurdatum *"
              filled
              stack-label
              :rules="[val => !!val || 'Erforderlich']"
            />

            <q-input
              v-model.number="inventurForm.INVENTURFUELLMENGE"
              type="number"
              label="Inventurmenge (kg) *"
              filled
              stack-label
              :rules="[val => val !== null && val !== '' || 'Erforderlich']"
            />

            <div class="row justify-end q-mt-xl q-gutter-x-sm">
              <q-btn label="Abbrechen" color="negative" outline rounded v-close-popup />
              <q-btn label="Speichern" type="submit" color="primary" rounded unelevated padding="xs xl" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
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

const extractFloat = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Float64' in val) return Number(val.Float64) || 0;
  return Number(val) || 0;
};

const formatNumberLocalized = (val: any, decimals = 2) => {
  const n = typeof val === 'number' ? val : extractFloat(val);
  if (isNaN(n)) return '';
  return n.toLocaleString('de-DE', { minimumFractionDigits: decimals, maximumFractionDigits: decimals });
};

const parseNumberLocalized = (val: string) => {
  if (!val) return 0;
  const sanitized = val.toString().replace(/\./g, '').replace(',', '.');
  return parseFloat(sanitized) || 0;
};

import { ref, reactive, onMounted, computed, watch } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import type { QTableProps } from 'quasar';
import FormHeader from '../components/FormHeader.vue';
import { useSessionStore } from '../stores/session';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const sessionStore = useSessionStore();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Futter');

const loading = ref(false);

interface Futter {
  ID: number;
  ID_SILO: number;
  ID_HERDEN?: number;
  ID_PERSON: number;
  LIEFERDATUM: string;
  LIEFERMENGE: number;
  NETTO: number;
  BRUTTO: number;
  SILONUMMER?: number;
  PREISDT?: number;
  RABATTPROZ?: number;
  MWSTPROZ?: number;
  MWSTKZ?: string;
  DATUM?: string;
  ZEITSTEMPEL?: string;
  FUTTERSORTE_TEXT?: string;
  ID_FUTTERSORTEN?: number;
}

const rows = ref<Futter[]>([]);
const siloOptions = ref<{ label: string; id: number; silonummer?: number }[]>([]);
const rawSilos = ref<any[]>([]);
const firmenparameter = ref<any>(null);
const futterinventurActive = computed(() => {
  if (!firmenparameter.value) return false;
  const p = firmenparameter.value.data || firmenparameter.value;
  return extractInt(p.futterinventur || p.FUTTERINVENTUR) === 1;
});
const showInventurDialog = ref(false);
const inventurForm = reactive({
  ID_SILO: null as number | null,
  INVENTURDATUMNEU: '',
  INVENTURFUELLMENGE: null as number | null
});
const personOptions = ref<{ LABEL: string; ID: number }[]>([]);
const mwstOptions = ref<{ LABEL: string; MWSTKZ: string; PROZENT: number }[]>([]);
const filterSilo = ref<number | null>(null);

const textOptions = ref<any[]>([]);
const filterText = ref<number | null>(null);

const filteredRows = computed(() => {
  let list = rows.value;

  if (filterSilo.value) {
    list = list.filter(row => row.ID_SILO === filterSilo.value);
  }

  if (filterText.value) {
    const textId = Number(filterText.value);
    list = list.filter(row => extractInt((row as any).id_futtersorten || (row as any).ID_FUTTERSORTEN) === textId);
  }

  return list;
});

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  { name: 'LIEFERDATUM', align: 'left', label: 'Datum', field: (row: any) => extractString(row.lieferdatum || row.LIEFERDATUM) || '-', sortable: true },
  { name: 'SILONUMMER', align: 'right', label: 'Silo-Nr', field: (row: any) => extractInt(row.silonummer || row.SILONUMMER) || '-', sortable: true },
  { name: 'FUTTERSORTE_TEXT', align: 'left', label: 'Grund / Sorte', field: (row: any) => extractString(row.futtersorte_text || row.FUTTERSORTE_TEXT) || '-', sortable: true },
  { name: 'LIEFERMENGE', align: 'right', label: 'Menge (kg)', field: (row: any) => extractFloat(row.liefermenge || row.LIEFERMENGE), sortable: true },
  { name: 'NETTO', align: 'right', label: 'Netto (€)', field: (row: any) => extractFloat(row.netto || row.NETTO), sortable: true },
  { name: 'BRUTTO', align: 'right', label: 'Brutto (€)', field: (row: any) => extractFloat(row.brutto || row.BRUTTO), sortable: true }
];

interface Pagination {
  rowsPerPage: number;
  sortBy: string | null;
  descending: boolean;
  page: number;
}
const pagination = ref<Pagination>({ 
  rowsPerPage: 15, 
  sortBy: 'LIEFERDATUM', 
  descending: true,
  page: 1
});

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = reactive({
  ID_SILO: null as number | null,
  SILONUMMER: null as number | null,
  ID_HERDEN: null as number | null,
  ID_PERSON: null as number | null,
  LIEFERDATUM: '',
  fullTimestamp: '',
  LIEFERMENGE: null as number | null,
  PREISDT: null as number | null,
  RABATTPROZ: null as number | null,
  NETTO: null as number | null,
  BRUTTO: null as number | null,
  MWSTPROZ: null as number | null,
  MWSTKZ: '',
  DATUM: '',
  ZEITSTEMPEL: '',
  ID_FUTTERSORTEN: null as number | null
});

const displayForm = reactive({
  LIEFERMENGE: '',
  PREISDT: '',
  NETTO: '',
  BRUTTO: '',
  RABATTPROZ: ''
});

const originalFormState = ref('');
const isDirty = computed(() => JSON.stringify(form) !== originalFormState.value);
const cancelBtn = ref<{ $el: HTMLElement } | null>(null);
const saveBtn = ref<{ $el: HTMLElement } | null>(null);

function onDialogShow() {
  originalFormState.value = JSON.stringify(form);
  setTimeout(() => { (cancelBtn.value)?.$el?.focus(); }, 50);
}

watch(isDirty, (dirty: boolean) => {
  if (dirty && (document.activeElement === (cancelBtn.value)?.$el || document.activeElement === document.body)) {
    (saveBtn.value)?.$el?.focus();
  }
});

onMounted(async () => {
  initWidths(columns);
  try {
    await Promise.all([loadData(), fetchSilos(), fetchPersonen(), fetchMwst(), fetchFutterSorten()]);
    const resF = await api.get('/api/firmenparameter/get-or-create/-1');
    firmenparameter.value = resF.data;
  } catch (err) {
    console.error('Initial load failed:', err);
  }
});

const dialogTextOptions = ref<any[]>([]);
async function fetchFutterSorten() {
  try {
    const res = await api.get('/api/futtersorten');
    const sorten = res.data || [];
    textOptions.value = [
      { id: 0, bezeichnung: 'Alle' },
      ...sorten.map((s: any) => ({
        id: s.id || s.ID,
        bezeichnung: s.bezeichnung || s.BEZEICHNUNG
      }))
    ];
    dialogTextOptions.value = sorten.map((s: any) => ({
      id: s.id || s.ID,
      bezeichnung: s.bezeichnung || s.BEZEICHNUNG
    }));
  } catch (err) {
    console.error('Error fetching futter sorten:', err);
  }
}

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/futter');
    rows.value = res.data || [];
  } catch (err) {
    console.error('Load data error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Daten' });
  } finally {
    loading.value = false;
  }
}

async function fetchSilos() {
  const res = await api.get('/api/silo');
  rawSilos.value = res.data || [];
  siloOptions.value = rawSilos.value.map((s: any) => ({
    label: `${s.silonummer || s.SILONUMMER || '?'}: ${s.bezeichnung || s.BEZEICHNUNG || ''}`,
    id: s.id || s.ID,
    silonummer: s.silonummer || s.SILONUMMER
  }));
}

async function fetchPersonen() {
  const res = await api.get('/api/person/lieferant');
  personOptions.value = (res.data || []).map((p: any) => ({
    LABEL: `${p.PERSONENNUMMER || '?'}: ${p.FIRMA || p.NAME || ''}`,
    ID: p.ID || 0
  }));
}

async function fetchMwst() {
  const res = await api.get('/api/mwst');
  mwstOptions.value = (res.data || []).map((m: any) => ({
    LABEL: `${m.MWSTKZ || ''} (${m.PROZENT || 0}%)`,
    MWSTKZ: m.MWSTKZ || '',
    PROZENT: m.PROZENT || 0
  }));
}

function onSiloChange(val: number) {
  const selected = siloOptions.value.find(s => s.id === val);
  if (selected) {
    form.SILONUMMER = selected.silonummer ?? null;
  }
}

function onMwstChange(val: string) {
  const selected = mwstOptions.value.find(m => m.MWSTKZ === val);
  if (selected) {
    form.MWSTPROZ = selected.PROZENT;
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  resetForm();
  form.fullTimestamp = sessionStore.workingTimestamp;
  
  // Reset displayForm
  displayForm.LIEFERMENGE = '';
  displayForm.PREISDT = '';
  displayForm.NETTO = '';
  displayForm.BRUTTO = '';
  displayForm.RABATTPROZ = '';

  showDialog.value = true;
}

function onEdit(row: Futter) {
  isEditing.value = true;
  editId.value = row.ID || (row as any).id;
  
  Object.assign(form, {
    ID_SILO: extractInt(row.ID_SILO || (row as any).id_silo) || null,
    SILONUMMER: extractInt(row.SILONUMMER || (row as any).silonummer) || null,
    ID_HERDEN: extractInt(row.ID_HERDEN || (row as any).id_herden) || 0,
    ID_PERSON: extractInt(row.ID_PERSON || (row as any).id_person) || null,
    LIEFERDATUM: row.LIEFERDATUM || (row as any).lieferdatum || '',
    fullTimestamp: `${row.LIEFERDATUM || (row as any).lieferdatum || ''} ${(row.ZEITSTEMPEL || (row as any).zeitstempel)?.substring(11, 16) || '12:00'}`,
    LIEFERMENGE: extractFloat(row.LIEFERMENGE || (row as any).liefermenge) || null,
    PREISDT: extractFloat(row.PREISDT || (row as any).preisdt) || null,
    RABATTPROZ: extractFloat(row.RABATTPROZ || (row as any).rabattproz) || null,
    NETTO: extractFloat(row.NETTO || (row as any).netto) || null,
    BRUTTO: extractFloat(row.BRUTTO || (row as any).brutto) || null,
    MWSTPROZ: extractFloat(row.MWSTPROZ || (row as any).mwstproz) || null,
    MWSTKZ: extractString(row.MWSTKZ || (row as any).mwstkz) || '',
    DATUM: row.DATUM || (row as any).datum || '',
    ZEITSTEMPEL: row.ZEITSTEMPEL || (row as any).zeitstempel || '',
    ID_FUTTERSORTEN: extractInt((row as any).ID_FUTTERSORTEN || (row as any).id_futtersorten) || null
  });

  // Sync displayForm
  displayForm.LIEFERMENGE = formatNumberLocalized(form.LIEFERMENGE, 2);
  displayForm.PREISDT = formatNumberLocalized(form.PREISDT, 2);
  displayForm.NETTO = formatNumberLocalized(form.NETTO, 2);
  displayForm.BRUTTO = formatNumberLocalized(form.BRUTTO, 2);
  displayForm.RABATTPROZ = formatNumberLocalized(form.RABATTPROZ, 2);
  
  showDialog.value = true;
}

function closeDialog() {
  showDialog.value = false;
}

function resetForm() {
  Object.keys(form).forEach(key => {
    const k = key as keyof typeof form;
    (form as any)[k] = (typeof (form as any)[k] === 'string' ? '' : null);
  });
}

function onDelete(row: Futter) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diese Futterbuchung wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(() => {
    loading.value = true;
    api.delete(`/api/futter/${row.ID || (row as any).id}`)
      .then(() => {
        $q.notify({ type: 'positive', message: 'Buchung gelöscht' });
        void loadData();
      })
      .catch((err: unknown) => {
        console.error('Error deleting:', err);
        $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
      })
      .finally(() => {
        loading.value = false;
      });
  });
}

async function onSubmit() {
  try {
    const payload = {
      ...form,
      ID_SILO: Number(form.ID_SILO),
      SILONUMMER: Number(form.SILONUMMER),
      ID_HERDEN: 0,
      ID_PERSON: Number(form.ID_PERSON),
      LIEFERDATUM: form.fullTimestamp.split(' ')[0],
      LIEFERMENGE: Number(form.LIEFERMENGE),
      PREISDT: Number(form.PREISDT),
      RABATTPROZ: Number(form.RABATTPROZ),
      NETTO: Number(form.NETTO),
      BRUTTO: Number(form.BRUTTO),
      MWSTPROZ: Number(form.MWSTPROZ),
      DATUM: form.fullTimestamp.split(' ')[0],
      ZEITSTEMPEL: form.fullTimestamp + ':00Z',
      ID_FUTTERSORTEN: Number(form.ID_FUTTERSORTEN || 0)
    };
    
    if (isEditing.value && editId.value) {
      await api.put(`/api/futter/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Futterbuchung aktualisiert' });
    } else {
      await api.post('/api/futter', payload);
      $q.notify({ type: 'positive', message: 'Futterbuchung erstellt' });
    }
    showDialog.value = false;
    void loadData();
  } catch (err) {
    console.error('Error saving:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}

function openInventur() {
  inventurForm.ID_SILO = null;
  const workingDate = (sessionStore.workingTimestamp ?? '').split(' ')[0];
  const fallbackDate = new Date().toISOString().split('T')[0] || '';
  inventurForm.INVENTURDATUMNEU = workingDate || fallbackDate;
  inventurForm.INVENTURFUELLMENGE = null;
  showInventurDialog.value = true;
}

async function onSubmitInventur() {
  try {
    const silo = rawSilos.value.find(s => (s.id || s.ID) === inventurForm.ID_SILO);
    if (!silo) {
      $q.notify({ type: 'negative', message: 'Silo nicht gefunden' });
      return;
    }

    const payload = {
      INVENTURDATUMNEU: inventurForm.INVENTURDATUMNEU || '0001-01-01',
      INVENTURFUELLMENGE: Number(inventurForm.INVENTURFUELLMENGE)
    };

    await api.post(`/api/silo/${silo.id || silo.ID}/inventur`, payload);
    $q.notify({ type: 'positive', message: 'Futterinventur erfolgreich gespeichert und berechnet' });
    showInventurDialog.value = false;
    await fetchSilos();
    void loadData();
  } catch (err: any) {
    console.error('Error saving Futterinventur:', err);
    const msg = err.response?.data?.error || err.message || 'Fehler beim Speichern der Futterinventur';
    $q.notify({ type: 'negative', message: msg });
  }
}
</script>

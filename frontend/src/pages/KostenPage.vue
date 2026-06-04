<template>
  <q-page padding>
    <div class="row items-center q-mb-md">
      <div class="text-h4 text-weight-bolder text-primary">Kostenmanagement</div>
    </div>

    <!-- Header Section: Kostentabkopf -->
    <q-card flat bordered class="q-mb-lg rounded-borders shadow-2" :class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'" style="border-radius: 16px;">
      <q-card-section class="bg-primary text-white row items-center q-pa-md">
        <q-icon name="analytics" size="md" class="q-mr-md" />
        <div class="text-h6 text-weight-bold">Stammdaten & Faktoren</div>
        <q-space />
        <q-btn
          color="white"
          :text-color="isEditingHead ? 'positive' : 'primary'"
          :icon="isEditingHead ? 'save' : 'edit'"
          :label="isEditingHead ? 'Speichern' : 'Bearbeiten'"
          @click="toggleHeadEdit"
          unelevated
          rounded
        />
      </q-card-section>

      <q-card-section class="q-pa-lg">
        <div class="row q-col-gutter-md">
          <div class="col-12 col-sm-4">
            <q-input
              v-model="displaySchlachterloes"
              label="Schlachterlös am Ende (€)"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
              suffix="€"
              class="text-weight-bold"
            />
          </div>
          <div class="col-12 col-sm-4">
            <q-input
              v-model.number="headForm.PRODDAUERGEPLANT"
              label="Produktionsdauer geplant (Tage)"
              type="number"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
            />
          </div>
          <div class="col-12 col-sm-4">
            <q-input
               v-model="displayGebaeudewert"
              label="Gebäudewert (€)"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
              suffix="€"
            />
          </div>
        </div>

        <div class="row q-col-gutter-md q-mt-xs">
          <div class="col-12 col-sm-4">
            <q-input
              v-model.number="headForm.ABSCHREIBUNG_G"
              label="Abschreibung G (%)"
              type="number"
              step="0.1"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
            />
          </div>
          <div class="col-12 col-sm-4">
            <q-input
               v-model="displayGeraetewert"
              label="Gerätewert (€)"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
              suffix="€"
            />
          </div>
          <div class="col-12 col-sm-4">
            <q-input
              v-model.number="headForm.ABSCHREIBUNG_R"
              label="Abschreibung R (%)"
              type="number"
              step="0.1"
              filled
              :readonly="!isEditingHead"
              dense
              stack-label
            />
          </div>
        </div>
      </q-card-section>
    </q-card>

    <!-- ANZEIGE KOSTEN PRO TAG (Berechnet) -->
    <div class="row q-mb-lg">
      <div class="col-12">
        <q-banner padded class="bg-primary text-white shadow-5 rounded-borders q-pa-md">
          <template v-slot:avatar>
            <q-icon name="payments" color="white" size="lg" />
          </template>
          <div class="row items-center justify-between no-wrap">
            <div>
              <div class="text-subtitle1 opacity-80">Berechnete Gesamtkosten pro Tag</div>
              <div class="text-h3 text-weight-bolder">{{ formatCurrency(totalKostenProTag) }} €</div>
            </div>
            <div class="text-right q-ml-xl gt-xs">
              <div class="text-caption opacity-90 text-weight-bold">Fixkosten (Gebäude/Geräte): {{ formatCurrency(fixKostenProTag) }} €/Tag</div>
              <div class="text-caption opacity-90 text-weight-bold">Variable Kosten: {{ formatCurrency(varKostenProTag) }} €/Tag</div>
            </div>
          </div>
        </q-banner>
      </div>
    </div>

    <!-- Detail Section: Kosten-Grid -->
    <q-table
      title="Kosten-Positionen"
      :rows="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      v-model:pagination="pagination"
      class="shadow-2 rounded-borders"
      :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
      :dark="$q.dark.isActive"
      table-header-class="bg-grey-3 text-weight-bolder"
      separator="cell"
    >
      <template v-slot:top-right>
        <q-btn color="primary" icon="add" label="Neu" @click="openCreate" rounded unelevated />
      </template>

      <template v-slot:body-cell-actions="props">
        <q-td :props="props" auto-width>
          <div class="row no-wrap q-gutter-x-xs justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated />
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated />
          </div>
        </q-td>
      </template>

      <template v-slot:body-cell-kosten="props">
        <q-td :props="props">
          <span class="text-weight-bold">{{ formatCurrency(props.value) }} €</span>
        </q-td>
      </template>

      <template v-slot:body-cell-kosten_pro_tag="props">
        <q-td :props="props" class="bg-primary text-white text-weight-bolder">
          {{ formatCurrency(props.value) }} €
        </q-td>
      </template>
    </q-table>

    <!-- Dialog for Kosten-Position -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 400px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('auto.edit_kosten') : t('auto.new_kosten') }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-md">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <div class="row q-col-gutter-sm">
              <div class="col-12">
                <q-input v-model="form.BUCHUNGSDATUM" type="date" label="Buchungsdatum" filled stack-label/>
              </div>
              <div class="col-12">
                <q-select
                  v-model="form.KOSTENTYP"
                  label="Kostentyp"
                  filled
                  :options="kostentypOptions"
                  emit-value
                  map-options
                />
              </div>
              <div class="col-12">
                <q-input v-model="form.BEZEICHNUNG" label="Bezeichnung" filled/>
              </div>
              <div class="col-6">
                <q-input v-model="displayFormKosten" label="Kosten (€)" filled />
              </div>
              <div class="col-6">
                <q-input v-model.number="form.TAGE" type="number" label="Anzahl Tage" filled/>
              </div>
            </div>

            <div class="row justify-end q-mt-md">
              <q-btn ref="cancelBtn" :label="t('form.cancel')" color="grey" flat v-close-popup />
              <q-btn ref="saveBtn" :label="isEditing ? t('form.update') : t('form.save')" type="submit" color="primary" unelevated rounded />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';

const $q = useQuasar();
const loading = ref(false);

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

const rows = ref<any[]>([]);
const isEditingHead = ref(false);

const headForm = reactive({
  ID: 1,
  SCHLACHTERLOES: 0,
  PRODDAUERGEPLANT: 0,
  GEBAEUDEWERT: 0,
  ABSCHREIBUNG_G: 0,
  GERAETEWERT: 0,
  ABSCHREIBUNG_R: 0
});

const form = reactive({
  KOSTENTYP: '',
  BEZEICHNUNG: '',
  KOSTEN: 0,
  TAGE: 0,
  BUCHUNGSDATUM: new Date().toISOString().split('T')[0]
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

const kostentypOptions = [
  { label: 'Variabel (VA)', value: 'VA' },
  { label: 'Fix (FI)', value: 'FI' }
];

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  {
    name: 'buchungsdatum',
    align: 'left',
    label: 'Vermittelt am',
    field: (row: any) => extractString(row.BUCHUNGSDATUM) || '-',
    sortable: true
  },
  {
    name: 'kostentyp',
    align: 'left',
    label: 'Typ',
    field: (row: any) => extractString(row.KOSTENTYP) || '-',
    sortable: true
  },
  {
    name: 'bezeichnung',
    align: 'left',
    label: 'Bezeichnung',
    field: (row: any) => extractString(row.BEZEICHNUNG) || '-',
    sortable: true
  },
  {
    name: 'kosten',
    align: 'right',
    label: 'Betrag',
    field: (row: any) => extractFloat(row.KOSTEN) || 0,
    sortable: true
  },
  {
    name: 'tage',
    align: 'right',
    label: 'Tage',
    field: (row: any) => extractInt(row.TAGE) || 0,
    sortable: true
  },
  {
    name: 'kosten_pro_tag',
    align: 'right',
    label: 'Kosten/Tag',
    field: (row: any) => {
      const k = extractFloat(row.KOSTEN);
      const t = extractInt(row.TAGE);
      return t > 0 ? (k / t) : 0;
    },
    sortable: true
  }
];

const pagination = ref({ rowsPerPage: 20 });
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

function formatCurrency(val: number) {
  return new Intl.NumberFormat('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);
}

function parseCurrency(val: string) {
  if (!val) return 0;
  const cleanStr = val.replace(/\./g, '').replace(',', '.');
  return parseFloat(cleanStr) || 0;
}

function createMoneyComputed(target: Record<string, any>, key: string) {
  return computed({
    get: () => formatCurrency(target[key] as number),
    set: (val: string) => { target[key] = parseCurrency(val); }
  });
}

const displaySchlachterloes = createMoneyComputed(headForm, 'SCHLACHTERLOES');
const displayGebaeudewert = createMoneyComputed(headForm, 'GEBAEUDEWERT');
const displayGeraetewert = createMoneyComputed(headForm, 'GERAETEWERT');
const displayFormKosten = createMoneyComputed(form, 'KOSTEN');

const fixKostenProTag = computed(() => {
  const gDepr = (headForm.GEBAEUDEWERT * headForm.ABSCHREIBUNG_G) / 100;
  const rDepr = (headForm.GERAETEWERT * headForm.ABSCHREIBUNG_R) / 100;
  return (gDepr + rDepr) / 365;
});

const debugVarItems = computed(() => {
  return rows.value.map(row => {
    const amount = extractFloat(row.KOSTEN);
    const days = extractInt(row.TAGE);
    const label = extractString(row.BEZEICHNUNG) || 'Unbenannt';
    const type = extractString(row.KOSTENTYP).trim().toUpperCase();
    return {
      id: extractInt(row.ID),
      label: `${label} (${type})`,
      val: amount,
      days,
      daily: days > 0 ? (amount / days) : 0
    };
  });
});

const varKostenProTag = computed(() => {
  return debugVarItems.value.reduce((sum, item) => sum + item.daily, 0);
});

const totalKostenProTag = computed(() => fixKostenProTag.value + varKostenProTag.value);

async function loadHead() {
  try {
    const res = await api.get('/api/kostentabkopf');
    Object.assign(headForm, {
      ID: extractInt(res.data.ID) || 1,
      SCHLACHTERLOES: extractFloat(res.data.SCHLACHTERLOES) || 0,
      PRODDAUERGEPLANT: extractInt(res.data.PRODDAUERGEPLANT) || 0,
      GEBAEUDEWERT: extractFloat(res.data.GEBAEUDEWERT) || 0,
      ABSCHREIBUNG_G: extractFloat(res.data.ABSCHREIBUNG_G) || 0,
      GERAETEWERT: extractFloat(res.data.GERAETEWERT) || 0,
      ABSCHREIBUNG_R: extractFloat(res.data.ABSCHREIBUNG_R) || 0
    });
  } catch { /**/ }
}

async function loadKosten() {
  loading.value = true;
  try {
    const res = await api.get('/api/kosten');
    rows.value = res.data || [];
  } catch { /**/ } finally { loading.value = false; }
}

async function toggleHeadEdit() {
  if (isEditingHead.value) {
    try {
      await api.put('/api/kostentabkopf', headForm);
      $q.notify({ type: 'positive', message: 'Stammdaten gespeichert' });
      isEditingHead.value = false;
    } catch { $q.notify({ type: 'negative', message: 'Fehler beim Speichern' }); }
  } else { isEditingHead.value = true; }
}

function openCreate() { isEditing.value = false; editId.value = null; resetForm(); showDialog.value = true; }
function resetForm() {
  form.KOSTENTYP = 'VA';
  form.BEZEICHNUNG = '';
  form.KOSTEN = 0;
  form.TAGE = 0;
  form.BUCHUNGSDATUM = new Date().toISOString().split('T')[0];
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = extractInt(row.ID);
  Object.assign(form, {
    KOSTENTYP: extractString(row.KOSTENTYP),
    BEZEICHNUNG: extractString(row.BEZEICHNUNG),
    KOSTEN: extractFloat(row.KOSTEN),
    TAGE: extractInt(row.TAGE),
    BUCHUNGSDATUM: extractString(row.BUCHUNGSDATUM)
  });
  showDialog.value = true;
}

function onDelete(row: Kosten) {
  $q.dialog({ title: 'Löschen', message: 'Eintrag wirklich löschen?', cancel: true }).onOk(() => {
    loading.value = true;
    api.delete(`/api/kosten/${row.ID}`)
      .then(() => void loadKosten())
      .catch((_err: unknown) => $q.notify({type: 'negative', message: 'Fehler beim Löschen'}))
      .finally(() => {
        loading.value = false;
      });
  });
}

async function onSubmit() {
  try {
    if (isEditing.value && editId.value) { await api.put(`/api/kosten/${editId.value}`, form); }
    else { await api.post('/api/kosten', form); }
    showDialog.value = false; void loadKosten();
  } catch { $q.notify({ type: 'negative', message: 'Fehler beim Speichern' }); }
}

onMounted(() => { void loadHead(); void loadKosten(); });
</script>

<style scoped>
.rounded-borders { border-radius: 16px; }
.opacity-90 { opacity: 0.9; }
</style>

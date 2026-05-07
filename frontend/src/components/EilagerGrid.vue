<template>
  <div class="q-pa-md">
    <div class="row q-gutter-md q-mb-md items-center">
      <q-btn color="primary" icon="add" label="Eierlager anlegen" @click="openCreate" />
      <q-input v-model="searchTerm" placeholder="Suchen..." dense outlined class="bg-white" style="width: 250px">
        <template v-slot:append><q-icon name="search" /></template>
      </q-input>
      <q-select
        v-model="kzFilter"
        :options="typeOptions"
        option-value="KZ"
        option-label="BETREFF"
        emit-value
        map-options
        label="Typ Filter"
        filled
        dense
        stack-label
        clearable
        style="width: 200px"
        class="bg-white"
      />
      <q-space />
      <q-btn icon="refresh" @click="loadData" flat round />
    </div>

    <q-table
      :rows="filteredRows"
      :columns="columns"
      row-key="ID"
      :loading="loading"
      class="huhnlite-grid-standard shadow-2 rounded-borders resizable-table"
      :pagination="{ rowsPerPage: 15 }"
      separator="cell"
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
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row as Eilager)" unelevated size="sm">
              <q-tooltip>Bearbeiten</q-tooltip>
            </q-btn>
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row as Eilager)" unelevated
                   size="sm">
              <q-tooltip>Löschen</q-tooltip>
            </q-btn>
          </div>
        </q-td>
      </template>

      <template v-slot:body-cell-BEZEICHNUNG="props">
        <q-td :props="props">
          {{ extractString((props.row as Eilager).BEZEICHNUNG) }}
        </q-td>
      </template>
    </q-table>

    <!-- Edit Dialog -->
    <q-dialog v-model="showDialog" persistent>
      <q-card style="min-width: 450px; border-radius: 12px;">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-primary">{{ isEditing ? 'Eierlager bearbeiten' : 'Neues Eierlager' }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input
              v-model.number="form.LAGERNUMMER"
              label="Lagernummer *"
              type="number"
              outlined
              dense
              :rules="[val => !!val || 'Pflichtfeld']"
            />
            <q-select
              v-model="form.KZ"
              :options="typeOptions"
              option-value="KZ"
              option-label="BETREFF"
              emit-value
              map-options
              label="Lagertyp *"
              outlined
              dense
              :rules="[val => !!val || 'Pflichtfeld']"
            />
            <q-input
              v-model="form.BEZEICHNUNG"
              label="Bezeichnung"
              outlined
              dense
            />

            <div class="row q-col-gutter-sm bg-grey-1 q-pa-sm rounded-borders border-grey-4">
              <div class="col-12 text-caption text-grey-7 q-mb-xs">
                <q-icon name="info" size="xs"/>
                Bestände (Read-Only - Verwaltung über Eilagerbuchungen)
              </div>
              <div class="col-6">
                <q-input v-model.number="form.JUMBOS" label="Jumbos" type="number" outlined dense readonly
                         bg-color="grey-2"/>
              </div>
              <div class="col-6">
                <q-input v-model.number="form.XL" label="XL" type="number" outlined dense readonly bg-color="grey-2"/>
              </div>
              <div class="col-6">
                <q-input v-model.number="form.LARGE" label="L" type="number" outlined dense readonly bg-color="grey-2"/>
              </div>
              <div class="col-6">
                <q-input v-model.number="form.MEDIUM" label="M" type="number" outlined dense readonly
                         bg-color="grey-2"/>
              </div>
              <div class="col-6">
                <q-input v-model.number="form.SMALL" label="S" type="number" outlined dense readonly bg-color="grey-2"/>
              </div>
              <div class="col-6">
                <q-input v-model.number="form.VOLLEIKG" label="Vollei (kg)" type="number" outlined dense readonly
                         bg-color="grey-2"/>
              </div>
            </div>

            <div class="row justify-end q-mt-md">
              <q-btn label="Abbrechen" color="grey" flat v-close-popup />
              <q-btn :label="isEditing ? 'Speichern' : 'Anlegen'" type="submit" color="primary" class="q-ml-sm" :loading="saving" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue';
import { api } from '../boot/api';
import { useQuasar } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Eilager');

const loading = ref(false);
const saving = ref(false);
const rows = ref<Eilager[]>([]);
const searchTerm = ref('');
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const kzFilter = ref<string | null>(null);
const typeOptions = ref<TypeOption[]>([]);

interface Eilager {
  ID: number;
  LAGERNUMMER: number;
  BEZEICHNUNG?: any;
  KZ?: any;
  LETZTE_BUCHUNG?: any;
  JUMBOS?: any;
  XL?: any;
  LARGE?: any;
  MEDIUM?: any;
  SMALL?: any;
  VOLLEIKG?: any;
}

interface TypeOption {
  KZ: string;
  BETREFF: string;
}



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

const form = reactive({
  LAGERNUMMER: 0,
  BEZEICHNUNG: '',
  JUMBOS: 0,
  XL: 0,
  LARGE: 0,
  MEDIUM: 0,
  SMALL: 0,
  VOLLEIKG: 0,
  KZ: ''
});

import type {QTableProps} from 'quasar';

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  {name: 'LAGERNUMMER', align: 'left', label: 'Nr.', field: (row: any) => row.LAGERNUMMER || row.lagernummer, sortable: true},
  {name: 'KZ', align: 'left', label: 'KZ', field: (row: any) => extractString(row.KZ || row.kz), sortable: true},
  {
    name: 'BEZEICHNUNG',
    align: 'left',
    label: 'Bezeichnung',
    field: (row: any) => extractString(row.BEZEICHNUNG || row.bezeichnung),
    sortable: true
  },
  {
    name: 'LETZTE_BUCHUNG',
    align: 'left',
    label: 'Letzte Buchung',
    field: (row: any) => extractString(row.LETZTE_BUCHUNG || row.letzte_buchung),
    sortable: true
  },
  {name: 'JUMBOS', align: 'right', label: 'Jumbos', field: (row: any) => extractInt(row.JUMBOS || row.jumbos)},
  {name: 'XL', align: 'right', label: 'XL', field: (row: any) => extractInt(row.XL || row.xl)},
  {name: 'LARGE', align: 'right', label: 'L', field: (row: any) => extractInt(row.LARGE || row.large)},
  {name: 'MEDIUM', align: 'right', label: 'M', field: (row: any) => extractInt(row.MEDIUM || row.medium)},
  {name: 'SMALL', align: 'right', label: 'S', field: (row: any) => extractInt(row.SMALL || row.small)},
  {name: 'VOLLEIKG', align: 'right', label: 'Vollei (kg)', field: (row: any) => extractInt(row.VOLLEIKG || row.volleikg)},
];

const filteredRows = computed(() => {
  let res = rows.value;
  if (kzFilter.value) {
    res = res.filter(r => extractString(r.KZ) === kzFilter.value);
  }
  if (!searchTerm.value) return res;
  const s = searchTerm.value.toLowerCase();
  return res.filter(r =>
    (r.LAGERNUMMER || (r as any).lagernummer || '').toString().includes(s) ||
    extractString(r.BEZEICHNUNG || (r as any).bezeichnung).toLowerCase().includes(s)
  );
});

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/eilager');
    rows.value = (res.data as Eilager[]) || [];
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Eilager konnten nicht geladen werden' });
  } finally {
    loading.value = false;
  }
}

async function loadTypeOptions() {
  try {
    const res = await api.get('/api/texte/typ/L');
    typeOptions.value = ((res.data as any[]) || []).map((t) => ({
      KZ: extractString(t.KZ),
      BETREFF: extractString(t.BETREFF) || extractString(t.KZ)
    }));
  } catch (err: unknown) {
    console.error('Fehler beim Laden der Lagertypen', err);
  }
}

function openCreate() {
  isEditing.value = false; editId.value = null;
  Object.assign(form, {
    LAGERNUMMER: 0,
    BEZEICHNUNG: '',
    JUMBOS: 0,
    XL: 0,
    LARGE: 0,
    MEDIUM: 0,
    SMALL: 0,
    VOLLEIKG: 0,
    KZ: ''
  });
  showDialog.value = true;
}

function onEdit(row: Eilager) {
  isEditing.value = true;
  editId.value = row.ID || (row as any).id;
  Object.assign(form, {
    LAGERNUMMER: row.LAGERNUMMER || (row as any).lagernummer,
    BEZEICHNUNG: extractString(row.BEZEICHNUNG || (row as any).bezeichnung),
    JUMBOS: extractInt(row.JUMBOS || (row as any).jumbos) || 0,
    XL: extractInt(row.XL || (row as any).xl) || 0,
    LARGE: extractInt(row.LARGE || (row as any).large) || 0,
    MEDIUM: extractInt(row.MEDIUM || (row as any).medium) || 0,
    SMALL: extractInt(row.SMALL || (row as any).small) || 0,
    VOLLEIKG: extractInt(row.VOLLEIKG || (row as any).volleikg) || 0,
    KZ: extractString(row.KZ || (row as any).kz)
  });
  showDialog.value = true;
}

function onDelete(row: Eilager) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: `Soll das Eilager ${row.LAGERNUMMER} gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/eilager/${row.ID || (row as any).id}`);
      $q.notify({ type: 'positive', message: 'Eilager gelöscht' });
      void loadData();
    } catch (error: any) {
      const msg = error.response?.data?.error || 'Fehler beim Löschen';
      $q.notify({type: 'negative', message: msg});
    }
  });
}

async function onSubmit() {
  saving.value = true;
  try {
    const payload = {...form};
    if (!payload.KZ) payload.KZ = 'x';
    if (isEditing.value) {
      await api.put(`/api/eilager/${editId.value}`, payload);
    } else {
      await api.post('/api/eilager', payload);
    }
    showDialog.value = false;
    $q.notify({ type: 'positive', message: 'Gespeichert' });
    void loadData();
  } catch (_err) {
    $q.notify({type: 'negative', message: 'Fehler beim Speichern'});
  }
  finally { saving.value = false; }
}

onMounted(() => {
  initWidths(columns);
  void loadData();
  void loadTypeOptions();
});
</script>

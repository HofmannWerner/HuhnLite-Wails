<template>
  <div class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="text-h6 text-primary">Silo-Verwaltung</div>
      <q-btn color="primary" icon="add" :label="t('grid.newSilo')" @click="openCreate" rounded unelevated />
    </div>

    <!-- Table -->
    <q-table separator="cell"
      :rows="rows"
      :columns="columns"
      row-key="ID"
      :loading="loading"
      v-model:pagination="pagination"
      class="q-mb-lg cursor-pointer shadow-2 rounded-borders overflow-hidden resizable-table"
      @row-dblclick="(evt, row) => onEdit(row)"
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
        <q-td :props="props" auto-width>
          <div class="row no-wrap q-gutter-x-xs justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm">
              <q-tooltip>Bearbeiten</q-tooltip>
            </q-btn>
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)"  unelevated size="sm">
              <q-tooltip>Löschen</q-tooltip>
            </q-btn>
          </div>
        </q-td>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 400px; max-width: 600px;">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6">{{ isEditing ? t('grid.editSilo') : t('grid.newSilo') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated />
        </q-card-section>
        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input v-model.number="form.SILONUMMER" type="number" :label="t('grid.siloNumberRequired')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => val !== null && val !== '' || 'Silonummer ist ein Pflichtfeld']" />
            <q-input v-model="form.BEZEICHNUNG" :label="t('grid.designationRequired')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => !!val || 'Bezeichnung darf nicht leer sein']" />
            <q-input v-model="form.INVENTURDATUMALT" type="date" :label="t('auto.inventur_datum_alt')" stack-label filled :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model="form.INVENTURDATUMNEU" type="date" :label="t('auto.inventur_datum_neu')" stack-label filled :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model.number="form.MAXFUELLMENGE" type="number" :label="t('auto.maximale_fuellmenge_kg')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model.number="form.MINFUELLMENGE" type="number" :label="t('auto.minimale_fuellmenge_kg')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model.number="form.INVENTURFUELLMENGE" type="number" :label="t('auto.inventur_fuellmenge_kg')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model.number="form.PERSONENNUMMER" type="number" :label="t('auto.personennummer')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <q-input v-model.number="form.ID_LIEFERANT" type="number" :label="t('auto.id_lieferant')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" />
            <div class="q-mt-md">
              <q-btn ref="saveBtn" :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated />
              <q-btn ref="cancelBtn" :label="t('form.cancel')" color="negative" class="q-ml-sm" @click="closeDialog" rounded unelevated />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
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

const formatNumber = (val: any) => {
  const n = Number(val);
  if (isNaN(n)) return '0';
  return n.toLocaleString('de-DE');
};

import { ref, reactive, onMounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import type { QTableProps } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Silo');

const loading = ref(false);

interface Silo {
  ID: number;
  SILONUMMER: { Int64: number; Valid: boolean };
  BEZEICHNUNG: { String: string; Valid: boolean };
  INVENTURDATUMALT: { String: string; Valid: boolean };
  INVENTURDATUMNEU: { String: string; Valid: boolean };
  MAXFUELLMENGE: { Int64: number; Valid: boolean };
  MINFUELLMENGE: { Int64: number; Valid: boolean };
  INVENTURFUELLMENGE: { Int64: number; Valid: boolean };
  PERSONENNUMMER: { Int64: number; Valid: boolean };
  ID_LIEFERANT: { Int64: number; Valid: boolean };
}

const rows = ref<Silo[]>([]);
const pagination = ref({ rowsPerPage: 50 });
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = reactive({
  SILONUMMER: 0,
  BEZEICHNUNG: '',
  INVENTURDATUMALT: '0001-01-01',
  INVENTURDATUMNEU: '0001-01-01',
  MAXFUELLMENGE: 0,
  MINFUELLMENGE: 0,
  INVENTURFUELLMENGE: 0,
  PERSONENNUMMER: 0,
  ID_LIEFERANT: 0
});

const originalFormState = ref('');
const isDirty = computed(() => JSON.stringify(form) !== originalFormState.value);
const cancelBtn = ref<{ $el: HTMLElement } | null>(null);
const saveBtn = ref<{ $el: HTMLElement } | null>(null);

function onDialogShow() {
  originalFormState.value = JSON.stringify(form);
  setTimeout(() => {
    (cancelBtn.value)?.$el?.focus();
  }, 50);
}

watch(isDirty, (dirty: boolean) => {
  if (dirty && (document.activeElement === (cancelBtn.value)?.$el || document.activeElement === document.body)) {
    (saveBtn.value)?.$el?.focus();
  }
});

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  { name: 'SILONUMMER', align: 'right', label: 'Silonr', field: (row: any) => extractInt(row.SILONUMMER || row.silonummer), sortable: true },
  { name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: (row: any) => extractString(row.BEZEICHNUNG || row.bezeichnung), sortable: true },
  { name: 'INVENTURDATUMALT', align: 'left', label: 'Inventur Alt', field: (row: any) => extractString(row.INVENTURDATUMALT || row.inventurdatumalt), sortable: true },
  { name: 'INVENTURDATUMNEU', align: 'left', label: 'Inventur Neu', field: (row: any) => extractString(row.INVENTURDATUMNEU || row.inventurdatumneu), sortable: true },
  { name: 'MAXFUELLMENGE', align: 'right', label: 'Max. Füllmenge', field: (row: any) => extractInt(row.MAXFUELLMENGE || row.maxfuellmenge), format: val => formatNumber(val), sortable: true },
  { name: 'MINFUELLMENGE', align: 'right', label: 'Min. Füllmenge', field: (row: any) => extractInt(row.MINFUELLMENGE || row.minfuellmenge), format: val => formatNumber(val), sortable: true },
  { name: 'INVENTURFUELLMENGE', align: 'right', label: 'Inventur Füllmenge', field: (row: any) => extractInt(row.INVENTURFUELLMENGE || row.inventurfuellmenge), format: val => formatNumber(val), sortable: true },
  { name: 'PERSONENNUMMER', align: 'right', label: 'Personennr', field: (row: any) => extractInt(row.PERSONENNUMMER || row.personennummer), sortable: true },
  { name: 'ID_LIEFERANT', align: 'right', label: 'ID Lieferant', field: (row: any) => extractInt(row.ID_LIEFERANT || row.id_lieferant), sortable: true }
];

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/silo');
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Silo)' });
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  resetForm();
  showDialog.value = true;
}

function resetForm() {
  form.SILONUMMER = 0;
  form.BEZEICHNUNG = '';
  form.INVENTURDATUMALT = '0001-01-01';
  form.INVENTURDATUMNEU = '0001-01-01';
  form.MAXFUELLMENGE = 0;
  form.MINFUELLMENGE = 0;
  form.INVENTURFUELLMENGE = 0;
  form.PERSONENNUMMER = 0;
  form.ID_LIEFERANT = 0;
}

function onEdit(row: Silo) {
  isEditing.value = true;
  editId.value = row.ID || (row as any).id;
  form.SILONUMMER = extractInt(row.SILONUMMER || (row as any).silonummer) || null;
  form.BEZEICHNUNG = extractString(row.BEZEICHNUNG || (row as any).bezeichnung) || '';
  form.INVENTURDATUMALT = extractString(row.INVENTURDATUMALT || (row as any).inventurdatumalt) || '';
  form.INVENTURDATUMNEU = extractString(row.INVENTURDATUMNEU || (row as any).inventurdatumneu) || '';
  form.MAXFUELLMENGE = extractInt(row.MAXFUELLMENGE || (row as any).maxfuellmenge) || null;
  form.MINFUELLMENGE = extractInt(row.MINFUELLMENGE || (row as any).minfuellmenge) || null;
  form.INVENTURFUELLMENGE = extractInt(row.INVENTURFUELLMENGE || (row as any).inventurfuellmenge) || null;
  form.PERSONENNUMMER = extractInt(row.PERSONENNUMMER || (row as any).personennummer) || null;
  form.ID_LIEFERANT = extractInt(row.ID_LIEFERANT || (row as any).id_lieferant) || null;
  showDialog.value = true;
}

function closeDialog() {
  showDialog.value = false;
  setTimeout(() => {
    isEditing.value = false;
    editId.value = null;
    resetForm();
  }, 300);
}

function onDelete(row: Silo) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diesen Eintrag wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(() => {
    loading.value = true;
    api.delete(`/api/silo/${row.ID || (row as any).id}`)
      .then(() => {
        $q.notify({ type: 'positive', message: 'Eintrag erfolgreich gelöscht' });
        void loadData();
      })
      .catch((error: any) => {
        const msg = error.response?.data?.error || 'Fehler beim Löschen';
        $q.notify({ type: 'negative', message: msg });
      })
      .finally(() => {
        loading.value = false;
      });
  });
}

async function onSubmit() {
  try {
    const payload = {
      SILONUMMER: Number(form.SILONUMMER),
      BEZEICHNUNG: form.BEZEICHNUNG,
      INVENTURDATUMALT: form.INVENTURDATUMALT || '0001-01-01',
      INVENTURDATUMNEU: form.INVENTURDATUMNEU || '0001-01-01',
      MAXFUELLMENGE: Number(form.MAXFUELLMENGE),
      MINFUELLMENGE: Number(form.MINFUELLMENGE),
      INVENTURFUELLMENGE: Number(form.INVENTURFUELLMENGE),
      PERSONENNUMMER: Number(form.PERSONENNUMMER),
      ID_LIEFERANT: Number(form.ID_LIEFERANT)
    };
    
    if (isEditing.value && editId.value) {
      await api.put(`/api/silo/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Silo erfolgreich aktualisiert' });
    } else {
      await api.post('/api/silo', payload);
      $q.notify({ type: 'positive', message: 'Silo erfolgreich hinzugefügt' });
    }
    
    closeDialog();
    void loadData();
  } catch (error: any) {
    console.error('Silo Save Error:', error);
    const msg = error.response?.data?.error || 'Fehler beim Speichern';
    $q.notify({ type: 'negative', message: msg });
  }
}

onMounted(() => {
  initWidths(columns);
  void loadData();
});
</script>

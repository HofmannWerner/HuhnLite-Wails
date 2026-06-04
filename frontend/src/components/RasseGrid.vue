<template>
  <div class="q-pa-md">
    <div style="max-width: 600px;">
      <q-table
        separator="cell"
        :rows="rows"
        :columns="columns"
        row-key="ID"
        :loading="loading"
        v-model:pagination="pagination"
        class="q-mb-lg cursor-pointer rounded-borders shadow-2 resizable-table"
        :class="$q.dark.isActive ? 'bg-dark-page text-white' : 'bg-grey-2 text-black'"
        :dark="$q.dark.isActive"
        wrap-cells
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

        <template v-slot:top-right>
          <q-btn color="primary" icon="add" :label="t('grid.newBreed')" @click="openCreate" rounded unelevated />
        </template>
        <template v-slot:body-cell-actions="props">
          <q-td :props="props" auto-width>
            <div class="row no-wrap q-gutter-x-xs justify-center">
              <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm" />
              <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm" />
            </div>
          </q-td>
        </template>
      </q-table>
    </div>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent>
      <q-card style="min-width: 400px; max-width: 600px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('grid.editBreed') : t('grid.newBreed') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
        </q-card-section>
        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input
              v-model="newRasse"
              :label="t('grid.breedDesignation')"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => !!val || 'Feld darf nicht leer sein']"
            />
            <div class="row justify-end q-mt-lg q-gutter-x-sm">
              <q-btn :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog" padding="xs lg" />
              <q-btn :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated padding="xs xl" />
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

import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import type { QTableProps } from 'quasar';
import { api } from '../boot/api';
import { useResizableColumns } from '../composables/useResizableColumns';

/* eslint-disable @typescript-eslint/no-explicit-any */
const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Rasse');

const loading = ref(false);
const rows = ref<Record<string, any>[]>([]);
const pagination = ref({ rowsPerPage: 50 });
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const newRasse = ref('');

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' },
  { name: 'rasse', align: 'left', label: 'Rasse', field: (row: any) => extractString(row.RASSE) || '-', sortable: true }
];

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/rasse');
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Rassen)' });
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  newRasse.value = '';
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = row.ID;
  newRasse.value = row.RASSE || '';
  showDialog.value = true;
}

function closeDialog() { showDialog.value = false; }

function onDelete(row: any) {
  $q.dialog({ title: 'Löschen', message: 'Rasse wirklich löschen?', cancel: true }).onOk(async () => {
    try {
      await api.delete(`/api/rasse/${row.ID}`);
      void loadData();
    } catch (error: any) {
      const msg = error.response?.data?.error || 'Fehler beim Löschen';
      $q.notify({ type: 'negative', message: msg });
    }
  });
}

async function onSubmit() {
  try {
    const payload = { RASSE: newRasse.value };
    if (isEditing.value && editId.value) {
      await api.put(`/api/rasse/${editId.value}`, payload);
    } else {
      await api.post('/api/rasse', payload);
    }
    closeDialog();
    void loadData();
  } catch { $q.notify({ type: 'negative', message: 'Fehler beim Speichern' }); }
}

onMounted(() => {
  initWidths(columns);
  loadData();
});
</script>

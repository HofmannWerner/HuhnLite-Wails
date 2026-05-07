<template>
  <div class="q-pa-md">
    <div style="max-width: 800px;">
      <div class="row items-center justify-between q-mb-md">
        <div class="text-h6 text-primary">Stall-Verwaltung</div>
        <q-btn color="primary" icon="add" label="Neuer Stall" @click="openCreate" rounded unelevated />
      </div>

      <!-- Table -->
      <q-table separator="cell"
        :rows="rows"
        :columns="columns"
        row-key="ID"
        :loading="loading"
        v-model:pagination="pagination"
        class="q-mb-lg cursor-pointer shadow-2 rounded-borders overflow-hidden resizable-table"
        wrap-cells
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
    </div>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 400px; max-width: 600px;">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6">{{ isEditing ? 'Stall bearbeiten' : 'Neuer Stall hinzufügen' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated />
        </q-card-section>
        <q-card-section>
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input v-model.number="form.STALLNUMMER" type="number" label="Stallnummer *" filled stack-label :rules="[val => !!val || 'Feld darf nicht leer sein']" />
            <q-input v-model="form.BEZEICHNUNG" label="Stallbezeichnung *" filled stack-label :rules="[val => !!val || 'Feld darf nicht leer sein']"
            />
            <div class="q-mt-md">
              <q-btn ref="saveBtn" :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated />
              <q-btn ref="cancelBtn" label="Abbrechen" color="negative" class="q-ml-sm" @click="closeDialog" rounded unelevated />
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

import { ref, onMounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import type { QTableProps } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Stall');

const loading = ref(false);

interface Stall {
  ID: number;
  STALLNUMMER: { Int64: number; Valid: boolean };
  BEZEICHNUNG: { String: string; Valid: boolean };
}

const rows = ref<Stall[]>([]);
const pagination = ref({ rowsPerPage: 50 });

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = ref({
  STALLNUMMER: 0,
  BEZEICHNUNG: ''
});

const originalFormState = ref('');
const isDirty = computed(() => JSON.stringify(form.value) !== originalFormState.value);
const cancelBtn = ref<{ $el: HTMLElement } | null>(null);
const saveBtn = ref<{ $el: HTMLElement } | null>(null);

function onDialogShow() {
  originalFormState.value = JSON.stringify(form.value);
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
  { name: 'STALLNUMMER', align: 'left', label: 'Stallnummer', field: (row: any) => extractInt(row.STALLNUMMER || row.stallnummer), sortable: true },
  { name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: (row: any) => extractString(row.BEZEICHNUNG || row.bezeichnung), sortable: true }
];

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/stall');
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Ställe)' });
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  form.value = {
    STALLNUMMER: 0,
    BEZEICHNUNG: ''
  };
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = row.ID || row.id;
  form.value = {
    STALLNUMMER: extractInt(row.STALLNUMMER || row.stallnummer),
    BEZEICHNUNG: extractString(row.BEZEICHNUNG || row.bezeichnung)
  };
  showDialog.value = true;
}

function closeDialog() {
  showDialog.value = false;
  setTimeout(() => {
    isEditing.value = false;
    editId.value = null;
    form.value = {
      STALLNUMMER: 0,
      BEZEICHNUNG: ''
    };
  }, 300);
}

function onDelete(row: any) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diesen Eintrag wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(() => {
    loading.value = true;
    api.delete(`/api/stall/${row.ID || row.id}`)
      .then(() => {
        $q.notify({ type: 'positive', message: 'Eintrag erfolgreich gelöscht' });
        void loadData();
      })
      .catch((error: any) => {
        console.error('Fehler beim Löschen des Stalls:', error);
        const msg = error.response?.data?.error || 'Fehler beim Löschen des Stalls';
        $q.notify({
          color: 'negative',
          message: msg,
          icon: 'error'
        });
      })
      .finally(() => {
        loading.value = false;
      });
  });
}

async function onSubmit() {
  try {
    const payload = {
      STALLNUMMER: Number(form.value.STALLNUMMER),
      BEZEICHNUNG: form.value.BEZEICHNUNG
    };
    if (isEditing.value && editId.value) {
      await api.put(`/api/stall/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Stall erfolgreich aktualisiert' });
    } else {
      await api.post('/api/stall', payload);
      $q.notify({ type: 'positive', message: 'Stall erfolgreich hinzugefügt' });
    }
    closeDialog();
    void loadData();
  } catch (error: any) {
    console.error('Stall Save Error:', error);
    const msg = error.response?.data?.error || 'Fehler beim Speichern';
    $q.notify({ type: 'negative', message: msg });
  }
}

onMounted(() => {
  initWidths(columns);
  void loadData();
});
</script>

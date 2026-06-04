<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">MwSt / Steuersätze</div>
    </div>

    <!-- Table -->
    <q-table
      separator="cell"
      :rows="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      v-model:pagination="pagination"
      class="shadow-2 rounded-borders overflow-hidden"
      :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
      wrap-cells
      table-style="table-layout: auto;"
      @row-dblclick="(evt, row) => onEdit(row)"
    >
      <template v-slot:top-right>
        <q-btn
          color="primary"
          icon="add"
          label="Neuen Steuersatz"
          @click="openCreate"
          rounded
          unelevated
          class="q-px-md"
        />
      </template>

      <template v-slot:body-cell-mwstkz="props">
        <q-td :props="props" class="text-weight-bold text-center">
          <q-badge color="blue-2" text-color="blue-10" class="q-px-sm q-py-xs text-subtitle2">
            {{ props.value }}
          </q-badge>
        </q-td>
      </template>

      <template v-slot:body-cell-prozent="props">
        <q-td :props="props" class="text-right">
          <div class="text-weight-medium">{{ props.value }} %</div>
        </q-td>
      </template>

      <template v-slot:body-cell-actions="props">
        <q-td :props="props" auto-width>
          <div class="row no-wrap q-gutter-x-xs justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated />
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated/>
          </div>
        </q-td>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 400px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('auto.edit_mwst') : t('auto.new_mwst') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat/>
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-input
              v-model="form.mwstkz"
              label="MwSt-Kennzeichen (z.B. A) *"
              maxlength="1"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => !!val || 'Kennzeichen ist ein Pflichtfeld']"
            />

            <q-input
              v-model.number="form.prozent"
              type="number"
              step="0.01"
              label="Prozentsatz (%) *"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => val !== null || 'Darf nicht leer sein']"
            />

            <q-input
              v-model="form.konto"
              label="Konto"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
            />

            <div class="row justify-end q-mt-lg q-gutter-x-sm">
              <q-btn ref="cancelBtn" :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog"
                     padding="xs lg"/>
              <q-btn ref="saveBtn" :label="isEditing ? t('form.update') : t('form.save')" type="submit" color="primary"
                     rounded unelevated padding="xs xl"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import {api} from 'src/boot/api';
import type { QTableProps } from 'quasar';

/* eslint-disable @typescript-eslint/no-explicit-any */

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const extractFloat = (val: any) => {
  if (val === null || val === undefined) return 0.0;
  if (typeof val === 'object' && 'Float64' in val) return Number(val.Float64) || 0.0;
  return Number(val) || 0.0;
};

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  return Number(val) || 0;
};

const $q = useQuasar();
const loading = ref(false);

const rows = ref<any[]>([]);
const pagination = ref({ rowsPerPage: 50 });

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = reactive({
  mwstkz: '',
  prozent: null as number | null,
  konto: ''
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
  {name: 'id', align: 'left', label: 'ID', field: (row: any) => extractInt(row.id), sortable: true},
  {
    name: 'mwstkz',
    align: 'center',
    label: 'Kennzeichen',
    field: (row: any) => extractString(row.mwstkz),
    sortable: true
  },
  {
    name: 'prozent',
    align: 'right',
    label: 'Prozent (%)',
    field: (row: any) => extractFloat(row.prozent),
    sortable: true
  },
  {name: 'konto', align: 'left', label: 'Konto', field: (row: any) => extractString(row.konto), sortable: true},
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions' }
];

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/mwst');
    rows.value = res.data || [];
  } catch (err) {
    console.error('Ladefehler MwSt:', err);
    $q.notify({type: 'negative', message: 'Fehler beim Laden der MwSt-Sätze'});
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  form.mwstkz = '';
  form.prozent = null;
  form.konto = '';
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = extractInt(row.id);
  form.mwstkz = extractString(row.mwstkz);
  form.prozent = extractFloat(row.prozent);
  form.konto = extractString(row.konto);
  showDialog.value = true;
}

function closeDialog() {
  showDialog.value = false;
}

function onDelete(row: any) {
  const rowId = extractInt(row.id);
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diesen Eintrag wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(async () => {
    loading.value = true;
    try {
      await api.delete(`/api/mwst/${rowId}`);
      $q.notify({type: 'positive', message: 'Eintrag erfolgreich gelöscht'});
      await loadData();
    } catch (err) {
      $q.notify({type: 'negative', message: 'Fehler beim Löschen'});
    } finally {
      loading.value = false;
    }
  });
}

async function onSubmit() {
  try {
    const payload = {
      mwstkz: form.mwstkz,
      prozent: Number(form.prozent),
      konto: form.konto
    };

    if (isEditing.value && editId.value) {
      await api.put(`/api/mwst/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'MwSt erfolgreich aktualisiert' });
    } else {
      await api.post('/api/mwst', payload);
      $q.notify({ type: 'positive', message: 'MwSt erfolgreich hinzugefügt' });
    }

    closeDialog();
    await loadData();
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}

onMounted(() => {
  void loadData();
});
</script>

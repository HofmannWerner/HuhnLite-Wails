<template>
  <div class="q-pa-none">
    <!-- Filter section -->
    <q-card flat bordered class="q-mb-lg shadow-1" style="border-radius: 12px;">
      <q-card-section class="row items-center q-gutter-md">
        <div class="row items-center q-gutter-sm">
          <q-icon name="filter_list" size="sm" color="primary" />
          <div class="text-subtitle1 text-weight-bold">Filter:</div>
        </div>

        <q-select
          v-model="filterEilager"
          :options="eilagerOptions"
          option-value="ID"
          option-label="label"
          label="Nach Eierlager filtern"
          filled
          stack-label
          dense
          clearable
          emit-value
          map-options
          style="min-width: 300px"
          :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
        >
          <template v-slot:prepend>
            <q-icon name="inventory_2" />
          </template>
        </q-select>

        <q-input
          v-model="searchTerm"
          placeholder="Lagerplatz suchen..."
          filled
          dense
          rounded
          stack-label
          style="width: 250px"
          :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
        >
          <template v-slot:append>
            <q-icon name="search" />
          </template>
        </q-input>

        <q-space />

        <q-btn
          color="primary"
          icon="add_circle"
          label="Neuer Lagerplatz"
          @click="openCreate"
          rounded
          unelevated
          class="shadow-2"
        />
      </q-card-section>
    </q-card>

    <!-- Data Grid -->
    <q-table
      :rows="filteredRows"
      :columns="columns"
      row-key="ID"
      :loading="loading"
      :pagination="{ rowsPerPage: 15 }"
      separator="cell"
      class="huhnlite-grid-standard shadow-2 rounded-borders overflow-hidden"
      :dark="$q.dark.isActive"
      :card-class="$q.dark.isActive ? 'bg-dark' : 'bg-white'"
      @row-dblclick="(evt, row) => onEdit(row)"
    >
      <template v-slot:body-cell-actions="props">
        <q-td :props="props" auto-width>
          <div class="row no-wrap q-gutter-x-sm justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm">
              <q-tooltip>Bearbeiten</q-tooltip>
            </q-btn>
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm">
              <q-tooltip>Löschen</q-tooltip>
            </q-btn>
          </div>
        </q-td>
      </template>

      <template v-slot:body-cell-eilager_bezeichnung="props">
        <q-td :props="props">
          <div class="row items-center no-wrap">
            <q-icon name="inventory_2" size="xs" color="grey-7" class="q-mr-xs" />
            <div class="text-weight-medium">{{ props.value }}</div>
          </div>
        </q-td>
      </template>

      <template v-slot:no-data>
        <div class="full-width row flex-center text-grey-7 q-gutter-sm q-pa-xl">
          <q-icon size="2em" name="sentiment_dissatisfied" />
          <span>Keine Lagerplätze gefunden.</span>
        </div>
      </template>
    </q-table>

    <!-- Dialog -->
    <q-dialog v-model="showDialog" persistent transition-show="scale" transition-hide="scale">
      <q-card style="min-width: 450px; border-radius: 16px;" :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-white'">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('auto.edit_lagerplatz') : t('auto.new_lagerplatz') }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup @click="closeDialog" />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <q-select
              v-model="form.ID_EILAGER"
              :options="eilagerOptions"
              option-value="ID"
              option-label="label"
              label="Eierlager auswählen *"
              filled
              stack-label
              emit-value
              map-options
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => !!val || 'Bitte wählen Sie ein Eierlager aus']"
            >
              <template v-slot:prepend>
                <q-icon name="inventory_2" />
              </template>
            </q-select>

            <q-input
              v-model="form.BEZEICHNUNG"
              label="Bezeichnung *"
              filled
              stack-label
              maxlength="40"
              counter
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => !!val || 'Pflichtfeld']"
            />

            <q-input
              v-model="form.BEMERKUNG"
              type="textarea"
              label="Bemerkung"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              rows="3"
            />

            <div class="row justify-end q-mt-xl q-gutter-sm">
              <q-btn :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog" padding="xs lg" />
              <q-btn :label="isEditing ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl" />
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

import { ref, reactive, onMounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {api} from 'src/boot/api';
import type { QTableProps } from 'quasar';

const { t } = useI18n();

interface Lagerplatz {
  ID: number;
  ID_EILAGER: number;
  BEZEICHNUNG: string;
  BEMERKUNG: string;
  EILAGER_BEZEICHNUNG: string;
}

interface EilagerOption {
  ID: number;
  label: string;
}

const $q = useQuasar();
const loading = ref(false);
const rows = ref<Lagerplatz[]>([]);
const eilagerOptions = ref<EilagerOption[]>([]);

const filterEilager = ref<number | null>(null);
const searchTerm = ref('');
const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = reactive({
  ID_EILAGER: null as number | null,
  BEZEICHNUNG: '',
  BEMERKUNG: ''
});

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions', style: 'width: 100px' },
  {name: 'EILAGER_BEZEICHNUNG', align: 'left', label: 'Eierlager', field: 'EILAGER_BEZEICHNUNG', sortable: true},
  {name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: 'BEZEICHNUNG', sortable: true},
  {name: 'BEMERKUNG', align: 'left', label: 'Bemerkung', field: 'BEMERKUNG', sortable: true}
];

const filteredRows = computed(() => {
  let result = rows.value;
  if (filterEilager.value) result = result.filter(r => r.ID_EILAGER === filterEilager.value);
  if (searchTerm.value) {
    const s = searchTerm.value.toLowerCase();
    result = result.filter(r =>
      (r.BEZEICHNUNG && r.BEZEICHNUNG.toLowerCase().includes(s)) ||
      (r.BEMERKUNG && r.BEMERKUNG.toLowerCase().includes(s)) ||
      (r.EILAGER_BEZEICHNUNG && r.EILAGER_BEZEICHNUNG.toLowerCase().includes(s))
    );
  }
  return result;
});

async function loadEilager() {
  try {
    const res = await api.get('/api/eilager');
    eilagerOptions.value = (res.data || []).map((e: {
      ID: number;
      LAGERNUMMER: string;
      BEZEICHNUNG: string | { String: string }
    }) => ({
      ID: e.ID,
      label: `${e.LAGERNUMMER} - ${typeof e.BEZEICHNUNG === 'object' ? e.BEZEICHNUNG : e.BEZEICHNUNG}`
    }));
  } catch (_err) {
    console.error(_err);
  }
}

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/lagerplatz');
    rows.value = (res.data || []).map((r: {
      ID: number;
      ID_EILAGER: number;
      BEZEICHNUNG: any;
      BEMERKUNG: any;
      EILAGER_BEZEICHNUNG: any
    }) => ({
      ID: r.ID,
      ID_EILAGER: r.ID_EILAGER,
      BEZEICHNUNG: r.BEZEICHNUNG ?? '',
      BEMERKUNG: r.BEMERKUNG ?? '',
      EILAGER_BEZEICHNUNG: r.EILAGER_BEZEICHNUNG ?? ''
    }));
  } catch (_err) {
    $q.notify({type: 'negative', message: 'Fehler beim Laden'});
  }
  finally { loading.value = false; }
}

function openCreate() {
  isEditing.value = false; editId.value = null;
  form.ID_EILAGER = filterEilager.value || null;
  form.BEZEICHNUNG = '';
  form.BEMERKUNG = '';
  showDialog.value = true;
}

function onEdit(row: Lagerplatz) {
  isEditing.value = true;
  editId.value = row.ID;
  form.ID_EILAGER = row.ID_EILAGER;
  form.BEZEICHNUNG = row.BEZEICHNUNG;
  form.BEMERKUNG = row.BEMERKUNG;
  showDialog.value = true;
}

function closeDialog() { showDialog.value = false; }

function onDelete(row: Lagerplatz) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: `Möchten Sie den Lagerplatz "${row.BEZEICHNUNG}" wirklich löschen?`,
    persistent: true,
    ok: { label: 'Löschen', color: 'negative', rounded: true, unelevated: true },
    cancel: { label: 'Abbrechen', outline: true, rounded: true, color: 'primary' }
  }).onOk(async () => {
    try {
      await api.delete(`/api/lagerplatz/${row.ID}`);
      $q.notify({ type: 'positive', message: 'Lagerplatz wurde gelöscht', icon: 'check_circle' });
      void loadData();
    } catch (_err) {
      $q.notify({type: 'negative', message: 'Fehler beim Löschen'});
    }
  });
}

async function onSubmit() {
  try {
    const payload = {
      ID_EILAGER: form.ID_EILAGER,
      BEZEICHNUNG: form.BEZEICHNUNG,
      BEMERKUNG: form.BEMERKUNG
    };
    if (isEditing.value && editId.value) {
      await api.put(`/api/lagerplatz/${editId.value}`, payload);
    } else {
      await api.post('/api/lagerplatz', payload);
    }
    $q.notify({ type: 'positive', message: 'Erfolgreich gespeichert', icon: 'check_circle' });
    showDialog.value = false;
    void loadData();
  } catch (err: unknown) {
    const errorMsg = (err as {
      response?: { data?: { error?: string } };
      message?: string
    })?.response?.data?.error || (err as { message?: string })?.message || 'Fehler beim Speichern';
    $q.notify({type: 'negative', message: 'Fehler beim Speichern', caption: errorMsg});
  }
}

onMounted(() => {
  void loadData();
  void loadEilager();
});
</script>

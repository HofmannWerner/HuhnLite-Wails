<template>
  <div class="q-pa-md">
    <div class="row items-center justify-between q-mb-md">
      <div class="row q-gutter-md">
        <q-btn color="negative" icon="add" :label="t('auto.neue_verluste')" @click="openCreate" :disable="!filterHerde" rounded unelevated />
      </div>
      <div class="text-h6 text-negative">{{ t('auto.verluste') }}</div>
    </div>

    <div class="row q-col-gutter-md q-mb-md items-center">
      <div class="col-12 col-sm-4 col-md-2">
        <q-input
          v-model.number="filterDays"
          type="number"
          :label="t('auto.zeitraum_tage')"
          filled
          stack-label
          min="1"
          :dark="$q.dark.isActive"
          :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
        >
          <template v-slot:prepend>
            <q-icon name="history" />
          </template>
        </q-input>
      </div>
      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterHerde"
          :options="herdeOptions"
          option-value="id"
          option-label="bezeichnung"
          emit-value
          map-options
          clearable
          :label="t('auto.herde_auswaehlen')"
          filled
          stack-label
          :dark="$q.dark.isActive"
          :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
        >
          <template v-slot:prepend>
            <q-icon name="pets" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterText"
          :options="textOptions"
          option-value="id"
          option-label="betreff"
          emit-value
          map-options
          clearable
          :label="t('auto.grund_verlust')"
          filled
          stack-label
          :dark="$q.dark.isActive"
          :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
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
      @update:pagination="(val: any) => { pagination = val }"
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

       <template v-slot:body-cell-typ="props">
        <q-td :props="props">
          <q-chip color="negative" text-color="white" dense>
            {{ t('auto.verlust') }}
          </q-chip>
        </q-td>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <TierbewegungDialog 
      v-model="showDialog" 
      :is-editing="isEditing" 
      :edit-id="editId" 
      :initial-herde-id="filterHerde" 
      initial-typ="V"
      :fixed-typ="true"
      @saved="loadData" 
    />
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

import { ref, onMounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';
import TierbewegungDialog from '../components/TierbewegungDialog.vue';

/* eslint-disable @typescript-eslint/no-explicit-any */

import { useSessionStore } from '../stores/session';
import { useResizableColumns } from '../composables/useResizableColumns';

const $q = useQuasar();
const sessionStore = useSessionStore();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Verluste');

const loading = ref(false);
const rows = ref<Record<string, any>[]>([]);
const herdeOptions = ref<any[]>([]);
const filterHerde = ref<number | null>(null);
const textOptions = ref<any[]>([]);
const filterText = ref<number | null>(null);
const filterDays = ref(7);

const filteredRows = computed(() => {
  // Hard filter for type 'V'
  let list = rows.value.filter(row => extractString(row.TYP || row.typ) === 'V');

  // Filter by date (Today - filterDays)
  if (filterDays.value > 0) {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const cutoff = new Date(today);
    cutoff.setDate(today.getDate() - filterDays.value);
    
    list = list.filter(row => {
      const dateStr = extractString(row.bewegungsdatum || row.BEWEGUNGSDATUM);
      if (!dateStr || dateStr === '0001-01-01') return false;
      const rowDate = new Date(dateStr);
      rowDate.setHours(0, 0, 0, 0);
      return rowDate >= cutoff;
    });
  }

  if (filterHerde.value !== null && filterHerde.value !== undefined) {
    const herdeId = Number(filterHerde.value);
    const herde = herdeOptions.value.find(h => h.id === herdeId);
    if (herde) {
      list = list.filter(row => extractInt(row.herdennummer || row.HERDENNUMMER) === herde.herdennummer);
    }
  }

  if (filterText.value) {
    const textId = Number(filterText.value);
    list = list.filter(row => extractInt(row.ID_TEXTE) === textId);
  }

  return list;
});

const columns: QTableProps['columns'] = [
  { name: 'actions', align: 'center', label: 'Aktion', field: 'actions', style: 'width: 80px' },
  { name: 'HERDEN_BEZEICHNUNG', align: 'left', label: 'Herde', field: (row: any) => extractString(row.herden_bezeichnung || row.HERDEN_BEZEICHNUNG) || row.herdennummer || row.HERDENNUMMER || '-', sortable: true },
  { name: 'BEWEGUNGSDATUM', align: 'left', label: 'Datum', field: (row: any) => extractString(row.bewegungsdatum || row.BEWEGUNGSDATUM) || '-', sortable: true },
  { name: 'BEWEGUNGEN', align: 'right', label: 'Anzahl', field: (row: any) => extractInt(row.bewegungen || row.BEWEGUNGEN) || 0, sortable: true },
  { name: 'KOSTEN', align: 'right', label: 'Kosten (€)', field: (row: any) => (row.kosten || row.KOSTEN || 0).toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }), sortable: true },
  { name: 'GRUND_TEXT', align: 'left', label: 'Grund', field: (row: any) => extractString(row.grund_text || row.GRUND_TEXT) || '-', sortable: true }
];

interface Pagination {
  rowsPerPage: number;
  sortBy: string | null;
  descending: boolean;
  page: number;
}
const pagination = ref<Pagination>({ 
  rowsPerPage: 15, 
  sortBy: 'BEWEGUNGSDATUM', 
  descending: true,
  page: 1
});

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

async function fetchVerlustTexte() {
  try {
    const res = await api.get('/api/texte/typ/V');
    textOptions.value = [
      { id: 0, betreff: 'Alle' },
      ...(res.data || []).map((t: any) => ({
        id: t.id || t.ID,
        betreff: t.betreff || t.BETREFF
      }))
    ];
  } catch (err) {
    console.error('Error fetching texts for type V:', err);
  }
}

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/tierbewegungen');
    rows.value = res.data || [];
  } catch (err) {
    console.error('Error loading movements:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Daten' });
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  initWidths(columns);
  await fetchVerlustTexte();
  await fetchHerden();
  await loadData();
});

async function fetchHerden() {
  const res = await api.get('/api/herden/lookup');
  herdeOptions.value = (res.data || []).map((h: any) => {
    const id = h.id || h.ID;
    const bez = h.bezeichnung || h.BEZEICHNUNG;
    const num = h.herdennummer || h.HERDENNUMMER;
    return {
      id: id,
      bezeichnung: bez ? bez : (num ? `Herde ${num}` : `Herde ${id}`),
      herdennummer: num
    };
  });
}

function openCreate() {
  if (!filterHerde.value) return;
  isEditing.value = false;
  editId.value = null;
  showDialog.value = true;
}

function onEdit(row: Record<string, any>) {
  isEditing.value = true;
  editId.value = row.ID;
  showDialog.value = true;
}

function onDelete(row: Record<string, any>) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diesen Verlusteintrag wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(() => {
    api.delete(`/api/tierbewegungen/${row.ID}`)
      .then(() => {
        $q.notify({ type: 'positive', message: 'Eintrag gelöscht' });
        void loadData();
      })
      .catch((err) => {
        console.error('Error deleting:', err);
        $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
      });
  });
}
</script>

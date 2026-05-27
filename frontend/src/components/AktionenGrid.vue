<template>
  <div class="q-pa-md">
    <!-- Header with Action on Left, Title on Right -->
    <div class="row items-center justify-between q-mb-md">
      <div class="row q-gutter-md">
        <q-btn color="primary" icon="add" label="Neue Aktion" @click="openCreate" rounded unelevated />
      </div>
      <div class="text-h6 text-primary">Aktionen</div>
    </div>

    <!-- Filters -->
    <div class="row q-col-gutter-md q-mb-md items-center">
      <div class="col-12 col-sm-3">
        <q-select
          v-model="filterUser"
          :options="userOptions"
          option-value="ID"
          option-label="LABEL"
          emit-value
          map-options
          label="Benutzer"
          filled
          stack-label
          @update:model-value="fetchData"
        >
          <template v-slot:prepend>
            <q-icon name="person" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-3">
        <q-select
          v-model="filterKz"
          :options="kzOptions"
          option-value="KZ"
          option-label="LABEL"
          emit-value
          map-options
          clearable
          label="Aktionstyp"
          filled
          stack-label
          @update:model-value="fetchData"
        >
          <template v-slot:prepend>
            <q-icon name="filter_list" />
          </template>
        </q-select>
      </div>

      <div class="col-12 col-sm-3">
        <q-input filled v-model="filterDateRangeText" label="Zeitraum" stack-label readonly dense>
          <template v-slot:prepend>
            <q-icon name="event" class="cursor-pointer">
              <q-popup-proxy cover transition-show="scale" transition-hide="scale">
                <q-date v-model="filterDateRange" range @update:model-value="fetchData">
                  <div class="row items-center justify-end">
                    <q-btn v-close-popup label="Schließen" color="primary" flat />
                  </div>
                </q-date>
              </q-popup-proxy>
            </q-icon>
          </template>
          <template v-slot:append v-if="filterDateRange">
            <q-icon name="close" @click.stop="filterDateRange = null; fetchData()" class="cursor-pointer" />
          </template>
        </q-input>
      </div>

      <div class="col-12 col-sm-2">
        <q-select
          v-model="filterStatus"
          :options="[{value: 0, label: 'Offen'}, {value: 1, label: 'Erledigt'}, {value: 2, label: 'Alle'}]"
          emit-value
          map-options
          label="Status"
          filled
          stack-label
          @update:model-value="fetchData"
        />
      </div>

      <q-space />
      
      <div class="col-auto">
        <q-btn flat round icon="refresh" @click="fetchData" />
      </div>
    </div>

    <!-- Grid -->
    <q-table
      :rows="rows"
      :columns="columns"
      row-key="id"
      :loading="loading"
      :pagination="pagination"
      class="huhnlite-grid-standard shadow-2 cursor-pointer"
      :dark="$q.dark.isActive"
      separator="cell"
      @row-dblclick="(evt, row) => onEdit(row)"
    >
      <template v-slot:body-cell-erledigt="props">
        <q-td :props="props" class="text-center">
          <q-checkbox 
            :model-value="extractInt(props.row.erledigt) === 1" 
            @update:model-value="(val) => toggleErledigt(props.row, val)"
            color="positive"
            :disable="extractString(props.row.aktionen_kz) === 'W'"
          />
        </q-td>
      </template>

      <template v-slot:body-cell-actions="props">
        <q-td :props="props" class="text-center">
          <div class="row no-wrap q-gutter-x-xs justify-center">
            <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm"/>
            <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm"/>
          </div>
        </q-td>
      </template>
      
      <template v-slot:body-cell-aktionsdatum="props">
        <q-td :props="props">
          {{ formatDateLocalized(props.row.aktionsdatum) }}
        </q-td>
      </template>

      <template v-slot:body-cell-username="props">
        <q-td :props="props">
          {{ extractString(props.row.username) || getUserLabel(extractInt(props.row.id_user)) }}
        </q-td>
      </template>

      <template v-slot:body-cell-username_erledigt="props">
        <q-td :props="props">
          {{ extractString(props.row.username_erledigt) || (extractInt(props.row.id_user_erledigt) > 0 ? getUserLabel(extractInt(props.row.id_user_erledigt)) : '') }}
        </q-td>
      </template>

      <template v-slot:body-cell-aktionen_kz="props">
        <q-td :props="props">
          {{ getKzLabel(extractString(props.row.aktionen_kz)) }}
        </q-td>
      </template>

      <template v-slot:body-cell-bezeichnung="props">
        <q-td :props="props">
          {{ extractString(props.row.bezeichnung) }}
        </q-td>
      </template>

      <template v-slot:body-cell-intervall_tage="props">
        <q-td :props="props">
          {{ extractInt(props.row.intervall_tage) }}
        </q-td>
      </template>

      <template v-slot:body-cell-anzahl_intervalle="props">
        <q-td :props="props">
          {{ extractInt(props.row.anzahl_intervalle) }}
        </q-td>
      </template>

      <template v-slot:body-cell-erledigt_am="props">
        <q-td :props="props">
          {{ extractString(props.row.erledigt_am) || '-' }}
        </q-td>
      </template>
    </q-table>

    <!-- Create/Edit Dialog -->
    <q-dialog v-model="showDialog" persistent>
      <q-card style="width: 500px; max-width: 95vw; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? 'Aktion bearbeiten' : 'Neue Aktion' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup unelevated color="white" flat />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-y-md">
            <div class="row q-col-gutter-sm">
              <div class="col-12">
                <q-select
                  v-model="form.AKTIONEN_KZ"
                  :options="kzOptions"
                  option-value="KZ"
                  option-label="LABEL"
                  emit-value
                  map-options
                  label="Aktionstyp *"
                  filled
                  stack-label
                  :rules="[val => !!val || 'Erforderlich']"
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-12 col-sm-6">
                <q-input
                  v-model="form.AKTIONSDATUM"
                  type="date"
                  label="Datum *"
                  filled
                  stack-label
                  :rules="[val => !!val || 'Erforderlich']"
                />
              </div>
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="form.ID_USER"
                  :options="userOptions"
                  option-value="ID"
                  option-label="LABEL"
                  emit-value
                  map-options
                  label="Benutzer"
                  filled
                  stack-label
                />
              </div>
            </div>

            <q-input
              v-model="form.BEZEICHNUNG"
              label="Bezeichnung / Beschreibung *"
              filled
              stack-label
              type="textarea"
              rows="3"
              :rules="[val => !!val || 'Erforderlich']"
            />

            <div class="row q-col-gutter-sm">
              <div class="col-6">
                <q-input
                  v-model.number="form.INTERVALL_TAGE"
                  type="number"
                  label="Intervall (Tage)"
                  filled
                  stack-label
                  hint="0 = Einmalig"
                  :disable="form.AKTIONEN_KZ !== 'W'"
                />
              </div>
              <div class="col-6">
                <q-input
                  v-model.number="form.ANZAHL_INTERVALLE"
                  type="number"
                  label="Anzahl Intervalle"
                  filled
                  stack-label
                  hint="0 = Unendlich"
                  :disable="form.AKTIONEN_KZ !== 'W'"
                />
              </div>
            </div>

            <div class="row" v-if="form.AKTIONEN_KZ === 'W'">
              <div class="col-12">
                <q-btn 
                  label="Erstelle Anzahl Intervalle" 
                  color="secondary" 
                  class="full-width" 
                  rounded 
                  outline
                  :disable="!(form.ANZAHL_INTERVALLE > 0 && form.INTERVALL_TAGE > 0)"
                  @click="generateIntervals"
                >
                  <q-tooltip v-if="!(form.ANZAHL_INTERVALLE > 0 && form.INTERVALL_TAGE > 0)">
                    Bitte Intervall (Tage) und Anzahl eingeben.
                  </q-tooltip>
                </q-btn>
              </div>
            </div>

            <div class="row">
              <div class="col-12 flex items-center">
                <q-checkbox 
                  v-model="form.ERLEDIGT_BOOL" 
                  label="Bereits Erledigt" 
                  color="positive" 
                  :disable="form.AKTIONEN_KZ === 'W'"
                />
              </div>
            </div>

            <div class="row justify-end q-mt-md q-gutter-x-sm">
              <q-btn label="Abbrechen" color="negative" outline rounded v-close-popup padding="xs lg" />
              <q-btn :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated padding="xs xl" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useQuasar, date } from 'quasar';
import { api } from '../boot/api';
import { useSessionStore } from '../stores/session';

/* eslint-disable @typescript-eslint/no-explicit-any */

const $q = useQuasar();
const route = useRoute();
const sessionStore = useSessionStore();

const loading = ref(false);
const rows = ref<any[]>([]);
const kzOptions = ref<any[]>([]);
const userOptions = ref<any[]>([{ ID: 0, LABEL: 'Alle' }]);

const filterKz = ref<string | null>(null);
const filterUser = ref<number>(0);
const filterDateRange = ref<{ from: string; to: string } | null>(null);
const filterStatus = ref<number>(0); // 0: Offen, 1: Erledigt, 2: Alle

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);

const form = reactive({
  AKTIONEN_KZ: '',
  ID_USER: 0,
  AKTIONSDATUM: date.formatDate(new Date(), 'YYYY-MM-DD'),
  BEZEICHNUNG: '',
  INTERVALL_TAGE: 0,
  ANZAHL_INTERVALLE: 0,
  ERLEDIGT: 0,
  ERLEDIGT_BOOL: false,
  ID_USER_ERLEDIGT: 0,
  ERLEDIGT_AM: ''
});

const columns = [
  { name: 'erledigt', label: 'Status', field: 'erledigt', align: 'center', sortable: true },
  { name: 'aktionen_kz', label: 'Typ', field: 'aktionen_kz', align: 'left', sortable: true },
  { name: 'aktionsdatum', label: 'Datum', field: 'aktionsdatum', align: 'left', sortable: true },
  { name: 'username', label: 'Zuständig', field: 'username', align: 'left', sortable: true },
  { name: 'bezeichnung', label: 'Bezeichnung', field: 'bezeichnung', align: 'left', sortable: true },
  { name: 'intervall_tage', label: 'Tage', field: 'intervall_tage', align: 'right', sortable: true },
  { name: 'anzahl_intervalle', label: 'Anzahl', field: 'anzahl_intervalle', align: 'right', sortable: true },
  { name: 'username_erledigt', label: 'Erledigt von', field: 'username_erledigt', align: 'left', sortable: true },
  { name: 'erledigt_am', label: 'Erledigt am', field: 'erledigt_am', align: 'left', sortable: true },
  { name: 'actions', label: 'Aktionen', align: 'center' }
];

const pagination = ref({
  sortBy: 'aktionsdatum',
  descending: true,
  page: 1,
  rowsPerPage: 15
});

const filterDateRangeText = computed(() => {
  if (!filterDateRange.value) return '';
  if (typeof filterDateRange.value === 'string') {
    return (filterDateRange.value as string).split('/').reverse().join('.');
  }
  const from = filterDateRange.value.from.split('/').reverse().join('.');
  const to = filterDateRange.value.to.split('/').reverse().join('.');
  return `${from} - ${to}`;
});

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  return Number(val) || 0;
};

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const formatDateLocalized = (val: any) => {
  const s = extractString(val);
  if (!s || s === '0001-01-01') return '-';
  return date.formatDate(s, 'DD.MM.YYYY');
};

const getUserLabel = (id: number) => {
  if (id === 0) return 'Alle';
  const u = userOptions.value.find(o => o.ID === id);
  return u ? u.LABEL : `User ${id}`;
};

const getKzLabel = (kz: string) => {
  if (!kz) return '-';
  const opt = kzOptions.value.find(o => o.KZ === kz);
  return opt ? opt.LABEL : kz;
};

const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = {
      show_erledigt: filterStatus.value,
      id_user: filterUser.value
    };
    if (filterKz.value) params.kz = filterKz.value;

    if (filterDateRange.value) {
      if (typeof filterDateRange.value === 'string') {
        const d = (filterDateRange.value as string).replace(/\//g, '-');
        params.start = d;
        params.end = d;
      } else {
        params.start = filterDateRange.value.from.replace(/\//g, '-');
        params.end = filterDateRange.value.to.replace(/\//g, '-');
      }
    }

    console.log('[DEBUG] Fetching actions with params:', params);
    const res = await api.get('/api/aktionen', { params });
    console.log('[DEBUG] Actions received:', res.data.length);
    rows.value = res.data || [];
  } catch (err) {
    console.error('Error fetching actions:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Aktionen' });
  } finally {
    loading.value = false;
  }
}

async function fetchOptions() {
  try {
    // Aktionen Typen
    const resTexte = await api.get('/api/texte');
    kzOptions.value = (resTexte.data || [])
      .filter((t: any) => t.TEXT_TYP_KZ === 'A')
      .map((t: any) => ({
        KZ: t.KZ_OBJEKT || t.KZ || '',
        LABEL: t.BETREFF || t.INHALT || t.KZ_OBJEKT || t.KZ || '?'
      }));

    // Benutzer
    const resUsers = await api.get('/api/benutzer');
    userOptions.value = [{ ID: 0, LABEL: 'Alle' }, ...(resUsers.data || []).map((u: any) => ({
      ID: extractInt(u.ID),
      LABEL: extractString(u.KLARNAME || u.USERNAME)
    }))];
  } catch (err) {
    console.error('Error fetching options:', err);
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  Object.assign(form, {
    AKTIONEN_KZ: kzOptions.value.length > 0 ? kzOptions.value[0].KZ : '',
    ID_USER: 0,
    AKTIONSDATUM: date.formatDate(new Date(), 'YYYY-MM-DD'),
    BEZEICHNUNG: '',
    INTERVALL_TAGE: 0,
    ANZAHL_INTERVALLE: 0,
    ERLEDIGT: 0,
    ERLEDIGT_BOOL: false,
    ID_USER_ERLEDIGT: 0,
    ERLEDIGT_AM: ''
  });
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = extractInt(row.id);
  Object.assign(form, {
    AKTIONEN_KZ: extractString(row.aktionen_kz),
    ID_USER: extractInt(row.id_user),
    AKTIONSDATUM: extractString(row.aktionsdatum),
    BEZEICHNUNG: extractString(row.bezeichnung),
    INTERVALL_TAGE: extractInt(row.intervall_tage),
    ANZAHL_INTERVALLE: extractInt(row.anzahl_intervalle),
    ERLEDIGT: extractInt(row.erledigt),
    ERLEDIGT_BOOL: extractInt(row.erledigt) === 1,
    ID_USER_ERLEDIGT: extractInt(row.id_user_erledigt),
    ERLEDIGT_AM: extractString(row.erledigt_am)
  });
  showDialog.value = true;
}

async function onSubmit(stayOpen = false) {
  try {
    const payload = {
      aktionen_kz: form.AKTIONEN_KZ,
      id_user: form.ID_USER,
      aktionsdatum: form.AKTIONSDATUM,
      bezeichnung: form.BEZEICHNUNG,
      intervall_tage: form.INTERVALL_TAGE,
      anzahl_intervalle: form.ANZAHL_INTERVALLE,
      erledigt: form.ERLEDIGT_BOOL ? 1 : 0,
      id_user_erledigt: form.ERLEDIGT_BOOL ? (form.ID_USER_ERLEDIGT || sessionStore.userId || 0) : 0,
      erledigt_am: form.ERLEDIGT_AM
    };

    let result;
    if (isEditing.value && editId.value) {
      result = await api.put(`/api/aktionen/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Aktion aktualisiert' });
    } else {
      result = await api.post('/api/aktionen', payload);
      $q.notify({ type: 'positive', message: 'Aktion erstellt' });
      // Nach Erstellen in den Bearbeiten-Modus wechseln, falls wir offen bleiben
      if (stayOpen && result.data) {
        isEditing.value = true;
        editId.value = extractInt(result.data.id);
      }
    }
    
    if (!stayOpen) {
      showDialog.value = false;
    }
    fetchData();
    return result.data;
  } catch (err) {
    console.error('Error saving action:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
    return null;
  }
}

async function generateIntervals() {
  // Wenn neu, erst speichern um eine ID zu bekommen
  if (!isEditing.value) {
    const saved = await onSubmit(true);
    if (!saved) return;
  }

  if (!isEditing.value || !editId.value) return;
  
  $q.loading.show({ message: 'Generiere Intervalle...' });
  try {
    const baseDate = new Date(form.AKTIONSDATUM);
    const count = form.ANZAHL_INTERVALLE;
    const interval = form.INTERVALL_TAGE;
    
    // 1. Create N actions of type 'B'
    // Die erste Aktion beginnt mit dem Datum im Satz Typ 'W' (i=0)
    for (let i = 0; i < count; i++) {
      const nextDate = new Date(baseDate);
      nextDate.setDate(nextDate.getDate() + (i * interval));
      
      const payload = {
        aktionen_kz: 'B',
        id_user: form.ID_USER,
        aktionsdatum: date.formatDate(nextDate, 'YYYY-MM-DD'),
        bezeichnung: form.BEZEICHNUNG,
        intervall_tage: 0,
        anzahl_intervalle: 0,
        erledigt: 0,
        id_user_erledigt: 0,
        erledigt_am: ''
      };
      await api.post('/api/aktionen', payload);
    }
    
    // 2. Update the original 'W' action to the next date after the generated ones
    const lastDate = new Date(baseDate);
    lastDate.setDate(lastDate.getDate() + (count * interval));
    form.AKTIONSDATUM = date.formatDate(lastDate, 'YYYY-MM-DD');
    
    const updatePayload = {
      aktionen_kz: 'W',
      id_user: form.ID_USER,
      aktionsdatum: form.AKTIONSDATUM,
      bezeichnung: form.BEZEICHNUNG,
      intervall_tage: interval,
      anzahl_intervalle: 0,
      erledigt: 0,
      id_user_erledigt: 0,
      erledigt_am: ''
    };
    await api.put(`/api/aktionen/${editId.value}`, updatePayload);
    
    $q.notify({ type: 'positive', message: `${count} Intervalle generiert` });
    showDialog.value = false; // Formular jetzt verlassen
    fetchData();
  } catch (err) {
    console.error('Error generating intervals:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Generieren der Intervalle' });
  } finally {
    $q.loading.hide();
  }
}

async function onDelete(row: any) {
  const id = extractInt(row.id);
  $q.dialog({
    title: 'Löschen bestätigen',
    message: 'Möchten Sie diese Aktion wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/aktionen/${id}`);
      $q.notify({ type: 'positive', message: 'Aktion gelöscht' });
      fetchData();
    } catch (err) {
      console.error('Error deleting action:', err);
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
    }
  });
}

async function toggleErledigt(row: any, val: boolean) {
  try {
    const id = extractInt(row.id);
    const payload = {
      aktionen_kz: extractString(row.aktionen_kz),
      id_user: extractInt(row.id_user),
      aktionsdatum: extractString(row.aktionsdatum),
      bezeichnung: extractString(row.bezeichnung),
      intervall_tage: extractInt(row.intervall_tage),
      anzahl_intervalle: extractInt(row.anzahl_intervalle),
      erledigt: val ? 1 : 0,
      id_user_erledigt: val ? (sessionStore.userId || 0) : 0,
      erledigt_am: extractString(row.erledigt_am)
    };
    await api.put(`/api/aktionen/${id}`, payload);
    row.erledigt = val ? 1 : 0; // Optimistic update
    $q.notify({ type: 'positive', message: val ? 'Aktion erledigt' : 'Aktion wieder offen', timeout: 1000 });
    await fetchData(); // Refresh grid to apply filters
  } catch (err) {
    console.error('Error toggling erledigt:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Aktualisieren' });
  }
}

watch(filterKz, () => fetchData());

const applyQueryFilters = () => {
  if (route.query.tab === 'aktionen' && route.query.filterKz) {
    filterKz.value = route.query.filterKz as string;
    filterUser.value = Number(route.query.filterUser) || 0;
    if (route.query.filterStartDate && route.query.filterEndDate) {
      const from = (route.query.filterStartDate as string).replace(/-/g, '/');
      const to = (route.query.filterEndDate as string).replace(/-/g, '/');
      filterDateRange.value = { from, to };
    } else if (route.query.filterDate) {
      const d = (route.query.filterDate as string).replace(/-/g, '/');
      filterDateRange.value = { from: d, to: d };
    }
    filterStatus.value = 0; // Offen
  } else if (route.query.tab === 'aktionen' && !route.query.filterKz) {
    // Wenn kein Filter in der URL, Standard-Ansicht (Alle Typen, Alle Zeiträume, Offen)
    filterKz.value = null;
    filterStatus.value = 0;
    filterDateRange.value = null; // Zeige alles Offene
  }
};

onMounted(async () => {
  await fetchOptions();
  applyQueryFilters();
  await fetchData();
});

// Watch für Route-Änderungen (z.B. wenn man von Dashboard auf Aktionen klickt)
watch(() => route.fullPath, () => {
  if (route.query.tab === 'aktionen') {
    applyQueryFilters();
    fetchData();
  }
});
</script>

<style scoped>
.huhnlite-grid-standard {
  border-radius: 12px;
}
</style>

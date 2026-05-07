<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">Textverwaltung</div>
    </div>

    <div class="row q-col-gutter-lg">
      <!-- Grid A: Texttypen -->
      <div class="col-12 col-md-5">
        <q-table
          title="Texttypen"
          :rows="typenRows"
          :columns="typenColumns"
          row-key="id"
          :loading="loadingTypen"
          v-model:pagination="paginationTypen"
          :rows-per-page-options="[10, 15, 20, 25, 50, 100, 0]"
          class="shadow-2 rounded-borders overflow-hidden"
          :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
        >
          <template v-slot:top-right>
            <q-btn color="primary" icon="add" label="Neu" @click="openTypeDialog()" rounded unelevated class="q-px-md"/>
          </template>

          <template v-slot:body="props">
            <q-tr
              :props="props"
              @click="onTypeClick(props.row)"
              :class="selectedTypeKz === extractString(props.row.kz) ? ($q.dark.isActive ? 'bg-primary text-white' : 'bg-blue-1') : ''"
              class="cursor-pointer"
            >
              <q-td key="actions" :props="props">
                <div class="row no-wrap q-gutter-x-xs">
                  <q-btn dense round icon="edit" color="primary" @click.stop="openTypeDialog(props.row)" unelevated/>
                  <q-btn v-if="isAdmin" dense round icon="delete" color="negative" @click.stop="deleteType(props.row)"
                         unelevated/>
                </div>
              </q-td>
              <q-td key="kz" :props="props">
                <div class="row no-wrap items-center">
                  <q-icon v-if="extractInt(props.row.system) === 1" name="lock" color="orange" size="xs"
                          class="q-mr-xs">
                    <q-tooltip>System-Eintrag (geschützt)</q-tooltip>
                  </q-icon>
                  <span class="text-weight-bold">{{ extractString(props.row.kz) || '-' }}</span>
                </div>
              </q-td>
              <q-td key="bezeichnung" :props="props">
                {{ extractString(props.row.bezeichnung) || '-' }}
              </q-td>
              <q-td key="system" :props="props" class="text-center">
                <q-badge :color="extractInt(props.row.system) === 1 ? 'orange' : 'transparent'"
                         :text-color="extractInt(props.row.system) === 1 ? 'white' : 'transparent'">
                  {{ extractInt(props.row.system) === 1 ? 'S' : '' }}
                </q-badge>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </div>

      <!-- Grid B: Texte -->
      <div class="col-12 col-md-7">
        <q-table
          title="Texte"
          :rows="filteredTexte"
          :columns="texteColumns"
          row-key="id"
          :loading="loadingTexte"
          v-model:pagination="paginationTexte"
          :rows-per-page-options="[10, 15, 20, 25, 50, 100, 0]"
          class="shadow-2 rounded-borders overflow-hidden"
          :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
        >
          <template v-slot:top-left>
            <q-select
              v-model="selectedTypeKz"
              :options="allTypesOption"
              option-label="label"
              option-value="kz"
              emit-value
              map-options
              dense
              filled
              rounded
              label="Typ filtern"
              style="min-width: 220px"
              class="q-mr-md"
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-3'"
            />
          </template>
          <template v-slot:top-right>
            <q-btn
              color="primary"
              icon="add"
              label="Neuer Text"
              @click="openTextDialog()"
              rounded
              unelevated
              :disabled="!selectedTypeKz"
              class="q-px-md"
            />
          </template>

          <template v-slot:body-cell-actions="props">
            <q-td :props="props" auto-width>
              <div class="row no-wrap q-gutter-x-xs">
                <q-btn dense round icon="edit" color="primary" @click="openTextDialog(props.row)" unelevated />
                <q-btn v-if="isAdmin" dense round icon="delete" color="negative" @click="deleteText(props.row)"
                       unelevated/>
              </div>
            </q-td>
          </template>

          <template v-slot:body-cell-kz="props">
            <q-td :props="props">
              <div class="row no-wrap items-center">
                <q-icon v-if="extractInt(props.row.system) === 1" name="lock" color="orange" size="xs" class="q-mr-xs">
                  <q-tooltip>System-Eintrag (geschützt)</q-tooltip>
                </q-icon>
                <q-badge v-if="props.value" color="blue-2" text-color="blue-10" class="text-weight-bold">
                  {{ props.value }}
                </q-badge>
              </div>
            </q-td>
          </template>

          <template v-slot:body-cell-system="props">
            <q-td :props="props" class="text-center">
              <q-badge :color="props.value === 1 ? 'orange' : 'transparent'"
                       :text-color="props.value === 1 ? 'white' : 'transparent'">
                {{ props.value === 1 ? 'S' : '' }}
              </q-badge>
            </q-td>
          </template>
        </q-table>
      </div>
    </div>

    <!-- Dialog Texttyp -->
    <q-dialog v-model="showTypeDialog" persistent @show="onTypeDialogShow">
      <q-card style="min-width: 400px; border-radius: 16px;">
        <q-card-section class="bg-primary text-white q-pa-md row items-center">
          <div class="text-h6 text-weight-bold">{{ editTypeData.id ? 'Texttyp bearbeiten' : 'Neuer Texttyp' }}</div>
          <q-space/>
          <q-btn icon="close" flat round dense v-close-popup/>
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="saveType" class="q-gutter-md">
            <q-input
              v-model="editTypeData.kz"
              label="Kürzel (KZ) *"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :rules="[val => !!val || 'Pflichtfeld']"
              :readonly="!!editTypeData.id && !isAdmin"
            />
            <q-input
              v-model="editTypeData.bezeichnung"
              label="Bezeichnung"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :readonly="!!editTypeData.id && !isAdmin"
            />

            <q-checkbox
              v-if="isAdmin"
              v-model="editTypeData.system"
              :false-value="0"
              :true-value="1"
              label="System-Eintrag (geschützt)"
              color="orange"
            />

            <div class="row justify-end q-mt-lg q-gutter-x-sm">
              <q-btn ref="typeCancelBtn" label="Abbrechen" color="negative" outline rounded v-close-popup
                     padding="xs lg"/>
              <q-btn ref="typeSaveBtn" label="Speichern" type="submit" color="primary" rounded unelevated
                     padding="xs xl"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Dialog Texte -->
    <q-dialog v-model="showTextDialog" persistent @show="onTextDialogShow">
      <q-card style="min-width: 550px; border-radius: 16px;">
        <q-card-section class="bg-primary text-white q-pa-md row items-center">
          <div class="text-h6 text-weight-bold">{{ editTextData.id ? 'Text bearbeiten' : 'Neuer Text' }}</div>
          <q-space/>
          <q-btn icon="close" flat round dense v-close-popup/>
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="saveText" class="q-gutter-md">
            <q-select
              v-model="editTextData.text_typ_kz"
              :options="typeOptions"
              option-label="label"
              option-value="kz"
              emit-value
              map-options
              label="Texttyp"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-3'"
              :rules="[val => !!val || 'Pflichtfeld']"
              readonly
            />
            <div class="row q-col-gutter-md">
              <div class="col-4">
                <q-input
                  v-model="editTextData.kz"
                  label="Kennz. (Funktion)"
                  filled
                  maxlength="1"
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                  :readonly="!!editTextData.id && !isAdmin"
                />
              </div>
              <div class="col-8">
                <q-input
                  v-model="editTextData.betreff"
                  label="Betreff"
                  filled
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                />
              </div>
            </div>
            <q-input
              v-model="editTextData.inhalt"
              label="Inhalt"
              type="textarea"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
            />

            <q-checkbox
              v-if="isAdmin"
              v-model="editTextData.system"
              :false-value="0"
              :true-value="1"
              label="System-Eintrag (geschützt)"
              color="orange"
            />

            <div class="row justify-end q-mt-lg q-gutter-x-sm">
              <q-btn ref="textCancelBtn" label="Abbrechen" color="negative" outline rounded v-close-popup
                     padding="xs lg"/>
              <q-btn ref="textSaveBtn" label="Speichern" type="submit" color="primary" rounded unelevated
                     padding="xs xl"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useQuasar } from 'quasar';
import type { QTableProps } from 'quasar';
import { api } from 'src/boot/api';
import { useSessionStore } from '../stores/session';

/* eslint-disable @typescript-eslint/no-explicit-any */

const $q = useQuasar();
const sessionStore = useSessionStore();

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  return Number(val) || 0;
};

const isAdmin = computed(() => sessionStore.profile_kz === 'A');

// Grid Typen
const typenRows = ref<any[]>([]);
const loadingTypen = ref(false);
const typenColumns: QTableProps['columns'] = [
  { name: 'actions', label: 'Aktion', align: 'center', field: 'actions' },
  {name: 'kz', label: 'KZ', align: 'left', field: (row: any) => extractString(row.kz), sortable: true},
  {
    name: 'bezeichnung',
    label: 'Bezeichnung',
    align: 'left',
    field: (row: any) => extractString(row.bezeichnung),
    sortable: true
  },
  {name: 'system', label: 'System', align: 'center', field: (row: any) => extractInt(row.system), sortable: true}
];

// Grid Texte
const texteRows = ref<any[]>([]);
const loadingTexte = ref(false);
const texteColumns: QTableProps['columns'] = [
  { name: 'actions', label: 'Aktion', align: 'center', field: 'actions' },
  {
    name: 'text_typ_kz',
    label: 'Typ',
    align: 'left',
    field: (row: any) => extractString(row.text_typ_kz),
    sortable: true
  },
  {name: 'kz', label: 'Kennz.', align: 'center', field: (row: any) => extractString(row.kz), sortable: true},
  {name: 'betreff', label: 'Betreff', align: 'left', field: (row: any) => extractString(row.betreff), sortable: true},
  {name: 'inhalt', label: 'Inhalt', align: 'left', field: (row: any) => extractString(row.inhalt), sortable: true},
  {name: 'system', label: 'System', align: 'center', field: (row: any) => extractInt(row.system), sortable: true}
];

// Dialog State
const showTypeDialog = ref(false);
const showTextDialog = ref(false);
const editTypeData = reactive({id: null as number | null, kz: '', bezeichnung: '', system: 0});
const editTextData = reactive({id: null as number | null, text_typ_kz: '', kz: '', betreff: '', inhalt: '', system: 0});

const originalTypeState = ref('');
const isTypeDirty = computed(() => JSON.stringify(editTypeData) !== originalTypeState.value);
const typeCancelBtn = ref<any>(null);
const typeSaveBtn = ref<any>(null);

function onTypeDialogShow() {
  originalTypeState.value = JSON.stringify(editTypeData);
  setTimeout(() => { (typeCancelBtn.value)?.$el?.focus(); }, 50);
}

watch(isTypeDirty, (dirty) => {
  if (dirty && (document.activeElement === (typeCancelBtn.value)?.$el || document.activeElement === document.body)) {
    (typeSaveBtn.value)?.$el?.focus();
  }
});

const originalTextState = ref('');
const isTextDirty = computed(() => JSON.stringify(editTextData) !== originalTextState.value);
const textCancelBtn = ref<any>(null);
const textSaveBtn = ref<any>(null);

function onTextDialogShow() {
  originalTextState.value = JSON.stringify(editTextData);
  setTimeout(() => { (textCancelBtn.value)?.$el?.focus(); }, 50);
}

watch(isTextDirty, (dirty) => {
  if (dirty && (document.activeElement === (textCancelBtn.value)?.$el || document.activeElement === document.body)) {
    (textSaveBtn.value)?.$el?.focus();
  }
});

const selectedTypeKz = ref<string | null>(null);
const paginationTypen = ref({rowsPerPage: 15});
const paginationTexte = ref({rowsPerPage: 15});

const typeOptions = computed(() => {
  return typenRows.value.map(t => ({
    label: `${extractString(t.kz)} - ${extractString(t.bezeichnung || '')}`,
    kz: extractString(t.kz)
  }));
});

const allTypesOption = computed(() => {
  return [
    { label: 'Alle Typen', kz: null },
    ...typeOptions.value
  ];
});

const filteredTexte = computed(() => {
  if (!selectedTypeKz.value) return texteRows.value;
  return texteRows.value.filter(t => extractString(t.text_typ_kz) === selectedTypeKz.value);
});

onMounted(async () => {
  await loadTypen();
  await loadTexte();
});

async function loadTypen() {
  loadingTypen.value = true;
  try {
    const res = await api.get('/api/texttypen');
    typenRows.value = res.data || [];

    if (typenRows.value.length > 0 && selectedTypeKz.value === null) {
      selectedTypeKz.value = extractString(typenRows.value[0].kz);
    }
  } catch (err) {
    console.error('Ladefehler Typen:', err);
  } finally {
    loadingTypen.value = false;
  }
}

async function loadTexte() {
  loadingTexte.value = true;
  try {
    const res = await api.get('/api/texte');
    texteRows.value = res.data || [];
  } catch (err) {
    console.error('Ladefehler Texte:', err);
  } finally {
    loadingTexte.value = false;
  }
}

function onTypeClick(row: any) {
  selectedTypeKz.value = extractString(row.kz);
}

function openTypeDialog(row: any = null) {
  if (row) {
    editTypeData.id = extractInt(row.id);
    editTypeData.kz = extractString(row.kz);
    editTypeData.bezeichnung = extractString(row.bezeichnung);
    editTypeData.system = extractInt(row.system);
  } else {
    editTypeData.id = null;
    editTypeData.kz = '';
    editTypeData.bezeichnung = '';
    editTypeData.system = 0;
  }
  showTypeDialog.value = true;
}

async function saveType() {
  try {
    if (!isAdmin.value) editTypeData.system = 0;
    if (editTypeData.id) {
      await api.put(`/api/texttypen/${editTypeData.id}`, editTypeData);
    } else {
      await api.post('/api/texttypen', editTypeData);
    }
    showTypeDialog.value = false;
    await loadTypen();
    $q.notify({ type: 'positive', message: 'Texttyp gespeichert' });
  } catch (err: any) {
    $q.notify({ type: 'negative', message: 'Fehler: ' + (err.response?.data?.error || err.message) });
  }
}

function deleteType(row: any) {
  const rowId = extractInt(row.id);
  const rowKz = extractString(row.kz);
  $q.dialog({
    title: 'Löschen',
    message: `Soll der Texttyp "${rowKz}" wirklich gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/texttypen/${rowId}`);
      await loadTypen();
      $q.notify({type: 'positive', message: 'Texttyp gelöscht'});
    } catch (err: any) {
      $q.notify({
        type: 'negative',
        message: 'Fehler (evtl. hängen noch Texte daran): ' + (err.response?.data?.error || err.message)
      });
    }
  });
}

function openTextDialog(row: any = null) {
  if (row) {
    editTextData.id = extractInt(row.id);
    editTextData.text_typ_kz = extractString(row.text_typ_kz);
    editTextData.kz = extractString(row.kz);
    editTextData.betreff = extractString(row.betreff);
    editTextData.inhalt = extractString(row.inhalt);
    editTextData.system = extractInt(row.system);
  } else {
    editTextData.id = null;
    editTextData.text_typ_kz = selectedTypeKz.value || '';
    editTextData.kz = '';
    editTextData.betreff = '';
    editTextData.inhalt = '';
    editTextData.system = 0;
  }
  showTextDialog.value = true;
}

async function saveText() {
  try {
    if (!isAdmin.value) editTextData.system = 0;
    if (editTextData.id) {
      await api.put(`/api/texte/${editTextData.id}`, editTextData);
    } else {
      await api.post('/api/texte', editTextData);
    }
    showTextDialog.value = false;
    await loadTexte();
    $q.notify({ type: 'positive', message: 'Text gespeichert' });
  } catch (err: any) {
    $q.notify({ type: 'negative', message: 'Fehler: ' + (err.response?.data?.error || err.message) });
  }
}

function deleteText(row: any) {
  const rowId = extractInt(row.id);
  $q.dialog({
    title: 'Löschen',
    message: 'Soll dieser Text wirklich gelöscht werden?',
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/texte/${rowId}`);
      await loadTexte();
      $q.notify({type: 'positive', message: 'Text gelöscht'});
    } catch (err: any) {
      $q.notify({type: 'negative', message: 'Fehler: ' + (err.response?.data?.error || err.message)});
    }
  });
}
</script>

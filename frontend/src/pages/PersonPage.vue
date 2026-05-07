<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">Personenverwaltung</div>
    </div>

    <!-- Main Table in Grid Mode (Cardview) -->
    <q-table
      :rows="filteredRows"
      :columns="columns"
      row-key="id"
      grid
      hide-header
      :loading="loading"
      :filter="filter"
      v-model:pagination="pagination"
      class="bg-transparent"
    >
      <template v-slot:top>
        <div class="full-width row items-center justify-between q-gutter-md">
          <div class="text-h6 text-grey-8">Übersicht</div>
          <div class="row q-gutter-sm items-center">
            <q-select
              ref="typeSelectRef"
              v-model="typeFilter"
              :options="personTypOptions"
              label="Typ filtern"
              option-value="id"
              option-label="betreff"
              emit-value
              map-options
              dense
              filled
              rounded
              stack-label
              clearable
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :dark="$q.dark.isActive"
              style="min-width: 180px"
              @clear="onClearType"
            />
            <q-input
              v-model="filter"
              label="Suchen im Namen"
              dense
              filled
              rounded
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              :dark="$q.dark.isActive"
              style="width: 250px"
              clearable
            >
              <template v-slot:append>
                <q-icon name="search" />
              </template>
            </q-input>
            <q-btn color="primary" icon="add" label="Neue Person" @click="openCreate" rounded unelevated />
          </div>
        </div>
      </template>

      <!-- Card View Definition -->
      <template v-slot:item="props">
        <div class="q-pa-md col-xs-12 col-sm-6 col-md-4">
          <q-card
            :class="$q.dark.isActive ? 'bg-dark-page text-white' : 'bg-grey-2 text-black hover-shadow'"
            class="shadow-2 overflow-hidden cursor-pointer non-selectable"
            :style="{ borderRadius: '16px', border: $q.dark.isActive ? '1px solid #424242' : '1px solid #e0e0e0', transition: 'all 0.3s ease' }"
            flat
            v-ripple
            @click="onEdit(props.row)"
          >
            <!-- Header section with Name & Type -->
            <q-card-section class="bg-primary text-white q-py-sm q-px-md row items-center justify-between">
              <div class="row items-center q-gutter-x-md">
                <q-avatar size="48px" class="shadow-2 border-white" style="border: 2px solid white;">
                  <img v-if="props.row.FOTO" :src="'data:image/jpeg;base64,' + props.row.FOTO"/>
                  <q-icon v-else name="person" />
                </q-avatar>
                <div class="column">
                  <div class="text-subtitle2 text-grey-3" style="font-size: 11px; line-height: 1;">
                    {{ getAnredeLabel(extractInt(props.row.ID_ANREDE)) }}
                  </div>
                  <div class="text-h6 text-weight-bold">{{ extractString(props.row.NAME) || '-' }}</div>
                </div>
              </div>
              <q-badge color="white" text-color="primary" class="text-weight-bold" rounded outline shadow-1 @click.stop>
                {{ getPersonTypLabel(extractInt(props.row.ID_TEXTE)) }}
              </q-badge>
            </q-card-section>

            <q-card-section class="q-pa-none">
              <q-list dense>
                <!-- Firma -->
                <q-item v-if="extractString(props.row.FIRMA)" class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="business" color="grey-7" size="20px" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Firma</q-item-label>
                    <q-item-label class="text-weight-medium">{{ extractString(props.row.FIRMA) }}</q-item-label>
                  </q-item-section>
                </q-item>

                <!-- Anschrift -->
                <q-item class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="location_on" color="grey-7" size="20px" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Anschrift</q-item-label>
                    <q-item-label>{{ extractString(props.row.STRASSE) || '-' }}</q-item-label>
                    <q-item-label v-if="extractString(props.row.PLZ)" class="text-grey-8">
                      {{ extractString(props.row.PLZ) }} {{ extractString(props.row.ORT) }}
                    </q-item-label>
                  </q-item-section>
                </q-item>

                <!-- ID & Typ -->
                <q-item class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="badge" color="grey-7" size="20px" />
                  </q-item-section>
                  <q-item-section>
                    <div class="row q-col-gutter-x-md">
                      <div class="col-auto">
                        <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Nr.</q-item-label>
                        <q-item-label class="text-weight-medium">{{ extractInt(props.row.PERSONENNUMMER || props.row.personennummer) || '-' }}</q-item-label>
                      </div>
                      <div class="col">
                        <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Typ</q-item-label>
                        <q-item-label>{{ getPersonTypLabel(extractInt(props.row.ID_TEXTE || props.row.id_texte)) }}</q-item-label>
                      </div>
                    </div>
                  </q-item-section>
                </q-item>

                <q-separator inset class="q-my-xs" />

                <!-- Mobil -->
                <q-item v-if="extractString(props.row.MOBILTELEPHON)" class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="phone_android" color="primary" size="20px" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Mobil</q-item-label>
                    <q-item-label>
                      <a :href="'tel:' + extractString(props.row.MOBILTELEPHON)" class="text-primary text-weight-medium"
                         style="text-decoration: none;" @click.stop>
                        {{ extractString(props.row.MOBILTELEPHON) }}
                      </a>
                    </q-item-label>
                  </q-item-section>
                </q-item>

                <!-- Email -->
                <q-item v-if="extractString(props.row.EMAIL)" class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="email" color="primary" size="20px" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">E-Mail</q-item-label>
                    <q-item-label>
                      <a :href="'mailto:' + extractString(props.row.EMAIL)" class="text-primary text-weight-medium"
                         style="text-decoration: none;" @click.stop>
                        {{ extractString(props.row.EMAIL) }}
                      </a>
                    </q-item-label>
                  </q-item-section>
                </q-item>

                <!-- Homepage -->
                <q-item v-if="extractString(props.row.HOMEPAGE)" class="q-py-sm">
                  <q-item-section avatar>
                    <q-icon name="language" color="primary" size="20px"/>
                  </q-item-section>
                  <q-item-section>
                    <q-item-label caption class="text-weight-bold text-uppercase" style="font-size: 10px;">Homepage
                    </q-item-label>
                    <q-item-label>
                      <a
                        :href="extractString(props.row.HOMEPAGE).startsWith('http') ? extractString(props.row.HOMEPAGE) : 'https://' + extractString(props.row.HOMEPAGE)"
                        target="_blank" class="text-primary text-weight-medium" style="text-decoration: none;"
                        @click.stop>
                        {{ extractString(props.row.HOMEPAGE) }}
                      </a>
                    </q-item-label>
                  </q-item-section>
                </q-item>
              </q-list>
            </q-card-section>

            <!-- Card Actions Footer -->
            <q-card-actions align="right" class="q-px-md q-pb-md">
              <q-btn dense round icon="edit" color="primary" flat @click.stop="onEdit(props.row)" />
              <q-btn dense round icon="delete" color="negative" flat @click.stop="onDelete(props.row)" />
            </q-card-actions>
          </q-card>
        </div>
      </template>

      <!-- Empty State -->
      <template v-slot:no-data>
        <div class="full-width column items-center q-pa-xl text-grey-6">
          <q-icon name="group" size="84px" />
          <div class="text-h6 q-mt-md">Keine Personen gefunden</div>
          <q-btn color="primary" label="Erste Person anlegen" class="q-mt-md" @click="openCreate" unelevated rounded />
        </div>
      </template>
    </q-table>

    <!-- Dialog Form (Existing logic preserved) -->
    <q-dialog v-model="showDialog" persistent @show="onDialogShow">
      <q-card style="min-width: 450px; max-width: 600px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? 'Person bearbeiten' : 'Neue Person hinzufügen' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <!-- Photo Upload -->
            <div class="row items-center q-gutter-x-md q-mb-md">
              <q-avatar size="80px" class="shadow-1">
                <img v-if="form.FOTO"
                     :src="form.FOTO.startsWith('data:') ? form.FOTO : 'data:image/jpeg;base64,' + form.FOTO"/>
                <q-icon v-else name="add_a_photo" class="text-grey-5" />
              </q-avatar>
              <div class="column q-gutter-y-xs">
                <q-file
                  v-model="photoFile"
                  label="Bild auswählen"
                  filled
                  dense
                  accept="image/*"
                  @update:model-value="onPhotoSelected"
                  class="bg-grey-2"
                  style="width: 155px"
                >
                  <template v-slot:prepend>
                    <q-icon name="attach_file" />
                  </template>
                </q-file>
                <q-btn v-if="form.FOTO" label="Foto entfernen" color="negative" flat dense size="sm"
                       @click="form.FOTO = ''; photoFile = null"/>
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-4">
                <q-input v-model.number="form.PERSONENNUMMER" type="number" label="Personen-Nr" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
              <div class="col-2">
                <q-input v-model="form.KZ" label="KZ" filled stack-label readonly
                         :bg-color="$q.dark.isActive ? 'grey-10' : 'grey-3'"/>
              </div>
              <div class="col-6">
                 <q-select
                   v-model="form.ID_TEXTE"
                  :options="personTypOptions"
                  option-value="id"
                  option-label="betreff"
                  emit-value
                  map-options
                  label="Personen-Typ"
                  filled
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-6">
                <q-select
                  v-model="form.ID_ANREDE"
                  :options="anredeOptions"
                  option-value="id"
                  option-label="betreff"
                  emit-value
                  map-options
                  label="Anrede"
                  filled
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                />
              </div>
              <div class="col-6">
                <q-input v-model="form.NAME" label="Name *" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => !!val || 'Pflichtfeld']"/>
              </div>
            </div>

            <q-input v-model="form.FIRMA" label="Firma" filled stack-label
                     :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
            <q-input v-model="form.STRASSE" label="Straße" filled stack-label
                     :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>

            <div class="row q-col-gutter-sm">
              <div class="col-4">
                <q-input v-model="form.PLZ" label="PLZ" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
              <div class="col-8">
                <q-input v-model="form.ORT" label="Ort" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
            </div>

            <q-input v-model="form.POSTFACH" label="Postfach" filled stack-label
                     :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>

            <div class="row q-col-gutter-sm">
              <div class="col-6">
                <q-input v-model="form.TELEFON" label="Telefon" type="tel" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
              <div class="col-6">
                <q-input v-model="form.MOBILTELEPHON" label="Mobiltelefon" type="tel" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-6">
                <q-input v-model="form.EMAIL" label="Email" type="email" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
              <div class="col-6">
                <q-input v-model="form.EMAIL2" label="Email 2" type="email" filled stack-label
                         :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
              </div>
            </div>

            <q-input v-model="form.HOMEPAGE" label="Homepage" filled stack-label
                     :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" placeholder="https://..."/>

            <div class="row justify-end q-mt-lg q-gutter-x-sm">
              <q-btn ref="cancelBtn" label="Abbrechen" color="negative" outline rounded @click="closeDialog" padding="xs lg" />
              <q-btn ref="saveBtn" :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated padding="xs xl" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';

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

interface LookupOption {
  id: number;
  betreff: string;
  kz?: string;
  inhalt?: string;
}

const $q = useQuasar();
const loading = ref(false);
const rows = ref<any[]>([]);
const pagination = ref({ rowsPerPage: 50 });
const filter = ref('');
const typeFilter = ref<number | null>(null);
const typeSelectRef = ref<InstanceType<typeof QSelect> | null>(null);

function onClearType() {
  typeSelectRef.value?.hidePopup();
}

// Computed filtered rows based on typeFilter
const filteredRows = computed(() => {
  if (!typeFilter.value) return rows.value;
  return rows.value.filter(r => extractInt(r.ID_TEXTE) === typeFilter.value);
});

const showDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const form = ref({
  PERSONENNUMMER: null as number | null,
  ID_TEXTE: null as number | null,
  ID_ANREDE: null as number | null,
  KZ: '',
  NAME: '',
  FIRMA: '',
  STRASSE: '',
  PLZ: '',
  ORT: '',
  POSTFACH: '',
  TELEFON: '',
  MOBILTELEPHON: '',
  EMAIL: '',
  EMAIL2: '',
  FOTO: '',
  HOMEPAGE: ''
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

// Watch id_personentyp selection to sync KZ
watch(() => form.value.ID_TEXTE, (newVal) => {
  if (newVal) {
    const opt = personTypOptions.value.find(o => o.id === newVal);
    if (opt) {
      form.value.KZ = opt.kz || '';
    }
  }
});

const photoFile = ref<File | null>(null);

function onPhotoSelected(file: File | null) {
  if (!file) return;
  const reader = new FileReader();
  reader.onload = (e) => {
    form.value.FOTO = e.target?.result as string;
  };
  reader.readAsDataURL(file);
}

// Definitions for search filtering
const columns: QTableProps['columns'] = [
  {
    name: 'name',
    align: 'left',
    label: 'Name',
    field: (row: any) => extractString(row.name || row.NAME),
    sortable: true
  },
  {
    name: 'firma',
    align: 'left',
    label: 'Firma',
    field: (row: any) => extractString(row.firma || row.FIRMA),
    sortable: true
  },
  {name: 'ort', align: 'left', label: 'Ort', field: (row: any) => extractString(row.ort || row.ORT), sortable: true},
  {
    name: 'mobiltelephon',
    align: 'left',
    label: 'Mobil',
    field: (row: any) => extractString(row.mobiltelephon || row.MOBILTELEPHON)
  },
  {name: 'email', align: 'left', label: 'Email', field: (row: any) => extractString(row.email || row.EMAIL)}
];

const anredeOptions = ref<LookupOption[]>([]);
const personTypOptions = ref<LookupOption[]>([]);

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get<any[]>('/api/person');
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Personen)' });
  } finally {
    loading.value = false;
  }
}

function getAnredeLabel(id: number | undefined) {
  if (!id) return '';
  const opt = anredeOptions.value.find(o => o.id === id);
  return opt ? opt.betreff : '';
}

function getPersonTypLabel(id: number | undefined) {
  if (!id) return '-';
  const opt = personTypOptions.value.find(o => o.id === id);
  return opt ? opt.betreff : '-';
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  form.value = {
    PERSONENNUMMER: null,
    ID_TEXTE: null,
    ID_ANREDE: null,
    KZ: '',
    NAME: '',
    FIRMA: '',
    STRASSE: '',
    PLZ: '',
    ORT: '',
    POSTFACH: '',
    TELEFON: '',
    MOBILTELEPHON: '',
    EMAIL: '',
    EMAIL2: '',
    FOTO: '',
    HOMEPAGE: ''
  };
  photoFile.value = null;
  showDialog.value = true;
}

function onEdit(row: any) {
  isEditing.value = true;
  editId.value = extractInt(row.ID);
  form.value = {
    PERSONENNUMMER: extractInt(row.PERSONENNUMMER) || null,
    ID_TEXTE: extractInt(row.ID_TEXTE) || null,
    ID_ANREDE: extractInt(row.ID_ANREDE) || null,
    KZ: extractString(row.KZ),
    NAME: extractString(row.NAME),
    FIRMA: extractString(row.FIRMA),
    STRASSE: extractString(row.STRASSE),
    PLZ: extractString(row.PLZ),
    ORT: extractString(row.ORT),
    POSTFACH: extractString(row.POSTFACH),
    TELEFON: extractString(row.TELEFON),
    MOBILTELEPHON: extractString(row.MOBILTELEPHON),
    EMAIL: extractString(row.EMAIL),
    EMAIL2: extractString(row.EMAIL2),
    FOTO: extractString(row.FOTO),
    HOMEPAGE: extractString(row.HOMEPAGE)
  };
  photoFile.value = null;
  showDialog.value = true;
}

function closeDialog() {
  showDialog.value = false;
  setTimeout(() => {
    isEditing.value = false;
    editId.value = null;
    form.value = {
      PERSONENNUMMER: null,
      ID_TEXTE: null,
      ID_ANREDE: null,
      KZ: '',
      NAME: '',
      FIRMA: '',
      STRASSE: '',
      PLZ: '',
      ORT: '',
      POSTFACH: '',
      TELEFON: '',
      MOBILTELEPHON: '',
      EMAIL: '',
      EMAIL2: '',
      FOTO: '',
      HOMEPAGE: ''
    };
    photoFile.value = null;
  }, 300);
}

function onDelete(row: any) {
  const rowId = extractInt(row.ID);
  const rowName = extractString(row.NAME);
  $q.dialog({
    title: 'Löschen bestätigen',
    message: `Soll der Eintrag "${rowName}" wirklich gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(() => {
    api.delete(`/api/person/${rowId}`)
      .then(() => {
        $q.notify({ type: 'positive', message: 'Eintrag erfolgreich gelöscht' });
        void loadData();
      })
      .catch(() => {
        $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
      });
  });
}

async function onSubmit() {
  if (form.value.KZ?.trim().toUpperCase() === 'F') {
    $q.notify({ type: 'negative', message: 'Das Kennzeichen "F" ist für die Firmenverwaltung reserviert und darf hier nicht verwendet werden.' });
    return;
  }
  try {
    const payload = {
      ID_TEXTE: Number(form.value.ID_TEXTE),
      ID_ANREDE: Number(form.value.ID_ANREDE),
      PERSONENNUMMER: Number(form.value.PERSONENNUMMER),
      KZ: form.value.KZ,
      POSTFACH: form.value.POSTFACH,
      NAME: form.value.NAME,
      FIRMA: form.value.FIRMA,
      STRASSE: form.value.STRASSE,
      PLZ: form.value.PLZ,
      ORT: form.value.ORT,
      TELEFON: form.value.TELEFON,
      MOBILTELEPHON: form.value.MOBILTELEPHON,
      EMAIL: form.value.EMAIL,
      EMAIL2: form.value.EMAIL2,
      FOTO: form.value.FOTO,
      HOMEPAGE: form.value.HOMEPAGE
    };
    if (isEditing.value && editId.value) {
      await api.put(`/api/person/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Person erfolgreich aktualisiert' });
    } else {
      await api.post('/api/person', payload);
      $q.notify({ type: 'positive', message: 'Person erfolgreich hinzugefügt' });
    }
    closeDialog();
    void loadData();
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}

async function fetchLookups() {
  try {
    interface RawLookup {
      ID: number;
      BETREFF?: any;
      INHALT?: any;
      KZ?: any;
    }
    const [anrederes, typres] = await Promise.all([
      api.get<RawLookup[]>('/api/texte/typ/N'),
      api.get<RawLookup[]>('/api/texte/typ/P')
    ]);

    const mapLookup = (t: any) => {
      // Handle both uppercase (legacy/manual) and lowercase (SQLC default) keys
      const b = t.betreff || t.BETREFF;
      const i = t.inhalt || t.INHALT;
      const kz = t.kz || t.KZ;
      const id = t.id || t.ID;

      const kzv = typeof kz === 'string' ? kz : (kz?.String || '');
      const bv = typeof b === 'string' ? b : (b?.String || (typeof i === 'string' ? i : i?.String) || `Eintrag ${id}`);

      return {
        id: Number(id),
        kz: kzv,
        betreff: bv
      };
    };

    anredeOptions.value = (anrederes.data || []).map(mapLookup);
    personTypOptions.value = (typres.data || []).map(mapLookup).filter(o => o.KZ !== 'A' && o.KZ !== 'F');
  } catch (err) {
    console.error('Fehler beim Laden der Lookups', err);
  }
}

onMounted(() => {
  void loadData();
  void fetchLookups();
});
</script>

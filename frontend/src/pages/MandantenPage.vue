<template>
  <q-page padding class="q-pa-md">
    <div class="row items-center justify-between q-mb-lg">
      <div>
        <h1 class="text-h4 q-my-none text-weight-bold row items-center">
          <q-icon name="business" class="q-mr-sm text-primary" />
          {{ t('menu.mandanten') }}
        </h1>
        <div class="text-subtitle1 text-grey-7">
          Verwalten Sie Ihre verschiedenen Betriebsmandanten und deren Datenbanken.
        </div>
      </div>
      <q-btn
        color="primary"
        icon="add"
        label="Neuer Mandant"
        @click="showCreateDialog = true"
        unelevated
      />
    </div>

    <q-card flat bordered class="rounded-borders shadow-2">
      <q-card-section class="q-pa-none">
        <q-table
          :rows="tenants"
          :columns="columns"
          row-key="id"
          flat
          bordered
          :loading="loading"
          binary-state-sort
          :pagination="{ rowsPerPage: 0 }"
          hide-pagination
        >
          <template v-slot:body-cell-active="props">
            <q-td :props="props" class="text-center">
              <q-badge
                v-if="props.row.id === activeMandant"
                color="positive"
                text-color="white"
                label="Aktiv"
                class="q-px-sm q-py-xs text-weight-bold"
              />
              <q-btn
                v-else
                color="secondary"
                size="sm"
                label="Umschalten"
                @click="switchTenant(props.row.id)"
                :loading="switchingId === props.row.id"
                unelevated
              />
            </q-td>
          </template>

          <template v-slot:body-cell-name="props">
            <q-td :props="props">
              <div class="row items-center no-wrap">
                <span class="text-weight-bold text-body1">{{ props.row.name }}</span>
                <q-btn
                  flat
                  round
                  color="grey-6"
                  icon="edit"
                  size="sm"
                  class="q-ml-sm"
                  @click="editTenant(props.row)"
                >
                  <q-tooltip>Mandant bearbeiten</q-tooltip>
                </q-btn>
              </div>
              <div class="text-caption text-grey-7">
                Auto-Backup: {{ getAutoBackupLabel(props.row.autobackup) }}<span v-if="props.row.backuptime"> (um {{ props.row.backuptime }})</span>
                <span v-if="props.row.waittime"> • WaitTime: {{ props.row.waittime }}</span>
              </div>
            </q-td>
          </template>

          <template v-slot:body-cell-system="props">
            <q-td :props="props" class="text-center">
              <q-toggle
                v-model="props.row.system"
                :true-value="1"
                :false-value="0"
                color="orange"
                @update:model-value="updateTenant(props.row)"
              />
            </q-td>
          </template>

          <template v-slot:body-cell-test="props">
            <q-td :props="props" class="text-center">
              <q-toggle
                v-model="props.row.test"
                :true-value="1"
                :false-value="0"
                color="red"
                @update:model-value="updateTenant(props.row)"
              />
            </q-td>
          </template>

          <template v-slot:body-cell-actions="props">
            <q-td :props="props" class="text-center">
              <q-btn
                color="blue-8"
                size="sm"
                icon="cloud_upload"
                label="Export an PWA"
                @click="exportToPWA(props.row.id)"
                :loading="exportingId === props.row.id"
                unelevated
              />
            </q-td>
          </template>
        </q-table>
      </q-card-section>
    </q-card>

    <!-- Backup Path Configuration Card -->
    <q-card flat bordered class="rounded-borders shadow-2 q-mt-lg">
      <q-card-section>
        <div class="row items-center justify-between">
          <div class="row items-center">
            <q-icon name="folder_zip" size="sm" color="primary" class="q-mr-sm" />
            <span class="text-h6 text-weight-bold">Alternativer Backup-Pfad</span>
          </div>
          <q-btn
            color="primary"
            icon="save"
            label="Pfad Speichern"
            @click="saveBackupPath"
            :loading="savingPath"
            unelevated
          />
        </div>
        <div class="text-caption text-grey-7 q-mt-xs">
          Geben Sie hier ein übergeordnetes Verzeichnis (z. B. <code>Z:/HuhnLite</code>) an. Die Backups der Mandanten werden automatisch in Unterordnern <code>/mandant_n/Backups/</code> gespeichert. Wenn das Feld leer gelassen wird, werden die Backups im Standard-Datenbank-Verzeichnis angelegt.
        </div>
        <div class="row items-center q-mt-md q-col-gutter-sm">
          <div class="col">
            <q-input
              v-model="backupPath"
              label="Backup-Zielverzeichnis (z.B. Z:\HuhnLite)"
              outlined
              dense
              clearable
              placeholder="z.B. Z:\HuhnLite"
            >
              <template v-slot:append>
                <q-btn
                  flat
                  round
                  dense
                  icon="folder_open"
                  color="primary"
                  @click="browseFolder"
                >
                  <q-tooltip>Ordner über Explorer auswählen (Desktop)</q-tooltip>
                </q-btn>
                <q-btn
                  flat
                  round
                  dense
                  icon="dns"
                  color="secondary"
                  class="q-ml-xs"
                  @click="openRemoteFolderBrowser"
                >
                  <q-tooltip>Server-Ordner durchsuchen (Web / Server)</q-tooltip>
                </q-btn>
              </template>
            </q-input>
          </div>
        </div>
      </q-card-section>
    </q-card>

    <!-- Create Tenant Dialog -->
    <q-dialog v-model="showCreateDialog" persistent>
      <q-card style="min-width: 400px;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">Neuen Mandanten anlegen</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <q-input
            v-model="newTenantName"
            label="Name des Mandanten"
            placeholder="z.B. Otto Dotter"
            outlined
            autofocus
            @keyup.enter="createTenant"
            :rules="[val => !!val || 'Name ist erforderlich']"
          />
          <div class="text-caption text-grey-7 q-mt-xs">
            Ein neues Verzeichnis <code>mandant_n</code> wird angelegt, und die Referenz-Datenbank (HuhnLite.db) wird dorthin kopiert.
          </div>
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Anlegen & Umschalten"
            color="primary"
            @click="createTenant"
            :loading="creating"
            :disabled="!newTenantName.trim()"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Edit Tenant Name Dialog -->
    <q-dialog v-model="showEditDialog" persistent>
      <q-card style="min-width: 400px;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">Mandant bearbeiten</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <q-input
            v-model="editTenantName"
            label="Name des Mandanten"
            outlined
            autofocus
            @keyup.enter="saveTenantName"
            :rules="[val => !!val || 'Name ist erforderlich']"
          />
          <q-select
            v-model="editTenantAutoBackup"
            label="Automatisches Backup"
            outlined
            emit-value
            map-options
            :options="autoBackupOptions"
            class="q-mt-md"
          />
          <q-input
            v-model="editTenantBackupTime"
            label="Backup Uhrzeiten (z.B. 12:00, 20:00)"
            placeholder="Kommasepariert z.B. 12:00, 20:00"
            outlined
            class="q-mt-md"
            hint="Mehrere Zeiten mit Komma trennen (z.B. 12:00, 20:00)"
          />
          <q-input
            v-model="editTenantWaitTime"
            label="Server WaitTime (z.B. 00:01 oder 45)"
            placeholder="00:01"
            outlined
            class="q-mt-md"
            hint="Timeout für Inaktivität / Auto-Shutdown des Servers (z.B. 00:01 für 1 Min, 45 für 45 Sek)"
          />
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Speichern"
            color="primary"
            @click="saveTenantName"
            :disabled="!editTenantName.trim()"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Remote Server Directory Picker Dialog -->
    <q-dialog v-model="showRemoteFolderDialog" persistent>
      <q-card style="min-width: 550px; max-width: 90vw;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold row items-center">
            <q-icon name="dns" color="primary" class="q-mr-sm" />
            Server-Ordner durchsuchen
          </div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <div class="row items-center q-col-gutter-xs q-mb-md">
            <div class="col-auto" v-if="availableDrives.length > 0">
              <q-select
                v-model="selectedDrive"
                :options="availableDrives"
                label="Laufwerk"
                outlined
                dense
                style="min-width: 100px;"
                @update:model-value="onDriveChange"
              />
            </div>
            <div class="col">
              <q-input
                v-model="currentBrowsePath"
                label="Aktueller Serverpfad"
                outlined
                dense
                @keyup.enter="fetchRemoteDirs(currentBrowsePath)"
              >
                <template v-slot:append>
                  <q-btn flat round dense icon="arrow_forward" @click="fetchRemoteDirs(currentBrowsePath)" />
                </template>
              </q-input>
            </div>
            <div class="col-auto" v-if="parentBrowsePath">
              <q-btn
                outline
                color="primary"
                icon="arrow_upward"
                label="Nach oben"
                dense
                class="q-px-sm"
                @click="fetchRemoteDirs(parentBrowsePath)"
              >
                <q-tooltip>Eine Ebene höher</q-tooltip>
              </q-btn>
            </div>
          </div>

          <div v-if="browseLoading" class="row justify-center q-my-md">
            <q-spinner color="primary" size="2em" />
          </div>

          <q-scroll-area v-else style="height: 280px;" class="bordered rounded-borders bg-grey-1 q-pa-xs">
            <q-list separator dense v-if="folderList.length > 0">
              <q-item
                v-for="folder in folderList"
                :key="folder.path"
                clickable
                v-ripple
                @click="fetchRemoteDirs(folder.path)"
              >
                <q-item-section avatar min-width="32px">
                  <q-icon name="folder" color="amber-8" />
                </q-item-section>
                <q-item-section>
                  <q-item-label class="text-weight-medium">{{ folder.name }}</q-item-label>
                </q-item-section>
                <q-item-section side>
                  <q-icon name="chevron_right" size="xs" color="grey-6" />
                </q-item-section>
              </q-item>
            </q-list>

            <div v-else class="text-center text-grey-7 q-pa-md">
              Keine Unterordner in diesem Verzeichnis vorhanden.
            </div>
          </q-scroll-area>
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Diesen Ordner auswählen"
            color="primary"
            icon="check"
            @click="confirmRemoteFolder"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import { SelectBackupDirectory } from '../../wailsjs/go/main/App';

const { t } = useI18n();
const $q = useQuasar();

interface Tenant {
  id: number;
  name: string;
  system: number;
  test: number;
  autobackup: number;
  backuptime: string;
  waittime?: string;
}

const tenants = ref<Tenant[]>([]);
const activeMandant = ref<number | null>(null);
const loading = ref(false);
const creating = ref(false);
const switchingId = ref<number | null>(null);
const exportingId = ref<number | null>(null);

const backupPath = ref('');
const savingPath = ref(false);

interface FolderItem {
  name: string;
  path: string;
}

const showRemoteFolderDialog = ref(false);
const currentBrowsePath = ref('');
const parentBrowsePath = ref('');
const availableDrives = ref<string[]>([]);
const selectedDrive = ref('');
const folderList = ref<FolderItem[]>([]);
const browseLoading = ref(false);

const showCreateDialog = ref(false);
const newTenantName = ref('');

const showEditDialog = ref(false);
const editTenantId = ref<number | null>(null);
const editTenantName = ref('');
const editTenantSystem = ref(0);
const editTenantTest = ref(0);
const editTenantAutoBackup = ref(0);
const editTenantBackupTime = ref('');
const editTenantWaitTime = ref('00:01');

const autoBackupOptions = [
  { label: 'Deaktiviert (kein Backup)', value: 0 },
  { label: 'Beim Starten der Anwendung', value: 1 },
  { label: 'Beim Beenden der Anwendung', value: 2 },
  { label: 'Sowohl beim Starten als auch Beenden', value: 3 }
];

function getAutoBackupLabel(val: number) {
  switch (val) {
    case 1: return 'Beim Start';
    case 2: return 'Beim Ende';
    case 3: return 'Start & Ende';
    default: return 'Deaktiviert';
  }
}

const columns = [
  { name: 'id', label: 'ID', field: 'id', align: 'left' as const, sortable: true, style: 'width: 50px' },
  { name: 'name', label: 'Mandanten-Name', field: 'name', align: 'left' as const, sortable: true },
  { name: 'system', label: 'Systemverwaltung', field: 'system', align: 'center' as const },
  { name: 'test', label: 'Test-Modus (Test-DB)', field: 'test', align: 'center' as const },
  { name: 'active', label: 'Status / Aktion', field: 'id', align: 'center' as const },
  { name: 'actions', label: 'PWA-Export', field: 'id', align: 'center' as const }
];

async function loadTenants() {
  loading.value = true;
  try {
    const res = await api.get('/api/tenants');
    tenants.value = res.data.tenants || [];
    activeMandant.value = res.data.active_mandant;
    backupPath.value = res.data.backup_path || '';
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Laden der Mandanten: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    loading.value = false;
  }
}

async function saveBackupPath() {
  savingPath.value = true;
  try {
    await api.post('/api/tenants/backup-path', { backup_path: backupPath.value.trim() });
    $q.notify({
      type: 'positive',
      message: 'Backup-Pfad erfolgreich gespeichert.'
    });
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Speichern des Backup-Pfads: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    savingPath.value = false;
  }
}

async function fetchRemoteDirs(pathStr?: string) {
  browseLoading.value = true;
  try {
    const res = await api.post('/api/system/browse-dirs', { path: pathStr || '' });
    currentBrowsePath.value = res.data.current || '';
    parentBrowsePath.value = res.data.parent || '';
    availableDrives.value = res.data.drives || [];
    folderList.value = res.data.folders || [];

    if (currentBrowsePath.value && currentBrowsePath.value.includes(':')) {
      const drive = currentBrowsePath.value.split(':')[0] + ':';
      if (availableDrives.value.includes(drive)) {
        selectedDrive.value = drive;
      }
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Laden des Serverpfads: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    browseLoading.value = false;
  }
}

function onDriveChange(drive: string) {
  if (drive) {
    fetchRemoteDirs(drive + '/');
  }
}

function openRemoteFolderBrowser() {
  showRemoteFolderDialog.value = true;
  fetchRemoteDirs(backupPath.value);
}

function confirmRemoteFolder() {
  if (currentBrowsePath.value) {
    backupPath.value = currentBrowsePath.value;
  }
  showRemoteFolderDialog.value = false;
}

async function browseFolder() {
  const isWailsDesktop = !!((window as any).go?.main?.App?.SelectBackupDirectory);

  if (isWailsDesktop) {
    try {
      let selected = '';
      if (typeof SelectBackupDirectory === 'function') {
        selected = await SelectBackupDirectory(backupPath.value);
      } else {
        selected = await (window as any).go.main.App.SelectBackupDirectory(backupPath.value);
      }
      if (selected) {
        backupPath.value = selected;
      }
      return;
    } catch (err: any) {
      console.warn('Wails desktop dialog error, falling back to server backend:', err);
    }
  }

  // In Server Mode (Web Browser): Try backend REST API for native Explorer dialog on Server host
  try {
    const res = await api.post('/api/system/select-dir', { path: backupPath.value });
    if (res.data.status === 'success' && res.data.path) {
      backupPath.value = res.data.path;
      return;
    }
  } catch (err: any) {
    console.warn('Server native folder selection API failed:', err);
  }

  // Fallback to web server folder browser modal
  openRemoteFolderBrowser();
}

async function switchTenant(id: number) {
  switchingId.value = id;
  try {
    await api.post('/api/tenants/switch', { id });
    $q.notify({
      type: 'positive',
      message: 'Erfolgreich auf Mandant umgeschaltet.',
      caption: 'Die Anwendung wird neu geladen...'
    });
    setTimeout(() => {
      window.location.reload();
    }, 1000);
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Umschalten: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    switchingId.value = null;
  }
}

async function createTenant() {
  if (!newTenantName.value.trim()) return;
  creating.value = true;
  try {
    const res = await api.post('/api/tenants/create', { name: newTenantName.value.trim() });
    showCreateDialog.value = false;
    newTenantName.value = '';
    
    if (res.data.switched) {
      $q.notify({
        type: 'positive',
        message: 'Mandant erfolgreich angelegt & umgeschaltet.',
        caption: 'Die Anwendung wird neu geladen...'
      });
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } else {
      $q.notify({
        type: 'positive',
        message: 'Mandant erfolgreich angelegt.'
      });
      loadTenants();
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Anlegen: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    creating.value = false;
  }
}

function editTenant(tenant: Tenant) {
  editTenantId.value = tenant.id;
  editTenantName.value = tenant.name;
  editTenantSystem.value = tenant.system;
  editTenantTest.value = tenant.test;
  editTenantAutoBackup.value = tenant.autobackup || 0;
  editTenantBackupTime.value = tenant.backuptime || '';
  editTenantWaitTime.value = tenant.waittime || '00:01';
  showEditDialog.value = true;
}

async function saveTenantName() {
  if (!editTenantName.value.trim() || editTenantId.value === null) return;
  try {
    await api.post('/api/tenants/update', {
      id: editTenantId.value,
      name: editTenantName.value.trim(),
      system: editTenantSystem.value,
      test: editTenantTest.value,
      autobackup: editTenantAutoBackup.value,
      backuptime: editTenantBackupTime.value.trim(),
      waittime: editTenantWaitTime.value.trim()
    });
    $q.notify({
      type: 'positive',
      message: 'Mandant erfolgreich aktualisiert.'
    });
    showEditDialog.value = false;
    
    if (editTenantId.value === activeMandant.value) {
      window.location.reload();
    } else {
      loadTenants();
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Speichern: ' + (err.response?.data?.error || err.message)
    });
  }
}

async function updateTenant(tenant: Tenant) {
  try {
    await api.post('/api/tenants/update', {
      id: tenant.id,
      name: tenant.name,
      system: tenant.system,
      test: tenant.test,
      autobackup: tenant.autobackup || 0,
      backuptime: tenant.backuptime || '',
      waittime: tenant.waittime || '00:01'
    });
    $q.notify({
      type: 'positive',
      message: `Mandant ${tenant.name} erfolgreich aktualisiert.`
    });
    
    if (tenant.id === activeMandant.value) {
      $q.notify({
        type: 'info',
        message: 'Aktiver Mandant geändert. Lade Verbindung neu...'
      });
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Aktualisieren: ' + (err.response?.data?.error || err.message)
    });
    loadTenants();
  }
}

async function exportToPWA(id: number) {
  exportingId.value = id;
  try {
    const res = await api.post('/api/tenants/export', { id });
    $q.notify({
      type: 'positive',
      message: 'Export erfolgreich!',
      caption: res.data.message || 'Dateien wurden in DatenAustausch erstellt.',
      timeout: 5000,
      actions: [{ label: 'OK', color: 'white' }]
    });
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Exportieren: ' + (err.response?.data?.error || err.message),
      timeout: 0,
      actions: [{ label: 'OK', color: 'white' }]
    });
  } finally {
    exportingId.value = null;
  }
}

onMounted(() => {
  loadTenants();
});
</script>

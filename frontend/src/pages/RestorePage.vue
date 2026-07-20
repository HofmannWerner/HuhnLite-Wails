<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="col">
        <h1 class="text-h4 q-my-none">{{ t('auto.sicherungsverwaltung_title') }}</h1>
        <div class="text-subtitle1 text-grey-7">{{ t('auto.sicherungsverwaltung_subtitle') }}</div>
      </div>
      <div class="col-auto">
        <q-btn color="primary" icon="refresh" :label="t('auto.liste_aktualisieren')" @click="loadBackups" :loading="loading" unelevated />
      </div>
    </div>

    <!-- Active DB Info Card -->
    <div class="row q-col-gutter-md q-mb-lg">
      <div class="col-12">
        <q-card flat bordered class="bg-blue-grey-1 text-blue-grey-10">
          <q-card-section class="row items-center justify-between q-py-sm">
            <div class="row items-center col">
              <q-icon name="info" size="sm" class="q-mr-sm text-primary" />
              <div class="text-subtitle2 font-weight-bold">
                {{ t('auto.aktive_datenbank') }} <code class="bg-white q-px-sm q-py-xs rounded word-break-all">{{ currentDB }}</code>
              </div>
            </div>
            <div class="col-auto">
              <q-btn
                color="secondary"
                icon="backup"
                :label="t('menu.backup') || 'Backup erstellen'"
                :loading="backingUp"
                @click="runManualBackup"
                unelevated
              />
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>

    <!-- Actions Toolbar -->
    <div class="row q-gutter-sm q-mb-md items-center">
      <q-btn
        color="negative"
        icon="delete"
        :label="t('auto.delete_selected')"
        :disable="selected.length === 0 || processing"
        @click="confirmDelete"
        unelevated
      />
      <q-btn
        color="orange-9"
        icon="zip"
        :label="t('auto.compress_selected')"
        :disable="!canCompress || processing"
        @click="confirmCompress"
        unelevated
      />
      <q-btn
        color="primary"
        icon="settings_backup_restore"
        :label="t('auto.restore_selected')"
        :disable="selected.length !== 1 || processing"
        @click="confirmRestore"
        unelevated
      />
      <q-space />
      <div class="text-caption text-grey-8" v-if="selected.length > 0">
        {{ selected.length }} {{ selected.length === 1 ? 'Element ausgewählt' : 'Elemente ausgewählt' }}
      </div>
    </div>

    <!-- Grid Table -->
    <q-card flat bordered class="rounded-borders">
      <q-table
        :rows="backups"
        :columns="columns"
        row-key="path"
        selection="multiple"
        v-model:selected="selected"
        :loading="loading"
        flat
        bordered
        :pagination="{ rowsPerPage: 15 }"
        no-data-label="Keine Sicherungen gefunden"
      >
        <!-- Type Column Custom Slot -->
        <template v-slot:body-cell-type="props">
          <q-td :props="props">
            <q-badge :color="props.value === 'test' ? 'negative' : 'positive'" class="text-weight-bold">
              {{ props.value === 'test' ? t('auto.test_badge') : t('auto.prod_badge') }}
            </q-badge>
          </q-td>
        </template>

        <!-- Format Size slot -->
        <template v-slot:body-cell-size="props">
          <q-td :props="props">
            {{ formatBytes(props.value) }}
          </q-td>
        </template>

        <!-- Format Date slot -->
        <template v-slot:body-cell-time="props">
          <q-td :props="props">
            {{ formatDate(props.value) }}
          </q-td>
        </template>
      </q-table>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';

const { t } = useI18n();
const $q = useQuasar();

interface BackupFile {
  name: string;
  path: string;
  size: number;
  time: string;
  type: string;
}

const loading = ref(false);
const processing = ref(false);
const backingUp = ref(false);
const backups = ref<BackupFile[]>([]);
const selected = ref<BackupFile[]>([]);
const currentDB = ref('');

async function runManualBackup() {
  backingUp.value = true;
  $q.loading.show({
    message: 'Datenbanksicherung wird erstellt...',
    boxClass: 'bg-grey-2 text-grey-9',
    spinnerColor: 'primary'
  });
  try {
    const res = await api.post('/api/backup');
    $q.notify({
      type: 'positive',
      message: t('auto.backup_success') || 'Sicherung erfolgreich erstellt',
      caption: res.data.filename,
      timeout: 4000
    });
    loadBackups();
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: 'Sicherung fehlgeschlagen: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    $q.loading.hide();
    backingUp.value = false;
  }
}

const columns = [
  { name: 'name', align: 'left', label: t('auto.file_name'), field: 'name', sortable: true },
  { name: 'type', align: 'center', label: t('auto.db_type'), field: 'type', sortable: true },
  { name: 'size', align: 'right', label: t('auto.file_size'), field: 'size', sortable: true },
  { name: 'time', align: 'left', label: t('auto.file_date'), field: 'time', sortable: true },
  { name: 'path', align: 'left', label: t('auto.file_path'), field: 'path', sortable: true }
];

const canCompress = computed(() => {
  return selected.value.length > 0 && selected.value.some(f => !f.name.endsWith('.zip'));
});

async function loadBackups() {
  loading.value = true;
  selected.value = [];
  try {
    const res = await api.get('/api/db/list');
    currentDB.value = res.data.current;
    
    // Grid displays all backups
    backups.value = res.data.files || [];
    
    // Sort descending by date
    backups.value.sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime());
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: t('auto.error_loading_backups') + ': ' + (err.response?.data?.error || err.message)
    });
  } finally {
    loading.value = false;
  }
}

function confirmDelete() {
  if (selected.value.length === 0) return;
  
  $q.dialog({
    title: t('auto.confirm_delete_title'),
    message: t('auto.confirm_delete_msg', { count: selected.value.length }),
    cancel: true,
    persistent: true,
    ok: {
      label: t('form.delete') || 'Löschen',
      color: 'negative',
      unelevated: true
    }
  }).onOk(() => {
    runDelete();
  });
}

async function runDelete() {
  processing.value = true;
  $q.loading.show({
    message: 'Dateien werden gelöscht...',
    boxClass: 'bg-grey-2 text-grey-9',
    spinnerColor: 'negative'
  });
  try {
    await api.post('/api/db/delete', {
      paths: selected.value.map(f => f.path)
    });
    $q.notify({
      type: 'positive',
      message: t('auto.delete_success')
    });
    loadBackups();
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: 'Löschen fehlgeschlagen: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    $q.loading.hide();
    processing.value = false;
  }
}

function confirmCompress() {
  if (selected.value.length === 0) return;
  
  $q.dialog({
    title: t('auto.confirm_compress_title'),
    message: t('auto.confirm_compress_msg', { count: selected.value.filter(f => !f.name.endsWith('.zip')).length }),
    cancel: true,
    persistent: true,
    ok: {
      label: t('auto.compress_selected') || 'Komprimieren',
      color: 'warning',
      unelevated: true
    }
  }).onOk(() => {
    runCompress();
  });
}

async function runCompress() {
  processing.value = true;
  $q.loading.show({
    message: 'Dateien werden komprimiert...',
    boxClass: 'bg-grey-2 text-grey-9',
    spinnerColor: 'warning'
  });
  try {
    await api.post('/api/db/compress', {
      paths: selected.value.filter(f => !f.name.endsWith('.zip')).map(f => f.path)
    });
    $q.notify({
      type: 'positive',
      message: t('auto.compress_success')
    });
    loadBackups();
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: 'Komprimierung fehlgeschlagen: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    $q.loading.hide();
    processing.value = false;
  }
}

function confirmRestore() {
  if (selected.value.length !== 1) return;
  
  const target = selected.value[0];
  $q.dialog({
    title: t('auto.confirm_restore_title') || 'Restore bestätigen',
    message: t('auto.confirm_restore_msg', { filename: target.name, current: currentDB.value }),
    html: true,
    cancel: true,
    persistent: true,
    ok: {
      label: t('auto.yes_restore') || 'Ja, Wiederherstellen',
      color: 'negative',
      unelevated: true
    }
  }).onOk(() => {
    runRestore();
  });
}

async function runRestore() {
  if (selected.value.length !== 1) return;
  
  const target = selected.value[0];
  processing.value = true;
  $q.loading.show({
    message: 'Datenbank wird wiederhergestellt. Bitte warten...',
    boxClass: 'bg-grey-2 text-grey-9',
    spinnerColor: 'negative'
  });
  try {
    const res = await api.post('/api/db/restore', {
      PATH: target.path
    });
    
    $q.notify({
      type: 'positive',
      message: t('auto.restore_success') || 'Wiederherstellung erfolgreich',
      caption: (t('auto.safety_backup_created') || 'Sicherheits-Backup erstellt') + ': ' + res.data.safety,
      timeout: 5000
    });
    
    setTimeout(() => {
      window.location.reload();
    }, 2000);
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: (t('auto.restore_failed') || 'Restore fehlgeschlagen') + ': ' + (err.response?.data?.error || err.message),
      timeout: 0,
      actions: [{ label: 'OK', color: 'white' }]
    });
  } finally {
    $q.loading.hide();
    processing.value = false;
  }
}

function formatBytes(bytes: number, decimals = 2) {
  if (!bytes) return '0 Bytes';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

function formatDate(dateStr: string) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

onMounted(() => {
  loadBackups();
});
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
</style>

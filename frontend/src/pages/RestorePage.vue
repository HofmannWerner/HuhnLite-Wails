<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="col">
        <h1 class="text-h4 q-my-none">{{ t('auto.datenbank_wiederherstellung_restore') }}</h1>
        <div class="text-subtitle1 text-grey-7">{{ t('auto.waehlen_sie_eine_backup_datei_aus_um_den') }}</div>
      </div>
    </div>

    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-8">
        <q-card flat bordered class="rounded-borders">
          <q-card-section class="bg-warning text-black row items-center">
            <q-icon name="warning" size="md" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">{{ t('auto.achtung_ueberschreiben_von_daten') }}</div>
          </q-card-section>
          
          <q-card-section>
            <p>
              {{ t('auto.beim_wiederherstellen_wird_die_aktuelle_') }} <strong>{{ t('auto.vollstaendig_durch_das_ausgewaehlte_back') }}</strong>{{ t('auto.nicht_gespeicherte_aenderungen_gehen_ver') }}
            </p>
          </q-card-section>

          <q-separator />

          <q-card-section v-if="loading" class="text-center q-pa-xl">
            <q-spinner color="primary" size="3em" />
            <div class="q-mt-md">{{ t('auto.lade_verfuegbare_backups') }}</div>
          </q-card-section>

          <q-list v-else-if="backups.length > 0" separator>
            <q-item-label header>{{ t('auto.verfuegbare_backup_dateien') }}</q-item-label>
            <q-item v-for="file in backups" :key="file.path" clickable @click="selectedFile = file" :active="selectedFile?.path === file.path" active-class="bg-blue-1 text-primary">
              <q-item-section avatar>
                <q-icon name="storage" />
              </q-item-section>
              <q-item-section>
                <q-item-label class="text-weight-bold">{{ file.name }}</q-item-label>
                <q-item-label caption>{{ file.path }}</q-item-label>
              </q-item-section>
              <q-item-section side v-if="selectedFile?.path === file.path">
                <q-icon name="check_circle" color="primary" />
              </q-item-section>
            </q-item>
          </q-list>

          <q-card-section v-else class="text-center q-pa-xl text-grey-7">
            <q-icon name="folder_off" size="3em" />
            <div class="q-mt-md">{{ t('auto.keine_backup_dateien_gefunden') }}</div>
          </q-card-section>

          <q-separator />

          <q-card-actions align="right" class="q-pa-md">
            <q-btn flat :label="t('form.cancel')" color="grey-7" to="/" :disable="processing" />
            <q-btn flat :label="t('auto.liste_aktualisieren')" icon="refresh" @click="loadBackups" :disable="processing" />
            <q-btn color="negative" :label="t('auto.wiederherstellen')" icon="settings_backup_restore" @click="confirmRestore" :disable="!selectedFile || processing" :loading="processing" unelevated />
          </q-card-actions>
        </q-card>
      </div>

      <div class="col-12 col-md-4">
        <q-card flat bordered class="rounded-borders">
          <q-card-section class="bg-primary text-white">
            <div class="text-h6">{{ t('auto.informationen') }}</div>
          </q-card-section>
          <q-card-section>
            <div class="text-subtitle2 q-mb-xs">{{ t('auto.aktive_datenbank') }}</div>
            <div class="text-caption text-grey-8 word-break-all">{{ currentDB }}</div>
            
            <q-separator class="q-my-md" />
            
            <div class="text-subtitle2 q-mb-xs">{{ t('auto.hinweis') }}</div>
            <div class="text-caption">
              {{ t('auto.backups_befinden_sich_normalerweise_im_u') }} <code>{{ t('auto.backups') }}</code> {{ t('auto.des_datenbank_verzeichnisses_sollte_eine') }}
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';

const { t } = useI18n();
const $q = useQuasar();
const loading = ref(false);
const processing = ref(false);
const backups = ref<{ name: string, path: string }[]>([]);
const selectedFile = ref<{ name: string, path: string } | null>(null);
const currentDB = ref('');

async function loadBackups() {
  loading.value = true;
  try {
    const res = await api.get('/api/db/list');
    currentDB.value = res.data.current;
    
    // Wir zeigen nur Dateien an, die nicht die aktuelle DB sind und vorzugsweise aus dem Backup-Ordner kommen
    // Oder wir zeigen einfach alle an, die .db im Namen haben
    const files = res.data.files as string[];
    backups.value = files
      .filter(p => p !== currentDB.value)
      .map(p => ({
        name: p.split('/').pop() || p,
        path: p
      }))
      .sort((a, b) => b.name.localeCompare(a.name));
      
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: t('auto.error_loading_backups') + ': ' + (err.response?.data?.error || err.message)
    });
  } finally {
    loading.value = false;
  }
}

function confirmRestore() {
  if (!selectedFile.value) return;
  
  $q.dialog({
    title: t('auto.confirm_restore_title'),
    message: t('auto.confirm_restore_msg', { filename: selectedFile.value.name, current: currentDB.value }),
    html: true,
    cancel: true,
    persistent: true,
    ok: {
      label: t('auto.yes_restore'),
      color: 'negative',
      unelevated: true
    }
  }).onOk(() => {
    runRestore();
  });
}

async function runRestore() {
  if (!selectedFile.value) return;
  
  processing.value = true;
  try {
    const res = await api.post('/api/db/restore', {
      PATH: selectedFile.value.path
    });
    
    $q.notify({
      type: 'positive',
      message: t('auto.restore_success'),
      caption: t('auto.safety_backup_created') + ': ' + res.data.safety,
      timeout: 5000
    });
    
    setTimeout(() => {
      window.location.reload();
    }, 2000);
    
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: t('auto.restore_failed') + ': ' + (err.response?.data?.error || err.message),
      timeout: 0,
      actions: [{ label: 'OK', color: 'white' }]
    });
  } finally {
    processing.value = false;
  }
}

onMounted(() => {
  loadBackups();
});
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
.word-break-all {
  word-break: break-all;
}
</style>

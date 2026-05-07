<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="col">
        <h1 class="text-h4 q-my-none">Datenbank Wiederherstellung (Restore)</h1>
        <div class="text-subtitle1 text-grey-7">Wählen Sie eine Backup-Datei aus, um den aktuellen Datenbestand zu überschreiben.</div>
      </div>
    </div>

    <div class="row q-col-gutter-md">
      <div class="col-12 col-md-8">
        <q-card flat bordered class="rounded-borders">
          <q-card-section class="bg-warning text-black row items-center">
            <q-icon name="warning" size="md" class="q-mr-md" />
            <div class="text-h6 text-weight-bold">Achtung: Überschreiben von Daten</div>
          </q-card-section>
          
          <q-card-section>
            <p>
              Beim Wiederherstellen wird die aktuelle Datenbank <strong>vollständig durch das ausgewählte Backup ersetzt</strong>.
              Nicht gespeicherte Änderungen gehen verloren. Vor dem Restore wird automatisch ein Sicherheits-Backup der aktuellen Datenbank erstellt.
            </p>
          </q-card-section>

          <q-separator />

          <q-card-section v-if="loading" class="text-center q-pa-xl">
            <q-spinner color="primary" size="3em" />
            <div class="q-mt-md">Lade verfügbare Backups...</div>
          </q-card-section>

          <q-list v-else-if="backups.length > 0" separator>
            <q-item-label header>Verfügbare Backup-Dateien</q-item-label>
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
            <div class="q-mt-md">Keine Backup-Dateien gefunden.</div>
          </q-card-section>

          <q-separator />

          <q-card-actions align="right" class="q-pa-md">
            <q-btn flat label="Abbrechen" color="grey-7" to="/" :disable="processing" />
            <q-btn flat label="Liste aktualisieren" icon="refresh" @click="loadBackups" :disable="processing" />
            <q-btn color="negative" label="Wiederherstellen" icon="settings_backup_restore" @click="confirmRestore" :disable="!selectedFile || processing" :loading="processing" unelevated />
          </q-card-actions>
        </q-card>
      </div>

      <div class="col-12 col-md-4">
        <q-card flat bordered class="rounded-borders">
          <q-card-section class="bg-primary text-white">
            <div class="text-h6">Informationen</div>
          </q-card-section>
          <q-card-section>
            <div class="text-subtitle2 q-mb-xs">Aktive Datenbank:</div>
            <div class="text-caption text-grey-8 word-break-all">{{ currentDB }}</div>
            
            <q-separator class="q-my-md" />
            
            <div class="text-subtitle2 q-mb-xs">Hinweis:</div>
            <div class="text-caption">
              Backups befinden sich normalerweise im Unterordner <code>backups/</code> des Datenbank-Verzeichnisses. 
              Sollte eine Datei fehlen, prüfen Sie den Backup-Pfad in den Einstellungen oder auf dem Server.
            </div>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';

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
      .filter(p => p !== currentDB.value) // Aktuelle DB nicht als Restore-Quelle (macht keinen Sinn)
      .map(p => ({
        name: p.split('/').pop() || p,
        path: p
      }))
      .sort((a, b) => b.name.localeCompare(a.name)); // Neueste (nach Zeitstempel im Namen) meist oben
      
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: 'Fehler beim Laden der Backups: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    loading.value = false;
  }
}

function confirmRestore() {
  if (!selectedFile.value) return;
  
  $q.dialog({
    title: 'Wiederherstellung bestätigen',
    message: `Möchten Sie die Datenbank wirklich durch <b>${selectedFile.value.name}</b> ersetzen? <br><br>Dies wird die aktuelle Datenbank überschreiben!`,
    html: true,
    cancel: true,
    persistent: true,
    ok: {
      label: 'Ja, Wiederherstellen',
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
      message: 'Wiederherstellung erfolgreich abgeschlossen!',
      caption: `Sicherheitskopie erstellt: ${res.data.safety}`,
      timeout: 5000
    });
    
    // Seite neu laden nach 2 Sekunden um DB-State im Frontend zu refreshen
    setTimeout(() => {
      window.location.reload();
    }, 2000);
    
  } catch (err: any) {
    $q.notify({
      color: 'negative',
      message: 'Fehler bei der Wiederherstellung: ' + (err.response?.data?.error || err.message),
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

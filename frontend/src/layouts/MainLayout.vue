<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated class="bg-primary text-white">
      <q-toolbar>
        <q-btn dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" unelevated />

        <q-toolbar-title class="text-weight-bold" style="letter-spacing: 0.5px;">
          Huhn-Lite
        </q-toolbar-title>

        <!-- Global Date/Time Selector -->
        <div class="row items-center q-gutter-x-md q-px-md bg-white-opacity-10 rounded-borders q-ml-md">
           <q-icon name="schedule" size="xs" />
           <q-input
            v-model="sessionDate"
            type="date"
            dark
            dense
            borderless
            input-class="text-weight-bold"
            style="width: 130px"
          />
          <q-input
            v-model="sessionTime"
            type="time"
            dark
            dense
            borderless
            input-class="text-weight-bold"
            style="width: 80px"
          />
        </div>

        <q-space />

        <q-btn round dense :icon="sessionStore.darkMode ? 'dark_mode' : 'light_mode'" @click="sessionStore.setDarkMode(!sessionStore.darkMode)" aria-label="Toggle Dark Mode" unelevated />

        <div v-if="session.isLoggedIn" class="q-ml-md row items-center gt-xs">
          <q-icon name="person" size="xs" class="q-mr-xs"/>
          <span class="text-weight-bold text-grey-3">{{ session.klarname || session.username }}</span>
        </div>

        <q-btn v-if="session.isLoggedIn && session.authEnabled" flat round dense icon="logout" @click="handleLogout"
               class="q-ml-sm" unelevated>
          <q-tooltip>Abmelden</q-tooltip>
        </q-btn>
      </q-toolbar>
    </q-header>


    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
      :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-grey-1'"
    >
      <q-list padding :class="$q.dark.isActive ? 'text-grey-3' : 'text-grey-9'">
        <q-item v-if="session.can('dashboard')" clickable v-ripple to="/" exact
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar>
            <q-icon name="dashboard" />
          </q-item-section>
          <q-item-section class="text-weight-medium">Dashboard</q-item-section>
        </q-item>

        <q-separator class="q-my-md" />

        <!-- Stammdaten -->
        <q-item-label header class="text-weight-bold text-uppercase" :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">Stammdaten</q-item-label>

        <q-item v-if="session.can('herden_verwalten')" clickable v-ripple to="/herden"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="pets" /></q-item-section>
          <q-item-section>Herden & Rassen</q-item-section>
        </q-item>

        <q-item v-if="session.can('einrichtungen_verwalten')" clickable v-ripple to="/einrichtungen"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar><q-icon name="warehouse" /></q-item-section>
          <q-item-section>Einrichtungen</q-item-section>
        </q-item>


        <q-item v-if="session.can('personen_verwalten')" clickable v-ripple to="/person"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="people" /></q-item-section>
          <q-item-section>Personen</q-item-section>
        </q-item>


        <q-separator class="q-my-md" />

        <!-- Bewegungsdaten -->
        <q-item-label header class="text-weight-bold text-uppercase" :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">Bewegungsdaten</q-item-label>

        <q-item v-if="session.can('buchungen_erfassen')" clickable v-ripple to="/buchungen"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar><q-icon name="receipt_long" /></q-item-section>
          <q-item-section>Buchungen</q-item-section>
        </q-item>


        <q-item v-if="session.can('auswertungen_anzeigen')" clickable v-ripple to="/reports"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="assessment"/>
          </q-item-section>
          <q-item-section>Reports</q-item-section>
        </q-item>
        <q-separator class="q-my-md"/>

        <!-- Kosten & Einstellungen -->
        <q-item-label header class="text-weight-bold text-uppercase"
                      :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">Kosten & Einstellungen
        </q-item-label>

        <q-item v-if="session.can('parameter_editieren')" clickable v-ripple to="/settings"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="settings"/>
          </q-item-section>
          <q-item-section>Parameter</q-item-section>
        </q-item>

        <q-item v-if="session.can('kosten_verwalten')" clickable v-ripple to="/kosten"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="attach_money" /></q-item-section>
          <q-item-section>Kosten</q-item-section>
        </q-item>
        <q-item v-if="session.can('tabellen_anzeigen')" clickable v-ripple to="/mist"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="table_chart"/>
          </q-item-section>
          <q-item-section>Tabellen</q-item-section>
        </q-item>
        <q-item v-if="session.can('texte_verwalten')" clickable v-ripple to="/textverwaltung"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="description" /></q-item-section>
          <q-item-section>Textverwaltung</q-item-section>
        </q-item>

        <q-separator class="q-my-md" />

        <q-expansion-item
          v-if="session.can('system_verwaltung')"
          icon="admin_panel_settings"
          label="Systemverwaltung"
          header-class="text-weight-bold"
          expand-separator
          :content-inset-level="0.5"
          toggle-aria-label="Systemverwaltung öffnen"
        >

          <q-item v-if="session.can('benutzer_profile')" clickable v-ripple to="/benutzer"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

            <q-item-section avatar>
              <q-icon name="manage_accounts"/>
            </q-item-section>
            <q-item-section>Benutzerverwaltung</q-item-section>
          </q-item>

          <q-item v-if="session.can('benutzer_profile')" clickable v-ripple to="/profile"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="security"/>
            </q-item-section>
            <q-item-section>Profilverwaltung</q-item-section>
          </q-item>



          <q-item v-if="session.can('sql_struktur_verwalten')" clickable v-ripple to="/admin"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="storage"/>
            </q-item-section>
            <q-item-section>Wartung / Stunden-Shift</q-item-section>
          </q-item>

          <q-item v-if="session.permissions.backup_erstellen" clickable v-ripple @click="runBackup"
                  :disable="backupLoading">
            <q-item-section avatar>
              <q-icon v-if="!backupLoading" name="backup"/>
              <q-spinner v-else color="primary" size="2em"/>
            </q-item-section>
            <q-item-section>Backup erstellen</q-item-section>
          </q-item>

            <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/showtv"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="visibility"/>
            </q-item-section>
            <q-item-section>Anzeige Steuerung</q-item-section>
          </q-item>

          <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/restore"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="settings_backup_restore"/>
            </q-item-section>
            <q-item-section>Restore (Wiederherstellung)</q-item-section>
          </q-item>
        </q-expansion-item>

      </q-list>

      <div class="q-pa-md q-mt-auto">
        <q-separator class="q-my-md" />
        <q-item clickable v-ripple @click="shutdownServer" class="text-negative rounded-borders shadow-1 bg-white-opacity-10">
          <q-item-section avatar>
            <q-icon name="power_settings_new" color="negative" />
          </q-item-section>
          <q-item-section class="text-weight-bold">Programm beenden</q-item-section>
        </q-item>
      </div>
    </q-drawer>

    <q-page-container :class="$q.dark.isActive ? 'bg-dark' : 'bg-blue-grey-1'">
      <router-view />
    </q-page-container>

    <!-- Login Modal - Blocks UI if auth enabled -->
    <LoginModal/>
  </q-layout>
</template>


<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import {api} from 'src/boot/api';
import {onMounted} from 'vue';

import { useSessionStore } from '../stores/session';
import LoginModal from '../components/LoginModal.vue';


const $q = useQuasar();
const sessionStore = useSessionStore();
const session = sessionStore;
const leftDrawerOpen = ref(false);
const backupLoading = ref(false);

async function runBackup() {
  backupLoading.value = true;
  try {
    const res = await api.post('/api/backup');
    $q.notify({
      type: 'positive',
      message: `Backup ${String((res.data as { filename: string }).filename)} erstellt!`,
      caption: `Pfad: ${String((res.data as { path: string }).path)}`,
      icon: 'check_circle'
    });
  } catch (err: unknown) {
    const errorMsg = (err as {
      response?: { data?: { error?: string } };
      message?: string
    })?.response?.data?.error || (err as { message?: string })?.message || 'Unbekannter Fehler';
    $q.notify({
      type: 'negative',
      message: 'Fehler: ' + errorMsg
    });
  } finally {
    backupLoading.value = false;
  }
}

async function shutdownServer() {
  $q.dialog({
    title: 'Programm beenden',
    message: 'Möchten Sie die Anwendung wirklich schließen? Der Hintergrund-Dienst wird beendet.',
    cancel: {
      label: 'Abbrechen',
      flat: true
    },
    ok: {
      label: 'Beenden',
      color: 'negative',
      unelevated: true
    },
    persistent: true
  }).onOk(async () => {
    try {
      // Fenstergröße speichern
      if (window.go && window.go.main && window.go.main.App) {
        await window.go.main.App.SaveWindowState(sessionStore.username || 'default');
      }

      // Backend informieren (für Logs etc.)
      await api.post('/api/system/shutdown');
      
      $q.notify({
        type: 'warning',
        message: 'Anwendung wird geschlossen...',
        position: 'center',
        timeout: 1000
      });

      // Sauber über Wails beenden
      if (window.go && window.go.main && window.go.main.App) {
        setTimeout(() => {
          window.go.main.App.Quit();
        }, 500);
      } else {
        window.close();
      }
    } catch (err) {
      console.error('Shutdown error:', err);
      if (window.go && window.go.main && window.go.main.App) {
        window.go.main.App.Quit();
      } else {
        window.close();
      }
    }
  });
}

const sessionDate = computed({
  get: () => sessionStore.workingTimestamp.split(' ')[0] || '',
  set: (val) => {
    const time = sessionStore.workingTimestamp.split(' ')[1] || '12:00';
    sessionStore.workingTimestamp = `${val} ${time}`;
  }
});

const sessionTime = computed({
  get: () => (sessionStore.workingTimestamp.split(' ')[1] || '').substring(0, 5),
  set: (val) => {
    const date = sessionStore.workingTimestamp.split(' ')[0] || '';
    sessionStore.workingTimestamp = `${date} ${val}`;
  }
});

function toggleLeftDrawer () {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}
const handleLogout = () => {
  sessionStore.logout();
};

onMounted(async () => {
  console.log('MainLayout: Starte Initialisierung...');
  
  let success = false;
  let retries = 5;

  while (!success && retries > 0) {
    try {
      const res = await api.get('/api/system-settings/auth_required');
      const val = res.data.value;
      sessionStore.authEnabled = (val === true || val === 'true' || val === '1');
      console.log('MainLayout: Auth-Status geladen:', sessionStore.authEnabled);
      
      if (!sessionStore.authEnabled) {
        sessionStore.setAdminSession();
      }
      success = true;
    } catch (err) {
      console.log(`MainLayout: Server noch nicht bereit, Retry (${retries})...`);
      retries--;
      if (retries > 0) await new Promise(resolve => setTimeout(resolve, 500));
    }
  }

  // Fallback falls alles fehlschlägt
  if (!success) {
    console.warn('MainLayout: Konnte Auth-Status nicht laden, nutze Default.');
    // Wenn wir gar keine Antwort bekommen, loggen wir sicherheitshalber ein,
    // damit der User nicht vor einem leeren Fenster steht.
    sessionStore.setAdminSession();
  }
});

</script>

<style scoped>
.bg-white-opacity-10 {
  background: rgba(255, 255, 255, 0.15);
}
.rounded-borders {
  border-radius: 8px;
}
</style>

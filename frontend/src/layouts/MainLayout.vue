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

      // Zuerst dem Backend sagen, dass es sich beenden soll
      await api.post('/api/system/shutdown');
      
      $q.notify({
        type: 'warning',
        message: 'Anwendung wird geschlossen...',
        position: 'center',
        timeout: 2000
      });

      // Kurz warten, dann versuchen das Fenster zu schließen
      setTimeout(() => {
        window.close();
        // Fallback falls window.close() blockiert wird
        document.body.innerHTML = '<div style="background:#111; color:#eee; height:100vh; display:flex; flex-direction:column; align-items:center; justify-content:center; font-family:sans-serif;"><h1>HuhnLite wurde beendet</h1><p>Sie können dieses Fenster nun schließen.</p></div>';
      }, 1000);
    } catch (err) {
      console.error('Shutdown error:', err);
      // Wenn der Server schon weg ist oder ein Fehler kam
      window.close();
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
  try {
    // 1. Zuerst den DB-Parameter prüfen (Priorität)
    console.log('MainLayout: Prüfe auth_required Parameter in DB...');
    const res = await api.get('/api/system-settings/auth_required');
    const dbVal = String(res.data.value || '').toLowerCase();
    sessionStore.authEnabled = (dbVal === 'true' || dbVal === '1');
    console.log('MainLayout: DB auth_required =', sessionStore.authEnabled);
    if (!sessionStore.authEnabled) {
      sessionStore.setAdminSession();
    }
  } catch (_err: unknown) {
    // 2. Fallback auf .env / Config, falls DB-Wert (noch) nicht da
    console.log('MainLayout: DB Parameter nicht gefunden, nutze /api/config Fallback');
    try {
      const configRes = await api.get('/api/config');
      sessionStore.authEnabled = (configRes.data as { auth_enabled: boolean }).auth_enabled;
      console.log('MainLayout: Config auth_enabled =', sessionStore.authEnabled);
      if (!sessionStore.authEnabled) {
        sessionStore.setAdminSession();
      }
    } catch (err2: unknown) {
      console.error('Config-Fehler:', err2);
      sessionStore.authEnabled = true; // Sicherer Default
    }
  }

  // Wenn deaktiviert -> Auto-Login als Admin
  if (!sessionStore.authEnabled) {
    console.log('MainLayout: Auto-Login als Test-Admin...');
    sessionStore.setAdminSession();
  } else {
    console.log('MainLayout: Login erforderlich.');
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

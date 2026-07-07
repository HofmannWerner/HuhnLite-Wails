<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated class="bg-primary text-white">
      <q-toolbar>
        <q-btn dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" unelevated />

        <q-toolbar-title class="text-weight-bold" style="letter-spacing: 0.5px;">
          Huhn-Lite
        </q-toolbar-title>

        <!-- Test Mode Toggle -->
        <q-checkbox
          v-model="isTestDb"
          label="Test"
          color="amber-9"
          dark
          class="q-ml-sm text-weight-bold text-amber-9"
          @update:model-value="onToggleTestDb"
        >
          <q-tooltip>Datenbank umschalten (Haupt- vs. Test-Datenbank)</q-tooltip>
        </q-checkbox>

        <!-- Global Date/Time Selector -->
        <div class="row items-center q-gutter-x-md q-px-md bg-white rounded-borders q-ml-md date-time-selector">
           <q-icon name="schedule" size="xs" class="text-black" />
           <q-input
            v-model="sessionDate"
            type="date"
            dense
            borderless
            readonly
            input-class="text-weight-bold text-black"
            style="width: 130px; background: transparent;"
          />
          <q-input
            v-model="sessionTime"
            type="time"
            dense
            borderless
            readonly
            input-class="text-weight-bold text-black"
            style="width: 80px; background: transparent;"
          />
        </div>

        <q-space />

        <q-btn round dense :icon="sessionStore.darkMode ? 'dark_mode' : 'light_mode'" @click="sessionStore.setDarkMode(!sessionStore.darkMode)" aria-label="Toggle Dark Mode" unelevated />

        <q-btn flat round dense class="q-ml-sm" aria-label="Language">
          <q-avatar size="24px" square>
            <img :src="flags[sessionStore.selectedLanguage as 'de' | 'en' | 'it']" style="border: 1px solid rgba(255,255,255,0.2); object-fit: cover;" />
          </q-avatar>
          <q-menu auto-close>
            <q-list style="min-width: 120px">
              <q-item clickable :active="sessionStore.selectedLanguage === 'de'" @click="sessionStore.setLanguage('de')">
                <q-item-section avatar class="q-pr-none" style="min-width: auto">
                  <img :src="flags.de" style="width: 24px; height: 16px; border: 1px solid rgba(0,0,0,0.1); object-fit: cover;" />
                </q-item-section>
                <q-item-section class="q-pl-sm">Deutsch</q-item-section>
              </q-item>
              <q-item clickable :active="sessionStore.selectedLanguage === 'en'" @click="sessionStore.setLanguage('en')">
                <q-item-section avatar class="q-pr-none" style="min-width: auto">
                  <img :src="flags.en" style="width: 24px; height: 16px; border: 1px solid rgba(0,0,0,0.1); object-fit: cover;" />
                </q-item-section>
                <q-item-section class="q-pl-sm">English</q-item-section>
              </q-item>
              <q-item clickable :active="sessionStore.selectedLanguage === 'it'" @click="sessionStore.setLanguage('it')">
                <q-item-section avatar class="q-pr-none" style="min-width: auto">
                  <img :src="flags.it" style="width: 24px; height: 16px; border: 1px solid rgba(0,0,0,0.1); object-fit: cover;" />
                </q-item-section>
                <q-item-section class="q-pl-sm">Italiano</q-item-section>
              </q-item>
            </q-list>
          </q-menu>
        </q-btn>

        <div v-if="session.isLoggedIn" class="q-ml-md row items-center gt-xs">
          <q-icon name="person" size="xs" class="q-mr-xs"/>
          <span class="text-weight-bold text-grey-3">{{ session.klarname || session.username }}</span>
        </div>

        <q-btn v-if="session.isLoggedIn && session.authEnabled" flat round dense icon="logout" @click="handleLogout"
               class="q-ml-sm" unelevated>
          <q-tooltip>{{ t('layout.logout') }}</q-tooltip>
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
          <q-item-section class="text-weight-medium">{{ t('menu.dashboard') }}</q-item-section>
        </q-item>

        <q-separator class="q-my-md" />

        <!-- Stammdaten -->
        <q-item-label header class="text-weight-bold text-uppercase" :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">{{ t('menu.stammdaten') }}</q-item-label>

        <q-item v-if="session.can('herden_verwalten')" clickable v-ripple to="/herden"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="pets" /></q-item-section>
          <q-item-section>{{ t('menu.herden') }}</q-item-section>
        </q-item>

        <q-item v-if="session.can('einrichtungen_verwalten')" clickable v-ripple to="/einrichtungen"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar><q-icon name="warehouse" /></q-item-section>
          <q-item-section>{{ t('menu.einrichtungen') }}</q-item-section>
        </q-item>


        <q-item v-if="session.can('personen_verwalten')" clickable v-ripple to="/person"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="people" /></q-item-section>
          <q-item-section>{{ t('menu.personen') }}</q-item-section>
        </q-item>


        <q-separator class="q-my-md" />

        <!-- Bewegungsdaten -->
        <q-item-label header class="text-weight-bold text-uppercase" :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">{{ t('menu.bewegungsdaten') }}</q-item-label>

        <q-item v-if="session.can('buchungen_erfassen')" clickable v-ripple to="/buchungen"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar><q-icon name="receipt_long" /></q-item-section>
          <q-item-section>{{ t('menu.buchungen') }}</q-item-section>
        </q-item>


        <q-item v-if="session.can('auswertungen_anzeigen')" clickable v-ripple to="/reports"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="assessment"/>
          </q-item-section>
          <q-item-section>{{ t('menu.reports') }}</q-item-section>
        </q-item>
        <q-separator class="q-my-md"/>

        <!-- Kosten & Einstellungen -->
        <q-item-label header class="text-weight-bold text-uppercase"
                      :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">{{ t('menu.kosten_einstellungen') }}
        </q-item-label>

        <q-item v-if="session.can('parameter_editieren')" clickable v-ripple to="/settings"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="settings"/>
          </q-item-section>
          <q-item-section>{{ t('menu.parameter') }}</q-item-section>
        </q-item>

        <q-item v-if="session.can('kosten_verwalten')" clickable v-ripple to="/kosten"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="attach_money" /></q-item-section>
          <q-item-section>{{ t('menu.kosten') }}</q-item-section>
        </q-item>
        <q-item v-if="session.can('tabellen_anzeigen')" clickable v-ripple to="/mist"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

          <q-item-section avatar>
            <q-icon name="table_chart"/>
          </q-item-section>
          <q-item-section>{{ t('menu.tabellen') }}</q-item-section>
        </q-item>
        <q-item v-if="session.can('texte_verwalten')" clickable v-ripple to="/textverwaltung"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">

        <q-item-section avatar><q-icon name="description" /></q-item-section>
          <q-item-section>{{ t('menu.textverwaltung') }}</q-item-section>
        </q-item>

        <q-separator class="q-my-md" />

        <q-expansion-item
          v-if="session.can('system_verwaltung')"
          icon="admin_panel_settings"
          :label="t('menu.systemverwaltung')"
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
            <q-item-section>{{ t('menu.benutzerverwaltung') }}</q-item-section>
          </q-item>

          <q-item v-if="session.can('benutzer_profile')" clickable v-ripple to="/profile"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="security"/>
            </q-item-section>
            <q-item-section>{{ t('menu.profilverwaltung') }}</q-item-section>
          </q-item>



          <q-item v-if="session.can('sql_struktur_verwalten') && isTestDb" clickable v-ripple to="/admin"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="storage"/>
            </q-item-section>
            <q-item-section>{{ t('menu.wartung') }}</q-item-section>
          </q-item>

          <q-item v-if="session.permissions.backup_erstellen" clickable v-ripple @click="runBackup"
                  :disable="backupLoading">
            <q-item-section avatar>
              <q-icon v-if="!backupLoading" name="backup"/>
              <q-spinner v-else color="primary" size="2em"/>
            </q-item-section>
            <q-item-section>{{ t('menu.backup') }}</q-item-section>
          </q-item>

            <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/showtv"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="visibility"/>
            </q-item-section>
            <q-item-section>{{ t('menu.anzeige') }}</q-item-section>
          </q-item>

          <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/restore"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="settings_backup_restore"/>
            </q-item-section>
            <q-item-section>{{ t('menu.restore') }}</q-item-section>
          </q-item>
        </q-expansion-item>

        <q-separator class="q-my-md" />

        <q-item clickable v-ripple @click="openHelp"
                :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
          <q-item-section avatar>
            <q-icon name="help_outline"/>
          </q-item-section>
          <q-item-section>{{ t('menu.hilfe') }}</q-item-section>
        </q-item>
      </q-list>


      <div class="q-pa-md q-mt-auto">
        <q-separator class="q-my-md" />
        <q-item clickable v-ripple @click="shutdownServer" class="text-negative rounded-borders shadow-1 bg-white-opacity-10">
          <q-item-section avatar>
            <q-icon name="power_settings_new" color="negative" />
          </q-item-section>
          <q-item-section class="text-weight-bold">{{ t('menu.beenden') }}</q-item-section>
        </q-item>
      </div>
    </q-drawer>

    <q-page-container :class="$q.dark.isActive ? 'bg-dark' : 'bg-blue-grey-1'">
      <router-view />
    </q-page-container>

    <!-- Login Modal - Blocks UI if auth enabled -->
    <LoginModal/>

    <!-- Help Modal -->
    <q-dialog v-model="helpDialogOpen" maximized transition-show="slide-up" transition-hide="slide-down">
      <q-card class="column no-wrap" style="height: 100vh;">
        <q-bar class="bg-primary text-white q-py-md">
          <q-icon name="help_outline" />
          <div class="text-weight-bold">{{ t('layout.helpTitle') }}</div>
          <q-space />
          <q-btn dense flat icon="close" v-close-popup>
            <q-tooltip>{{ t('layout.close') }}</q-tooltip>
          </q-btn>
        </q-bar>

        <q-card-section class="col q-pa-none relative-position bg-white">
          <div v-if="helpLoading" class="absolute-center text-center">
            <q-spinner-gears size="50px" color="primary" />
            <div class="q-mt-md text-grey-7">{{ t('layout.loadingHelp') }}</div>
          </div>
          <iframe
            v-else-if="helpUrl"
            :src="helpUrl"
            style="width: 100%; height: 100%; border: none;"
          ></iframe>
        </q-card-section>
      </q-card>
    </q-dialog>
  </q-layout>
</template>


<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar, date } from 'quasar';
import { api } from 'src/boot/api';
import { useI18n } from 'vue-i18n';

import { useSessionStore } from '../stores/session';
import LoginModal from '../components/LoginModal.vue';

const flags = {
  de: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 5 3"><rect width="5" height="3" fill="%23ffce00"/><rect width="5" height="2" fill="%23dd0000"/><rect width="5" height="1" fill="%23000"/></svg>',
  en: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 60 30"><rect width="60" height="30" fill="%2300247d"/><path d="M0,0 L60,30 M60,0 L0,30" stroke="%23fff" stroke-width="6"/><path d="M0,0 L60,30 L60,0 L0,30" stroke="%23c8102e" stroke-width="4"/><path d="M30,0 v30 M0,15 h60" stroke="%23fff" stroke-width="10"/><path d="M30,0 v30 M0,15 h60" stroke="%23c8102e" stroke-width="6"/></svg>',
  it: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 3 2"><rect width="1" height="2" fill="%23009246"/><rect x="1" width="1" height="2" fill="%23fff"/><rect x="2" width="1" height="2" fill="%23ce2b37"/></svg>'
};


const $q = useQuasar();
const router = useRouter();
const sessionStore = useSessionStore();
const session = sessionStore;
const { t, locale } = useI18n({ useScope: 'global' });

locale.value = sessionStore.selectedLanguage || 'de';
watch(() => sessionStore.selectedLanguage, (val) => {
  locale.value = val || 'de';
});

const leftDrawerOpen = ref(false);
const backupLoading = ref(false);

const helpDialogOpen = ref(false);
const helpLoading = ref(false);
const helpUrl = ref('');

const isTestDb = ref(false);

async function checkTestDbStatus() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const active = await window.go.main.App.IsTestDB();
      isTestDb.value = active;
    } catch (err) {
      console.error('Failed to get test DB status:', err);
    }
  }
}

async function onToggleTestDb(val: any) {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      const newDsn = await window.go.main.App.ToggleTestDB(val);
      $q.notify({
        type: val ? 'warning' : 'positive',
        message: val ? 'Umgeschaltet auf Test-Datenbank' : 'Umgeschaltet auf Haupt-Datenbank',
        caption: `Verbindung: ${newDsn}`,
        icon: val ? 'warning' : 'check_circle',
        timeout: 1000
      });
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } catch (err: any) {
      $q.notify({
        type: 'negative',
        message: 'Fehler beim Umschalten der Datenbank',
        caption: String(err?.message || err || 'Unbekannter Fehler'),
        icon: 'error',
        timeout: 5000
      });
      isTestDb.value = !val; // Rollback
    }
  } else {
    $q.notify({
      type: 'info',
      message: 'Datenbank-Umschaltung ist im Webbrowser nicht verfügbar.',
      icon: 'info'
    });
    isTestDb.value = !val; // Rollback
  }
}

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
    title: t('layout.shutdownTitle'),
    message: t('layout.shutdownMessage'),
    cancel: {
      label: t('layout.shutdownCancel'),
      flat: true
    },
    ok: {
      label: t('layout.shutdownConfirm'),
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
        message: t('layout.closing'),
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

async function openHelp() {
  helpDialogOpen.value = true;
  helpLoading.value = true;
  helpUrl.value = '';

  const activeLang = sessionStore.selectedLanguage || 'de';

  if (window.go && window.go.main && window.go.main.App) {
    try {
      // Retrieve file content directly via Wails
      const htmlContent = await window.go.main.App.GetHelpContent(activeLang);

      // Inject <base> tag to resolve relative image paths against the static files server
      const baseApiUrl = api.defaults.baseURL || 'http://localhost:8080';
      const baseTag = `<base href="${baseApiUrl}/help/">`;

      let modifiedHtml = htmlContent;
      if (htmlContent.includes('<head>')) {
        modifiedHtml = htmlContent.replace('<head>', `<head>${baseTag}`);
      } else if (htmlContent.includes('<HEAD>')) {
        modifiedHtml = htmlContent.replace('<HEAD>', `<HEAD>${baseTag}`);
      } else {
        modifiedHtml = `${baseTag}${htmlContent}`;
      }

      // Replace anchor links to prevent them from navigating the iframe via the base URL
      modifiedHtml = modifiedHtml.replace(/href="#([^"]+)"/g, 'href="javascript:void(0);" onclick="const el = document.getElementById(\'$1\'); if (el) el.scrollIntoView({behavior: \'smooth\'});"');

      helpUrl.value = 'data:text/html;charset=utf-8,' + encodeURIComponent(modifiedHtml);
    } catch (err) {
      console.error('Error loading help content:', err);
      // Fallback: Try to open natively
      const errMsg = await window.go.main.App.OpenHelp(activeLang);
      if (errMsg) {
        $q.notify({
          type: 'negative',
          message: 'Hilfe konnte nicht geladen werden: ' + String(errMsg),
          position: 'top',
          timeout: 5000
        });
      }
      // Close the in-app help dialog since it's blank or opened natively
      helpDialogOpen.value = false;
    } finally {
      helpLoading.value = false;
    }
  } else {
    // Im Webbrowser: Versuche, das echte Hilfe-Dokument vom Server zu laden
    const baseApiUrl = api.defaults.baseURL || window.location.origin;
    const fileUrl = `${baseApiUrl}/help/HuhnLite-${activeLang}.html`;

    try {
      // Prüfen, ob die Hilfedatei auf dem Server existiert
      await api.get(`/help/HuhnLite-${activeLang}.html`);
      // Falls sie existiert, im Iframe laden
      helpUrl.value = fileUrl;
    } catch (err) {
      console.warn('Help file not found on server, showing mockup:', err);
      // Fallback: Lokaler Entwicklungs-Mockup
      const mockHtml = `
        <html>
          <head>
            <style>
              body { font-family: sans-serif; padding: 20px; color: #333; line-height: 1.6; }
              h1 { color: #027be3; }
            </style>
          </head>
          <body>
            <h1>Hilfe-Dokument (Entwicklungsmodus - ${activeLang})</h1>
            <p>Die Wails-Laufzeitumgebung ist nicht verfügbar. Im Live-System wird hier das Handbuch geladen.</p>
          </body>
        </html>
      `;
      helpUrl.value = 'data:text/html;charset=utf-8,' + encodeURIComponent(mockHtml);
    } finally {
      helpLoading.value = false;
    }
  }
}

onMounted(async () => {
  console.log('MainLayout: Starte Initialisierung...');
  await checkTestDbStatus();
  
  let success = false;
  let retries = 10;

  while (!success && retries > 0) {
    try {
      const res = await api.get('/api/config');
      sessionStore.authEnabled = res.data.auth_enabled;
      sessionStore.systemEditEnabled = res.data.system_edit_enabled;
      console.log('MainLayout: Config geladen:', { auth: sessionStore.authEnabled, system: sessionStore.systemEditEnabled });
      
      if (!sessionStore.authEnabled) {
        sessionStore.setAdminSession();
      }
      success = true;
    } catch (err) {
      console.log(`MainLayout: Server noch nicht bereit, Retry (${retries})...`);
      retries--;
      if (retries > 0) await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }

  // Fallback falls alles fehlschlägt
  if (!success) {
    console.warn('MainLayout: Konnte Auth-Status nicht laden, nutze Default.');
    sessionStore.setAdminSession();
  }
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

let aktionenChecked = false;

async function checkAktionen() {
  if (aktionenChecked) return;
  aktionenChecked = true;

  try {
    const todayStr = date.formatDate(new Date(), 'YYYY-MM-DD');
    const params = {
      kz: 'B',
      show_erledigt: 0, // Nur offene
      // Wir lassen 'start' leer, um auch alle vergangenen offenen Aktionen zu sehen
      end: todayStr,
      id_user: 0 // Alle abrufen
    };
    
    console.log('[DEBUG] MainLayout checkAktionen params:', params);
    const res = await api.get('/api/aktionen', { params });
    const allActions = res.data || [];
    
    let relevantActions = [];
    const currentUserId = sessionStore.userId || 0;
    const isAdmin = sessionStore.profile_kz === 'A';
    
    if (sessionStore.authEnabled && sessionStore.isLoggedIn) {
      if (isAdmin) {
        // Administratoren sehen alle offenen Aktionen
        relevantActions = allActions;
      } else {
        // Reguläre Benutzer sehen eigene + allgemeine (ID 0) Aktionen
        relevantActions = allActions.filter((a: any) => {
          const actionUserId = extractInt(a.id_user);
          return actionUserId === currentUserId || actionUserId === 0;
        });
      }
    } else {
      // Wenn keine Anmeldung erforderlich, alle Typ B Aktionen anzeigen
      relevantActions = allActions;
    }
    
    if (relevantActions.length > 0) {
      let oldestDateStr = todayStr;
      relevantActions.forEach((a: any) => {
        const d = extractString(a.aktionsdatum);
        if (d && d < oldestDateStr) {
          oldestDateStr = d;
        }
      });

      $q.dialog({
        title: t('message.openActions'),
        message: t('message.openActionsMsg').replace('{count}', String(relevantActions.length)),
        ok: {
          label: t('message.show'),
          color: 'primary',
          unelevated: true
        },
        cancel: {
          label: t('message.later'),
          flat: true,
          color: 'grey-7'
        },
        persistent: true
      }).onOk(() => {
        router.push({ 
          path: '/buchungen', 
          query: { 
            tab: 'aktionen',
            filterKz: 'B',
            filterStartDate: oldestDateStr,
            filterEndDate: todayStr,
            filterUser: currentUserId // Zeige Aktionen für diesen User (oder alle Typ B für heute)
          } 
        });
      });
    }
  } catch (err) {
    console.error('MainLayout: Fehler beim Prüfen der Aktionen:', err);
  }
}

// Watch für Login
watch(() => sessionStore.isLoggedIn, (newVal) => {
  if (newVal) {
    checkAktionen();
  } else {
    aktionenChecked = false;
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
:deep(input[type="date"]),
:deep(input[type="time"]),
:deep(input[readonly]),
:deep(input:read-only),
:deep(.q-field--readonly .q-field__control),
:deep(.q-field--readonly .q-field__control::before),
:deep(.q-field--readonly .q-field__control::after) {
  background: transparent !important;
  background-color: transparent !important;
  border: none !important;
  -webkit-appearance: none;
  opacity: 1 !important;
}
.date-time-selector :deep(input) {
  color: #000000 !important;
}
.date-time-selector {
  color: #000000 !important;
}
</style>

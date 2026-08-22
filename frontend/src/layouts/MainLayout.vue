<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated class="bg-primary text-white">
      <q-toolbar>
        <q-btn dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" unelevated />

        <q-toolbar-title class="text-weight-bold" style="letter-spacing: 0.5px;">
          Huhn-Lite
        </q-toolbar-title>

        <!-- Active Tenant Info -->
        <div v-if="activeTenantName" class="row items-center q-ml-md q-px-sm bg-white-opacity-10 rounded-borders text-white" style="background: rgba(255, 255, 255, 0.1); padding: 4px 8px; border-radius: 4px;">
          <q-icon name="business" class="q-mr-xs text-amber-9" />
          <span class="text-weight-bold text-caption">{{ activeTenantName }}</span>
          <q-badge v-if="isTestDb" color="negative" class="q-ml-sm text-weight-bold text-caption">TEST</q-badge>
          <q-badge v-else color="positive" class="q-ml-sm text-weight-bold text-caption">PROD</q-badge>
        </div>

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

        <!-- Broadcast Message Button (Server Mode) -->
        <q-btn
          v-if="isServerMode"
          round
          dense
          flat
          icon="campaign"
          color="warning"
          class="q-ml-sm"
          @click="showBroadcastSendModal = true"
        >
          <q-tooltip>Rundnachricht an alle aktiven Benutzer senden</q-tooltip>
        </q-btn>


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

          <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/restore"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="settings_backup_restore"/>
            </q-item-section>
            <q-item-section>{{ t('menu.sicherungsverwaltung') }}</q-item-section>
          </q-item>

          <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/showtv"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="visibility"/>
            </q-item-section>
            <q-item-section>{{ t('menu.anzeige') }}</q-item-section>
          </q-item>

          <q-item v-if="session.can('system_verwaltung')" clickable v-ripple to="/mandanten"
                  :active-class="$q.dark.isActive ? 'text-primary bg-grey-9' : 'text-primary bg-blue-1'">
            <q-item-section avatar>
              <q-icon name="business"/>
            </q-item-section>
            <q-item-section>{{ t('menu.mandanten') }}</q-item-section>
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
          <q-btn
            flat
            dense
            color="white"
            icon="open_in_new"
            label="Im System-Viewer öffnen"
            @click="openHelpNatively"
            class="q-mr-md text-weight-medium"
          >
            <q-tooltip>Öffnet die PDF-Datei in Ihrem Standard-PDF-Programm</q-tooltip>
          </q-btn>

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

    <!-- Send Broadcast Dialog -->
    <q-dialog v-model="showBroadcastSendModal" persistent>
      <q-card style="min-width: 450px;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold row items-center">
            <q-icon name="campaign" color="warning" class="q-mr-sm" size="md" />
            Rundnachricht an alle aktiven User
          </div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <div class="text-subtitle2 q-mb-xs">Meldungstyp:</div>
          <div class="row q-gutter-sm q-mb-md">
            <q-btn
              :color="broadcastType === 'info' ? 'primary' : 'grey-4'"
              :text-color="broadcastType === 'info' ? 'white' : 'black'"
              icon="info"
              label="Information"
              @click="broadcastType = 'info'"
              unelevated
            />
            <q-btn
              :color="broadcastType === 'warning' ? 'negative' : 'grey-4'"
              :text-color="broadcastType === 'warning' ? 'white' : 'black'"
              icon="warning"
              label="Warnung"
              @click="broadcastType = 'warning'"
              unelevated
            />
          </div>

          <q-input
            v-model="broadcastText"
            type="textarea"
            rows="3"
            label="Ihre Nachricht an alle angemeldeten Benutzer"
            placeholder="z.B. Wartungsarbeiten am Server beginnen in 10 Minuten..."
            outlined
            autofocus
          />
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Nachricht senden"
            :color="broadcastType === 'warning' ? 'negative' : 'primary'"
            icon="send"
            @click="sendBroadcastMessage"
            :loading="sendingBroadcast"
            :disabled="!broadcastText.trim()"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Received Broadcast Popup Dialog -->
    <q-dialog v-model="showReceivedBroadcastModal" persistent>
      <q-card style="min-width: 450px;" class="q-pa-sm shadow-10">
        <q-card-section
          :class="receivedBroadcast?.type === 'warning' ? 'bg-negative text-white' : 'bg-primary text-white'"
          class="row items-center"
        >
          <q-icon
            :name="receivedBroadcast?.type === 'warning' ? 'warning' : 'info'"
            size="md"
            class="q-mr-sm"
          />
          <div class="text-h6 text-weight-bold">
            {{ receivedBroadcast?.type === 'warning' ? 'Wichtige Warnung' : 'System-Information' }}
          </div>
        </q-card-section>

        <q-card-section class="q-pt-lg q-pb-md">
          <div class="text-body1 text-weight-bold q-mb-sm" style="white-space: pre-wrap;">
            {{ receivedBroadcast?.message }}
          </div>
          <div class="text-caption text-grey-7 q-mt-md">
            Absender: {{ receivedBroadcast?.sender || 'System' }} <span v-if="receivedBroadcast?.timestamp">| {{ formatBroadcastTime(receivedBroadcast?.timestamp) }}</span>
          </div>
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn
            label="Verstanden / Schließen"
            :color="receivedBroadcast?.type === 'warning' ? 'negative' : 'primary'"
            v-close-popup
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-layout>
</template>


<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
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
const activeTenantName = ref('');

interface BroadcastMsg {
  id: string;
  type: string;
  message: string;
  sender: string;
  timestamp: string;
}

const isServerMode = ref(false);
const showBroadcastSendModal = ref(false);
const broadcastType = ref<'info' | 'warning'>('info');
const broadcastText = ref('');
const sendingBroadcast = ref(false);

const showReceivedBroadcastModal = ref(false);
const receivedBroadcast = ref<BroadcastMsg | null>(null);
const acknowledgedBroadcastId = ref<string>('');

function formatBroadcastTime(ts?: string) {
  if (!ts) return '';
  try {
    const d = new Date(ts);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  } catch {
    return ts;
  }
}

async function sendBroadcastMessage() {
  if (!broadcastText.value.trim()) return;
  sendingBroadcast.value = true;
  try {
    const senderName = sessionStore.klarname || sessionStore.username || 'Administrator';
    await api.post('/api/system/broadcast', {
      type: broadcastType.value,
      message: broadcastText.value.trim(),
      sender: senderName
    });
    $q.notify({
      type: 'positive',
      message: 'Rundnachricht erfolgreich an alle aktiven Benutzer gesendet.'
    });
    showBroadcastSendModal.value = false;
    broadcastText.value = '';
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Senden: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    sendingBroadcast.value = false;
  }
}


async function fetchActiveTenant() {
  try {
    const res = await api.get('/api/tenants');
    const activeId = res.data.active_mandant;
    const active = res.data.tenants?.find((t: any) => t.id === activeId);
    if (active) {
      activeTenantName.value = active.name;
    } else {
      activeTenantName.value = '';
    }
  } catch (err) {
    console.error('Failed to fetch active tenant:', err);
  }
}

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

      if (heartbeatTimer) {
        clearInterval(heartbeatTimer);
      }

      $q.notify({
        type: 'warning',
        message: t('layout.closing'),
        position: 'center',
        timeout: 1000
      });

      await exitSessionAndRedirect();
    } catch (err) {
      console.error('Shutdown error:', err);
      if (window.go && window.go.main && window.go.main.App) {
        window.go.main.App.Quit();
      } else {
        const targetUrl = formatBasePortUrl(9000);
        window.location.href = targetUrl;
      }
    }
  });
}

async function exitSessionAndRedirect() {
  let launcherPort = 9000;
  try {
    const shutdownRes = await api.post('/api/system/shutdown');
    if (shutdownRes && shutdownRes.data && shutdownRes.data.port) {
      launcherPort = shutdownRes.data.port;
    }
  } catch (e) {
    // Safe fallback if server endpoint responds slowly or missing
  }

  if (launcherPort === 8080 || launcherPort === 9000) {
    try {
      const protocol = window.location.protocol || 'http:';
      const hostname = window.location.hostname || 'localhost';
      const portPart = window.location.port ? `:${window.location.port}` : '';
      const res = await fetch(`${protocol}//${hostname}${portPart}/api/launcher-port`);
      if (res.ok) {
        const data = await res.json();
        if (data && data.port) {
          launcherPort = data.port;
        }
      }
    } catch (e) {
      console.error('Failed to fetch launcher port:', e);
    }
  }

  // Sauber über Wails beenden oder zur IP-Adresse mit Baseport verzweigen im Browser
  if (window.go && window.go.main && window.go.main.App) {
    setTimeout(() => {
      window.go.main.App.Quit();
    }, 500);
  } else {
    const targetUrl = formatBasePortUrl(launcherPort);
    window.location.href = targetUrl;
  }
}

function formatBasePortUrl(launcherPort: number): string {
  const protocol = window.location.protocol || 'http:';
  const hostname = window.location.hostname || 'localhost';
  const port = launcherPort || 9000;
  return `${protocol}//${hostname}:${port}/`;
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
  $q.dialog({
    title: t('layout.logout') || 'Abmelden',
    message: 'Möchten Sie die Sitzung beenden?',
    cancel: {
      label: t('layout.shutdownCancel') || 'Abbrechen',
      flat: true
    },
    ok: {
      label: t('layout.logout') || 'Abmelden',
      color: 'primary',
      unelevated: true
    },
    persistent: true
  }).onOk(async () => {
    sessionStore.logout();

    if (isServerMode.value || !(window as any).go) {
      $q.notify({
        type: 'warning',
        message: t('layout.closing') || 'Sitzung wird beendet...',
        position: 'center',
        timeout: 1000
      });
      await exitSessionAndRedirect();
    }
  });
};

async function loadHelp() {
  helpLoading.value = true;
  helpUrl.value = '';

  const activeLang = sessionStore.selectedLanguage || 'de';
  const pdfName = activeLang === 'de' ? 'HuhnLite_de.pdf' : `HuhnLite_${activeLang}.pdf`;
  helpUrl.value = `/help/pdfjs/web/viewer.html?file=/help/${pdfName}#pagemode=bookmarks`;

  helpLoading.value = false;
}

async function openHelp() {
  helpDialogOpen.value = true;
  await loadHelp();
}

async function openHelpNatively() {
  const activeLang = sessionStore.selectedLanguage || 'de';
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.OpenHelpFile) {
    const errMsg = await window.go.main.App.OpenHelpFile(activeLang, true);
    if (errMsg) {
      $q.notify({
        type: 'negative',
        message: 'Hilfe konnte nicht geöffnet werden: ' + String(errMsg),
        position: 'top',
        timeout: 5000
      });
    } else {
      // Close the dialog since it was successfully opened natively
      helpDialogOpen.value = false;
    }
  } else {
    // Web browser fallback
    const baseApiUrl = api.defaults.baseURL || window.location.origin;
    const pdfName = activeLang === 'de' ? 'HuhnLite_de.pdf' : `HuhnLite_${activeLang}.pdf`;
    window.open(`${baseApiUrl}/help/${pdfName}`, '_blank');
  }
}

let heartbeatTimer: any = null;

onMounted(async () => {
  console.log('MainLayout: Starte Initialisierung...');

  // Start periodic heartbeat so server tracks active browser sessions and checks for broadcast popups
  const checkHeartbeat = async () => {
    try {
      const res = await api.get('/api/system/heartbeat');
      if (res.data) {
        if (res.data.mode === 'server') {
          isServerMode.value = true;
        }
        if (res.data.broadcast && res.data.broadcast.id) {
          const b: BroadcastMsg = res.data.broadcast;
          if (b.id !== acknowledgedBroadcastId.value) {
            acknowledgedBroadcastId.value = b.id;
            receivedBroadcast.value = b;
            showReceivedBroadcastModal.value = true;
          }
        }
      }
    } catch (e) {}
  };

  checkHeartbeat();
  heartbeatTimer = setInterval(checkHeartbeat, 5000);


  await checkTestDbStatus();
  await fetchActiveTenant();
  
  let success = false;
  let retries = 10;

  while (!success && retries > 0) {
    try {
      const res = await api.get('/api/config');
      sessionStore.authEnabled = res.data.auth_enabled;
      sessionStore.systemEditEnabled = res.data.system_edit_enabled;
      if (res.data.language) {
        const isCli = !!res.data.cli_language;
        sessionStore.setLanguage(res.data.language, isCli);
      }
      console.log('MainLayout: Config geladen:', { auth: sessionStore.authEnabled, system: sessionStore.systemEditEnabled, language: res.data.language, cliLanguage: res.data.cli_language });
      
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

onUnmounted(() => {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
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

onMounted(() => {
  fetchActiveTenant();
  checkTestDbStatus();
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

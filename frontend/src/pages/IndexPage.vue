<template>
  <q-page class="column items-center justify-center q-pa-lg">
    <!-- Title Section -->
    <div class="text-center q-mb-xl" style="line-height: 1.1;">
      <div class="text-weight-bolder text-primary title-text">
        {{ t('auto.huhnlite') }}
      </div>
      <div class="text-grey-7 subtitle-text">
        {{ t('auto.software_loesung_fuer_den_legehennenhalt') }}
      </div>
    </div>

    <!-- Hero Image Section -->
    <div class="q-mb-xl text-center">
      <img
        :src="landingHero"
        style="max-width: 420px; width: 100%; border-radius: 20px; box-shadow: 0 15px 40px rgba(0,0,0,0.15); border: 1px solid rgba(255,255,255,0.1);"
        alt="HuhnLite Farmer"
        class="hover-shadow-premium"
      />
    </div>

    <!-- Cards Layout -->
    <div class="row q-col-gutter-xl justify-center full-width" style="max-width: 1200px;">

      <!-- Stammdaten -->
      <div v-if="session.can('herden_verwalten')" class="col-12 col-sm-6 col-md-3">
        <q-card class="my-card cursor-pointer hover-scale full-height column" @click="router.push('/herden')">
          <q-card-section class="text-center q-pa-lg bg-primary text-white">
            <q-icon name="storage" size="3rem"/>
          </q-card-section>
          <q-card-section class="q-pt-md text-center col" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'">
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">{{ t('auto.stammdaten') }}
            </div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">{{ t('auto.herden_einrichtungen_personen') }}
            </div>
          </q-card-section>
        </q-card>
      </div>

      <!-- Bewegungsdaten -->
      <div v-if="session.can('buchungen_erfassen')" class="col-12 col-sm-6 col-md-3">
        <q-card class="my-card cursor-pointer hover-scale full-height column" @click="router.push('/buchungen')">
          <q-card-section class="text-center q-pa-lg bg-secondary text-white">
            <q-icon name="trending_up" size="3rem"/>
          </q-card-section>
          <q-card-section class="q-pt-md text-center col" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'">
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">
              {{ t('auto.bewegungsdaten') }}
            </div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">{{ t('auto.buchungen_futter_verluste') }}
            </div>
          </q-card-section>
        </q-card>
      </div>

      <!-- Tabellen -->
      <div v-if="session.can('tabellen_anzeigen')" class="col-12 col-sm-6 col-md-3">
        <q-card class="my-card cursor-pointer hover-scale full-height column" @click="router.push('/mist')">
          <q-card-section class="text-center q-pa-lg bg-info text-white">
            <q-icon name="table_chart" size="3rem"/>
          </q-card-section>
          <q-card-section class="q-pt-md text-center col" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'">
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">{{ t('auto.tabellen') }}</div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">{{ t('auto.kosten_parameter_mwst') }}
            </div>
          </q-card-section>
        </q-card>
      </div>

      <!-- Reports -->
      <div v-if="session.can('auswertungen_anzeigen')" class="col-12 col-sm-6 col-md-3">
        <q-card class="my-card cursor-pointer hover-scale full-height column" @click="router.push('/reports')">
          <q-card-section class="text-center q-pa-lg bg-accent text-white">
            <q-icon name="assessment" size="3rem"/>
          </q-card-section>
          <q-card-section class="q-pt-md text-center col" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'">
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">{{ t('auto.reports') }}</div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">{{ t('auto.dynamische_auswertungen') }}
            </div>
          </q-card-section>
        </q-card>
      </div>

      <!-- Login/Logout -->
      <div v-if="session.authEnabled" class="col-12 col-sm-6 col-md-3">
        <q-card class="my-card cursor-pointer hover-scale full-height column auth-card" @click="handleAuthClick">
          <q-card-section class="text-center q-pa-lg text-white bg-auth">
            <q-icon :name="session.isLoggedIn ? 'logout' : 'login'" size="3rem"/>
          </q-card-section>
          <q-card-section class="q-pt-md text-center col" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'">
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">
              {{ userDisplayName }}
            </div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">
              {{ session.isLoggedIn ? 'Sitzung beenden' : 'System-Anmeldung öffnen' }}
            </div>
          </q-card-section>
        </q-card>
      </div>

    </div>

    <div v-if="dbStatus.engine === 'offline'" class="q-mt-lg full-width" style="max-width: 800px;">
      <q-banner rounded class="bg-negative text-white q-pa-md shadow-3">
        <template v-slot:avatar>
          <q-icon name="error" color="white" size="md" />
        </template>
        <div class="text-subtitle1 text-weight-bold">Datenbankverbindung fehlgeschlagen (Offline)</div>
        <div class="text-body2 q-mt-xs">{{ dbStatus.error || 'Es konnte keine Verbindung zur Datenbank hergestellt werden.' }}</div>
        <div class="text-caption text-grey-2 q-mt-xs">Detaillierte Fehlerprotokolle finden Sie in der Log-Datei unter %APPDATA%\HuhnLite\app.log</div>
      </q-banner>
    </div>

    <div class="q-mt-xl column items-center q-gutter-y-sm">
      <div class="row items-center q-gutter-md">
        <div class="text-grey-6 text-caption">
          Version vom {{ buildTime }}
        </div>
        <q-badge v-if="dbStatus.engine" :color="dbStatus.engine === 'postgres' ? 'indigo-7' : (dbStatus.engine === 'mysql' ? 'orange-9' : (dbStatus.engine === 'sqlite' ? 'blue-8' : 'negative'))" class="q-pa-sm text-weight-bold">
          <q-icon :name="dbStatus.engine === 'offline' ? 'link_off' : 'database'" class="q-mr-xs" />
          {{ dbStatus.engine === 'postgres' ? 'PostgreSQL' : (dbStatus.engine === 'mysql' ? 'MariaDB' : (dbStatus.engine === 'sqlite' ? 'SQLite' : 'Offline')) }}
          <q-tooltip>
            {{ dbStatus.engine === 'offline' ? 'FEHLER: ' + (dbStatus.error || 'Keine Verbindung') : 'Verbindung: ' + dbStatus.host }}
          </q-tooltip>
        </q-badge>
      </div>
      <div v-if="dbStatus.engine && dbStatus.engine !== 'offline'" class="text-grey-6 text-caption text-center" style="max-width: 800px; word-break: break-all;">
        {{ t('auto.aktive_db') }} <span :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-8'" class="text-weight-bold">{{ dbStatus.host }}</span>
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { useRouter } from 'vue-router';
import {computed, onMounted, reactive} from 'vue';
import { useSessionStore } from '../stores/session';
import landingHero from '../assets/landing_page.jpg';
import { GetDBStatus } from '../../wailsjs/go/main/App';

const router = useRouter();
const sessionStore = useSessionStore();
const session = sessionStore;

const dbStatus = reactive({
  engine: '',
  host: '',
  error: ''
});

onMounted(async () => {
  try {
    const status = await GetDBStatus();
    dbStatus.engine = status.engine;
    dbStatus.host = status.host;
    dbStatus.error = status.error || '';
  } catch (e) {
    console.error('Failed to get DB status', e);
  }
});

const userDisplayName = computed(() => {
  if (!session.isLoggedIn) return 'Anmelden';
  return session.klarname || session.username || 'Benutzer';
});

const handleAuthClick = () => {
  if (session.isLoggedIn) {
    $q.dialog({
      title: 'Sitzung beenden',
      message: 'Möchten Sie die aktuelle Sitzung wirklich beenden?',
      cancel: {
        label: 'Abbrechen',
        flat: true
      },
      ok: {
        label: 'Abmelden',
        color: 'primary',
        unelevated: true
      },
      persistent: true
    }).onOk(async () => {
      session.logout();
      if (!(window as any).go) {
        $q.notify({
          type: 'warning',
          message: 'Sitzung wird beendet...',
          position: 'center',
          timeout: 1000
        });
        try {
          const res = await api.post('/api/system/shutdown');
          let port = res?.data?.port || 9000;
          const protocol = window.location.protocol || 'http:';
          const hostname = window.location.hostname || 'localhost';
          window.location.href = `${protocol}//${hostname}:${port}/`;
        } catch (e) {
          window.location.href = `${window.location.protocol || 'http:'}//${window.location.hostname || 'localhost'}:9000/`;
        }
      }
    });
  } else {
    session.triggerLogin();
  }
};

const buildTime = process.env.BUILD_TIME;
</script>

<style scoped>
.title-text {
  font-size: clamp(2.5rem, 5vw, 4.5rem);
  letter-spacing: -1px;
}

.subtitle-text {
  font-size: clamp(1.1rem, 2vw, 1.6rem);
}

.hover-scale {
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
}

.hover-scale:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.15) !important;
}

.hover-shadow-premium {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.hover-shadow-premium:hover {
  transform: translateY(-4px) scale(1.01);
  box-shadow: 0 20px 45px rgba(0, 0, 0, 0.25) !important;
}

.bg-auth {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>

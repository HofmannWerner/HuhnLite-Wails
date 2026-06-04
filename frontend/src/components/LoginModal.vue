<template>
  <q-dialog v-model="internalVisible" persistent transition-show="scale" transition-hide="scale" @show="setFocus">
    <q-card v-if="internalVisible" :dark="sessionStore.darkMode" :class="sessionStore.darkMode ? 'bg-grey-10 text-white' : 'bg-white'" style="width: 400px; max-width: 90vw;" class="q-pa-md shadow-24">
      <q-card-section class="flex flex-center j-center">
        <q-avatar size="100px" font-size="52px" color="primary" text-color="white" icon="lock" />
      </q-card-section>

      <q-card-section>
        <div class="text-h5 text-center text-weight-bold">{{ t('auto.huhnlite_anmeldung') }}</div>
        <div class="text-subtitle2 text-center" :class="sessionStore.darkMode ? 'text-grey-4' : 'text-grey-7'">{{ t('auto.bitte_identifizieren_sie_sich') }}</div>
      </q-card-section>

      <q-card-section class="q-gutter-y-md">
        <q-input
          id="login-username-input"
          ref="usernameInput"
          v-model="username"
          :label="t('auto.benutzername')"
          outlined
          autocomplete="off"
          :dark="sessionStore.darkMode"
          label-color="primary"
          @keyup.enter="handleLogin"
        >
          <template v-slot:prepend>
            <q-icon name="person" :color="sessionStore.darkMode ? 'white' : 'primary'" />
          </template>
        </q-input>

        <q-input
          v-model="password"
          :label="t('auto.passwort')"
          outlined
          autocomplete="new-password"
          :dark="sessionStore.darkMode"
          label-color="primary"
          :type="showPassword ? 'text' : 'password'"
          @keyup.enter="handleLogin"
        >
          <template v-slot:prepend>
            <q-icon name="key" :color="sessionStore.darkMode ? 'white' : 'primary'" />
          </template>
          <template v-slot:append>
            <q-icon
              :name="showPassword ? 'visibility_off' : 'visibility'"
              :color="sessionStore.darkMode ? 'white' : 'grey-7'"
              class="cursor-pointer"
              @click="showPassword = !showPassword"
            />
          </template>
        </q-input>
      </q-card-section>

      <q-card-actions align="center" class="q-pb-lg q-gutter-x-sm">
        <q-btn
          :label="t('form.cancel')"
          color="grey-7"
          flat
          class="col"
          @click="handleCancel"
          :disable="loading"
        />
        <q-btn
          :label="t('auto.anmelden')"
          color="primary"
          class="col q-py-sm"
          unelevated
          :loading="loading"
          @click="handleLogin"
        />
      </q-card-actions>
      <div class="text-caption text-center q-pb-md" :class="sessionStore.darkMode ? 'text-grey-4' : 'text-grey-6'">
        Anwendungs-Build: {{ buildTime }}
      </div>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, computed, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar, date } from 'quasar';
import { api } from 'src/boot/api';
import { useSessionStore } from '../stores/session';

const $q = useQuasar();
const router = useRouter();
const sessionStore = useSessionStore();
const usernameInput = ref<any>(null);
const username = ref('');
const password = ref('');
const showPassword = ref(false);
const loading = ref(false);

// Computed mit Setter, um v-model Fehler zu vermeiden
const internalVisible = computed({
  get: () => sessionStore.authEnabled && !sessionStore.isLoggedIn && !sessionStore.loginDismissed,
  set: (val) => {
    if (!val) sessionStore.dismissLogin();
  }
});

const setFocus = () => {
  loading.value = false;
  username.value = '';
  password.value = '';
  
  // "Shotgun"-Ansatz: Wir probieren es 10 mal alle 200ms.
  // Das fängt Verzögerungen beim App-Start sicher ab.
  for (let i = 1; i <= 10; i++) {
    setTimeout(() => {
      const el = document.getElementById('login-username-input');
      if (el) {
        const input = el.querySelector('input');
        if (input) {
          input.focus();
          input.select();
        }
      }
      if (usernameInput.value) {
        usernameInput.value.focus();
      }
    }, i * 200);
  }
};

// Reset bei jedem Öffnen sicherstellen
watch(internalVisible, (val) => {
  if (val) {
    username.value = '';
    password.value = '';
    setFocus();
  }
});

// Fokus wird jetzt primär über @show im q-dialog gesteuert

const handleCancel = () => {
  sessionStore.dismissLogin();
};

const handleLogin = async () => {
  if (!username.value || !password.value) {
    $q.notify({ color: 'warning', message: 'Bitte Benutzername und Passwort eingeben.' });
    return;
  }

  loading.value = true;
  try {
    // Backend erwartet UPPERCASE Keys
    const response = await api.post('/api/login', {
      USERNAME: username.value,
      PASSWORT: password.value
    });
    
    sessionStore.setSession(response.data);

    $q.notify({ 
      color: 'positive', 
      message: `Willkommen zurück, ${response.data.klarname || response.data.username}!`, 
      icon: 'check' 
    });

    // Direkt in die Aktionen verzweigen (Standard-Ansicht für den Benutzer)
    const todayStr = date.formatDate(new Date(), 'YYYY-MM-DD');
    let oldestDateStr = todayStr;
    try {
      const currentUserId = response.data.id || 0;
      const isAdmin = response.data.profile_kz === 'A';
      const res = await api.get('/api/aktionen', {
        params: {
          kz: 'B',
          show_erledigt: 0,
          end: todayStr,
          id_user: 0
        }
      });
      const allActions = res.data || [];
      let relevantActions = [];
      if (isAdmin) {
        relevantActions = allActions;
      } else {
        relevantActions = allActions.filter((a: any) => {
          const val = a.id_user;
          let actionUserId = 0;
          if (val !== null && val !== undefined) {
            if (typeof val === 'object' && 'Int64' in val) actionUserId = Number(val.Int64) || 0;
            else if (typeof val === 'object' && 'Int32' in val) actionUserId = Number(val.Int32) || 0;
            else actionUserId = Number(val) || 0;
          }
          return actionUserId === currentUserId || actionUserId === 0;
        });
      }
      relevantActions.forEach((a: any) => {
        const val = a.aktionsdatum;
        let d = '';
        if (val !== null && val !== undefined) {
          if (typeof val === 'object' && 'String' in val) d = String(val.String);
          else d = String(val);
        }
        if (d && d < oldestDateStr) {
          oldestDateStr = d;
        }
      });
    } catch (err) {
      console.error('Fehler beim Ermitteln des ältesten Aktionsdatums in LoginModal:', err);
    }

    await router.push({ 
      path: '/buchungen', 
      query: { 
        tab: 'aktionen',
        filterKz: 'B',
        filterStartDate: oldestDateStr,
        filterEndDate: todayStr,
        filterUser: response.data.id || 0
      } 
    });
  } catch (error: any) {
    const msg = error.response?.data?.error || 'Anmeldung fehlgeschlagen';
    $q.notify({ color: 'negative', message: msg, icon: 'error' });
  } finally {
    loading.value = false;
  }
};

const buildTime = process.env.BUILD_TIME;
</script>

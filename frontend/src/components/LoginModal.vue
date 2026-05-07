<template>
  <q-dialog v-model="internalVisible" persistent transition-show="scale" transition-hide="scale">
    <q-card :dark="sessionStore.darkMode" :class="sessionStore.darkMode ? 'bg-grey-10 text-white' : 'bg-white'" style="width: 400px; max-width: 90vw;" class="q-pa-md shadow-24">
      <q-card-section class="flex flex-center j-center">
        <q-avatar size="100px" font-size="52px" color="primary" text-color="white" icon="lock" />
      </q-card-section>

      <q-card-section>
        <div class="text-h5 text-center text-weight-bold">HuhnLite Anmeldung</div>
        <div class="text-subtitle2 text-center" :class="sessionStore.darkMode ? 'text-grey-4' : 'text-grey-7'">Bitte identifizieren Sie sich</div>
      </q-card-section>

      <q-card-section class="q-gutter-y-md">
        <q-input
          ref="usernameInput"
          v-model="username"
          label="Benutzername"
          outlined
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
          label="Passwort"
          outlined
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
          label="Abbrechen"
          color="grey-7"
          flat
          class="col"
          @click="handleCancel"
          :disable="loading"
        />
        <q-btn
          label="Anmelden"
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
import { ref, computed, watch, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import { useSessionStore } from '../stores/session';

const $q = useQuasar();
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
  // Kleiner Delay für den Fokus
  setTimeout(() => {
    if (usernameInput.value) usernameInput.value.focus();
  }, 100);
};

watch(internalVisible, (val) => {
  if (val) {
    setFocus();
  }
});

onMounted(() => {
  if (internalVisible.value) {
    setFocus();
  }
});

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
  } catch (error: any) {
    const msg = error.response?.data?.error || 'Anmeldung fehlgeschlagen';
    $q.notify({ color: 'negative', message: msg, icon: 'error' });
  } finally {
    loading.value = false;
  }
};

const buildTime = process.env.BUILD_TIME;
</script>

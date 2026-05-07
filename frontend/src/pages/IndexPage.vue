<template>
  <q-page class="column items-center justify-center q-pa-lg">
    <!-- Title -->
    <!-- Title Section -->
    <div class="text-center q-mb-xl" style="line-height: 1.1;">
      <div class="text-weight-bolder text-primary title-text">
        HuhnLite
      </div>
      <div class="text-grey-7 subtitle-text">
        Software-Lösung für den Legehennenhalter
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
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">Stammdaten
            </div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">Herden,
              Einrichtungen, Personen...
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
              Bewegungsdaten
            </div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">Buchungen,
              Futter, Verluste...
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
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">Tabellen</div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">Kosten,
              Parameter, MwSt...
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
            <div class="text-h5 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : 'text-dark'">Reports</div>
            <div class="text-subtitle2 text-grey-7 q-mt-sm" :class="$q.dark.isActive ? 'text-grey-4' : ''">Dynamische
              Auswertungen...
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
    
    <div class="q-mt-xl text-grey-6 text-caption">
      Version vom {{ buildTime }}
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import {computed} from 'vue';
import { useSessionStore } from '../stores/session';
import landingHero from '../assets/landing_page.jpg';

const router = useRouter();
const sessionStore = useSessionStore();
const session = sessionStore;

const userDisplayName = computed(() => {
  if (!session.isLoggedIn) return 'Anmelden';
  return session.klarname || session.username || 'Benutzer';
});

const handleAuthClick = () => {
  if (session.isLoggedIn) {
    session.logout();
  } else {
    session.triggerLogin();
  }
};

const buildTime = process.env.BUILD_TIME;
</script>

<style scoped>
.title-text {
  font-size: 4rem;
  font-family: 'Inter', 'Roboto', sans-serif;
  letter-spacing: -2px;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.1);
  line-height: 1;
}
.subtitle-text {
  font-size: 2rem;
  font-weight: 500;
  letter-spacing: 1px;
  margin-top: -0.2rem;
}
.hover-scale {
  transition: transform 0.3s cubic-bezier(0.25, 0.8, 0.25, 1), box-shadow 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
  border-radius: 16px;
  overflow: hidden;
}
.hover-scale:hover {
  transform: translateY(-8px);
  box-shadow: 0 14px 28px rgba(0,0,0,0.25), 0 10px 10px rgba(0,0,0,0.22) !important;
}
.bg-primary {
  background: linear-gradient(135deg, #1976D2 0%, #1565C0 100%);
}
.bg-secondary {
  background: linear-gradient(135deg, #26A69A 0%, #00897B 100%);
}
.bg-info {
  background: linear-gradient(135deg, #31CCEC 0%, #00ACC1 100%);
}
.bg-accent {
  background: linear-gradient(135deg, #9C27B0 0%, #7B1FA2 100%);
}

.bg-auth {
  background: linear-gradient(135deg, #FF5722 0%, #E64A19 100%);
}

.hover-shadow-premium {
  transition: transform 0.4s ease, box-shadow 0.4s ease;
}

.hover-shadow-premium:hover {
  transform: scale(1.01);
  box-shadow: 0 30px 60px rgba(0, 0, 0, 0.2) !important;
}
</style>

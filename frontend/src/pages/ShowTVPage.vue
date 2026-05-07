<template>
  <q-page padding class="showtv-page">
    <div class="max-width-container">
      <!-- Header -->
      <div class="row items-center q-mb-xl">
        <div class="col">
          <h1 class="text-h3 text-weight-bolder text-primary q-my-none">
            Anzeige Steuerung
            <span class="text-subtitle1 text-grey-7 block text-weight-medium" :class="$q.dark.isActive ? 'text-grey-4' : ''">Tierbewegungen Sichtbarkeit konfigurieren</span>
          </h1>
        </div>
        <div class="col-auto">
          <q-btn 
            color="primary" 
            icon="refresh" 
            round 
            flat 
            size="lg"
            @click="loadData"
          >
            <q-tooltip>Daten neu laden</q-tooltip>
          </q-btn>
        </div>
      </div>

      <!-- Main Content -->
      <div v-if="loading" class="row justify-center q-pa-xl">
        <q-spinner-dots color="primary" size="40px" />
      </div>

      <div v-else-if="items.length === 0" class="row justify-center q-pa-xl text-center">
        <q-card flat bordered class="q-pa-xl rounded-borders shadow-2" style="max-width: 500px">
          <q-icon name="visibility_off" size="xl" color="grey-5" />
          <div class="text-h6 q-mt-md text-grey-7" :class="$q.dark.isActive ? 'text-grey-3' : ''">Keine Einträge gefunden.</div>
          <p class="text-grey-6 q-mb-lg" :class="$q.dark.isActive ? 'text-grey-5' : ''">Es sind noch keine Steuerungselemente in der Datenbank hinterlegt.</p>
          <q-btn color="primary" outline label="Jetzt aktualisieren" icon="refresh" @click="loadData" />
        </q-card>
      </div>

      <div v-else class="row q-col-gutter-lg">
        <div v-for="item in items" :key="item.ID" class="col-12 col-md-6 col-lg-4">
          <q-card flat bordered class="showtv-card rounded-borders shadow-hover transition-all">
            <q-card-section class="q-pa-md">
              <div class="row items-center no-wrap q-gutter-x-sm">
                <div class="col">
                  <q-input
                    v-model="item.TVNAME"
                    filled
                    dense
                    placeholder="Bezeichnung"
                    @blur="saveItem(item)"
                  />
                </div>
                <div class="col-auto">
                  <q-checkbox
                    v-model="item.SHOWIT"
                    :true-value="1"
                    :false-value="0"
                    color="primary"
                    keep-color
                    @update:model-value="saveItem(item)"
                  >
                    <q-tooltip>Anzeigen ein/aus</q-tooltip>
                  </q-checkbox>
                </div>
              </div>
            </q-card-section>

            <q-inner-loading :showing="item._saving">
              <q-spinner-oval color="primary" size="24px" />
            </q-inner-loading>
          </q-card>
        </div>
      </div>

      <!-- Footer Info -->
      <div class="q-mt-xl text-center text-grey-6 text-italic" :class="$q.dark.isActive ? 'text-grey-5' : ''">
        <q-icon name="info" class="q-mr-xs" />
        Änderungen werden beim Verlassen des Feldes oder beim Umschalten automatisch gespeichert.
      </div>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { api } from 'src/boot/api';
import { useQuasar } from 'quasar';

interface ShowTVItem {
  ID: number;
  TVNAME: string;
  SHOWIT: number;
  _saving?: boolean;
}

const $q = useQuasar();
const loading = ref(true);
const items = ref<ShowTVItem[]>([]);

async function loadData() {
  loading.value = true;
  try {
    const response = await api.get('/api/showtv');
    items.value = response.data || [];
  } catch (error) {
    console.error('Fehler beim Laden von SHOWTV:', error);
    $q.notify({
      type: 'negative',
      message: 'Daten konnten nicht geladen werden.'
    });
  } finally {
    loading.value = false;
  }
}

async function saveItem(item: ShowTVItem) {
  // Verhindert doppeltes Speichern während eines laufenden Vorgangs
  if (item._saving) return;

  item._saving = true;
  try {
    await api.put(`/api/showtv/${item.ID}`, {
      TVNAME: item.TVNAME,
      SHOWIT: item.SHOWIT
    });
    // Erfolg leise anzeigen (optional)
  } catch (error) {
    console.error('Fehler beim Speichern:', error);
    $q.notify({
      type: 'negative',
      message: `Fehler beim Speichern von "${item.TVNAME}"`
    });
  } finally {
    item._saving = false;
  }
}

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.showtv-page {
  min-height: 100vh;
}

.max-width-container {
  max-width: 1200px;
  margin: 0 auto;
}

.rounded-borders {
  border-radius: 16px;
}

.showtv-card {
  border: 1px solid rgba(0, 0, 0, 0.05);
  position: relative;
  overflow: hidden;
}

body.body--dark .showtv-card {
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: var(--q-dark-page);
}

.shadow-hover:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.08) !important;
}

body.body--dark .shadow-hover:hover {
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5) !important;
}

.transition-all {
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}

/* Glassmorphism Effekt für Card-Section angedeutet */
.showtv-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  background: var(--q-primary);
  opacity: 0.7;
}
</style>

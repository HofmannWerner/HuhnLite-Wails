<template>
  <q-page padding class="q-pa-md">
    <div class="row items-center justify-between q-mb-lg">
      <div>
        <h1 class="text-h4 q-my-none text-weight-bold row items-center">
          <q-icon name="business" class="q-mr-sm text-primary" />
          {{ t('menu.mandanten') }}
        </h1>
        <div class="text-subtitle1 text-grey-7">
          Verwalten Sie Ihre verschiedenen Betriebsmandanten und deren Datenbanken.
        </div>
      </div>
      <q-btn
        color="primary"
        icon="add"
        label="Neuer Mandant"
        @click="showCreateDialog = true"
        unelevated
      />
    </div>

    <q-card flat bordered class="rounded-borders shadow-2">
      <q-card-section class="q-pa-none">
        <q-table
          :rows="tenants"
          :columns="columns"
          row-key="id"
          flat
          bordered
          :loading="loading"
          binary-state-sort
          :pagination="{ rowsPerPage: 0 }"
          hide-pagination
        >
          <template v-slot:body-cell-active="props">
            <q-td :props="props" class="text-center">
              <q-badge
                v-if="props.row.id === activeMandant"
                color="positive"
                text-color="white"
                label="Aktiv"
                class="q-px-sm q-py-xs text-weight-bold"
              />
              <q-btn
                v-else
                color="secondary"
                size="sm"
                label="Umschalten"
                @click="switchTenant(props.row.id)"
                :loading="switchingId === props.row.id"
                unelevated
              />
            </q-td>
          </template>

          <template v-slot:body-cell-name="props">
            <q-td :props="props">
              <div class="row items-center no-wrap">
                <span class="text-weight-bold text-body1">{{ props.row.name }}</span>
                <q-btn
                  flat
                  round
                  color="grey-6"
                  icon="edit"
                  size="sm"
                  class="q-ml-sm"
                  @click="editTenant(props.row)"
                >
                  <q-tooltip>Name ändern</q-tooltip>
                </q-btn>
              </div>
            </q-td>
          </template>

          <template v-slot:body-cell-system="props">
            <q-td :props="props" class="text-center">
              <q-toggle
                v-model="props.row.system"
                :true-value="1"
                :false-value="0"
                color="orange"
                @update:model-value="updateTenant(props.row)"
              />
            </q-td>
          </template>

          <template v-slot:body-cell-test="props">
            <q-td :props="props" class="text-center">
              <q-toggle
                v-model="props.row.test"
                :true-value="1"
                :false-value="0"
                color="red"
                @update:model-value="updateTenant(props.row)"
              />
            </q-td>
          </template>
        </q-table>
      </q-card-section>
    </q-card>

    <!-- Create Tenant Dialog -->
    <q-dialog v-model="showCreateDialog" persistent>
      <q-card style="min-width: 400px;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">Neuen Mandanten anlegen</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <q-input
            v-model="newTenantName"
            label="Name des Mandanten"
            placeholder="z.B. Otto Dotter"
            outlined
            autofocus
            @keyup.enter="createTenant"
            :rules="[val => !!val || 'Name ist erforderlich']"
          />
          <div class="text-caption text-grey-7 q-mt-xs">
            Ein neues Verzeichnis <code>mandant_n</code> wird angelegt, und die Referenz-Datenbank (HuhnLite.db) wird dorthin kopiert.
          </div>
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Anlegen & Umschalten"
            color="primary"
            @click="createTenant"
            :loading="creating"
            :disabled="!newTenantName.trim()"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Edit Tenant Name Dialog -->
    <q-dialog v-model="showEditDialog" persistent>
      <q-card style="min-width: 400px;" class="q-pa-sm">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-weight-bold">Mandantenname bearbeiten</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pt-md">
          <q-input
            v-model="editTenantName"
            label="Name des Mandanten"
            outlined
            autofocus
            @keyup.enter="saveTenantName"
            :rules="[val => !!val || 'Name ist erforderlich']"
          />
        </q-card-section>

        <q-card-actions align="right" class="q-px-md q-pb-md">
          <q-btn flat label="Abbrechen" color="grey-7" v-close-popup />
          <q-btn
            label="Speichern"
            color="primary"
            @click="saveTenantName"
            :disabled="!editTenantName.trim()"
            unelevated
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';

const { t } = useI18n();
const $q = useQuasar();

interface Tenant {
  id: number;
  name: string;
  system: number;
  test: number;
}

const tenants = ref<Tenant[]>([]);
const activeMandant = ref<number | null>(null);
const loading = ref(false);
const creating = ref(false);
const switchingId = ref<number | null>(null);

const showCreateDialog = ref(false);
const newTenantName = ref('');

const showEditDialog = ref(false);
const editTenantId = ref<number | null>(null);
const editTenantName = ref('');
const editTenantSystem = ref(0);
const editTenantTest = ref(0);

const columns = [
  { name: 'id', label: 'ID', field: 'id', align: 'left' as const, sortable: true, style: 'width: 50px' },
  { name: 'name', label: 'Mandanten-Name', field: 'name', align: 'left' as const, sortable: true },
  { name: 'system', label: 'Systemverwaltung', field: 'system', align: 'center' as const },
  { name: 'test', label: 'Test-Modus (Test-DB)', field: 'test', align: 'center' as const },
  { name: 'active', label: 'Status / Aktion', field: 'id', align: 'center' as const }
];

async function loadTenants() {
  loading.value = true;
  try {
    const res = await api.get('/api/tenants');
    tenants.value = res.data.tenants || [];
    activeMandant.value = res.data.active_mandant;
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Laden der Mandanten: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    loading.value = false;
  }
}

async function switchTenant(id: number) {
  switchingId.value = id;
  try {
    await api.post('/api/tenants/switch', { id });
    $q.notify({
      type: 'positive',
      message: 'Erfolgreich auf Mandant umgeschaltet.',
      caption: 'Die Anwendung wird neu geladen...'
    });
    setTimeout(() => {
      window.location.reload();
    }, 1000);
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Umschalten: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    switchingId.value = null;
  }
}

async function createTenant() {
  if (!newTenantName.value.trim()) return;
  creating.value = true;
  try {
    await api.post('/api/tenants/create', { name: newTenantName.value.trim() });
    $q.notify({
      type: 'positive',
      message: 'Mandant erfolgreich angelegt & umgeschaltet.',
      caption: 'Die Anwendung wird neu geladen...'
    });
    showCreateDialog.value = false;
    newTenantName.value = '';
    setTimeout(() => {
      window.location.reload();
    }, 1000);
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Anlegen: ' + (err.response?.data?.error || err.message)
    });
  } finally {
    creating.value = false;
  }
}

function editTenant(tenant: Tenant) {
  editTenantId.value = tenant.id;
  editTenantName.value = tenant.name;
  editTenantSystem.value = tenant.system;
  editTenantTest.value = tenant.test;
  showEditDialog.value = true;
}

async function saveTenantName() {
  if (!editTenantName.value.trim() || editTenantId.value === null) return;
  try {
    await api.post('/api/tenants/update', {
      id: editTenantId.value,
      name: editTenantName.value.trim(),
      system: editTenantSystem.value,
      test: editTenantTest.value
    });
    $q.notify({
      type: 'positive',
      message: 'Mandantenname erfolgreich aktualisiert.'
    });
    showEditDialog.value = false;
    
    if (editTenantId.value === activeMandant.value) {
      window.location.reload();
    } else {
      loadTenants();
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Speichern: ' + (err.response?.data?.error || err.message)
    });
  }
}

async function updateTenant(tenant: Tenant) {
  try {
    await api.post('/api/tenants/update', {
      id: tenant.id,
      name: tenant.name,
      system: tenant.system,
      test: tenant.test
    });
    $q.notify({
      type: 'positive',
      message: `Mandant ${tenant.name} erfolgreich aktualisiert.`
    });
    
    if (tenant.id === activeMandant.value) {
      $q.notify({
        type: 'info',
        message: 'Aktiver Mandant geändert. Lade Verbindung neu...'
      });
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    }
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Aktualisieren: ' + (err.response?.data?.error || err.message)
    });
    loadTenants();
  }
}

onMounted(() => {
  loadTenants();
});
</script>

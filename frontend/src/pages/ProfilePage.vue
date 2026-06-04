<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="col">
        <h1 class="text-h4 q-my-none">{{ t('auto.profil_berechtigungen') }}</h1>
        <div class="text-subtitle1 text-grey-7">{{ t('auto.verwalten_sie_die_zugriffsbeschraenkunge') }}</div>
      </div>
    </div>

    <div class="row q-col-gutter-lg">
      <!-- Linke Seite: Profil-Auswahl -->
      <div class="col-12 col-md-3">
        <q-list bordered separator padding :class="[$q.dark.isActive ? 'bg-grey-10' : 'bg-white', 'rounded-borders']">
          <q-item-label header class="row items-center justify-between">
            {{ t('auto.verfuegbare_profile') }}
            <q-btn flat round dense color="primary" icon="add" @click="openCreateDialog">
              <q-tooltip>{{ t('auto.neues_profil_anlegen') }}</q-tooltip>
            </q-btn>
          </q-item-label>
          <q-item
            v-for="p in profiles"
            :key="p.ID"
            clickable
            v-ripple
            :active="selectedProfileKz === p.PROFIL_KZ"
            active-class="bg-blue-1 text-primary"
            @click="selectProfile(p)"
          >

            <q-item-section avatar>
              <q-icon :name="p.PROFIL_KZ === 'A' ? 'admin_panel_settings' : 'person_outline'" />
            </q-item-section>
            <q-item-section>
              <q-item-label class="text-weight-bold">Profil {{ p.PROFIL_KZ }}</q-item-label>
              <q-item-label caption>{{ p.PROFIL_KZ === 'A' ? 'Administrator' : 'Benutzer' }}</q-item-label>
            </q-item-section>
            <q-item-section side v-if="p.PROFIL_KZ !== 'A'">
              <q-btn flat round dense size="sm" color="negative" icon="delete" @click.stop="confirmDelete(p)">
                <q-tooltip>{{ t('auto.profil_loeschen') }}</q-tooltip>
              </q-btn>
            </q-item-section>
          </q-item>

        </q-list>
      </div>

      <!-- Rechte Seite: Berechtigungs-Matrix -->
      <div class="col-12 col-md-9" v-if="selectedProfile">
        <q-card flat bordered :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-white'">
          <q-card-section :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'">
            <div class="row items-center justify-between q-mb-sm">
              <div class="text-h6">Berechtigungen für Profil: {{ selectedProfileKz }}</div>
              <q-checkbox
                v-model="allPermissions"
                :label="t('auto.alle_funktionen_ein_ausschalten')"
                color="primary"
                @update:model-value="toggleAll"
              />
            </div>
            <q-input
              v-model="form.BESCHREIBUNG"
              :label="t('auto.profilbeschreibung')"
              outlined
              dense
              :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
              :placeholder="t('auto.z_b_gastzugriff_ohne_schreibrechte')"
            />
          </q-card-section>


          <q-separator />

          <q-card-section class="q-pa-md">
            <div class="row q-col-gutter-md">
              <!-- Gruppe: Stammdaten & Dashboard -->
              <div class="col-12 col-sm-6">
                <div class="text-subtitle2 q-mb-sm text-primary">{{ t('auto.basis_stammdaten') }}</div>
                <div class="column q-gutter-y-xs">
                  <q-checkbox v-model="form.F_DASHBOARD" :label="t('auto.dashboard_anzeigen')"/>
                  <q-checkbox v-model="form.F_HERDEN_VERWALTEN" :label="t('auto.herden_verwalten')"/>
                  <q-checkbox v-model="form.F_EINRICHTUNGEN_VERWALTEN" :label="t('auto.einrichtungen_verwalten')"/>
                  <q-checkbox v-model="form.F_PERSONEN_VERWALTEN" :label="t('auto.personen_verwalten')"/>
                </div>
              </div>

              <!-- Gruppe: Buchungen & Reports -->
              <div class="col-12 col-sm-6">
                <div class="text-subtitle2 q-mb-sm text-primary">{{ t('auto.operativ_auswertung') }}</div>
                <div class="column q-gutter-y-xs">
                  <q-checkbox v-model="form.F_BUCHUNGEN_ERFASSEN" :label="t('auto.buchungen_erfassen')"/>
                  <q-checkbox v-model="form.F_AUSWERTUNGEN_ANZEIGEN" :label="t('auto.reports_auswertungen_anzeigen')"/>
                  <q-checkbox v-model="form.F_KOSTEN_VERWALTEN" :label="t('auto.kosten_verwalten')"/>
                  <q-checkbox v-model="form.F_TABELLEN_ANZEIGEN" :label="t('auto.roh_tabellen_anzeigen_mist')"/>
                </div>
              </div>

              <div class="col-12 q-my-sm"><q-separator /></div>

              <!-- Gruppe: Systemverwaltung -->
              <div class="col-12 col-sm-6">
                <div class="text-subtitle2 q-mb-sm text-warning text-uppercase">{{ t('auto.system_admin') }}</div>
                <div class="column q-gutter-y-xs">
                  <q-checkbox v-model="form.F_SYSTEM_VERWALTUNG" :label="t('auto.zugriff_systemverwaltung_panel')"/>
                  <q-checkbox v-model="form.F_BENUTZER_PROFILE" :label="t('auto.benutzer_profile_verwalten')"/>
                  <q-checkbox v-model="form.F_SQL_STRUKTUR_VERWALTEN" :label="t('auto.berichtsstruktur_verwalten')"/>
                  <q-checkbox v-model="form.F_PARAMETER_EDITIEREN" :label="t('auto.parameter_editieren')"/>
                  <q-checkbox v-model="form.F_TEXTE_VERWALTEN" :label="t('auto.texte_verwalten')"/>
                  <q-checkbox v-model="form.F_BACKUP_ERSTELLEN" :label="t('auto.backups_erstellen')"/>
                </div>
              </div>
            </div>
          </q-card-section>

          <q-card-actions align="right" :class="['q-pa-md', $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-1']">
            <q-btn :label="t('form.save')" color="primary" icon="save" @click="savePermissions" unelevated :loading="saving" />
          </q-card-actions>
        </q-card>
      </div>

      <div v-else :class="['col-12 col-md-9 flex flex-center j-center rounded-borders', $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2']" style="min-height: 300px">
        <div class="text-h6 text-grey-6 text-center">
          <q-icon name="arrow_back" size="lg" class="q-mb-md" /><br />
          {{ t('auto.bitte_waehlen_sie_links_ein_profil_zur_b') }}
        </div>
      </div>
    </div>
    <!-- Dialog für neues Profil -->
    <q-dialog v-model="createDialogVisible" persistent>
      <q-card style="min-width: 350px">
        <q-card-section>
          <div class="text-h6">{{ t('auto.neues_profil_anlegen') }}</div>
        </q-card-section>

        <q-card-section class="q-pt-none">
          <q-input
            v-model="newProfileKz"
            :label="t('auto.profil_kennzeichen_1_zeichen')"
            outlined
            dense
            autofocus
            maxlength="1"
            counter
            @keyup.enter="createProfile"
          />
          <q-input
            v-model="newProfileDesc"
            :label="t('auto.beschreibung')"
            outlined
            dense
            class="q-mt-md"
            @keyup.enter="createProfile"
          />
        </q-card-section>


        <q-card-actions align="right" class="text-primary">
          <q-btn flat :label="t('form.cancel')" v-close-popup />
          <q-btn flat :label="t('auto.anlegen')" @click="createProfile" :disable="!newProfileKz" />
        </q-card-actions>
      </q-card>
    </q-dialog>

  </q-page>
</template>


<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, onMounted, watch } from 'vue';
import { api } from 'src/boot/api';
import { useQuasar } from 'quasar';

const $q = useQuasar();
const profiles = ref<any[]>([]);
const selectedProfile = ref<any>(null);
const selectedProfileKz = ref<string | null>(null);
const saving = ref(false);
const allPermissions = ref(false);

const createDialogVisible = ref(false);
const newProfileKz = ref('');
const newProfileDesc = ref('');



const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  return Number(val) || 0;
};

const form = ref<Record<string, any>>({
  BESCHREIBUNG: '',
  F_DASHBOARD: false,
  F_HERDEN_VERWALTEN: false,
  F_EINRICHTUNGEN_VERWALTEN: false,
  F_PERSONEN_VERWALTEN: false,
  F_BUCHUNGEN_ERFASSEN: false,
  F_AUSWERTUNGEN_ANZEIGEN: false,
  F_SQL_STRUKTUR_VERWALTEN: false,
  F_BENUTZER_PROFILE: false,
  F_PARAMETER_EDITIEREN: false,
  F_KOSTEN_VERWALTEN: false,
  F_TABELLEN_ANZEIGEN: false,
  F_TEXTE_VERWALTEN: false,
  F_SYSTEM_VERWALTUNG: false,
  F_BACKUP_ERSTELLEN: false
});


const loadProfiles = async () => {
  try {
    const res = await api.get('/api/userprofiles');
    profiles.value = res.data;
  } catch (error) {
    $q.notify({ color: 'negative', message: 'Fehler beim Laden der Profile' });
  }
};

const openCreateDialog = () => {
  newProfileKz.value = '';
  newProfileDesc.value = '';
  createDialogVisible.value = true;
};

const createProfile = async () => {
  if (!newProfileKz.value) return;

  try {
    await api.post('/api/userprofiles', {
      profil_kz: newProfileKz.value.toUpperCase(),
      beschreibung: newProfileDesc.value
    });


    $q.notify({ color: 'positive', message: `Profil ${newProfileKz.value} erfolgreich angelegt.` });
    createDialogVisible.value = false;
    await loadProfiles();
  } catch (error) {
    console.error(error);
    $q.notify({ color: 'negative', message: 'Fehler beim Anlegen: Profil-Kennzeichen existiert evtl. schon.' });
  }
};

const confirmDelete = (p: any) => {
  $q.dialog({
    title: 'Profil löschen',
    message: `Soll das Profil '${p.PROFIL_KZ}' wirklich gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/userprofiles/${p.PROFIL_KZ}`);
      $q.notify({ color: 'positive', message: 'Profil erfolgreich gelöscht.' });
      if (selectedProfileKz.value === p.PROFIL_KZ) {
        selectedProfile.value = null;
        selectedProfileKz.value = null;
      }
      await loadProfiles();
    } catch (error) {
      console.error(error);
      $q.notify({ color: 'negative', message: 'Fehler beim Löschen des Profils.' });
    }
  });
};

const selectProfile = (p: any) => {
  selectedProfile.value = p;
  selectedProfileKz.value = extractString(p.PROFIL_KZ);

  // Mapping von DB (0/1) zu Boolean
  Object.keys(form.value).forEach(key => {
    if (key === 'BESCHREIBUNG') {
      form.value[key] = extractString(p[key]);
      return;
    }
    const val = extractInt(p[key]);
    form.value[key] = val === 1;
  });

  updateMasterSwitchState();
};

const updateMasterSwitchState = () => {
  const { BESCHREIBUNG, ...perms } = form.value;
  const values = Object.values(perms);
  allPermissions.value = values.length > 0 && values.every(v => v === true);
};


// Master-Switch Logik
const toggleAll = (val: boolean | string | number | null) => {
  const boolVal = !!val;
  Object.keys(form.value).forEach(key => {
    if (key !== 'BESCHREIBUNG') {
      form.value[key] = boolVal;
    }
  });
};


// Automatisches Update des Master-Switches wenn einzelne Checkboxen geändert werden
watch(form, () => {
  updateMasterSwitchState();
}, { deep: true });

const savePermissions = async () => {
  saving.value = true;
  try {
    // Mapping von Boolean zu DB (0/1)
    const payload: any = {};
    Object.keys(form.value).forEach(key => {
      if (key === 'BESCHREIBUNG') {
        payload[key] = form.value[key];
      } else {
        payload[key] = form.value[key] ? 1 : 0;
      }
    });



    await api.put(`/api/userprofiles/${selectedProfileKz.value}`, payload);

    $q.notify({ color: 'positive', message: `Berechtigungen für Profil ${selectedProfileKz.value} gespeichert.` });
    await loadProfiles(); // Profile neu laden

    // Aktuelles Profil in der Liste finden und State updaten
    const updated = profiles.value.find(p => p.PROFIL_KZ === selectedProfileKz.value);
    if (updated) selectedProfile.value = updated;

  } catch (error) {
    $q.notify({ color: 'negative', message: 'Fehler beim Speichern der Berechtigungen' });
  } finally {
    saving.value = false;
  }
};

onMounted(loadProfiles);
</script>

<style scoped>
.q-checkbox {
  width: 100%;
}
</style>

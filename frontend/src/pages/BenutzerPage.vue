<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="col">
        <h1 class="text-h4 q-my-none">{{ t('auto.benutzerverwaltung') }}</h1>
        <div class="text-subtitle1 text-grey-7">{{ t('auto.hier_koennen_sie_benutzer_anlegen_und_pr') }}</div>
      </div>
      <div class="col-auto">
        <q-btn color="primary" icon="person_add" :label="t('auto.neuer_benutzer')" @click="openDialog()" unelevated />
      </div>
    </div>

    <q-card flat bordered>
      <q-table
        :rows="users"
        :columns="columns"
        row-key="ID"
        flat
        :loading="loading"
        no-data-label="Keine Benutzer gefunden"
      >
        <template v-slot:body-cell-actions="props">
          <q-td :props="props" class="q-gutter-x-sm">
            <q-btn flat round color="primary" icon="edit" @click="openDialog(props.row)" />
            <q-btn flat round color="negative" icon="delete" @click="deleteUser(props.row)" />
          </q-td>
        </template>
      </q-table>
    </q-card>

    <!-- User Dialog -->
    <q-dialog v-model="dialogVisible" persistent>
      <q-card style="min-width: 400px" class="q-pa-md">
        <q-card-section>
          <div class="text-h6">{{ editMode ? 'Benutzer bearbeiten' : 'Neuer Benutzer' }}</div>
        </q-card-section>

        <q-card-section class="q-gutter-y-md">
          <q-input v-model="form.USERNAME" :label="t('auto.username')" outlined dense :readonly="editMode" />
          <q-input v-model="form.KLARNAME" :label="t('auto.klarname_anzeige')" outlined dense />
          <q-input v-model="form.PASSWORT" :label="t('auto.passwort')" type="password" outlined dense />
          <q-select
            v-model="form.ID_BENUTZER_PROFILE"
            :options="profileOptions"
            :label="t('auto.profil')"
            emit-value
            map-options
            outlined
            dense
            option-value="ID"
            :option-label="opt => opt.PROFIL_KZ + (opt.BESCHREIBUNG ? ' - ' + (typeof opt.BESCHREIBUNG === 'string' ? opt.BESCHREIBUNG : opt.BESCHREIBUNG) : '')"


          />
        </q-card-section>

        <q-card-actions align="right" class="q-mt-md">
          <q-btn flat :label="t('form.cancel')" color="grey" v-close-popup />
          <q-btn unelevated :label="t('form.save')" color="primary" @click="saveUser" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, onMounted } from 'vue';
import { api } from 'src/boot/api';
import { useQuasar } from 'quasar';


const $q = useQuasar();
const users = ref([]);
const profiles = ref([]);
const profileOptions = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const editMode = ref(false);

const form = ref({
  ID: null as number | null,
  USERNAME: '',
  KLARNAME: '',
  PASSWORT: '',
  ID_BENUTZER_PROFILE: null as number | null
});

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};

const columns: any[] = [
  { name: 'actions', label: 'Aktionen', field: 'actions', align: 'left' },
  { name: 'USERNAME', label: 'Username', field: 'USERNAME', align: 'left', sortable: true },
  { name: 'KLARNAME', label: 'Klarname', field: (row: any) => extractString(row.KLARNAME), align: 'left', sortable: true },
  { name: 'PROFIL', label: 'Profil', field: 'PROFIL_KZ', align: 'left' }
];

const loadData = async () => {
  loading.value = true;
  try {
    const resUsers = await api.get('/api/benutzer');
    users.value = resUsers.data;

    const resProfiles = await api.get('/api/userprofiles');

    profileOptions.value = resProfiles.data;
  } catch (error) {
    console.error(error);
    $q.notify({ color: 'negative', message: 'Fehler beim Laden der Daten' });
  } finally {
    loading.value = false;
  }
};

const openDialog = (row: any = null) => {
  if (row) {
    editMode.value = true;
    form.value = { 
      ID: row.ID,
      USERNAME: row.USERNAME,
      KLARNAME: extractString(row.KLARNAME),
      PASSWORT: '', 
      ID_BENUTZER_PROFILE: row.ID_BENUTZER_PROFILE 
    };
  } else {
    editMode.value = false;
    form.value = { ID: null, USERNAME: '', KLARNAME: '', PASSWORT: '', ID_BENUTZER_PROFILE: null };
  }
  dialogVisible.value = true;
};


const saveUser = async () => {
  try {
    if (editMode.value) {
      await api.put(`/api/benutzer/${form.value.ID}`, form.value);
    } else {
      await api.post('/api/benutzer', form.value);
    }

    $q.notify({ color: 'positive', message: 'Benutzer gespeichert' });
    dialogVisible.value = false;
    loadData();
  } catch (error) {

    console.error(error);
    $q.notify({ color: 'negative', message: 'Fehler beim Speichern' });
  }
};

const deleteUser = (row: any) => {
  $q.dialog({
    title: 'Löschen',
    message: `Möchten Sie den Benutzer ${row.USERNAME} wirklich löschen?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/benutzer/${row.ID}`);
      $q.notify({ color: 'positive', message: 'Benutzer gelöscht' });

      loadData();
    } catch (error) {
      console.error(error);
      $q.notify({ color: 'negative', message: 'Fehler beim Löschen' });
    }
  });
};

onMounted(loadData);

</script>

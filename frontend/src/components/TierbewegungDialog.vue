<template>
  <q-dialog :model-value="modelValue" @update:model-value="$emit('update:modelValue', $event)" persistent @show="onDialogShow">
    <q-card style="width: 500px; max-width: 95vw; border-radius: 16px;">
      <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
        <div class="text-h6 text-weight-bold">{{ isEditing ? 'Bewegung bearbeiten' : 'Neue Bewegung' }}</div>
        <q-space />
        <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
      </q-card-section>

      <q-card-section class="q-pa-lg">
        <q-form @submit="onSubmit" class="q-gutter-y-md">

          <!-- Zeile 1: Herde -->
          <div class="row q-col-gutter-sm">
            <div class="col-12">
               <q-select
                 v-model="form.ID_HERDEN"
                :options="herdeOptions"
                 option-value="ID"
                 option-label="BEZEICHNUNG"
                emit-value
                map-options
                label="Herde"
                filled
                stack-label
                readonly
                hide-bottom-space
                :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              />
            </div>
          </div>

          <!-- Zeile 2: Datum & Typ -->
          <div class="row q-col-gutter-sm">
            <div class="col-12 col-sm-6">
               <q-input
                :model-value="form.fullTimestamp ? form.fullTimestamp.split(' ')[0] : ''"
                type="date"
                :label="t('auto.bewegungsdatum')"
                filled
                stack-label
                readonly
                hide-bottom-space
                :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
              />
            </div>
            <div class="col-12 col-sm-6">
               <q-select
                 v-model="form.TYP"
                :options="typOptions"
                :label="t('auto.bewegungsart')"
                filled
                stack-label
                emit-value
                map-options
                :rules="[val => !!val || 'Erforderlich']"
                :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                :readonly="fixedTyp"
              />
            </div>
          </div>

          <div class="row q-col-gutter-sm">
            <div class="col-12 col-sm-6">
               <q-input
                 v-model.number="form.BEWEGUNGEN"
                type="number"
                :label="t('auto.anzahl_stueck')"
                filled
                stack-label
                :rules="[
                  val => val !== null && val > 0 || 'Geben Sie eine Anzahl ein',
                  val => (form.TYP !== 'V' || currentBestand === null || val <= (currentBestand || 0)) || `Anzahl (${val}) übersteigt den Tierbestand (${currentBestand})!`
                ]"
                :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                lazy-rules
              >
                 <template v-slot:hint v-if="form.TYP === 'V' && currentBestand !== null">
                  Verfügbarer Bestand: {{ currentBestand.toLocaleString('de-DE') }} Tiere
                </template>
              </q-input>
            </div>
            <div class="col-12 col-sm-6">
              <q-input
                v-model.number="form.KOSTEN"
                type="number"
                step="0.01"
                :label="t('menu.kosten')"
                filled
                stack-label
                prefix="€"
                :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
              />
            </div>
          </div>

          <!-- Zeile 4: Grund / Bemerkung -->
          <div class="row q-col-gutter-sm">
            <div class="col-12">
               <q-select
                 v-model="form.ID_TEXTE"
                :options="filteredTexteOptions"
                 option-value="ID"
                 option-label="LABEL"
                emit-value
                map-options
                :label="t('auto.grund_bemerkung')"
                filled
                stack-label
                :rules="[val => !!val || 'Erforderlich']"
                 :disable="!form.TYP"
                 :bg-color="$q.dark.isActive ? 'grey-9' : (form.TYP ? 'white' : 'grey-3')"
                class="full-width"
              />
            </div>
          </div>

          <div class="row justify-end q-mt-md q-gutter-x-sm">
            <q-btn ref="cancelBtn" :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog" padding="xs lg" />
            <q-btn ref="saveBtn" :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated padding="xs xl" />
          </div>
        </q-form>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
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

import { ref, reactive, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import { useSessionStore } from '../stores/session';

/* eslint-disable @typescript-eslint/no-explicit-any */

const props = defineProps({
  modelValue: Boolean,
  isEditing: Boolean,
  editId: Number as any,
  initialHerdeId: Number as any,
  initialTyp: String,
  fixedTyp: Boolean // Falls der Typ nicht geändert werden darf (z.B. aus BuchungPage)
});

const emit = defineEmits(['update:modelValue', 'saved']);

const $q = useQuasar();
const sessionStore = useSessionStore();
const herdeOptions = ref<any[]>([]);
const typOptions = ref<any[]>([]);
const texteAll = ref<any[]>([]);
const cancelBtn = ref<any>(null);

const form = reactive({
  ID_HERDEN: null as number | null,
  HERDENNUMMER: 0,
  ID_BUCHUNG: 0,
  TYP: '',
  ID_TEXTE: null as number | null,
  BEWEGUNGSDATUM: '',
  fullTimestamp: '',
  BEWEGUNGEN: null as number | null,
  ID_HERDEN_VON: 0,
  ID_HERDEN_NACH: 0,
  KOSTEN: 0
});

const currentBestand = ref<number | null>(null);

watch(() => form.ID_HERDEN, async (newVal) => {
  if (newVal) {
    try {
      const res = await api.get(`/api/buchung/last-info/${newVal}`);
      currentBestand.value = res.data.tierbestand || 0;
      console.log('TierbewegungDialog - Census for herde', newVal, 'is', currentBestand.value);
    } catch (err) {
      console.error('Bestand konnte nicht geladen werden', err);
      currentBestand.value = 0;
    }
  } else {
    currentBestand.value = null;
  }
}, { immediate: true });

const filteredTexteOptions = computed(() => {
  if (!form.TYP) return [];
  return texteAll.value
    .filter(t => (t.TEXT_TYP_KZ || '') === form.TYP)
    .map(t => ({
      LABEL: (t.BETREFF || t.INHALT || `Eintrag ${t.ID}`),
      ID: t.ID
    }));
});

async function onDialogShow() {
  await Promise.all([fetchHerden(), fetchTexte()]);

  if (props.isEditing && props.editId) {
    await loadEditData(props.editId);
  } else {
    resetForm();
    form.ID_HERDEN = props.initialHerdeId || null;
    form.TYP = props.initialTyp || '';
    form.fullTimestamp = sessionStore.workingTimestamp;
  }
}

async function loadEditData(id: number) {
  try {
    const res = await api.get(`/api/tierbewegungen`);
    const row = res.data.find((r: any) => r.ID === id || r.id === id);
    if (row) {
      const hNum = extractInt(row.herdennummer || row.HERDENNUMMER);
      // Find the ID for this herdennummer
      const herde = herdeOptions.value.find(h => h.HERDENNUMMER === hNum);
      
      Object.assign(form, {
        ID_HERDEN: herde ? herde.ID : null,
        HERDENNUMMER: hNum,
        ID_BUCHUNG: extractInt(row.id_buchung || row.ID_BUCHUNG),
        TYP: extractString(row.typ || row.TYP),
        ID_TEXTE: extractInt(row.id_texte || row.ID_TEXTE),
        BEWEGUNGSDATUM: extractString(row.bewegungsdatum || row.BEWEGUNGSDATUM),
        fullTimestamp: `${extractString(row.bewegungsdatum || row.BEWEGUNGSDATUM)} 12:00`,
        BEWEGUNGEN: extractInt(row.bewegungen || row.BEWEGUNGEN),
        ID_HERDEN_VON: extractInt(row.id_herden_von || row.ID_HERDEN_VON),
        ID_HERDEN_NACH: extractInt(row.id_herden_nach || row.ID_HERDEN_NACH),
        KOSTEN: row.kosten || row.KOSTEN || 0
      });
    }
  } catch (err) {
    console.error('Error loading edit data:', err);
  }
}

function resetForm() {
  form.ID_HERDEN = null;
  form.HERDENNUMMER = 0;
  form.ID_BUCHUNG = 0;
  form.TYP = '';
  form.ID_TEXTE = null;
  form.BEWEGUNGEN = null;
  form.BEWEGUNGSDATUM = '';
  form.ID_HERDEN_VON = 0;
  form.ID_HERDEN_NACH = 0;
  form.KOSTEN = 0;
}

async function fetchHerden() {
  const res = await api.get('/api/herden/lookup');
  herdeOptions.value = (res.data || []).map((h: any) => {
    const id = h.ID || h.id;
    const num = h.HERDENNUMMER || h.herdennummer;
    const bez = h.BEZEICHNUNG || h.bezeichnung;
    return {
      ID: id,
      HERDENNUMMER: num,
      BEZEICHNUNG: bez ? bez : (num ? `Herde ${num}` : `Herde ${id}`)
    };
  });
}

async function fetchTexte() {
  const [resTexte, resTypen] = await Promise.all([
    api.get('/api/texte'),
    api.get('/api/texttypen')
  ]);

  texteAll.value = resTexte.data || [];

  const filteredTypen = (resTypen.data || []).filter((t: any) => t.KZ === 'T' || t.KZ === 'V' || t.KZ === 'Z');
  typOptions.value = filteredTypen.map((t: any) => ({
    label: t.BEZEICHNUNG || t.KZ,
    value: t.KZ
  }));

  if (typOptions.value.length === 0) {
    typOptions.value = [
      { label: 'Abgang/Tod', value: 'T' },
      { label: 'Zugang/Verkauf', value: 'V' }
    ];
  }
}

function closeDialog() {
  emit('update:modelValue', false);
}

async function onSubmit() {
  try {
    const dateStr = form.fullTimestamp.split(' ')[0];
    
    // Find herdennummer for the selected herd
    const herde = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
    const herdenNummer = herde ? herde.HERDENNUMMER : 0;

    const payload = {
      HERDENNUMMER: herdenNummer,
      ID_BUCHUNG: form.ID_BUCHUNG || 0,
      TYP: form.TYP,
      ID_TEXTE: Number(form.ID_TEXTE),
      BEWEGUNGSDATUM: dateStr,
      BEWEGUNGEN: Number(form.BEWEGUNGEN),
      ID_HERDEN_VON: form.ID_HERDEN_VON || 0,
      ID_HERDEN_NACH: form.ID_HERDEN_NACH || 0,
      KOSTEN: Number(form.KOSTEN) || 0
    };

    if (props.isEditing && props.editId) {
      await api.put(`/api/tierbewegungen/${props.editId}`, payload);
      $q.notify({ type: 'positive', message: 'Eintrag aktualisiert' });
    } else {
      await api.post('/api/tierbewegungen', payload);
      $q.notify({ type: 'positive', message: 'Eintrag erstellt' });
    }
    emit('saved');
    closeDialog();
  } catch (err) {
    console.error('Error saving movement:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}
</script>

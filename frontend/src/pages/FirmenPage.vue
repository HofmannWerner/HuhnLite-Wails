<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">{{ t('grid.companyAdmin') }}</div>
    </div>

    <!-- BEREICH 0: SYSTEM-SICHERHEIT (Parametermodus) -->
    <q-card flat bordered class="q-mb-xl rounded-borders shadow-2"
            :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-amber-1'" style="border-radius: 16px;">
      <q-card-section class="row items-center q-pa-md">
        <q-icon name="security" size="md" color="amber-9" class="q-mr-md"/>
        <div>
          <div class="text-h6 text-weight-bold" :class="$q.dark.isActive ? 'text-white' : ''">{{ t('grid.systemSecurity') }}</div>
          <div class="text-caption text-grey-8" :class="$q.dark.isActive ? 'text-grey-4' : ''">{{ t('auto.zentrale_steuerung_der_login_pflicht_bei') }}
          </div>
        </div>
        <q-space/>
        <q-checkbox
          v-model="authRequired"
          :label="t('auto.anmeldung_beim_start_erforderlich')"
          color="amber-9"
          size="lg"
          left-label
          class="text-weight-bold"
          :disable="session.profile_kz !== 'A'"
          @update:model-value="saveAuthSetting"
        >
          <q-tooltip v-if="session.profile_kz !== 'A'">{{ t('auto.nur_administratoren_koennen_diese_einste') }}</q-tooltip>
        </q-checkbox>
      </q-card-section>
    </q-card>

    <!-- BEREICH 1: FIRMENSTAMMDATEN -->
    <q-card flat bordered class="q-mb-xl rounded-borders shadow-2" style="border-radius: 16px;">
      <q-card-section class="bg-primary text-white row items-center q-pa-md">
        <q-icon name="business" size="md" class="q-mr-md" />
        <div class="text-h6 text-weight-bold">{{ t('grid.companyMasterData') }}</div>
        <q-space />
        <q-btn ref="companyCancelBtn" icon="undo" :label="t('form.cancel')" color="white" flat rounded @click="resetCompany" autofocus />
        <q-btn ref="companySaveBtn" :disable="!isCompanyDirty" icon="save" :label="t('form.save')" color="white" outline rounded @click="saveCompany" />
      </q-card-section>

      <q-card-section class="q-pa-lg">
        <q-form class="q-gutter-md">
          <div class="row q-col-gutter-x-sm q-col-gutter-y-md">
            <div class="col-12 col-md-3 text-center">
              <q-avatar size="140px" :style="$q.dark.isActive ? 'background: #333;' : 'background: #eee;'"
                        class="shadow-2">
                <img v-if="companyForm.FOTO"
                     :src="companyForm.FOTO.startsWith('data:') ? companyForm.FOTO : 'data:image/jpeg;base64,' + companyForm.FOTO"/>
                <img v-else src="img/logo_huhnlite.png" style="object-fit: cover;" />
                <q-btn round dense color="primary" icon="photo_camera" class="absolute-bottom-right">
                  <q-file v-model="photoFile" class="absolute-full opacity-0" accept="image/*" borderless @update:model-value="onPhotoSelected" />
                </q-btn>
              </q-avatar>
            </div>
            <div class="col-12 col-md-9"
                 :style="($q.dark.isActive ? 'border-left: 1px solid #444; ' : 'border-left: 1px solid #ddd; ') + 'padding-left: 32px;'">
              <div class="row q-col-gutter-x-sm q-col-gutter-y-md">
                <div class="col-12 col-sm-6">
                  <q-input v-model="companyForm.NAME" :label="t('auto.name')" filled stack-label dense/>
                </div>
                <div class="col-12 col-sm-6">
                  <q-input v-model="companyForm.FIRMA" :label="t('auto.zusatz')" filled stack-label dense/>
                </div>
                <div class="col-12">
                  <q-input v-model="companyForm.STRASSE" :label="t('auto.strasse')" filled stack-label dense/>
                </div>
                <div class="col-4">
                  <q-input v-model="companyForm.PLZ" :label="t('auto.plz')" filled stack-label dense/>
                </div>
                <div class="col-8">
                  <q-input v-model="companyForm.ORT" :label="t('auto.ort')" filled stack-label dense/>
                </div>
                <div class="col-12 col-sm-6">
                  <q-input v-model="companyForm.MOBILTELEPHON" :label="t('auto.mobil')" filled stack-label dense prefix="📱"/>
                </div>
                <div class="col-12 col-sm-6">
                  <q-input v-model="companyForm.EMAIL" :label="t('auto.email')" filled stack-label dense prefix="✉️"/>
                </div>
              </div>
            </div>
          </div>
        </q-form>
      </q-card-section>
    </q-card>

    <!-- BEREICH 2: GLOBALE PARAMETER -->
    <q-card flat bordered class="rounded-borders shadow-2" style="border-radius: 16px;">
      <q-card-section class="bg-secondary text-white row items-center q-pa-md">
        <q-icon name="settings_suggest" size="md" class="q-mr-md" />
        <div class="text-h6 text-weight-bold">{{ t('grid.globalParameters') }}</div>
        <q-space />
        <q-btn ref="paramsCancelBtn" icon="undo" :label="t('form.cancel')" color="white" flat rounded @click="resetParams" />
        <q-btn ref="paramsSaveBtn" :disable="!isParamsDirty" icon="save" :label="t('form.save')" color="white" outline rounded @click="saveParams" />
      </q-card-section>

      <q-card-section class="q-pa-lg">
        <q-form class="q-gutter-md">
          <!-- Textfelder oben -->
          <div class="text-subtitle2 q-mb-sm text-secondary text-weight-bold">{{ t('grid.numericalBaseValues') }}</div>
          <div class="row q-col-gutter-md">
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model="paramForm.MASSVOLLEI" :label="t('auto.mass_vollei')" filled stack-label dense/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.ANZAHLKONTROLLW" type="number" :label="t('auto.anzahl_kontrolle')" filled
                       stack-label dense/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.LAUFZEITWOCHEN" type="number" :label="t('auto.laufzeit_w')" filled stack-label
                       dense/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.PRODUKTIONSDAUER" type="number" :label="t('auto.prod_dauer')" filled stack-label
                       dense/>
            </div>
            <div class="col-12 col-sm-4 col-md-1">
              <q-input v-model.number="paramForm.LEGEBEGINN_LW" type="number" :label="t('auto.legebeg_lw')" filled stack-label
                       dense/>
            </div>
            <div class="col-12 col-sm-3 col-md-1">
              <q-checkbox v-model="paramForm.BIO" :label="t('auto.bio')" dense color="primary"
                        class="full-height items-center q-pt-sm"/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.BIOAUFSCHLAG" type="number" step="any" :label="t('auto.bio_aufschlag')"
                       filled stack-label dense prefix="€"/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-select
                v-model="paramForm.HALTUNGSTYP"
                :options="haltungstypOptions"
                :label="t('auto.haltungstyp').replace(' *', '')"
                filled
                stack-label
                dense
                emit-value
                map-options
                @update:model-value="val => { if (val === '0') paramForm.BIO = true }"
              />
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.VERPACKUNGKG" type="number" step="any" label="Verpackung (kg)" filled
                       stack-label dense/>
            </div>
            <div class="col-12 col-sm-4 col-md-2">
              <q-input v-model.number="paramForm.MAXTAGEVERMITTELN" type="number" :label="t('auto.max_tage_vermittlung')" filled
                       stack-label dense/>
            </div>
            <div class="col-12 col-sm-6 col-md-3">
              <q-select
                v-model="paramForm.ID_TABELLEALTER"
                :options="alterTabellenOptions"
                option-value="ID"
                option-label="BEZEICHNUNG"
                emit-value
                map-options
                :label="t('auto.referenz_alterstabelle')"
                filled
                stack-label
                dense
              />
            </div>
            <div class="col-12 col-sm-6 col-md-3">
              <q-select
                v-model="paramForm.ID_TABELLEGEWICHT"
                :options="gewichtTabellenOptions"
                option-value="ID"
                option-label="BEZEICHNUNG"
                emit-value
                map-options
                :label="t('auto.referenz_gewichtstabelle')"
                filled
                stack-label
                dense
              />
            </div>
          </div>

          <q-separator class="q-my-lg" />

          <!-- Neue Rubrik: Chargen-Einstellungen -->
          <div class="text-subtitle2 q-mb-sm text-secondary text-weight-bold">{{ t('grid.batchSettings') }}</div>
          <div class="row q-col-gutter-md q-mb-sm items-start">
            <div class="col-12 col-sm-4">
              <q-input v-model="paramForm.CHARGEPREFIXFIRMA" :label="t('auto.praefix_firma')" filled stack-label dense/>
            </div>
            <div class="col-12 col-sm-1" style="min-width: 80px;">
              <q-input
                v-model="paramForm.CHARGETRENNUNG"
                :label="t('auto.trenn')"
                filled stack-label dense
                maxlength="1"
                input-class="text-center"
                :rules="[
                  val => !!val || '!',
                  val => /^[.\-_/ ]$/.test(val) || 'Erlaubt: . - _ / Leer'
                ]"
                hide-bottom-space
              />
            </div>
          </div>

          <div class="row q-col-gutter-md q-mb-md items-center">
            <div class="col-12 col-sm-4">
              <q-checkbox v-model="paramForm.CHARGELAGERNUMMER" :label="t('auto.lagernummer_einbeziehen')" dense
                          color="secondary"/>
            </div>
            <div class="col-12 col-sm-4">
              <q-checkbox v-model="paramForm.CHARGEPREFIXHERDENNUMMER" :label="t('auto.herdennummer_einbeziehen')" dense
                        color="secondary"/>
            </div>
            <div class="col-12 col-sm-4">
              <q-checkbox v-model="paramForm.CHARGEDATUM" :label="t('auto.datum_einbeziehen')" dense color="secondary"/>
            </div>
          </div>

          <div class="row q-col-gutter-md items-center q-mb-md">
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGEJUMBOS" :label="t('auto.jumbos')" dense color="secondary"/>
            </div>
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGEXL" label="XL" dense color="secondary"/>
            </div>
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGELARGE" :label="t('auto.large')" dense color="secondary"/>
            </div>
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGEMEDIUM" :label="t('auto.medium')" dense color="secondary"/>
            </div>
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGESMALL" :label="t('auto.small')" dense color="secondary"/>
            </div>
            <div class="col-auto">
              <q-checkbox v-model="paramForm.CHARGEVOLLEI" label="Vollei" dense color="secondary"/>
            </div>
          </div>

          <q-separator class="q-my-lg" />

          <!-- Optionen unten im 2/3-Spalten Raster -->
          <div class="text-subtitle2 q-mb-sm text-secondary text-weight-bold">{{ t('auto.optionen') }}</div>
          <div class="row q-col-gutter-sm">
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.JUMBOS" :label="t('auto.jumbos_erfassen')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.KLASSENERFASSEN"
                :label="t('auto.gewichtsklassen')"
                dense
                color="secondary"
                @update:model-value="val => val && (paramForm.KLASSEAERFASSEN = false)"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.KLASSEAERFASSEN"
                :label="t('auto.klassea_erfassen')"
                dense
                color="secondary"
                @update:model-value="val => val && (paramForm.KLASSENERFASSEN = false, paramForm.KLASSEAERRECHNEN = false)"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.KLASSEAERRECHNEN"
                :label="t('auto.klassea_errechnen')"
                dense
                color="secondary"
                @update:model-value="val => val && (paramForm.KLASSEAERFASSEN = false)"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.KLASSEAVERMITTELN"
                :label="t('auto.klassea_vermitteln')"
                dense
                color="secondary"
                @update:model-value="onKlasseAVermittelnChanged"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.ERFASSESCHMUTZEI" :label="t('auto.schmutzeier')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.ERFASSEKNICKEI" :label="t('auto.knickeier')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.ERFASSEBRUCHEI" :label="t('auto.brucheier')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.ERFASSEVOLLEI" :label="t('auto.vollei_stueck')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.ERFASSEVOLLEIKG" label="Vollei (kg)" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.AUFTEILUNGGEWICHT"
                :label="t('auto.aufteilung_gewicht')"
                dense
                color="secondary"
                @update:model-value="val => val && (paramForm.AUFTEILUNGALTER = false)"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.AUFTEILUNGALTER"
                :label="t('auto.aufteilung_alter')"
                dense
                color="secondary"
                @update:model-value="val => val && (paramForm.AUFTEILUNGGEWICHT = false)"
              />
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.KONTROLLWIEGUNG" :label="t('auto.kontrollwiegung')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.VERLUSTEBEIBUCHUNG" :label="t('auto.verluste_direkt_bei_leistungsbuchung_erf')"
                        dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.LAGERBUCHUNGBEIBUCHUNG"
                        :label="t('auto.lagerbuchungen_automatisch_bei_leistung_')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.BEIVERMITTELNDATUMAKTUELL" :label="t('auto.aktuelles_datum_vorschlagen')" dense
                        color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox v-model="paramForm.PSEUDOLAGER" :label="t('auto.pseudolager_erlauben')" dense color="secondary"/>
            </div>
            <div class="col-6 col-md-4">
              <q-checkbox
                v-model="paramForm.FUTTERINVENTUR"
                :label="t('auto.futterinventur')"
                dense
                color="secondary"
                @update:model-value="onFutterinventurChanged"
              />
            </div>
          </div>
        </q-form>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, reactive, computed, watch, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { api } from 'src/boot/api';
import { useSessionStore } from '../stores/session';

const $q = useQuasar();
const sessionStore = useSessionStore();
const session = sessionStore;
const authRequired = ref(true);

const companyId = ref<number | null>(null);
const companyForm = reactive({
  NAME: '',
  FIRMA: '',
  STRASSE: '',
  PLZ: '',
  ORT: '',
  MOBILTELEPHON: '',
  EMAIL: '',
  FOTO: '',
  ID_TEXTE: 0,
  ID_ANREDE: 0
});
const originalCompany = ref<string>('');
const isCompanyDirty = computed(() => JSON.stringify(companyForm) !== originalCompany.value);
const photoFile = ref<File | null>(null);

const paramForm = reactive({
  ID_HERDEN: -1,
  KZ: '',
  MASSVOLLEI: '',
  ANZAHLKONTROLLW: 0,
  LAUFZEITWOCHEN: 0,
  PRODUKTIONSDAUER: 0,
  VERPACKUNGKG: 0,
  SCHLACHTERLOESHENNE: 0,
  AUFTEILUNGGEWICHT: false,
  BIOAUFSCHLAG: 0,
  JUMBOS: false,
  KLASSENERFASSEN: false,
  KLASSEAERFASSEN: false,
  KLASSEAERRECHNEN: false,
  KLASSEAVERMITTELN: false,
  ERFASSESCHMUTZEI: false,
  ERFASSEKNICKEI: false,
  ERFASSEBRUCHEI: false,
  ERFASSEVOLLEI: false,
  KONTROLLWIEGUNG: false,
  ERFASSEVOLLEIKG: false,
  AUFTEILUNGALTER: false,
  ID_TABELLEALTER: null as number | null,
  ID_TABELLEGEWICHT: null as number | null,
  LEGEBEGINN_LW: 18,
  VERLUSTEBEIBUCHUNG: false,
  LAGERBUCHUNGBEIBUCHUNG: false,
  MAXTAGEVERMITTELN: 0,
  CHARGEJUMBOS: false,
  CHARGEXL: false,
  CHARGELARGE: false,
  CHARGEMEDIUM: false,
  CHARGESMALL: false,
  CHARGEVOLLEI: false,
  CHARGEPREFIXFIRMA: '',
  CHARGEPREFIXHERDENNUMMER: false,
  CHARGEDATUM: false,
  CHARGELAGERNUMMER: false,
  CHARGETRENNUNG: '-',
  BEIVERMITTELNDATUMAKTUELL: false,
  PSEUDOLAGER: false,
  BIO: false,
  HALTUNGSTYP: '3',
  FUTTERINVENTUR: false
});
const originalParams = ref<string>('');
const isParamsDirty = computed(() => JSON.stringify(paramForm) !== originalParams.value);

const alterTabellenOptions = ref<{tabellennummer: number, bezeichnung: string}[]>([]);
const gewichtTabellenOptions = ref<{tabellennummer: number, bezeichnung: string}[]>([]);
const haltungstypOptions = ref<{ label: string, value: string }[]>([]);


const companyCancelBtn = ref<{ $el: HTMLElement } | null>(null);
const companySaveBtn = ref<{ $el: HTMLElement } | null>(null);
const paramsCancelBtn = ref<{ $el: HTMLElement } | null>(null);
const paramsSaveBtn = ref<{ $el: HTMLElement } | null>(null);

watch(isCompanyDirty, (dirty: boolean) => {
  if (dirty && (document.activeElement === (companyCancelBtn.value)?.$el || document.activeElement === document.body)) {
    (companySaveBtn.value)?.$el?.focus();
  }
});

watch(isParamsDirty, (dirty: boolean) => {
  if (dirty && (document.activeElement === (paramsCancelBtn.value)?.$el || document.activeElement === document.body)) {
    (paramsSaveBtn.value)?.$el?.focus();
  }
});

onMounted(() => {
  void (async () => {
    await Promise.all([loadCompany(), loadParams(), fetchTabellen(), loadAuthSetting()]);
  })();
});

async function loadAuthSetting() {
  try {
    const res = await api.get('/api/system-settings/auth_required');
    const val = res.data.value;
    authRequired.value = (val === true || val === 'true' || val === '1');
  } catch (err) {
    // Falls noch nicht in DB, Fallback auf aktuellen Config-Status
    try {
      const config = await api.get('/api/config');
      authRequired.value = config.data.auth_enabled;
    } catch (e) {
      authRequired.value = true;
    }
  }
}

async function saveAuthSetting(val: boolean) {
  try {
    await api.post('/api/system-settings', {
      name: 'auth_required',
      value: val ? 'true' : 'false'
    });
    $q.notify({
      type: 'positive',
      message: val ? 'Login-Pflicht aktiviert' : 'Login-Pflicht deaktiviert (Auto-Login aktiv)',
      caption: 'Wird beim nächsten App-Start wirksam',
      icon: 'security'
    });
  } catch (err) {
    $q.notify({type: 'negative', message: 'Fehler beim Speichern der Auth-Einstellung'});
    authRequired.value = !val; // Rollback
  }
}
const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  return Math.floor(Number(val)) || 0;
};

const extractFloat = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Float64' in val) return Number(val.Float64) || 0;
  if (typeof val === 'string') {
    return parseFloat(val.replace(',', '.')) || 0;
  }
  return Number(val) || 0;
};

async function fetchTabellen() {
  try {
    const [alterRes, gewichtRes, haltungRes] = await Promise.all([
      api.get('/api/tabellenkopf/typ/A'),
      api.get('/api/tabellenkopf/typ/G'),
      api.get('/api/texte/typ/H')
    ]);
    alterTabellenOptions.value = (alterRes.data as any[] || []).map((o) => ({
      ID: extractInt(o.ID || o.id),
      BEZEICHNUNG: o.BEZEICHNUNG || o.bezeichnung || ''
    }));
    gewichtTabellenOptions.value = (gewichtRes.data as any[] || []).map((o) => ({
      ID: extractInt(o.ID || o.id),
      BEZEICHNUNG: o.BEZEICHNUNG || o.bezeichnung || ''
    }));
    haltungstypOptions.value = (haltungRes.data as {
      kz: string,
      inhalt?: string,
      betreff?: string
    }[] || []).map((o) => ({
      label: o.kz + ' - ' + (o.inhalt || o.betreff || ''),
      value: o.kz
    }));
  } catch (err) {
    console.error('Fehler beim Laden der Tabellen-Lookups', err);
  }
}

async function loadCompany() {
  try {
    const res = await api.get('/api/company/person');
    companyId.value = res.data.ID;
    const data = {
      NAME: res.data.NAME || '',
      FIRMA: res.data.FIRMA || '',
      STRASSE: res.data.STRASSE || '',
      PLZ: res.data.PLZ || '',
      ORT: res.data.ORT || '',
      MOBILTELEPHON: res.data.MOBILTELEPHON || '',
      EMAIL: res.data.EMAIL || '',
      FOTO: res.data.FOTO || '',
      ID_TEXTE: res.data.ID_TEXTE || 0,
      ID_ANREDE: res.data.ID_ANREDE || 0
    };
    Object.assign(companyForm, data); originalCompany.value = JSON.stringify(data);
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Firmendaten' });
  }
}
async function loadParams() {
  try {
    const res = await api.get('/api/company/params');
    const d = res.data;
    const data = {
      ID_HERDEN: -1,
      KZ: typeof (d.KZ || d.kz) === 'string' ? (d.KZ || d.kz) : (d.KZ || d.kz || 'F'),
      MASSVOLLEI: String(d.MASSVOLLEI || d.massvollei || ''),
      ANZAHLKONTROLLW: extractInt(d.ANZAHLKONTROLLW || d.anzahlkontrollw),
      LAUFZEITWOCHEN: extractInt(d.LAUFZEITWOCHEN || d.laufzeitwochen),
      PRODUKTIONSDAUER: extractInt(d.PRODUKTIONSDAUER || d.produktionsdauer),
      VERPACKUNGKG: extractFloat(d.VERPACKUNGKG || d.verpackungkg),
      SCHLACHTERLOESHENNE: extractFloat(d.SCHLACHTERLOESHENNE || d.schlachterloeshenne),
      AUFTEILUNGGEWICHT: (extractInt(d.AUFTEILUNGGEWICHT || d.aufteilunggewicht) === 1),
      JUMBOS: (extractInt(d.JUMBOS || d.jumbos) === 1),
      KLASSENERFASSEN: (extractInt(d.KLASSENERFASSEN || d.klassenerfassen) === 1),
      KLASSEAERFASSEN: (extractInt(d.KLASSEAERFASSEN || d.klasseaerfassen) === 1),
      KLASSEAERRECHNEN: (extractInt(d.KLASSEAERRECHNEN || d.klasseaerrechnen) === 1),
      KLASSEAVERMITTELN: (extractInt(d.KLASSEAVERMITTELN || d.klasseavermitteln) === 1),
      ERFASSESCHMUTZEI: (extractInt(d.ERFASSESCHMUTZEI || d.erfasseschmutzei) === 1),
      ERFASSEKNICKEI: (extractInt(d.ERFASSEKNICKEI || d.erfasseknickei) === 1),
      ERFASSEBRUCHEI: (extractInt(d.ERFASSEBRUCHEI || d.erfassebruchei) === 1),
      ERFASSEVOLLEI: (extractInt(d.ERFASSEVOLLEI || d.erfassevollei) === 1),
      KONTROLLWIEGUNG: (extractInt(d.KONTROLLWIEGUNG || d.kontrollwiegung) === 1),
      ERFASSEVOLLEIKG: (extractInt(d.ERFASSEVOLLEIKG || d.erfassevolleikg) === 1),
      AUFTEILUNGALTER: (extractInt(d.AUFTEILUNGALTER || d.aufteilungalter) === 1),
      ID_TABELLEALTER: extractInt(d.ID_TABELLEALTER || d.id_tabellealter) || null,
      ID_TABELLEGEWICHT: extractInt(d.ID_TABELLEGEWICHT || d.id_tabellegewicht) || null,
      LEGEBEGINN_LW: extractInt(d.LEGEBEGINN_LW || d.legebeginn_lw) || 18,
      VERLUSTEBEIBUCHUNG: (extractInt(d.VERLUSTEBEIBUCHUNG || d.verlustebeibuchung) === 1),
      LAGERBUCHUNGBEIBUCHUNG: (extractInt(d.LAGERBUCHUNGBEIBUCHUNG || d.lagerbuchungbeibuchung) === 1),
      MAXTAGEVERMITTELN: extractInt(d.MAXTAGEVERMITTELN || d.maxtagevermitteln),
      CHARGEJUMBOS: (extractInt(d.CHARGEJUMBOS || d.chargejumbos) === 1),
      CHARGEXL: (extractInt(d.CHARGEXL || d.chargexl) === 1),
      CHARGELARGE: (extractInt(d.CHARGELARGE || d.chargelarge) === 1),
      CHARGEMEDIUM: (extractInt(d.CHARGEMEDIUM || d.chargemedium) === 1),
      CHARGESMALL: (extractInt(d.CHARGESMALL || d.chargesmall) === 1),
      CHARGEVOLLEI: (extractInt(d.CHARGEVOLLEI || d.chargevollei) === 1),
      CHARGEPREFIXFIRMA: typeof (d.CHARGEPREFIXFIRMA || d.chargeprefixfirma) === 'string' ? (d.CHARGEPREFIXFIRMA || d.chargeprefixfirma) : (d.CHARGEPREFIXFIRMA || d.chargeprefixfirma || ''),
      CHARGEPREFIXHERDENNUMMER: (extractInt(d.CHARGEPREFIXHERDENNUMMER || d.chargeprefixherdennummer) === 1),
      CHARGEDATUM: (extractInt(d.CHARGEDATUM || d.chargedatum) === 1),
      CHARGELAGERNUMMER: (extractInt(d.CHARGELAGERNUMMER || d.chargelagernummer) === 1),
      CHARGETRENNUNG: typeof (d.CHARGETRENNUNG || d.chargetrennung) === 'string' ? (d.CHARGETRENNUNG || d.chargetrennung) : (d.CHARGETRENNUNG || d.chargetrennung || '-'),
      BEIVERMITTELNDATUMAKTUELL: (extractInt(d.BEIVERMITTELNDATUMAKTUELL || d.beivermittelndatumaktuell) === 1),
      PSEUDOLAGER: (extractInt(d.PSEUDOLAGER || d.pseudolager) === 1),
      BIO: (extractInt(d.BIO || d.bio) === 1),
      HALTUNGSTYP: typeof (d.HALTUNGSTYP || d.haltungstyp) === 'string' ? (d.HALTUNGSTYP || d.haltungstyp) : ((d.HALTUNGSTYP || d.haltungstyp) || '3'),
      BIOAUFSCHLAG: Number(d.BIOAUFSCHLAG || d.bioaufschlag) || 0,
      FUTTERINVENTUR: (extractInt(d.FUTTERINVENTUR || d.futterinventur) === 1)
    };
    Object.assign(paramForm, data);
    // Ensure all fields are in the baseline for comparison, including unchanged ones like id_herden
    originalParams.value = JSON.stringify(paramForm);
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Parameter' });
  }
}
function resetCompany() { Object.assign(companyForm, JSON.parse(originalCompany.value)); }
function resetParams() { Object.assign(paramForm, JSON.parse(originalParams.value)); }
function onPhotoSelected(file: File | null) {
  if (!file) return;
  const reader = new FileReader();
  reader.onload = (e) => {
    companyForm.FOTO = e.target?.result as string;
  };
  reader.readAsDataURL(file);
}

function onFutterinventurChanged(val: boolean) {
  if (val && !paramForm.KLASSEAVERMITTELN) {
    $q.dialog({
      title: 'KlasseA vermitteln erforderlich',
      message: 'Die Futterinventur kann nur aktiviert werden, wenn "KlasseA vermitteln" ebenfalls aktiv ist. Möchten Sie "KlasseA vermitteln" jetzt aktivieren?',
      cancel: {
        label: 'Abbrechen',
        flat: true,
        color: 'grey-7'
      },
      ok: {
        label: 'Aktivieren',
        color: 'primary',
        unelevated: true
      },
      persistent: true
    }).onOk(() => {
      paramForm.KLASSEAVERMITTELN = true;
    }).onCancel(() => {
      paramForm.FUTTERINVENTUR = false;
    });
  }
}

function onKlasseAVermittelnChanged(val: boolean) {
  if (!val && paramForm.FUTTERINVENTUR) {
    paramForm.FUTTERINVENTUR = false;
    $q.notify({
      type: 'info',
      message: 'Da "KlasseA vermitteln" deaktiviert wurde, wurde auch die "Futterinventur" deaktiviert.',
      icon: 'info',
      timeout: 4000
    });
  }
}

async function saveCompany() {
  try {
    await api.put(`/api/person/${companyId.value}`, {
      ...companyForm,
      KZ: 'F',
      FOTO: companyForm.FOTO.includes('base64,') ? companyForm.FOTO.split('base64,')[1] : companyForm.FOTO
    });
    originalCompany.value = JSON.stringify(companyForm);
  } catch (_err: unknown) { /**/
  }
}
async function saveParams() {
  try {
    const payload: any = {
      ...paramForm,
      ID_HERDEN: -1,
      KZ: 'F'
    };

    Object.keys(payload).forEach(targetKey => {
      let val = payload[targetKey];

      // Convert Booleans to 0/1
      if (typeof val === 'boolean') {
        val = val ? 1 : 0;
      }

      const numericFields = [
        'MASSVOLLEI', 'ANZAHLKONTROLLW', 'VERPACKUNGKG', 'LAUFZEITWOCHEN',
        'SCHLACHTERLOESHENNE', 'PRODUKTIONSDAUER', 'ID_TABELLEALTER',
        'ID_TABELLEGEWICHT', 'MAXTAGEVERMITTELN', 'LEGEBEGINN_LW', 'JUMBOS'
      ];

      if (numericFields.includes(targetKey)) {
        if (val === '' || val === null || val === undefined) {
          val = null;
        } else {
          val = Number(val);
        }
      }

      if (targetKey === 'BIOAUFSCHLAG') {
        val = Number(val) || 0;
      }

      payload[targetKey] = val;
    });

    await api.put(`/api/firmenparameter/-1`, payload);
    originalParams.value = JSON.stringify(paramForm);
    $q.notify({ type: 'positive', message: 'Parameter gespeichert' });
  } catch (_err: unknown) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}
</script>

<style scoped>
.rounded-borders { border-radius: 16px; }
.opacity-0 { opacity: 0; }
</style>

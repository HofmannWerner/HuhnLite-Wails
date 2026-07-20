<template>
  <div class="q-pa-md">
    <!-- Header with Action on Left, Title on Right -->
    <div class="row items-center justify-between q-mb-md">
      <div class="row q-gutter-md">
        <q-btn color="primary" icon="add" :label="t('auto.neue_buchung')" @click="openCreate" rounded unelevated />
      </div>
      <div class="text-h6 text-primary">{{ t('auto.leistungs_buchung') }}</div>
    </div>

    <div class="row q-col-gutter-md q-mb-md items-center">
      <div class="col-12 col-sm-4 col-md-3">
        <q-select
          v-model="filterHerde"
          :options="filteredHerdeDropdownOptions"
          option-value="ID"
          option-label="BEZEICHNUNG"
          emit-value
          map-options
          clearable
          :label="t('auto.herde_filtern')"
          filled
          stack-label
          :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
          :dark="$q.dark.isActive"
        >
          <template v-slot:prepend>
            <q-icon name="filter_list" />
          </template>
        </q-select>
      </div>
      <div class="col-auto">
        <q-checkbox v-model="nurAktiveFilter" :label="t('auto.nur_aktive_herden')" color="positive" :dark="$q.dark.isActive" />
      </div>

      <div class="col-12 col-sm-4 col-md-3">
        <q-input filled v-model="filterDateRangeText" :label="t('auto.zeitraum')" stack-label readonly dense :dark="$q.dark.isActive" :bg-color="$q.dark.isActive ? 'grey-9' : undefined">
          <template v-slot:prepend>
            <q-icon name="event" class="cursor-pointer">
              <q-popup-proxy cover transition-show="scale" transition-hide="scale">
                <q-date v-model="filterDateRange" range :dark="$q.dark.isActive">
                  <div class="row items-center justify-end">
                    <q-btn v-close-popup :label="t('form.close')" color="primary" flat />
                  </div>
                </q-date>
              </q-popup-proxy>
            </q-icon>
          </template>
        </q-input>
      </div>
    </div>

    <!-- Table -->
    <q-table
      :rows="filteredRows"
      :columns="columns"
      row-key="ID"
      :loading="loading"
      :pagination="pagination"
      @update:pagination="(val: any) => { pagination = val }"
      class="huhnlite-grid-standard resizable-table q-mb-lg shadow-2 cursor-pointer"
      :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
      :dark="$q.dark.isActive"
      table-header-class="text-weight-bold"
      separator="cell"
      @row-dblclick="(evt, row) => onEdit(row)"
    >

      <!-- Resizable Header Cells -->
      <template v-slot:header-cell="props">
        <q-th :props="props" 
              class="resizable-column" 
              :style="{ width: (columnWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
          <div class="ellipsis">{{ props.col.label }}</div>
          <div class="resizer" 
               :class="{ 'is-resizing': isResizing === props.col.name }"
               @pointerdown.stop.prevent.capture="startResize($event, props.col.name)">
          </div>
        </q-th>
      </template>

      <!-- Body mit Zeilen-Highlighting -->
      <template v-slot:body="props">
        <q-tr :props="props" @dblclick="onEdit(props.row)"
              :class="{
                'vermittelt-row': extractString(props.row.VERMITTELT) === 'V',
                'source-row': extractString(props.row.VERMITTELT) === 'S'
              }">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <!-- Aktionen -->
            <template v-if="col.name === 'actions'">
              <div class="row no-wrap q-gutter-x-xs justify-center">
                <q-btn dense round icon="edit" color="primary" @click="onEdit(props.row)" unelevated size="sm"/>
                <q-btn dense round icon="delete" color="negative" @click="onDelete(props.row)" unelevated size="sm"/>
              </div>
            </template>

            <!-- Vermittelt-Flag -->
            <template v-else-if="col.name === 'VERMITTELT'">
              <div v-if="extractString(props.row.VERMITTELT) === 'V'" class="text-weight-bold text-primary">
                V
                <q-icon name="info" size="xs" color="grey-6" class="q-ml-xs">
                  <q-tooltip>{{ t('auto.automatisch_vermittelte_buchung_verteilt') }}</q-tooltip>
                </q-icon>
              </div>
              <div v-else-if="extractString(props.row.VERMITTELT) === 'S'" class="text-weight-bold text-amber-8">
                {{ t('auto.s') }}
                <q-icon name="history" size="xs" color="amber-7" class="q-ml-xs">
                  <q-tooltip>{{ t('auto.original_sammelsatz_referenz_wird_nicht_') }}</q-tooltip>
                </q-icon>
              </div>
              <div v-else class="text-grey-5">{{ t('auto.n') }}</div>
            </template>

            <!-- Standard-Werte mit Tausender-Trennzeichen -->
            <template v-else>
              <template v-if="['KONTROLLGEWICHT', 'DGEWICHTEI', 'EIMASSE', 'VERPACKUNG', 'VERPACKUNGKG', 'KONTROLLWIEGUNG'].includes(col.name)">
                {{ formatWeight(col.value) }}
              </template>
              <template v-else-if="typeof col.value === 'number'">
                {{ Number.isInteger(col.value) ? col.value.toLocaleString('de-DE') : col.value.toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
              </template>
              <template v-else>
                {{ col.value }}
              </template>
            </template>
          </q-td>
        </q-tr>
      </template>
    </q-table>

    <!-- Dialog Form -->
    <q-dialog v-model="showDialog" persistent full-width @show="onDialogShow">
      <q-card style="max-width: 1000px; margin: auto; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? 'Buchung bearbeiten' : 'Neue Buchung' }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-y-lg">

            <!-- Neuer Form Header für Datum & Uhrzeit (nur bei Neuanlage) -->
            <FormHeader v-if="!isEditing" v-model="form.fullTimestamp" />

            <!-- Stammdaten der Buchung -->
            <div class="row q-col-gutter-md">
              <div class="col-12 col-sm-4">
                <q-select
                  v-model="form.ID_HERDEN"
                  ref="herdeSelect"
                  :options="selectableHerdeOptions"
                  option-value="ID"
                  option-label="BEZEICHNUNG"
                  emit-value
                  map-options
                  :label="t('auto.herde')"
                  filled
                  stack-label
                  :readonly="isEditing"
                  hide-bottom-space
                  class="col"
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :rules="[val => !!val || 'Bitte eine Herde auswählen']"
                  @update:model-value="onHerdeChange"
                />
              </div>
              <div class="col-auto" style="width: 180px">
                <q-input
                  v-model="bookingDateOnly"
                  type="date"
                  :label="t('auto.buchungsdatum')"
                  filled
                  stack-label
                  :readonly="isEditing"
                  hide-bottom-space
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :rules="[
                    val => !!val || 'Erforderlich',
                    val => new Date(val).getTime() <= new Date().setHours(23,59,59,999) || 'Datum darf nicht in der Zukunft liegen',
                    val => {
                      const herde = herdeOptions.find((h: any) => h.ID === form.ID_HERDEN);
                      if (!herde || !herde.LEGEDATUM) return true;
                      return new Date(val).getTime() >= new Date(herde.LEGEDATUM).getTime() || `Datum liegt vor Legebeginn (${date.formatDate(herde.LEGEDATUM, 'DD.MM.YYYY')})`;
                    }
                  ]"
                />
              </div>
              <div class="col-auto" style="width: 80px">
                <q-input :model-value="calculatedLegewoche === '-' ? '' : calculatedLegewoche" :label="t('auto.lw')" filled stack-label readonly hide-bottom-space :bg-color="$q.dark.isActive ? 'grey-9' : undefined">
                  <q-tooltip>{{ t('auto.leistungs_woche_seit_legebeginn') }}</q-tooltip>
                </q-input>
              </div>
              <div class="col-auto" style="width: 80px">
                <q-input :model-value="calculatedAlterswoche === '-' ? '' : calculatedAlterswoche" :label="t('auto.aw')" filled stack-label readonly hide-bottom-space :bg-color="$q.dark.isActive ? 'grey-9' : undefined">
                   <q-tooltip>{{ t('auto.alter_in_wochen') }}</q-tooltip>
                </q-input>
              </div>
              <div class="col-auto" style="width: 100px">
                <q-input v-model.number="form.TIERBESTAND" type="number" :label="t('auto.tierbestand')" filled stack-label readonly hide-bottom-space :bg-color="$q.dark.isActive ? 'grey-9' : undefined">
                   <q-tooltip>{{ t('auto.aktueller_hennenbestand') }}</q-tooltip>
                </q-input>
              </div>
            </div>

            <!-- Vermittlungs-Hinweis -->
            <div v-if="vermittlungDays > 1"
                 class="q-pa-md bg-info text-white rounded-borders q-mb-md flex items-center shadow-1">
              <q-icon name="info" size="sm" class="q-mr-sm"/>
              <div>
                <div class="text-weight-bold">
                  {{ vermittlungDays === 100 ? 'Vermittelte Buchung' : 'Vermittlung erforderlich' }}
                </div>
                <div class="text-caption">
                  <template v-if="vermittlungDays === 100">
                    {{ t('auto.diese_buchung_war_teil_einer_automatisch') }}
                  </template>
                  <template v-else>
                    Es besteht eine Lücke von {{ vermittlungDays }} Tagen zur letzten Buchung.
                    Die Mengen werden gleichmäßig auf diesen Zeitraum verteilt.
                  </template>
                </div>
              </div>
            </div>

            <q-separator />

            <div class="text-subtitle1 text-weight-bold text-primary q-mb-sm">{{ t('auto.leistung_gewicht') }}</div>
            <div class="row q-col-gutter-sm q-mb-md">
              <div class="col-auto" style="width: 140px">
                <q-input
                  v-model.number="form.GEWICHTPROBE"
                  type="number"
                  :label="t('auto.gewichtsprobe')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  step="any"
                  readonly
                />
              </div>
              <div class="col-auto" style="width: 140px">
                <q-input
                  v-model="displayKontrollgewicht"
                  :label="t('auto.kontrollgew_kg')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  @blur="onKontrollgewichtBlur"
                />
              </div>
              <div class="col-auto" style="width: 140px">
                <q-input
                  v-model="displayVerpackung"
                  :label="t('auto.verpackung_kg')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  @blur="onVerpackungBlur"
                />
              </div>
              <div class="col-6 col-sm-4 col-md-2">
                <q-input
                  ref="klasseAInput"
                  v-model.number="form.KLASSEA"
                  type="number"
                  :label="t('auto.klasse_a')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :readonly="!getParamBool('klasseaerfassen') && !getParamBool('aufteilunggewicht') && !getParamBool('aufteilungalter')"
                  @update:model-value="onKlasseAInput"
                  :rules="[validateEggs]"
                  hide-bottom-space
                />
              </div>

              <div class="col-auto" style="width: 140px">
                <q-input
                  v-model="displayVollei"
                  type="text"
                  :label="t('auto.vollei_kg')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :readonly="!getParamBool('erfassevolleikg') && !getParamBool('erfassevollei')"
                  @blur="onVolleiBlur"
                />
              </div>

              <div class="col-auto" style="width: 140px">
                <q-input
                  v-model="displayDgewicht"
                  type="text"
                  :label="t('auto.eigewicht_g')"
                  filled
                  stack-label
                  dense
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  readonly
                />
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-12">
                <div class="row q-col-gutter-sm">
                  <div class="col-auto" style="width: 140px" v-for="size in (['XL', 'LARGE', 'MEDIUM', 'SMALL'] as const)" :key="size">
                    <q-input
                      v-model.number="(form as any)[size]"
                      type="number"
                      :label="size.toUpperCase()"
                      filled
                      stack-label
                      dense
                      :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                      :readonly="!getParamBool(size === 'xl' ? 'chargexl' : size === 'large' ? 'chargelarge' : size === 'medium' ? 'chargemedium' : 'chargesmall') && !getParamBool('klassenerfassen')"
                      :rules="[validateEggs]"
                      hide-bottom-space
                    />
                  </div>
                </div>
              </div>
              <div class="col-12 q-mt-sm">
                <div class="row q-col-gutter-sm">
                  <div class="col-auto" style="width: 140px" v-for="size in (['SCHMUTZ', 'KNICKEIER', 'BRUCHEIER'] as const)" :key="size">
                    <q-input
                      v-model.number="(form as any)[size]"
                      type="number"
                      :label="size.charAt(0).toUpperCase() + size.slice(1)"
                      filled
                      stack-label
                      dense
                      :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                      :readonly="!getParamBool('erfasse' + size + (size === 'schmutz' ? 'ei' : ''))"
                      :rules="[validateEggs]"
                      hide-bottom-space
                    />
                  </div>
                </div>
              </div>
            </div>

            <q-separator />

            <!-- Eilager Indicators -->
            <div class="row q-col-gutter-sm q-mt-md" v-if="editId">
              <div class="col-12 text-subtitle2 text-grey-8">{{ t('auto.bereits_verbuchte_mengen_eilager_pseudo') }}</div>
              <div class="col-2" v-for="s in (['JUMBOS', 'XL', 'LARGE', 'MEDIUM', 'SMALL'] as const)" :key="s">
                <div class="bg-blue-1 text-blue-9 q-pa-xs rounded-borders text-center text-caption shadow-1">
                  <span class="text-weight-bold">{{ s.substring(0, 1) }}:</span> {{ (usedStock as any)[s] || 0 }}
                </div>
              </div>
              <div class="col-2" v-if="getParamBool('vollei')">
                <div class="bg-blue-1 text-blue-9 q-pa-xs rounded-borders text-center text-caption shadow-1">
                  <span class="text-weight-bold">{{ t('auto.v') }}</span> {{ usedStock.VOLLEIKG || 0 }}
                </div>
              </div>
            </div>

            <div class="row justify-between items-center q-mt-lg q-pt-md border-top-grey-3">
              <div class="row q-gutter-sm">
                <q-btn :label="t('auto.ins_eilager_uebertragen')" color="warning" icon="warehouse" @click="triggerLagerBuchung(false)"
                       v-if="!getParamBool('LAGERBUCHUNGBEIBUCHUNG')"
                       :disable="!((computedKlasseA || 0) > 0) || hasEggOverflow"
                       rounded unelevated>
                  <q-tooltip v-if="hasEggOverflow">{{ t('auto.menge_uebersteigt_tierbestand') }}</q-tooltip>
                  <q-tooltip v-else-if="!((computedKlasseA || 0) > 0)">
                    Deaktiviert, da keine Eier-Mengen erfasst wurden
                  </q-tooltip>
                </q-btn>

                <q-btn :label="t('auto.verluste_buchen')" color="negative" icon="remove_circle" @click="openVerlustDialog" :disable="!form.ID_HERDEN" rounded unelevated />

                <q-btn :label="t('auto.anderweitige_verwendung')" color="blue" icon="swap_horiz" @click="triggerLagerBuchung(true)"
                       :disable="isVermittelt || hasEggOverflow"
                       v-if="getParamBool('pseudolager')" rounded unelevated>
                  <q-tooltip v-if="hasEggOverflow">{{ t('auto.menge_uebersteigt_tierbestand') }}</q-tooltip>
                  <q-tooltip v-else-if="isVermittelt">
                    {{ t('auto.deaktiviert_bei_vermittelten_buchungen') }}
                  </q-tooltip>
                </q-btn>
              </div>
              <div class="row q-gutter-sm">
                <q-btn :label="t('form.cancel')" color="negative" outline @click="closeDialog" rounded padding="xs lg" />
                <q-btn :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated padding="xs xl" />
              </div>
            </div>

          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Sub-Dialog for Eilager Storage -->
    <q-dialog v-model="showEilagerDialog" persistent>
      <q-card style="min-width: 500px; border-radius: 16px;">
        <q-card-section class="bg-warning text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isPseudoBooking ? 'Anderweitige Verwendung' : 'Übertrag an Eierlager' }}</div>
          <div class="text-caption">{{ t('auto.charge_erstellen_und_bestand_einbuchen') }}</div>
        </q-card-section>

        <q-card-section class="q-pa-lg">
          <q-form @submit="submitLagerBuchung" class="q-gutter-y-md">

            <div class="row q-col-gutter-md">
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="eilagerForm.ID_EILAGER"
                  :options="eilagerList"
                  option-value="ID"
                  option-label="BEZEICHNUNG"
                  emit-value
                  map-options
                  :label="t('auto.ziel_lager_auswaehlen')"
                  filled stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                  :rules="[val => !!val || 'Bitte wählen Sie ein Lager aus']"
                  @update:model-value="onTargetLagerChange"
                />
              </div>
              <div class="col-12 col-sm-6" v-if="isPseudoBooking">
                <q-select
                  v-model="eilagerForm.BUCHUNGSTYP"
                  :options="verwendungsTexteOptions"
                  option-value="kz"
                  option-label="betreff"
                  emit-value
                  map-options
                  :label="t('auto.verwendungstext')"
                  filled stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                  :rules="[val => !!val || 'Bitte wählen Sie eine Verwendung aus']"
                />
              </div>
              <div class="col-12 col-sm-6">
                <q-select
                  v-model="eilagerForm.ID_FREMDESLAGER"
                  :options="filteredLagerplaetze"
                  option-value="ID"
                  option-label="BEZEICHNUNG"
                  emit-value
                  map-options
                  :label="t('auto.lagerplatz_ort')"
                  filled stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                  :rules="[val => !!val || 'Bitte wählen Sie einen Lagerplatz aus']"
                  :disable="!eilagerForm.ID_EILAGER"
                >
                  <template v-slot:no-option>
                    <q-item>
                      <q-item-section class="text-grey">
                        {{ t('auto.keine_plaetze_fuer_dieses_lager_gefunden') }}
                      </q-item-section>
                    </q-item>
                  </template>
                </q-select>
              </div>
            </div>

            <div class="bg-grey-2 q-pa-md rounded-borders q-mb-md" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'">
                <div class="text-subtitle2 q-mb-xs">{{ t('auto.aktueller_lagerbestand_referenz') }}</div>
                <div class="row q-col-gutter-sm text-caption">
                  <div class="col-2" v-for="s in (['JUMBOS', 'XL', 'LARGE', 'MEDIUM', 'SMALL'] as const)" :key="s">
                    <span class="text-weight-bold">{{ s.toUpperCase() }}:</span> {{ (vStock as any)[s] }}
                  </div>
                  <div class="col-2" v-if="getParamBool('vollei')">
                    <span class="text-weight-bold">{{ t('auto.vollei') }}</span> {{ vStock.VOLLEIKG }}kg
                  </div>
                </div>
            </div>

            <div class="q-pa-sm bg-blue-1 text-blue-9 rounded-borders q-mb-md">
              <div class="text-weight-bold q-mb-xs">{{ t('auto.verfuegbar_v_b_e') }}</div>
              <div class="row q-col-gutter-sm text-caption">
                <div class="col-4 col-sm-2" :class="v_jumbos < 0 ? 'text-negative text-weight-bold' : ''">
                  J: {{ v_jumbos }}
                </div>
                <div class="col-4 col-sm-2" :class="v_xl < 0 ? 'text-negative text-weight-bold' : ''">
                  XL: {{ v_xl }}
                </div>
                <div class="col-4 col-sm-2" :class="v_large < 0 ? 'text-negative text-weight-bold' : ''">
                  L: {{ v_large }}
                </div>
                <div class="col-4 col-sm-2" :class="v_medium < 0 ? 'text-negative text-weight-bold' : ''">
                  M: {{ v_medium }}
                </div>
                <div class="col-4 col-sm-2" :class="v_small < 0 ? 'text-negative text-weight-bold' : ''">
                  S: {{ v_small }}
                </div>
                <div class="col-4 col-sm-2" v-if="getParamBool('VOLLEI')"
                     :class="v_vollei < 0 ? 'text-negative text-weight-bold' : ''">
                  V: {{ v_vollei }}
                </div>
              </div>
              <div v-if="hasValidationError" class="text-negative text-weight-bold q-mt-xs">
                <q-icon name="warning"/>
                {{ t('auto.nicht_genuegend_eier_vorhanden') }}
              </div>
            </div>

            <div class="row q-col-gutter-sm">
              <div class="col-6 col-sm-4" v-for="s in (['JUMBOS', 'XL', 'LARGE', 'MEDIUM', 'SMALL'] as const)" :key="s">
                <q-input
                  v-model.number="(eilagerForm as any)[s]"
                  type="number"
                  :label="s.toUpperCase()"
                  filled dense stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :readonly="!getParamBool(s === 'jumbos' ? 'chargejumos' : (s === 'xl' ? 'chargexl' : (s === 'large' ? 'chargelarge' : (s === 'medium' ? 'chargemedium' : 'chargesmall'))))"
                  :error="(s === 'jumbos' ? v_jumbos : s === 'xl' ? v_xl : s === 'large' ? v_large : s === 'medium' ? v_medium : v_small) < 0"
                  hide-bottom-space
                />
              </div>
              <div class="col-6 col-sm-4" v-if="getParamBool('vollei')">
                <q-input
                  v-model.number="eilagerForm.VOLLEIKG"
                  type="number"
                  :label="t('auto.vollei_kg')"
                  filled dense stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : undefined"
                  :readonly="!getParamBool('charegevollei')"
                  :error="v_vollei < 0"
                  hide-bottom-space
                />
              </div>
            </div>

            <q-input
              v-model="eilagerForm.CHARGE"
              :label="t('auto.chargennummer_vorschlag')"
              filled
              stack-label
              :bg-color="$q.dark.isActive ? 'grey-9' : 'yellow-1'"
            >
              <template v-slot:append>
                <q-btn round flat icon="refresh" @click="generateChargeProposal" />
              </template>
            </q-input>

            <div class="row justify-end q-gutter-sm q-mt-lg">
              <q-btn :label="t('form.cancel')" color="negative" flat v-close-popup />
              <q-btn :label="t('auto.bestand_einbuchen')" color="warning" type="submit" unelevated :disable="hasValidationError"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Dialog Verluste buchen -->
    <q-dialog v-model="showVerlusteDialog" persistent>
      <q-card style="min-width: 350px; border-radius: 16px;">
        <q-card-section class="bg-negative text-white q-pa-md row items-center">
          <q-icon name="remove_circle" size="sm" class="q-mr-sm" />
          <div class="text-h6 text-weight-bold">{{ t('auto.verluste_buchen') }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup color="white" />
        </q-card-section>

        <q-card-section class="q-pa-lg column q-gutter-md">
          <div class="text-subtitle1 text-weight-bold text-primary">
            {{ herdeOptions.find(h => h.ID === verlustForm.ID_HERDEN)?.bezeichnung || 'Herde' }}
          </div>

          <q-input
            v-model="verlustForm.DATUM"
            label="Datum"
            type="date"
            filled
            stack-label
            readonly
          />

          <q-input
            v-model.number="verlustForm.VERLUSTE"
            :label="t('auto.anzahl_verendete_ausgeschiedene_tiere')"
            type="number"
            filled
            stack-label
            autofocus
            :rules="[val => val > 0 || 'Bitte geben Sie eine Zahl > 0 ein']"
          />

          <q-select
            v-model="verlustForm.ID_TEXTE"
            :options="verlustReasons"
            option-value="ID"
            option-label="BETREFF"
            :label="t('auto.grund_ursache')"
            filled
            emit-value
            map-options
            stack-label
            :rules="[val => !!val || 'Bitte wählen Sie einen Grund aus']"
          />

          <q-input
            v-model="verlustForm.MEMO"
            :label="t('auto.bemerkung')"
            filled
            stack-label
            type="textarea"
            rows="2"
          />
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md">
          <q-btn :label="t('form.cancel')" flat v-close-popup />
          <q-btn :label="t('form.save')" color="negative" rounded unelevated @click="submitVerlust" />
        </q-card-actions>
      </q-card>
    </q-dialog>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, reactive, onMounted, computed, watch, nextTick } from 'vue';
import { useQuasar, date } from 'quasar';
import { api } from '../boot/api';
import type { QTableProps } from 'quasar';
import FormHeader from '../components/FormHeader.vue';
import TierbewegungDialog from '../components/TierbewegungDialog.vue';
import { useSessionStore } from '../stores/session';
import { useResizableColumns } from '../composables/useResizableColumns';

/* eslint-disable @typescript-eslint/no-explicit-any */

const extractString = (val: any) => {
  if (val === null || val === undefined) return '';
  if (typeof val === 'object' && 'String' in val) return String(val.String);
  return String(val);
};;

const getInsensitive = (obj: any, key: string) => {
  if (!obj) return undefined;
  const foundKey = Object.keys(obj).find(k => k.toLowerCase() === key.toLowerCase());
  return foundKey ? obj[foundKey] : undefined;
};

const extractInt = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'object' && 'Int64' in val) return Number(val.Int64) || 0;
  if (typeof val === 'object' && 'Int32' in val) return Number(val.Int32) || 0;
  if (typeof val === 'string') {
    // Entfernt alle Tausender-Trenner (Punkte/Kommas), falls es eine Ganzzahl sein soll
    const cleaned = val.replace(/[^\d]/g, '');
    return parseInt(cleaned, 10) || 0;
  }
  return Math.floor(Number(val)) || 0;
};

const extractFloat = (val: any) => {
  if (val === null || val === undefined) return 0;
  if (typeof val === 'number') return val;
  if (typeof val === 'object' && 'Float64' in val) return Number(val.Float64) || 0;
  if (typeof val === 'string') {
    let s = val.trim();
    if (s.includes(',')) {
      s = s.replace(/\./g, '').replace(',', '.');
    }
    return parseFloat(s) || 0;
  }
  return Number(val) || 0;
};

const formatWeight = (val: any) => {
  const n = extractFloat(val);
  if (n === 0) return '0,00';
  return n.toFixed(2).replace('.', ',');
};

const formatNumberLocalized = (val: any, decimals = 2) => {
  if (val === null || val === undefined || val === '') return '';
  const num = Number(val);
  if (isNaN(num)) return '';
  return num.toLocaleString('de-DE', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  });
};

const parseNumberLocalized = (val: any) => {
  if (val === null || val === undefined || val === '') return 0;
  if (typeof val === 'number') return val;
  let s = String(val).trim();
  if (!s) return 0;
  // Deutsche Tausendertrennzeichen entfernen und Komma zu Punkt
  if (s.includes(',')) {
    s = s.replace(/\./g, '').replace(',', '.');
  }
  return parseFloat(s) || 0;
};

const extractBool = (val: any) => {
  if (val === null || val === undefined) return false;
  if (typeof val === 'boolean') return val;
  return val === 1 || val === '1' || val === 'true';
};

function addOneDay(dateStr: string): string {
  if (!dateStr || dateStr.startsWith('0001-01-01')) return '';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return '';
    d.setDate(d.getDate() + 1);
    return d.toISOString().split('T')[0];
  } catch {
    return '';
  }
}

const $q = useQuasar();
const sessionStore = useSessionStore();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Leistung');

const loading = ref(false);
const rows = ref<Record<string, unknown>[]>([]);

interface HerdeOption {
  ID: number;
  BEZEICHNUNG: string;
  LEGEDATUM?: string;
  GEBURTSDATUM?: string;
  ID_EILAGER?: number;
  AKTIV?: number;
}

const herdeOptions = ref<HerdeOption[]>([]);
const siloOptions = ref<any[]>([]);
const filterHerde = ref<number | null>(null);
const globalParams = ref<any>(null);
const herdParams = ref<any>(null);
const activeParameters = computed(() => {
  if (herdParams.value && Object.keys(herdParams.value).length > 0) {
    return herdParams.value;
  }
  return globalParams.value;
});

const selectableHerdeOptions = computed(() => {
  if (isEditing.value) {
    return herdeOptions.value;
  }
  const filtered = herdeOptions.value.filter(h => extractInt(h.AKTIV) === 1);
  return filtered;
});

const lastSelectedHerdeId = ref<number | null>(null);
const nurAktiveFilter = ref(true);
const filteredHerdeDropdownOptions = computed(() => {
  if (!nurAktiveFilter.value) return herdeOptions.value;
  return herdeOptions.value.filter(h => extractInt(h.AKTIV) === 1);
});

const filterDateRange = ref<{ from: string; to: string } | null>(null);
const filterDateRangeText = computed(() => {
  if (!filterDateRange.value) return '';
  const from = filterDateRange.value.from.split('/').reverse().join('.');
  const to = filterDateRange.value.to.split('/').reverse().join('.');
  return `${from} - ${to}`;
});

// Default 7 Tage
const today = new Date();
const lastWeek = new Date();
lastWeek.setDate(today.getDate() - 7);

// onMounted moved to end of script

// Lagerlisten
const isPseudoBooking = ref(false);
const showEilagerDialog = ref(false);
const fullEilagerRaw = ref<any[]>([]);
const eilagerList = computed(() => {
  if (isPseudoBooking.value) {
    return fullEilagerRaw.value
      .filter(l => extractString(l.KZ) !== 'E')
      .map(l => ({
        ID: l.ID,
        BEZEICHNUNG: `${l.LAGERNUMMER || '?'} - ${extractString(l.BEZEICHNUNG) || 'Eierlager'}`
      }));
  } else {
    return fullEilagerRaw.value
      .filter(l => extractString(l.KZ) === 'E')
      .map(l => ({
        ID: l.ID,
        BEZEICHNUNG: `${l.LAGERNUMMER || '?'} - ${extractString(l.BEZEICHNUNG) || 'Eierlager'}`
      }));
  }
});
const lagerplaetze = ref<any[]>([]);

const filteredRows = computed(() => {
  let list = rows.value;
  
  if (nurAktiveFilter.value) {
    list = list.filter(row => {
      // Check for HERDEN_AKTIV_REL (from join) or check herdeOptions fallback
      const rowAktiv = (row as any).herden_aktiv_rel ?? (row as any).HERDEN_AKTIV_REL;
      if (rowAktiv !== undefined) return extractInt(rowAktiv) === 1;
      
      const herd = herdeOptions.value.find(h => h.ID === extractInt(row.ID_HERDEN));
      return herd ? extractInt(herd.AKTIV) === 1 : true;
    });
  }

  if (!filterHerde.value && !filterDateRange.value) return list;

  if (filterHerde.value) {
    list = list.filter(row => {
      const rowHerdeId = extractInt(row.ID_HERDEN);
      return rowHerdeId === filterHerde.value;
    });
  }

  if (filterDateRange.value) {
    const from = new Date(filterDateRange.value.from.replace(/\//g, '-'));
    const to = new Date(filterDateRange.value.to.replace(/\//g, '-'));
    from.setHours(0, 0, 0, 0);
    to.setHours(23, 59, 59, 999);
    
    list = list.filter(row => {
      const dateStr = extractString(row.buchungsdatum || row.BUCHUNGSDATUM);
      if (!dateStr || dateStr === '0001-01-01') return false;
      const rowDate = new Date(dateStr);
      return rowDate >= from && rowDate <= to;
    });
  }

  return list;
});

interface Pagination {
  rowsPerPage: number;
  sortBy: string | null;
  descending: boolean;
}

const pagination = ref<Pagination>({
  rowsPerPage: 15,
  sortBy: 'BUCHUNGSDATUM',
  descending: true
});

const columns: QTableProps['columns'] = [
  { name: 'actions', label: 'Aktion', field: 'actions', align: 'center', style: 'width: 80px' },
  { name: 'VERMITTELT', label: 'V', field: (row: any) => extractString(row.vermittelt || row.VERMITTELT) || 'N', align: 'center', sortable: true },
  { name: 'BUCHUNGSDATUM', label: 'Datum', field: (row: any) => extractString(row.buchungsdatum || row.BUCHUNGSDATUM) || '-', align: 'left', sortable: true },
  { name: 'LW', label: 'LW', field: (row: any) => extractInt(row.lw || row.LW), align: 'center', sortable: true },
  { name: 'HERDENNUMMER', label: 'Herde', field: (row: any) => extractInt(row.herdennummer || row.HERDENNUMMER) || '-', align: 'left', sortable: true },
  { name: 'TIERBESTAND', label: 'Bestand', field: (row: any) => extractInt(row.tierbestand || row.TIERBESTAND), align: 'right', sortable: true },
  { name: 'KLASSEA', label: 'Klasse A', field: (row: any) => extractInt(row.klassea || row.KLASSEA), align: 'right', sortable: true },
  { name: 'XL', label: 'XL', field: (row: any) => extractInt(row.xl || row.XL), align: 'right' },
  { name: 'LARGE', label: 'L', field: (row: any) => extractInt(row.large || row.LARGE), align: 'right' },
  { name: 'MEDIUM', label: 'M', field: (row: any) => extractInt(row.medium || row.MEDIUM), align: 'right' },
  { name: 'SMALL', label: 'S', field: (row: any) => extractInt(row.small || row.SMALL), align: 'right' },
  { name: 'KNICKEIER', label: 'Knick', field: (row: any) => extractInt(row.knickeier || row.KNICKEIER), align: 'right' },
  { name: 'SCHMUTZ', label: 'Schm.', field: (row: any) => extractInt(row.schmutz || row.SCHMUTZ), align: 'right' },
  { name: 'KONTROLLGEWICHT', label: 'Gew. KG', field: (row: any) => extractFloat(row.kontrollgewicht || row.KONTROLLGEWICHT), align: 'right' },
  { name: 'VERPACKUNG', label: 'Vp. KG', field: (row: any) => extractFloat(row.verpackung || row.VERPACKUNG), align: 'right' },
  { name: 'GEWICHTPROBE', label: 'GP', field: (row: any) => extractFloat(row.gewichtprobe || row.GEWICHTPROBE), align: 'right' },
  { name: 'DGEWICHTEI', label: 'g/Ei', field: (row: any) => extractFloat(row.dgewichtei || row.DGEWICHTEI), align: 'right', sortable: true },
  { name: 'VERLUSTE', label: 'Verl.', field: (row: any) => extractInt(row.verluste || row.VERLUSTE), align: 'right', sortable: true },
];

const showDialog = ref(false);
const showVerlusteDialog = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const saving = ref(false);
const lastBookingDate = ref<string | null>(null);

const vermittlungDays = computed(() => {
  if (isEditing.value) {
    if (form.VERMITTELTAM && form.VERMITTELTAM !== form.BUCHUNGSDATUM) return 100;
    return 0;
  }
  if (!bookingDateOnly.value) return 0;
  if (!getParamBool('KLASSEAVERMITTELN')) return 0;

  let baseDate = lastBookingDate.value;
  if (!baseDate) {
    const sel = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
    if (sel?.LEGEDATUM) {
      const d = new Date(sel.LEGEDATUM);
      d.setDate(d.getDate() - 1);
      baseDate = date.formatDate(d, 'YYYY-MM-DD');
    }
  }

  if (!baseDate) return 0;

  const diffDays = date.getDateDiff(bookingDateOnly.value, baseDate, 'days');
  return diffDays > 1 ? diffDays : 0;
});

const isVermittelt = computed(() => {
  const v = extractString(form.VERMITTELT);
  return v === 'V' || v === 'S' || vermittlungDays.value > 0;
});

const herdeSelect = ref<any>(null);
const klasseAInput = ref<any>(null);
const gewichtProbeInput = ref<any>(null);
const kontrollGewichtInput = ref<any>(null);

const eilagerForm = reactive({
  ID_EILAGER: null as number | null,
  ID_FREMDESLAGER: null as number | null,
  JUMBOS: 0,
  XL: 0,
  LARGE: 0,
  MEDIUM: 0,
  SMALL: 0,
  VOLLEIKG: 0,
  SCHMUTZ: 0,
  KNICKEIER: 0,
  BRUCHEIER: 0,
  BUCHUNGSTYP: 'E',
  CHARGE: '',
  KZ_LAGER: 'E'
});

const verwendungsTexteOptions = ref<any[]>([]);

const usedStock = reactive({
  JUMBOS: 0,
  XL: 0,
  LARGE: 0,
  MEDIUM: 0,
  SMALL: 0,
  VOLLEIKG: 0
});

const v_jumbos = computed(() => (Number(form.KL6) || 0) - (Number(usedStock.JUMBOS) || 0) - (Number(eilagerForm.JUMBOS) || 0));
const v_xl = computed(() => (Number(form.XL) || 0) - (Number(usedStock.XL) || 0) - (Number(eilagerForm.XL) || 0));
const v_large = computed(() => (Number(form.LARGE) || 0) - (Number(usedStock.LARGE) || 0) - (Number(eilagerForm.LARGE) || 0));
const v_medium = computed(() => (Number(form.MEDIUM) || 0) - (Number(usedStock.MEDIUM) || 0) - (Number(eilagerForm.MEDIUM) || 0));
const v_small = computed(() => (Number(form.SMALL) || 0) - (Number(usedStock.SMALL) || 0) - (Number(eilagerForm.SMALL) || 0));
const v_vollei = computed(() => (Number(form.VOLLEI) || 0) - (Number(usedStock.VOLLEIKG) || 0) - (Number(eilagerForm.VOLLEIKG) || 0));

const hasValidationError = computed(() => {
  return v_xl.value < 0 || v_large.value < 0 || v_medium.value < 0 || v_small.value < 0 || v_vollei.value < 0;
});

const vStock = reactive({
  JUMBOS: 0,
  XL: 0,
  LARGE: 0,
  MEDIUM: 0,
  SMALL: 0,
  VOLLEIKG: 0
});

async function fetchUsedStock() {
  if (!editId.value) {
    Object.keys(usedStock).forEach(k => (usedStock as Record<string, number>)[k] = 0);
    return;
  }
  try {
    const res = await api.get(`/api/eilagerbuchungen/sum-by-buchung/${editId.value}`);
    if (res.data) {
      usedStock.JUMBOS = Number(res.data.jumbos) || 0;
      usedStock.XL = Number(res.data.xl) || 0;
      usedStock.LARGE = Number(res.data.large) || 0;
      usedStock.MEDIUM = Number(res.data.medium) || 0;
      usedStock.SMALL = Number(res.data.small) || 0;
      usedStock.VOLLEIKG = Number(res.data.volleikg) || 0;
    }
  } catch (_err) {
    console.error('Fehler beim Laden des verbrauchten Bestands:', _err);
  }
}

const verlustReasons = ref<any[]>([]);

const verlustForm = reactive({
  ID_HERDEN: 0,
  VERLUSTE: 1,
  ID_TEXTE: null as number | null,
  DATUM: '0001-01-01',
  MEMO: ''
});

function openVerlustDialog() {
  verlustForm.ID_HERDEN = form.ID_HERDEN;
  verlustForm.DATUM = bookingDateOnly.value || date.formatDate(new Date(), 'YYYY-MM-DD');
  verlustForm.VERLUSTE = 1;
  verlustForm.ID_TEXTE = verlustReasons.value.length > 0 ? verlustReasons.value[0].ID : null;
  verlustForm.MEMO = '';
  showVerlusteDialog.value = true;
}

async function submitVerlust() {
  if (verlustForm.VERLUSTE <= 0 || !verlustForm.ID_TEXTE) return;
  if (!verlustForm.DATUM) verlustForm.DATUM = '0001-01-01';

  try {
    const payload = {
      id_herden: verlustForm.ID_HERDEN,
      verluste: verlustForm.VERLUSTE,
      id_texte: verlustForm.ID_TEXTE,
      DATUM: verlustForm.DATUM,
      MEMO: verlustForm.MEMO
    };
    await api.post('/api/leistung/verlust', payload);
    $q.notify({
      type: 'positive',
      message: `${verlustForm.VERLUSTE} Verluste für Herde gebucht`,
      icon: 'check_circle'
    });
    showVerlusteDialog.value = false;
    void loadData(); 
  } catch (err: any) {
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Buchen der Verluste: ' + (err.response?.data?.error || err.message)
    });
  }
}

const form = reactive({
  ID_HERDEN: 0,
  BUCHUNGSDATUM: '0001-01-01',
  fullTimestamp: '',
  KL6: 0,
  XL: 0,
  LARGE: 0,
  MEDIUM: 0,
  SMALL: 0,
  KLASSEA: 0,
  SCHMUTZ: 0,
  KNICKEIER: 0,
  BRUCHEIER: 0,
  VOLLEI: 0,
  GEWICHTPROBE: 0,
  KONTROLLGEWICHT: 0,
  VERPACKUNG: 0,
  DGEWICHTEI: 0,
  VERMITTELTAM: '0001-01-01',
  VERMITTELT: 'N',
  HERDENNUMMER: 0,
  TIERBESTAND: 0
});

const displayKontrollgewicht = ref('');
const displayVerpackung = ref('');
const displayVollei = ref('');
const displayDgewicht = ref('');





const bookingDateOnly = computed({
  get: () => form.fullTimestamp ? form.fullTimestamp.split(' ')[0] : '',
  set: (val) => {
    const time = form.fullTimestamp ? (form.fullTimestamp.split(' ')[1] || '12:00') : '12:00';
    form.fullTimestamp = `${val} ${time}`;
  }
});

const calculatedLegewoche = computed(() => {
  if (!form.ID_HERDEN || !bookingDateOnly.value) return '-';
  const herde = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
  if (!herde || !herde.LEGEDATUM) return '-';

  const diffDays = date.getDateDiff(bookingDateOnly.value, herde.LEGEDATUM, 'days');
  if (diffDays < 0) return '-';
  return Math.floor(diffDays / 7) + 1;
});

const calculatedAlterswoche = computed(() => {
  if (!form.ID_HERDEN || !bookingDateOnly.value) return '-';
  const herde = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
  if (!herde || !herde.GEBURTSDATUM) return '-';

  const diffDays = date.getDateDiff(bookingDateOnly.value, herde.GEBURTSDATUM, 'days');
  if (diffDays < 0) return '-';
  return Math.floor(diffDays / 7) + 1;
});

const computedKlasseA = computed(() => {
  return (Number(form.XL) || 0) + (Number(form.LARGE) || 0) + (Number(form.MEDIUM) || 0) + (Number(form.SMALL) || 0);
});

const totalEggs = computed(() => {
  const kA = Math.max(extractInt(form.KLASSEA), extractInt(form.XL) + extractInt(form.LARGE) + extractInt(form.MEDIUM) + extractInt(form.SMALL));
  return kA + extractInt(form.SCHMUTZ) + extractInt(form.KNICKEIER) + extractInt(form.BRUCHEIER);
});

const hasEggOverflow = computed(() => {
  const total = Number(totalEggs.value) || 0;
  const limit = Number(form.TIERBESTAND) || 0;
  if (limit <= 0) return false;
  return total > limit;
});

const validateEggs = () => {
  const total = Number(totalEggs.value) || 0;
  const limit = Number(form.TIERBESTAND) || 0;
  // Wir sperren nur, wenn wir einen Bestand > 0 haben. 
  // Aber wir zeigen die Warnung immer an, wenn die Menge unrealistisch hoch ist.
  if (limit > 0 && total > limit) {
    return `Menge (${total}) > Bestand (${limit})`;
  }
  return true;
};

// watch(computedKlasseA, (val) => {
//   form.KLASSEA = val;
// });

function onKlasseAInput(val: number) {
  form.KLASSEA = val;
  // Trigger distribution if weight/age distribution is enabled
  if (getParamBool('aufteilungalter') || (getParamBool('aufteilunggewicht') && Number(form.KONTROLLGEWICHT) > 0)) {
    void triggerAutomaticDistribution();
  }
}

async function onHerdeChange(val: number) {
  if (!val) {
    herdParams.value = null;
    lastBookingDate.value = null;
    return;
  }

  // Sicherstellen, dass die Herdenliste geladen ist
  if (!herdeOptions.value || herdeOptions.value.length === 0) {
    await fetchHerden();
  }

  // Tierbestand SOFORT aus den Stammdaten der Herde nehmen
  const herde = herdeOptions.value.find(h => {
    const id = getInsensitive(h, 'id');
    return Number(id) === Number(val);
  });
  
  if (herde) {
    // Wir suchen nach 'anfangsbestand', aber falls das Feld anders heißt (z.B. BESTAND oder ANFANGS_BESTAND),
    // suchen wir nach dem Teilstring 'bestand'.
    let ab = getInsensitive(herde, 'anfangsbestand');
    if (ab === undefined || ab === null) {
      const altKey = Object.keys(herde).find(k => k.toLowerCase().includes('bestand'));
      if (altKey) ab = herde[altKey];
    }
    // Letzter Fallback: liefermenge
    if (ab === undefined || ab === null) {
      ab = getInsensitive(herde, 'liefermenge') ?? 0;
    }
    
    form.TIERBESTAND = extractInt(ab);
  } else {
    form.TIERBESTAND = 0;
  }

  try {
    const [pRes, bRes, latestRes] = await Promise.all([
      api.get(`/api/firmenparameter/H/${val}`).catch(() => ({ data: null })),
      api.get(`/api/buchung/last-info/${val}`).catch(() => ({ data: null })),
      api.get(`/api/herden/${val}/latest_booking`).catch(() => ({ data: null }))
    ]);
    
    if (pRes.data) {
      herdParams.value = pRes.data?.data || pRes.data;
    }
    
    if (bRes.data) {
      lastBookingDate.value = bRes.data.buchungsdatum;
    }

    // Falls wir neu anlegen: Versuche Tierbestand aus letzter Buchung zu berechnen
    if (!isEditing.value && latestRes.data) {
      const latest = latestRes.data;
      const prevBestand = extractInt(latest.tierbestand || latest.TIERBESTAND);
      const prevVerluste = extractInt(latest.verluste || latest.VERLUSTE);
      form.TIERBESTAND = prevBestand - prevVerluste;
    }

    if (!isEditing.value && !getParamBool('BEIVERMITTELNDATUMAKTUELL')) {
      let suggestedDate = '';
      if (lastBookingDate.value && !lastBookingDate.value.startsWith('0001-01-01')) {
        suggestedDate = addOneDay(lastBookingDate.value);
      } else {
        const herde = herdeOptions.value.find(h => h.ID === val);
        if (herde && herde.LEGEDATUM && !herde.LEGEDATUM.startsWith('0001-01-01')) {
          suggestedDate = addOneDay(herde.LEGEDATUM);
        }
      }
      if (suggestedDate) {
        form.fullTimestamp = `${suggestedDate} 12:00`;
      }
    }

    const weight = getParamNumber('ANZAHLKONTROLLW');
    const verpackung = getParamNumber('VERPACKUNGKG');
    if (weight > 0) {
      form.GEWICHTPROBE = weight;
    }
    if (verpackung > 0) {
      form.VERPACKUNG = verpackung;
    }
    syncDisplayRefs();
    calcDgewicht();
  } catch (err) {
    console.error('Fehler beim Laden der Herden-Info:', err);
    herdParams.value = null;
  }
}

function getParamBool(key: string): boolean {
  if (!activeParameters.value) return false;
  
  const obj = activeParameters.value;
  const val = obj[key] ?? obj[key.toLowerCase()] ?? obj[key.toUpperCase()];
  if (val !== undefined) return extractBool(val);
  
  if (Array.isArray(activeParameters.value)) {
    const p = activeParameters.value.find((x: any) => (x.BEZEICHNUNG || '').toLowerCase() === key.toLowerCase());
    return p ? extractBool(p.wert_bool ?? p.wert) : false;
  }
  return false;
}

function getParamNumber(key: string): number {
  if (!activeParameters.value) return 0;
  
  const obj = activeParameters.value;
  const val = obj[key] ?? obj[key.toLowerCase()] ?? obj[key.toUpperCase()];
  if (val !== undefined) return extractFloat(val);

  if (Array.isArray(activeParameters.value)) {
    const p = activeParameters.value.find((x: any) => (x.BEZEICHNUNG || '').toLowerCase() === key.toLowerCase());
    return p ? Number(p.wert_double ?? p.wert) || 0 : 0;
  }
  return 0;
}

function getParamString(key: string): string {
  if (!activeParameters.value) return '';
  
  const obj = activeParameters.value;
  const val = obj[key] ?? obj[key.toLowerCase()] ?? obj[key.toUpperCase()];
  if (val !== undefined) return extractString(val);

  if (Array.isArray(activeParameters.value)) {
    const p = activeParameters.value.find((x: any) => (x.BEZEICHNUNG || '').toLowerCase() === key.toLowerCase());
    return p ? extractString(p.wert_string ?? p.wert) : '';
  }
  return '';
}

async function loadData() {
  loading.value = true;
  try {
    const res = await api.get('/api/buchung');
    rows.value = res.data || [];
  } catch {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Leistung)' });
  } finally {
    loading.value = false;
  }
}

async function fetchHerden() {
  const res = await api.get('/api/herden/lookup');
  herdeOptions.value = (res.data || []).map((h: any) => {
    const id = h.ID ?? h.id;
    const bez = extractString(h.BEZEICHNUNG ?? h.bezeichnung);
    const hnr = extractInt(h.HERDENNUMMER ?? h.herdennummer);
    const aktiv = extractInt(h.AKTIV ?? h.aktiv);
    
    return {
      ID: id,
      BEZEICHNUNG: bez || (hnr ? `Herde ${hnr}` : `Herde ${id}`),
      HERDENNUMMER: hnr,
      LEGEDATUM: extractString(h.LEGEDATUM ?? h.legedatum),
      GEBURTSDATUM: extractString(h.GEBURTSDATUM ?? h.geburtstagsdatum),
      ANFANGSBESTAND: extractInt(h.ANFANGSBESTAND ?? h.anfangsbestand),
      ID_EILAGER: extractInt(h.ID_EILAGER ?? h.id_eilager),
      AKTIV: aktiv
    };
  });
}

async function fetchParameters() {
  try {
    const res = await api.get('/api/firmenparameter/F/0');
    globalParams.value = res.data?.data || res.data;
  } catch (err) {
    console.error('Fehler beim Laden der globalen Parameter', err);
  }
}

async function fetchVerwendungsTexte() {
  try {
    const res = await api.get('/api/texte/typ/L'); 
    verwendungsTexteOptions.value = (res.data || [])
      .filter((t: any) => extractString(t.KZ || t.kz) !== 'E') 
      .map((t: any) => ({
        kz: extractString(t.KZ || t.kz),
        betreff: extractString(t.BETREFF || t.betreff)
      }));
  } catch (err) {
    console.error('Fehler beim Laden der Verwendungstexte', err);
  }
}

async function fetchVerlustReasons() {
  try {
    const res = await api.get('/api/texte/typ/V');
    verlustReasons.value = res.data || [];
  } catch (err) {
    console.error('Fehler beim Laden der Verlustgründe', err);
  }
}

async function fetchEilager() {
  try {
    const res = await api.get('/api/eilager');
    fullEilagerRaw.value = res.data || [];
  } catch (err) {
    console.error('Fehler beim Laden der Eilager', err);
  }
}

async function fetchLagerplaetze() {
  try {
    const res = await api.get('/api/lagerplatz');
    lagerplaetze.value = (res.data || []).map((lp: any) => ({
      ID: lp.ID,
      ID_EILAGER: extractInt(lp.ID_EILAGER),
      BEZEICHNUNG: extractString(lp.BEZEICHNUNG)
    }));
  } catch (err) {
    console.error('Fehler beim Laden der Lagerplätze', err);
  }
}

const filteredLagerplaetze = computed(() => {
  if (!eilagerForm.ID_EILAGER) return [];
  return lagerplaetze.value.filter(lp => lp.ID_EILAGER === eilagerForm.ID_EILAGER);
});

function onTargetLagerChange() {
  eilagerForm.ID_FREMDESLAGER = null;
  const lager = fullEilagerRaw.value.find(l => l.ID === eilagerForm.ID_EILAGER);
  if (lager) {
    vStock.JUMBOS = extractInt(lager.JUMBOS) || 0;
    vStock.XL = extractInt(lager.XL) || 0;
    vStock.LARGE = extractInt(lager.LARGE) || 0;
    vStock.MEDIUM = extractInt(lager.MEDIUM) || 0;
    vStock.SMALL = extractInt(lager.SMALL) || 0;
    vStock.VOLLEIKG = extractFloat(lager.VOLLEIKG) || 0;
    
    if (lagerplaetze.value.length > 0) {
      const firstMatch = lagerplaetze.value.find(lp => lp.ID_EILAGER === eilagerForm.ID_EILAGER);
      if (firstMatch) {
        eilagerForm.ID_FREMDESLAGER = firstMatch.ID;
      }
    }
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  const prevHerdeId = lastSelectedHerdeId.value;
  resetForm();
  form.fullTimestamp = sessionStore.workingTimestamp;
  
  if (prevHerdeId) {
    form.ID_HERDEN = prevHerdeId;
    void onHerdeChange(prevHerdeId);
  } else if (herdeOptions.value.length === 1) {
    form.ID_HERDEN = herdeOptions.value[0].ID;
    void onHerdeChange(form.ID_HERDEN);
  }
  showDialog.value = true;
}

function onEdit(row: Record<string, unknown>) {
  isEditing.value = true;
  editId.value = (row.ID || row.id) as number;
  Object.assign(form, {
    ID_HERDEN: extractInt(row.ID_HERDEN || row.id_herden),
    fullTimestamp: `${extractString(row.BUCHUNGSDATUM || row.buchungsdatum) || ''} ${extractString(row.ZEITSTEMPEL || row.zeitstempel)?.substring(11, 16) || '12:00'}`,
    KL6: extractInt(row.KL6 || row.kl6) || 0,
    XL: extractInt(row.XL || row.xl) || 0,
    LARGE: extractInt(row.LARGE || row.large) || 0,
    MEDIUM: extractInt(row.MEDIUM || row.medium) || 0,
    SMALL: extractInt(row.SMALL || row.small) || 0,
    KLASSEA: extractInt(row.KLASSEA || row.klassea) || 0,
    SCHMUTZ: extractInt(row.SCHMUTZ || row.schmutz) || 0,
    KNICKEIER: extractInt(row.KNICKEIER || row.knickeier) || 0,
    BRUCHEIER: extractInt(row.BRUCHEIER || row.brucheier) || 0,
    VOLLEI: extractFloat(row.VOLLEI || row.vollei) || 0,
    GEWICHTPROBE: extractFloat(row.GEWICHTPROBE || row.gewichtprobe) || 0,
    KONTROLLGEWICHT: extractFloat(row.KONTROLLGEWICHT || row.kontrollgewicht) || 0,
    VERPACKUNG: 0, // Wird aus Parametern geladen
    DGEWICHTEI: extractFloat(row.DGEWICHTEI || row.dgewichtei) || 0,
    TIERBESTAND: 0, // Wird gleich über onHerdeChange aus Stammdaten geladen
    VERMITTELTAM: extractString(row.VERMITTELTAM || row.vermitteltam),
  });

  const vp = getParamNumber('VERPACKUNGKG');
  form.VERPACKUNG = vp;
  syncDisplayRefs();

  // Erst onHerdeChange aufrufen, um TIERBESTAND aus Stammdaten zu holen
  void onHerdeChange(form.ID_HERDEN!);
  void fetchUsedStock();
  showDialog.value = true;
}

function resetForm() {
  Object.assign(form, {
    ID_HERDEN: 0,
    fullTimestamp: '',
    KL6: 0,
    XL: 0,
    LARGE: 0,
    MEDIUM: 0,
    SMALL: 0,
    KLASSEA: 0,
    SCHMUTZ: 0,
    KNICKEIER: 0,
    BRUCHEIER: 0,
    VOLLEI: 0,
    GEWICHTPROBE: 0,
    KONTROLLGEWICHT: 0,
    VERPACKUNG: 0,
    DGEWICHTEI: 0,
    VERMITTELTAM: '0001-01-01',
    VERMITTELT: 'N'
  });
  Object.keys(usedStock).forEach(k => (usedStock as Record<string, number>)[k] = 0);
}

function closeDialog() {
  showDialog.value = false;
}

async function triggerLagerBuchung(pseudo = false) {
  if (!editId.value) {
    const newId = await onSubmit();
    if (!newId) return;
  }

  isPseudoBooking.value = pseudo;

  Object.assign(eilagerForm, {
    ID_FREMDESLAGER: 0,
    ID_EILAGER: 0,
    JUMBOS: 0, XL: 0, LARGE: 0, MEDIUM: 0, SMALL: 0, VOLLEIKG: 0,
    SCHMUTZ: 0, KNICKEIER: 0, BRUCHEIER: 0,
    BUCHUNGSTYP: pseudo ? null : 'E',
    CHARGE: '',
    KZ_LAGER: 'E'
  });

  eilagerForm.JUMBOS = 0;
  eilagerForm.XL = Number(form.XL) - usedStock.XL;
  eilagerForm.LARGE = Number(form.LARGE) - usedStock.LARGE;
  eilagerForm.MEDIUM = Number(form.MEDIUM) - usedStock.MEDIUM;
  eilagerForm.SMALL = Number(form.SMALL) - usedStock.SMALL;
  eilagerForm.VOLLEIKG = Number(form.VOLLEI) - usedStock.VOLLEIKG;

  if (!pseudo) {
    const herde = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
    if (herde && herde.ID_EILAGER) {
      eilagerForm.ID_EILAGER = herde.ID_EILAGER;
      void onTargetLagerChange();
    }
  }

  generateChargeProposal();
  showEilagerDialog.value = true;
}

function generateChargeProposal() {
  const sep = getParamString('CHARGETRENNUNG') || '-';
  const parts: string[] = [];

  // 1. Firmen-Prefix
  const prefixFirma = getParamString('CHARGEPREFIXFIRMA');
  if (prefixFirma) {
    parts.push(prefixFirma);
  }

  // 2. Herdennummer
  if (getParamBool('CHARGEPREFIXHERDENNUMMER')) {
    const herde = herdeOptions.value.find(h => h.ID === form.ID_HERDEN);
    if (herde && herde.HERDENNUMMER) {
      parts.push(String(herde.HERDENNUMMER));
    } else {
      parts.push(String(form.ID_HERDEN || '0'));
    }
  }

  // 3. Datum
  if (getParamBool('CHARGEDATUM')) {
    parts.push(date.formatDate(bookingDateOnly.value, 'DDMMYY'));
  }

  // 4. Lagernummer
  if (getParamBool('CHARGELAGERNUMMER')) {
    const lager = fullEilagerRaw.value.find(l => l.ID === eilagerForm.ID_EILAGER);
    if (lager && (lager.LAGERNUMMER || lager.lagernummer)) {
      parts.push(String(lager.LAGERNUMMER || lager.lagernummer));
    }
  }

  if (parts.length > 0) {
    eilagerForm.CHARGE = parts.join(sep);
  } else {
    // Fallback: Default-Format wenn nichts konfiguriert ist
    const dStr = date.formatDate(bookingDateOnly.value, 'DDMMYY');
    eilagerForm.CHARGE = `${dStr}-${form.ID_HERDEN || '0'}`;
  }
}

async function submitLagerBuchung() {
  if (!editId.value || !eilagerForm.ID_EILAGER || !eilagerForm.ID_FREMDESLAGER) return;
  try {
    const lager = fullEilagerRaw.value.find(l => l.ID === eilagerForm.ID_EILAGER);
    const payload = {
      ...eilagerForm,
      ID_BUCHUNG: editId.value,
      BUCHUNGSDATUM: bookingDateOnly.value,
      KZ_LAGER: extractString(lager?.KZ) || (isPseudoBooking.value ? 'P' : 'E')
    };
    await api.post('/api/eilagerbuchungen', payload);
    $q.notify({ type: 'positive', message: 'Bestand erfolgreich ins Eilager übernommen' });
    
    // Reset Eilager-Formular für weitere Buchungen
    eilagerForm.JUMBOS = 0;
    eilagerForm.XL = 0;
    eilagerForm.LARGE = 0;
    eilagerForm.MEDIUM = 0;
    eilagerForm.SMALL = 0;
    eilagerForm.VOLLEIKG = 0;
    eilagerForm.CHARGE = '';

    showEilagerDialog.value = false;
    void fetchUsedStock();
  } catch (err: unknown) {
    const errorMsg = (err as { response?: { data?: { error?: string } }; message?: string })?.response?.data?.error || (err as { message?: string })?.message || 'Unbekannter Fehler';
    $q.notify({ type: 'negative', message: 'Fehler beim Buchen im Eilager: ' + errorMsg });
  }
}

// onAnderweitigeVerwendung removed - functionality moved to triggerLagerBuchung(true)


function syncDisplayRefs() {
  displayKontrollgewicht.value = formatNumberLocalized(form.KONTROLLGEWICHT, 2);
  displayVerpackung.value = formatNumberLocalized(form.VERPACKUNG, 2);
  displayVollei.value = formatNumberLocalized(form.VOLLEI, 2);
  displayDgewicht.value = formatNumberLocalized(form.DGEWICHTEI, 2);
}

function onKontrollgewichtBlur() {
  form.KONTROLLGEWICHT = parseNumberLocalized(displayKontrollgewicht.value);
  syncDisplayRefs();
  calcDgewicht();
}

function onVerpackungBlur() {
  form.VERPACKUNG = parseNumberLocalized(displayVerpackung.value);
  syncDisplayRefs();
  calcDgewicht();
}

function onVolleiBlur() {
  form.VOLLEI = parseNumberLocalized(displayVollei.value);
  syncDisplayRefs();
}

const calcDgewicht = () => {
  const kg = Number(form.KONTROLLGEWICHT) || 0;
  const vp = Number(form.VERPACKUNG) || 0;
  const anzahl = extractInt(form.GEWICHTPROBE);
  if (kg > 0 && anzahl > 0) {
    form.DGEWICHTEI = Number(((kg - vp) * 1000 / anzahl).toFixed(2));
    displayDgewicht.value = formatNumberLocalized(form.DGEWICHTEI, 2);
  } else {
    form.DGEWICHTEI = 0;
    displayDgewicht.value = '';
  }
};

async function triggerAutomaticDistribution() {
  if (!form.ID_HERDEN || (Number(form.KONTROLLGEWICHT) <= 0 && !getParamBool('aufteilungalter'))) return;
  
  try {
    const payload = {
      ID_HERDEN: Number(form.ID_HERDEN),
      KONTROLLGEWICHT: Number(form.KONTROLLGEWICHT),
      GEWICHTPROBE: Number(form.GEWICHTPROBE),
      VERPACKUNG: Number(form.VERPACKUNG),
      KLASSEA: Number(form.KLASSEA),
      LW: Number(calculatedLegewoche.value),
      AW: Number(calculatedAlterswoche.value)
    };

    const res = await api.post('/api/calculate-distribution', payload);
    if (res.data && res.data.calculated) {
      form.KL6 = 0; // Jumbos zurücksetzen, da sie nicht Teil der Verteilungstabelle sind
      form.XL = res.data.xl || 0;
      form.LARGE = res.data.large || 0;
      form.MEDIUM = res.data.medium || 0;
      form.SMALL = res.data.small || 0;
      if (res.data.dgewicht) form.DGEWICHTEI = Number(res.data.dgewicht.toFixed(2));
    }
  } catch (err) {
    console.error('Fehler bei der automatischen Verteilung:', err);
  }
}

// Watch for changes that should trigger distribution (if enabled)
watch([() => form.KLASSEA, () => form.ID_HERDEN], () => {
  if (getParamBool('aufteilungalter') || (getParamBool('aufteilunggewicht') && Number(form.KONTROLLGEWICHT) > 0)) {
    void triggerAutomaticDistribution();
  }
});

// Sync herdennummer when id_herden changes
watch(() => form.ID_HERDEN, (newVal) => {
  if (newVal) {
    const herde = herdeOptions.value.find(h => h.ID === newVal);
    if (herde) {
      form.HERDENNUMMER = herde.HERDENNUMMER || 0;
    }
  }
});

function onDialogShow() {
  setTimeout(() => {
    (herdeSelect.value)?.$el?.focus();
  }, 100);
}

function onDelete(row: Record<string, unknown>) {
  $q.dialog({ title: 'Löschen', message: 'Buchung wirklich löschen?', cancel: true, persistent: true }).onOk(async () => {
    try {
      await api.delete(`/api/buchung/${row.ID || row.id}`);
      $q.notify({ type: 'positive', message: 'Buchung gelöscht' });
      void loadData();
    } catch (_err) {
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
    }
  });
}

async function onSubmit() {
  const payload = {
    ...form,
    ID_HERDEN: Number(form.ID_HERDEN),
    BUCHUNGSDATUM: bookingDateOnly.value,
    KL6: Number(form.KL6),
    XL: Number(form.XL),
    LARGE: Number(form.LARGE),
    MEDIUM: Number(form.MEDIUM),
    SMALL: Number(form.SMALL),
    SCHMUTZ: Number(form.SCHMUTZ),
    KNICKEIER: Number(form.KNICKEIER),
    BRUCHEIER: Number(form.BRUCHEIER),
    VOLLEI: Number(form.VOLLEI),
    GEWICHTPROBE: Number(form.GEWICHTPROBE),
    KONTROLLGEWICHT: Number(form.KONTROLLGEWICHT),
    VERPACKUNG: Number(form.VERPACKUNG),
    DGEWICHTEI: Number(form.DGEWICHTEI),
    LW: Number(calculatedLegewoche.value) || 0,
    AW: Number(calculatedAlterswoche.value) || 0,
    KLASSEA: computedKlasseA.value, // IMMER die Summe der Klassen speichern
    TIERBESTAND: Number(form.TIERBESTAND) || 0,
    HERDENNUMMER: Number(form.HERDENNUMMER) || 0,
    DATUM: bookingDateOnly.value || '0001-01-01',
    ZEITSTEMPEL: form.fullTimestamp ? (form.fullTimestamp + ':00Z') : '',
    VERMITTELT: form.VERMITTELT || 'N',
    VERMITTELTAM: form.VERMITTELTAM || '0001-01-01',
    _vermittlungDays: vermittlungDays.value
  };

  if (saving.value) return;
  saving.value = true;
  
  try {
    let savedId = editId.value;
    if (isEditing.value && editId.value) {
      await api.put(`/api/buchung/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Buchung aktualisiert' });
    } else {
      // Sequential entry logic: stay in herd if date was < today
      const bDateStr = bookingDateOnly.value;
      const todayStr = new Date().toISOString().split('T')[0];
      if (bDateStr < todayStr) {
        lastSelectedHerdeId.value = form.ID_HERDEN;
      } else {
        lastSelectedHerdeId.value = null;
      }

      const res = await api.post('/api/buchung', payload);
      savedId = res.data?.id;
      if (savedId) {
        editId.value = res.data.id;
        isEditing.value = true;
      }
      $q.notify({ type: 'positive', message: 'Buchung erfolgreich erstellt' });
    }

    // Refresh the local grid data
    void loadData();

    if (!showEilagerDialog.value) {
      showDialog.value = false;
    }
    return savedId;
  } catch (err: any) {
    const msg = err.response?.data?.error || 'Fehler beim Speichern';
    $q.notify({ type: 'negative', message: msg });
    return null;
  } finally {
    saving.value = false;
  }
}


onMounted(async () => {
  initWidths(columns);
  
  // Initialize date range
  filterDateRange.value = {
    from: lastWeek.toISOString().split('T')[0].replace(/-/g, '/'),
    to: today.toISOString().split('T')[0].replace(/-/g, '/')
  };

  try {
    await Promise.all([
      loadData(), 
      fetchParameters(),
      fetchHerden(), 
      fetchEilager(),
      fetchLagerplaetze(),
      fetchVerwendungsTexte(),
      fetchVerlustReasons()
    ]);
  } catch (err) {
    console.error('Initial load failed:', err);
  }
});
</script>

<style>
.q-table tbody tr.vermittelt-row {
  background-color: #e3f2fd !important; /* Hellblau für vermittelte Buchungen */
}
.body--dark .q-table tbody tr.vermittelt-row {
  background-color: #1976d233 !important; /* Kräftigeres Blau im Dunkelmodus (20% Alpha) */
}

.q-table tbody tr.source-row {
  background-color: #fff8e1 !important; /* Hellgelb/Bernstein für Quellsätze */
}
.body--dark .q-table tbody tr.source-row {
  background-color: #ffa00033 !important; /* Kräftigeres Bernstein im Dunkelmodus (20% Alpha) */
}

/* Hover-Effekt beibehalten aber leicht abdunkeln */
.q-table tbody tr.vermittelt-row:hover, .q-table tbody tr.source-row:hover {
  filter: contrast(1.1) brightness(0.95);
}
</style>

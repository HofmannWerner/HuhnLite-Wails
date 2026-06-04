<template>
  <q-page padding>
    <div class="row items-center q-mb-lg">
      <div class="text-h4 text-weight-bolder text-primary">{{ t('auto.tabellen') }}</div>
    </div>

    <q-tabs
      v-model="tab"
      dense
      class="text-grey bg-white shadow-2 rounded-borders"
      active-color="primary"
      indicator-color="primary"
      align="left"
      narrow-indicator
    >
      <q-tab name="alter" :label="t('auto.alterstabelle')" icon="calendar_today" />
      <q-tab name="gewicht" :label="t('auto.gewichtstabelle')" icon="straighten" />
      <q-tab name="mwst" :label="t('auto.mehrwertsteuer')" icon="percent" />
      <q-tab name="preise" :label="t('auto.preistabelle')" icon="payments"/>
      <q-tab name="futtersorten" :label="t('auto.futtersorten')" icon="egg_alt" />
      <q-tab name="feldnamen" :label="t('auto.feld_konfiguration')" icon="translate"/>
    </q-tabs>

    <q-separator class="q-my-md" />

    <q-tab-panels v-model="tab" animated class="bg-transparent">
      <!-- ALTERSTABELLE -->
      <q-tab-panel name="alter" class="q-pa-none">
        <div class="column q-gutter-y-md">
          <!-- 1. AUSWAHL (OBEN) -->
          <q-card flat bordered class="shadow-2 rounded-borders">
            <q-card-section class="q-pa-sm row items-center q-col-gutter-sm">
              <div class="col-12 col-sm-auto text-subtitle2 text-weight-bold text-primary">{{ t('auto.tabelle_waehlen') }}</div>
              <div class="col-12 col-sm">
                <q-select
                  v-model="selectedHeaderId"
                  :options="tabellenkopfOptions"
                  option-label="BEZEICHNUNG_VAL"
                  option-value="ID"
                  emit-value
                  map-options
                  :label="t('auto.vorhandene_alterstabellen_typ_a')"
                  filled
                  dense
                  :loading="loadingLookups"
                  :key="tabellenkopfOptions.length"
                  @update:model-value="onHeaderSelect"
                >
                  <template v-slot:no-option>
                    <q-item><q-item-section class="text-grey">{{ t('auto.keine_tabellen_gefunden') }}</q-item-section></q-item>
                  </template>
                </q-select>
              </div>
              <div class="col-12 col-sm-auto">
                <q-btn :label="t('auto.neuanlage')" color="secondary" icon="add" @click="onNewHeader" rounded unelevated dense padding="xs md" class="full-width" />
              </div>
            </q-card-section>
          </q-card>

          <!-- 2. KOPF-FORMULAR (MITTE) -->
          <q-card flat bordered class="shadow-2 rounded-borders">
            <q-card-section class="bg-primary text-white row items-center q-pa-sm">
              <!-- Action: Edit Button (Left) -->
              <q-btn
                v-if="!isEditingAlter && alterForm.ID"
                outline
                rounded
                dense
                icon="edit"
                :label="t('auto.kopfdaten_bearbeiten')"
                class="text-white q-px-md"
                @click="isEditingAlter = true"
              />
              <q-btn
                v-if="!isEditingAlter && alterForm.ID"
                outline
                rounded
                dense
                icon="delete"
                color="negative"
                :label="t('auto.tabelle_loeschen')"
                class="text-white q-px-md q-ml-sm"
                @click="onDeleteHeader"
              />
              <q-space />
              <!-- Title (Right) -->
              <div class="row items-center q-gutter-x-sm q-pr-sm">
                <q-icon :name="isEditingAlter ? 'edit' : 'calendar_today'" size="xs" />
                <div class="text-subtitle2 text-weight-bold">{{ isEditingAlter ? t('auto.kopfdaten_bearbeiten') : t('auto.kopfdaten_details') }} ({{ t('auto.alter') }})</div>
              </div>
            </q-card-section>
            <q-card-section class="q-pa-md">
              <q-form @submit="onSubmitAlterHeader" class="row q-col-gutter-md">
                <div class="col-12 col-sm-2">
                  <q-input v-model.number="alterForm.TABELLENNUMMER" :label="t('auto.nummer')" filled stack-label type="number"
                           :rules="[val => !!val || t('message.required')]" dense :disable="!!alterForm.ID || !isEditingAlter"/>
                </div>
                <div class="col-12 col-sm-4">
                  <q-input v-model="alterForm.BEZEICHNUNG" :label="t('grid.designationRequired')" filled stack-label
                           :rules="[val => !!val || t('message.required')]" dense :disable="!isEditingAlter"/>
                </div>
                <div class="col-12 col-sm-3">
                  <q-input v-model="alterForm.ANLAGEDATUM" :label="t('auto.anlagedatum')" filled stack-label dense
                           mask="####-##-##" :disable="!isEditingAlter">
                    <template v-slot:append>
                      <q-icon name="event" class="cursor-pointer">
                        <q-popup-proxy cover transition-show="scale" transition-hide="scale" :disabled="!isEditingAlter">
                          <q-date v-model="alterForm.ANLAGEDATUM" mask="YYYY-MM-DD">
                            <div class="row items-center justify-end">
                              <q-btn v-close-popup :label="t('form.close')" color="primary" flat />
                            </div>
                          </q-date>
                        </q-popup-proxy>
                      </q-icon>
                    </template>
                  </q-input>
                </div>
                <div class="col-12 col-sm-3">
                  <q-input v-model="alterForm.DATUM" :label="t('auto.gueltigkeit_datum')" filled stack-label dense
                           mask="####-##-##" :disable="!isEditingAlter">
                    <template v-slot:append>
                      <q-icon name="event" class="cursor-pointer">
                        <q-popup-proxy cover transition-show="scale" transition-hide="scale" :disabled="!isEditingAlter">
                          <q-date v-model="alterForm.DATUM" mask="YYYY-MM-DD">
                            <div class="row items-center justify-end">
                              <q-btn v-close-popup :label="t('form.close')" color="primary" flat />
                            </div>
                          </q-date>
                        </q-popup-proxy>
                      </q-icon>
                    </template>
                  </q-input>
                </div>

                <div class="col-12 row justify-end q-gutter-x-sm" v-if="isEditingAlter">
                  <q-btn :label="t('form.cancel')" color="grey-7" outline rounded @click="onCancelHeader" dense padding="xs lg" />
                  <q-btn :label="t('form.save')" type="submit" color="primary" rounded unelevated dense padding="xs xl" />
                </div>
              </q-form>
            </q-card-section>
          </q-card>

          <!-- 3. DATEN-GRID (UNTEN) -->
          <q-table
            :rows="alterRows"
            :columns="alterColumns"
            row-key="ID"
            class="huhnlite-grid-standard shadow-2 rounded-borders sticky-header-table"
            :loading="loadingAlter"
            style="height: 500px;"
            :pagination="{ rowsPerPage: 15 }"
          >
            <template v-slot:top>
              <div class="text-h6 text-weight-bold text-primary">{{ t('auto.referenzwerte_details') }}</div>
              <q-space />
              <q-btn v-if="alterForm.ID" :label="t('auto.zeile_hinzufuegen')" color="secondary" icon="add" outline rounded size="sm"
                     @click="openCreateAlter"/>
              <div v-if="alterForm.TABELLENNUMMER" class="text-caption text-grey q-ml-md">{{ t('auto.tab_nr') }}
                {{ alterForm.TABELLENNUMMER }}
              </div>
            </template>
            <template v-slot:body-cell-actions="props">
              <q-td :props="props" auto-width>
                <div class="row no-wrap q-gutter-x-xs">
                  <q-btn dense round icon="edit" color="primary" size="sm" @click="onEditAlter(props.row)" unelevated/>
                  <q-btn dense round icon="delete" color="negative" size="sm" @click="onDeleteAlter(props.row)"
                         unelevated/>
                </div>
              </q-td>
            </template>
          </q-table>
        </div>
      </q-tab-panel>

      <!-- GEWICHTSTABELLE -->
      <q-tab-panel name="gewicht" class="q-pa-none">
        <div class="column q-gutter-y-md">
          <!-- 1. AUSWAHL (OBEN) -->
          <q-card flat bordered class="shadow-2 rounded-borders">
            <q-card-section class="q-pa-sm row items-center q-col-gutter-sm">
              <div class="col-12 col-sm-auto text-subtitle2 text-weight-bold text-primary">{{ t('auto.tabelle_waehlen') }}</div>
              <div class="col-12 col-sm">
                <q-select
                  v-model="selectedGewichtHeaderId"
                  :options="tabellenkopfGewichtOptions"
                  option-label="BEZEICHNUNG_VAL"
                  option-value="ID"
                  emit-value
                  map-options
                  :label="t('auto.vorhandene_gewichtstabellen_typ_g')"
                  filled
                  dense
                  :loading="loadingLookups"
                  :key="tabellenkopfGewichtOptions.length"
                  @update:model-value="onGewichtHeaderSelect"
                >
                  <template v-slot:no-option>
                    <q-item><q-item-section class="text-grey">{{ t('auto.keine_tabellen_gefunden') }}</q-item-section></q-item>
                  </template>
                </q-select>
              </div>
              <div class="col-12 col-sm-auto">
                <q-btn :label="t('auto.neuanlage')" color="secondary" icon="add" @click="onNewGewichtHeader" rounded unelevated dense padding="xs md" class="full-width" />
              </div>
            </q-card-section>
          </q-card>

          <!-- 2. KOPF-FORMULAR (MITTE) -->
          <q-card flat bordered class="shadow-2 rounded-borders">
            <q-card-section class="bg-primary text-white row items-center q-pa-sm">
              <!-- Action: Edit Button (Left) -->
              <q-btn
                v-if="!isEditingGewicht && gewichtForm.ID"
                outline
                rounded
                dense
                icon="edit"
                :label="t('auto.kopfdaten_bearbeiten')"
                class="text-white q-px-md"
                @click="isEditingGewicht = true"
              />
              <q-btn
                v-if="!isEditingGewicht && gewichtForm.ID"
                outline
                rounded
                dense
                icon="delete"
                color="negative"
                :label="t('auto.tabelle_loeschen')"
                class="text-white q-px-md q-ml-sm"
                @click="onDeleteGewichtHeader"
              />
              <q-space />
              <!-- Title (Right) -->
              <div class="row items-center q-gutter-x-sm q-pr-sm">
                <q-icon :name="isEditingGewicht ? 'edit' : 'straighten'" size="xs" />
                <div class="text-subtitle2 text-weight-bold">{{ isEditingGewicht ? t('auto.kopfdaten_bearbeiten') : t('auto.kopfdaten_details') }} ({{ t('auto.gewicht') }})</div>
              </div>
            </q-card-section>
            <q-card-section class="q-pa-md">
              <q-form @submit="onSubmitGewichtHeader" class="row q-col-gutter-md">
                <div class="col-12 col-sm-2">
                  <q-input v-model.number="gewichtForm.TABELLENNUMMER" :label="t('auto.nummer')" filled stack-label type="number"
                           :rules="[val => !!val || t('message.required')]" dense
                           :disable="!!gewichtForm.ID || !isEditingGewicht"/>
                </div>
                <div class="col-12 col-sm-4">
                  <q-input v-model="gewichtForm.BEZEICHNUNG" :label="t('grid.designationRequired')" filled stack-label
                           :rules="[val => !!val || t('message.required')]" dense :disable="!isEditingGewicht"/>
                </div>
                <div class="col-12 col-sm-3">
                  <q-input v-model="gewichtForm.ANLAGEDATUM" :label="t('auto.anlagedatum')" filled stack-label dense
                           mask="####-##-##" :disable="!isEditingGewicht">
                    <template v-slot:append>
                      <q-icon name="event" class="cursor-pointer">
                        <q-popup-proxy cover transition-show="scale" transition-hide="scale" :disabled="!isEditingGewicht">
                          <q-date v-model="gewichtForm.ANLAGEDATUM" mask="YYYY-MM-DD">
                            <div class="row items-center justify-end">
                              <q-btn v-close-popup :label="t('form.close')" color="primary" flat />
                            </div>
                          </q-date>
                        </q-popup-proxy>
                      </q-icon>
                    </template>
                  </q-input>
                </div>
                <div class="col-12 col-sm-3">
                  <q-input v-model="gewichtForm.DATUM" :label="t('auto.gueltigkeit_datum')" filled stack-label dense
                           mask="####-##-##" :disable="!isEditingGewicht">
                    <template v-slot:append>
                      <q-icon name="event" class="cursor-pointer">
                        <q-popup-proxy cover transition-show="scale" transition-hide="scale" :disabled="!isEditingGewicht">
                          <q-date v-model="gewichtForm.DATUM" mask="YYYY-MM-DD">
                            <div class="row items-center justify-end">
                              <q-btn v-close-popup :label="t('form.close')" color="primary" flat />
                            </div>
                          </q-date>
                        </q-popup-proxy>
                      </q-icon>
                    </template>
                  </q-input>
                </div>

                <div class="col-12 row justify-end q-gutter-x-sm" v-if="isEditingGewicht">
                  <q-btn :label="t('auto.abbruch')" color="grey-7" outline rounded @click="onCancelGewichtHeader" dense padding="xs lg" />
                  <q-btn :label="t('form.save')" type="submit" color="primary" rounded unelevated dense padding="xs xl" />
                </div>
              </q-form>
            </q-card-section>
          </q-card>

          <!-- 3. DATEN-GRID (UNTEN) -->
          <q-table
            :rows="gewichtRows"
            :columns="gewichtColumns"
            row-key="ID"
            class="huhnlite-grid-standard shadow-2 rounded-borders sticky-header-table"
            :loading="loadingGewicht"
            style="height: 500px;"
            :pagination="{ rowsPerPage: 15 }"
          >
            <template v-slot:top>
              <div class="text-h6 text-weight-bold text-primary">{{ t('auto.gewichts_referenzwerte_details') }}</div>
              <q-space />
              <q-btn v-if="gewichtForm.ID" :label="t('auto.zeile_hinzufuegen')" color="secondary" icon="add" outline rounded
                     size="sm" @click="openCreateGewicht"/>
              <div v-if="gewichtForm.TABELLENNUMMER" class="text-caption text-grey q-ml-md">{{ t('auto.tab_nr') }}
                {{ gewichtForm.TABELLENNUMMER }}
              </div>
            </template>
            <template v-slot:body-cell-actions="props">
              <q-td :props="props" auto-width>
                <div class="row no-wrap q-gutter-x-xs">
                  <q-btn dense round icon="edit" color="primary" size="sm" @click="onEditGewicht(props.row)"
                         unelevated/>
                  <q-btn dense round icon="delete" color="negative" size="sm" @click="onDeleteGewicht(props.row)"
                         unelevated/>
                </div>
              </q-td>
            </template>
          </q-table>
        </div>
      </q-tab-panel>

      <!-- MEHRWERTSTEUER -->
      <q-tab-panel name="mwst" class="q-pa-none">
        <div style="max-width: 800px;">
          <q-table
            :rows="mwstRows"
            :columns="mwstColumns"
            row-key="ID"
            :loading="loadingMwst"
            class="huhnlite-grid-standard shadow-2 rounded-borders"
            @row-dblclick="(evt, row) => onEditMwst(row)"
          >
            <template v-slot:top>
              <div class="full-width row items-center justify-between">
                <div class="text-h6 text-weight-bold text-primary">{{ t('auto.mwst_steuersaetze') }}</div>
                <q-btn color="primary" icon="add" :label="t('auto.neuer_steuersatz')" @click="openCreateMwst" rounded unelevated/>
              </div>
            </template>
            <template v-slot:body-cell-actions="props">
              <q-td :props="props" auto-width>
                <div class="row no-wrap q-gutter-x-xs justify-center">
                  <q-btn dense round icon="edit" color="primary" @click="onEditMwst(props.row)" unelevated/>
                  <q-btn dense round icon="delete" color="negative" @click="onDeleteMwst(props.row)" unelevated/>
                </div>
              </q-td>
            </template>
          </q-table>
        </div>
      </q-tab-panel>

      <!-- PREISTABELLE -->
      <q-tab-panel name="preise" class="q-pa-none">
        <q-table
          :rows="preiseRows"
          :columns="preiseColumns"
          row-key="ID"
          :loading="loadingPreise"
          class="huhnlite-grid-standard shadow-2 rounded-borders sticky-header-table"
          style="height: 500px;"
          @row-dblclick="(evt, row) => onEditPreis(row)"
        >
          <template v-slot:top>
            <div class="full-width row items-center justify-between">
              <div class="text-h6 text-weight-bold text-primary">{{ t('auto.preistabelle_eierpreise') }}</div>
              <q-btn color="primary" icon="add" :label="t('auto.neu')" @click="openCreatePreis" rounded unelevated/>
            </div>
          </template>
          <template v-slot:body-cell-actions="props">
            <q-td :props="props" auto-width>
              <div class="row no-wrap q-gutter-x-xs justify-center">
                <q-btn dense round icon="edit" color="primary" @click="onEditPreis(props.row)" unelevated size="sm"/>
                <q-btn dense round icon="delete" color="negative" @click="onDeletePreis(props.row)" unelevated
                       size="sm"/>
              </div>
            </q-td>
          </template>
          <template v-slot:body-cell-kz_haltungstyp="props">
            <q-td :props="props">
              <q-badge
                :color="props.value === '0' ? 'green' : (props.value === 'F' || props.value === '1' ? 'blue' : (props.value === 'B' || props.value === '2' ? 'brown' : 'grey'))">
                {{
                  props.value === '0' ? 'Bio' : (props.value === 'F' || props.value === '1' ? 'Freiland' : (props.value === 'B' || props.value === '2' ? 'Boden' : (props.value === '3' ? 'Käfig' : props.value)))
                }}
              </q-badge>
            </q-td>
          </template>
        </q-table>

        <!-- Preis Dialog -->
        <q-dialog v-model="showPreiseDialog" persistent>
          <q-card style="min-width: 450px; border-radius: 12px;">
            <q-card-section class="bg-primary text-white q-pa-md row items-center">
              <div class="text-h6">{{ isEditingPreis ? t('auto.eintrag_bearbeiten') : t('auto.neuer_preiseintrag') }}</div>
              <q-space/>
              <q-btn icon="close" flat round dense v-close-popup/>
            </q-card-section>
            <q-card-section class="q-pa-lg">
              <q-form @submit="onSubmitPreis" class="q-gutter-md">
                <div class="row q-col-gutter-md">
                  <div class="col-12 col-sm-6">
                    <q-select
                      v-model="preisForm.KZ_HALTUNGSTYP"
                      :options="haltungstypOptions"
                      :label="t('auto.haltungstyp')"
                      filled
                      stack-label
                      dense
                      emit-value
                      map-options
                      :rules="[val => !!val || t('message.required')]"
                    />
                  </div>
                  <div class="col-12 col-sm-6">
                    <q-input v-model="preisForm.EIERKLASSE" :label="t('auto.eierklasse_z_b_xl')" filled stack-label dense
                             :rules="[val => !!val || t('message.required')]"/>
                  </div>
                  <div class="col-12 col-sm-6">
                    <q-input v-model.number="preisForm.GEWICHT_VON" type="number" step="0.01" :label="t('auto.gewicht_von_g')"
                             filled stack-label dense :rules="[val => val !== null || t('message.required')]"/>
                  </div>
                  <div class="col-12 col-sm-6">
                    <q-input v-model.number="preisForm.GEWICHT_BIS" type="number" step="0.01" :label="t('auto.gewicht_bis_g')"
                             filled stack-label dense :rules="[val => val !== null || t('message.required')]"/>
                  </div>
                  <div class="col-12 col-sm-6">
                    <q-input v-model.number="preisForm.PREIS_VON" type="number" step="0.01" :label="t('auto.preis_von')" filled
                             stack-label dense :rules="[val => val !== null || t('message.required')]"/>
                  </div>
                  <div class="col-12 col-sm-6">
                    <q-input v-model.number="preisForm.PREIS_BIS" type="number" step="0.01" :label="t('auto.preis_bis')" filled
                             stack-label dense :rules="[val => val !== null || t('message.required')]"/>
                  </div>
                </div>
                <div class="row justify-end q-mt-md q-gutter-x-sm">
                  <q-btn :label="t('form.cancel')" color="grey-7" flat v-close-popup/>
                  <q-btn :label="isEditingPreis ? t('form.update') : t('form.save')" type="submit" color="primary" rounded
                         unelevated padding="xs xl"/>
                </div>
              </q-form>
            </q-card-section>
          </q-card>
        </q-dialog>
      </q-tab-panel>

      <!-- FUTTERSORTEN -->
      <q-tab-panel name="futtersorten" class="q-pa-none">
        <div style="max-width: 600px;">
          <q-table
            :rows="futtersortenRows"
            :columns="futtersortenColumns"
            row-key="ID"
            :loading="loadingFuttersorten"
            class="huhnlite-grid-standard shadow-2 rounded-borders"
            @row-dblclick="(evt, row) => onEditFuttersorte(row)"
          >
            <template v-slot:top>
              <div class="full-width row items-center justify-between">
                <div class="text-h6 text-weight-bold text-primary">{{ t('auto.futtersorten') }}</div>
                <q-btn color="primary" icon="add" :label="t('auto.neue_sorte')" @click="openCreateFuttersorte" rounded unelevated/>
              </div>
            </template>
            <template v-slot:body-cell-actions="props">
              <q-td :props="props" auto-width>
                <div class="row no-wrap q-gutter-x-xs justify-center">
                  <q-btn dense round icon="edit" color="primary" @click="onEditFuttersorte(props.row)" unelevated size="sm"/>
                  <q-btn dense round icon="delete" color="negative" @click="onDeleteFuttersorte(props.row)" unelevated size="sm"/>
                </div>
              </q-td>
            </template>
          </q-table>
        </div>

        <!-- Futtersorte Dialog -->
        <q-dialog v-model="showFuttersortenDialog" persistent>
          <q-card style="min-width: 400px; border-radius: 12px;">
            <q-card-section class="bg-primary text-white q-pa-md row items-center">
              <div class="text-h6">{{ isEditingFuttersorte ? t('auto.sorte_bearbeiten') : t('auto.neue_futtersorte') }}</div>
              <q-space/>
              <q-btn icon="close" flat round dense v-close-popup/>
            </q-card-section>
            <q-card-section class="q-pa-lg">
              <q-form @submit="onSubmitFuttersorte" class="q-gutter-md">
                <q-input v-model="futtersorteForm.BEZEICHNUNG" :label="t('grid.designationRequired')" filled stack-label
                         :rules="[val => !!val || t('message.required')]"/>
                <div class="row justify-end q-mt-md q-gutter-x-sm">
                  <q-btn :label="t('form.cancel')" color="grey-7" flat v-close-popup/>
                  <q-btn :label="isEditingFuttersorte ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl"/>
                </div>
              </q-form>
            </q-card-section>
          </q-card>
        </q-dialog>
      </q-tab-panel>

      <!-- FELD-KONFIGURATION -->
      <q-tab-panel name="feldnamen" class="q-pa-none">
        <q-table
          :rows="fieldRows"
          :columns="fieldColumns"
          row-key="ID"
          :loading="loadingFields"
          class="huhnlite-grid-standard shadow-2 rounded-borders"
          :pagination="{ rowsPerPage: 50, sortBy: 'FELDNAME' }"
        >
          <template v-slot:top>
            <div class="full-width row items-center justify-between">
              <div class="text-h6 text-weight-bold text-primary">Feldnamen & Übersetzungen</div>
              <q-btn color="secondary" icon="search" label="Nach neuen Feldnamen suchen" @click="syncFieldNames"
                     outline rounded size="sm"/>
            </div>
          </template>
          <template v-slot:body-cell-NAMEINDB="props">
            <q-td :props="props" align="center">
              <q-icon :name="props.value === 1 ? 'check_circle' : 'calculate'"
                      :color="props.value === 1 ? 'green' : 'orange'" size="sm">
                <q-tooltip>{{ props.value === 1 ? 'Existiert in DB' : 'Berechnetes Feld / Alias' }}</q-tooltip>
              </q-icon>
            </q-td>
          </template>
          <template v-slot:body-cell-actions="props">
            <q-td :props="props" auto-width>
              <div class="row no-wrap q-gutter-x-xs justify-center">
                <q-btn dense round icon="edit" color="primary" @click="onEditField(props.row)" unelevated size="sm"/>
                <q-btn dense round icon="delete" color="negative" @click="onDeleteField(props.row)" unelevated
                       size="sm"/>
              </div>
            </q-td>
          </template>
        </q-table>

        <!-- Field Edit Dialog -->
        <q-dialog v-model="showFieldDialog" persistent>
          <q-card style="min-width: 450px; border-radius: 12px;">
            <q-card-section class="bg-primary text-white q-pa-md row items-center">
              <div class="text-h6">Feldnamen bearbeiten</div>
              <q-space/>
              <q-btn icon="close" flat round dense v-close-popup/>
            </q-card-section>
            <q-card-section class="q-pa-lg">
              <q-form @submit="onSubmitField" class="q-gutter-md">
                <div class="row q-col-gutter-md">
                  <div class="col-12">
                    <q-input v-model="fieldForm.FELDNAME" label="Technischer Name (DB)" filled dense readonly/>
                  </div>
                  <div class="col-12">
                    <q-input v-model="fieldForm.INHALT" label="Anzeige-Name / Übersetzung *" filled dense
                             :rules="[val => !!val || 'Pflichtfeld']"/>
                  </div>
                  <div class="col-12">
                    <q-input v-model="fieldForm.BETREFF" label="Kurztitel / Betreff" filled dense/>
                  </div>
                  <div class="col-12">
                    <q-checkbox v-model="fieldForm.NAMEINDB" label="Feld existiert in Datenbank-Tabelle" :true-value="1"
                              :false-value="0" color="secondary"/>
                  </div>
                </div>
                <div class="row justify-end q-mt-md q-gutter-x-sm">
                  <q-btn label="Abbrechen" color="grey-7" flat v-close-popup/>
                  <q-btn label="Speichern" type="submit" color="primary" rounded unelevated padding="xs xl"/>
                </div>
              </q-form>
            </q-card-section>
          </q-card>
        </q-dialog>
      </q-tab-panel>
    </q-tab-panels>


    <!-- Alter Dialog -->
    <q-dialog v-model="showAlterDialog" persistent>
      <q-card style="min-width: 500px; border-radius: 12px;">
        <q-card-section class="bg-primary text-white q-pa-md row items-center">
          <div class="text-h6">{{ isEditingAlterRow ? t('auto.lebenswoche_bearbeiten') : t('auto.neue_lebenswoche') }}</div>
          <q-space/>
          <q-btn icon="close" flat round dense v-close-popup/>
        </q-card-section>
        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmitAlterRow" class="row q-col-gutter-md">
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.ALTERINWOCHEN" :label="t('auto.lebenswoche')" type="number" filled
                       :rules="[val => !!val || t('message.required')]"/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.EIGEWICHTWO" :label="t('auto.ei_gewicht_woche_g')" type="number" step="0.1"
                       filled :rules="[val => !!val || t('message.required')]"/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.LEGERATEAH" :label="t('auto.legerate_ah')" type="number" step="0.1"
                       filled/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.LEGERATEDH" :label="t('auto.legerate_dh')" type="number" step="0.1"
                       filled/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.EIGEWICHTKUM" :label="t('auto.ei_gewicht_kum_g')" type="number" step="0.1"
                       filled/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.EIMASSEWO" :label="t('auto.ei_masse_woche_g')" type="number" step="0.1"
                       filled/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.EIMASSEKUM" :label="t('auto.ei_masse_kum_g')" type="number" step="0.1"
                       filled/>
            </div>
            <div class="col-12 col-sm-6">
              <q-input v-model.number="alterRowForm.EIZAHLKUM" :label="t('auto.eier_kum')" type="number" step="0.1" filled/>
            </div>
            <div class="col-12 row justify-end q-mt-md q-gutter-x-sm">
              <q-btn :label="t('form.cancel')" color="grey-7" flat v-close-popup/>
              <q-btn :label="isEditingAlterRow ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- Gewicht Dialog -->
    <q-dialog v-model="showGewichtDialog" persistent>
      <q-card style="min-width: 600px; border-radius: 12px;">
        <q-card-section class="bg-primary text-white q-pa-md row items-center">
          <div class="text-h6">{{ isEditingGewichtRow ? t('auto.referenzgewicht_bearbeiten') : t('auto.neues_referenzgewicht') }}</div>
          <q-space/>
          <q-btn icon="close" flat round dense v-close-popup/>
        </q-card-section>
        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmitGewichtRow" class="row q-col-gutter-md">
            <div class="col-12 col-sm-4">
              <q-input v-model.number="gewichtRowForm.EIGEWICHT" :label="t('auto.ei_gewicht_g')" type="number" step="0.1"
                       filled :rules="[val => !!val || t('message.required')]"/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE1" :label="t('auto.xl')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE2" :label="t('auto.l1')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE3" :label="t('auto.l2')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE4" :label="t('auto.m1')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE5" :label="t('auto.m2')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE6" :label="t('auto.s1')" type="number" step="0.1" filled/>
            </div>
            <div class="col-4 col-sm-2">
              <q-input v-model.number="gewichtRowForm.KLASSE7" :label="t('auto.s2')" type="number" step="0.1" filled/>
            </div>
            <div class="col-12 row justify-end q-mt-md q-gutter-x-sm">
              <q-btn :label="t('form.cancel')" color="grey-7" flat v-close-popup/>
              <q-btn :label="isEditingGewichtRow ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl"/>
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

    <!-- MwSt Dialog -->
    <q-dialog v-model="showMwstDialog" persistent @show="onMwstDialogShow">
      <q-card style="min-width: 400px; max-width: 600px; border-radius: 12px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6">{{ isEditingMwst ? t('auto.steuersatz_bearbeiten') : t('auto.neuer_steuersatz') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeMwstDialog" unelevated color="white" flat />
        </q-card-section>
        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmitMwst" class="q-gutter-md">
            <q-input v-model="mwstForm.MWSTKZ" label="Kennzeichen (z.B. A) *" filled stack-label maxlength="1"
                     :rules="[val => !!val || 'Pflichtfeld']"/>
            <q-input v-model.number="mwstForm.PROZENT" type="number" step="0.01" label="Prozentsatz (%) *" filled
                     stack-label :rules="[val => val !== null || 'Pflichtfeld']"/>
            <q-input v-model="mwstForm.KONTO" label="Konto" filled stack-label/>
            <div class="row justify-end q-mt-md q-gutter-x-sm">
              <q-btn ref="mwstCancelBtn" :label="t('form.cancel')" color="negative" outline rounded @click="closeMwstDialog" />
              <q-btn ref="mwstSaveBtn" :label="isEditingMwst ? t('form.update') : t('form.save')" type="submit" color="primary" rounded unelevated padding="xs xl" />
            </div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>

  </q-page>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { ref, reactive, onMounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { api } from 'src/boot/api';
import type { QTableProps } from 'quasar';

const $q = useQuasar();
const route = useRoute();
const router = useRouter();

interface Tabellenkopf {
  ID: number;
  TABELLENNUMMER: number;
  BEZEICHNUNG?: { String: string; Valid: boolean } | string;
  ANLAGEDATUM?: { String: string; Valid: boolean } | string;
  DATUM?: { String: string; Valid: boolean } | string | null;
  TABELLENNUMMER_VAL?: number;
  BEZEICHNUNG_VAL?: string;
}

interface LSLKlassik {
  ID: number;
  TABELLENNUMMER?: { Int64: number; Valid: boolean } | number;
  ALTERINWOCHEN?: { Int64: number; Valid: boolean } | number;
  EIZAHLKUM?: { Float64: number; Valid: boolean } | number;
  LEGERATEAH?: { Float64: number; Valid: boolean } | number;
  LEGERATEDH?: { Float64: number; Valid: boolean } | number;
  EIGEWICHTWO?: { Float64: number; Valid: boolean } | number;
  EIGEWICHTKUM?: { Float64: number; Valid: boolean } | number;
  EIMASSEWO?: { Float64: number; Valid: boolean } | number;
  EIMASSEKUM?: { Float64: number; Valid: boolean } | number;
}

interface GewichtTabelle {
  ID: number;
  TABELLENNUMMER?: { Int64: number; Valid: boolean } | number;
  EIGEWICHT?: { Float64: number; Valid: boolean } | number;
  KLASSE1?: { Float64: number; Valid: boolean } | number;
  KLASSE2?: { Float64: number; Valid: boolean } | number;
  KLASSE3?: { Float64: number; Valid: boolean } | number;
  KLASSE4?: { Float64: number; Valid: boolean } | number;
  KLASSE5?: { Float64: number; Valid: boolean } | number;
  KLASSE6?: { Float64: number; Valid: boolean } | number;
  KLASSE7?: { Float64: number; Valid: boolean } | number;
}

const getStr = (v: unknown): string => {
  if (v && typeof v === 'object') {
    if ('String' in v) return (v as any).String || '';
    if ('Float64' in v) return String((v as any).Float64 || 0);
  }
  return v ? String(v) : '';
};
const getNum = (v: unknown): number => {
  if (v && typeof v === 'object') {
    if ('Float64' in v) return Number((v as any).Float64 || 0);
    if ('Int64' in v) return Number((v as any).Int64 || 0);
    if ('Int32' in v) return Number((v as any).Int32 || 0);
  }
  if (typeof v === 'string') return parseFloat(v) || 0;
  return typeof v === 'number' ? v : Number(v || 0);
};

interface Mwst {
  ID: number;
  MWSTKZ?: { String: string; Valid: boolean } | string | null;
  PROZENT?: { Float64: number; Valid: boolean } | number | null;
  KONTO?: { String: string; Valid: boolean } | string | null;
}

// Tab Logic
const tab = ref<string>((route.params.tab as string) || 'mwst');
watch(() => route.params.tab, (newTab) => {
  if (newTab) tab.value = newTab as string;
});
watch(tab, (newVal) => {
  void router.push({ params: { tab: newVal } });
});

// Loading states
const loadingAlter = ref(false);
const loadingGewicht = ref(false);
const loadingMwst = ref(false);
const loadingLookups = ref(false);
const loadingFields = ref(false);
const loadingFuttersorten = ref(false);

interface FieldRow {
  ID: number;
  KZ: string;
  FELDNAME: string;
  INHALT: string;
  BETREFF: string;
  NAMEINDB: number;
}

// Placeholder Rows
const alterRows = ref([]);
const gewichtRows = ref([]);
const mwstRows = ref<Mwst[]>([]);
const fieldRows = ref<FieldRow[]>([]);
const futtersortenRows = ref<{ID: number, BEZEICHNUNG: string}[]>([]);

const fieldColumns = computed<QTableProps['columns']>(() => [
  {name: 'actions', align: 'left', label: t('grid.action'), field: 'actions'},
  {name: 'ID', align: 'left', label: 'ID', field: 'ID', sortable: true},
  {name: 'KZ', align: 'left', label: t('auto.kz'), field: (row: FieldRow) => getStr(row.KZ), sortable: true},
  {
    name: 'FELDNAME',
    align: 'left',
    label: t('auto.technischer_name'),
    field: (row: FieldRow) => getStr(row.FELDNAME),
    sortable: true
  },
  {
    name: 'INHALT',
    align: 'left',
    label: t('auto.anzeige_name_uebersetzung_col'),
    field: (row: FieldRow) => getStr(row.INHALT),
    sortable: true
  },
  {
    name: 'NAMEINDB',
    align: 'center',
    label: t('auto.in_db'),
    field: (row: FieldRow) => getNum(row.NAMEINDB),
    sortable: true
  }
]);

const mwstColumns = computed<QTableProps['columns']>(() => [
  {name: 'actions', align: 'center', label: t('grid.action'), field: 'actions'},
  {name: 'MWSTKZ', align: 'left', label: t('auto.kennzeichen'), field: 'MWSTKZ', sortable: true},
  {name: 'PROZENT', align: 'right', label: t('auto.prozent_percent'), field: (row: Mwst) => getNum(row.PROZENT), sortable: true},
  {name: 'KONTO', align: 'left', label: t('auto.konto'), field: (row: Mwst) => getStr(row.KONTO) || '-', sortable: true}
]);

const futtersortenColumns = computed<QTableProps['columns']>(() => [
  { name: 'actions', align: 'center', label: t('grid.action'), field: 'actions' },
  { name: 'BEZEICHNUNG', align: 'left', label: t('grid.designation'), field: 'BEZEICHNUNG', sortable: true }
]);

const alterColumns = computed<QTableProps['columns']>(() => [
  {name: 'actions', align: 'left', label: t('grid.action'), field: 'actions'},
  {
    name: 'ALTERINWOCHEN',
    align: 'left',
    label: t('auto.l_woche'),
    field: (row: LSLKlassik) => getNum(row.ALTERINWOCHEN),
    sortable: true
  },
  {
    name: 'EIGEWICHTWO',
    align: 'right',
    label: t('auto.ei_gew_woche_g'),
    field: (row: LSLKlassik) => getNum(row.EIGEWICHTWO),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'LEGERATEAH',
    align: 'right',
    label: t('auto.legerate_ah'),
    field: (row: LSLKlassik) => getNum(row.LEGERATEAH),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'LEGERATEDH',
    align: 'right',
    label: t('auto.legerate_dh'),
    field: (row: LSLKlassik) => getNum(row.LEGERATEDH),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'EIGEWICHTKUM',
    align: 'right',
    label: t('auto.ei_gew_kum_g'),
    field: (row: LSLKlassik) => getNum(row.EIGEWICHTKUM),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'EIMASSEWO',
    align: 'right',
    label: t('auto.ei_masse_woche_g'),
    field: (row: LSLKlassik) => getNum(row.EIMASSEWO),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'EIMASSEKUM',
    align: 'right',
    label: t('auto.ei_masse_kum_g'),
    field: (row: LSLKlassik) => getNum(row.EIMASSEKUM),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'EIZAHLKUM',
    align: 'right',
    label: t('auto.eier_kum'),
    field: (row: LSLKlassik) => getNum(row.EIZAHLKUM),
    format: (val: number) => val.toFixed(1)
  }
]);

const gewichtColumns = computed<QTableProps['columns']>(() => [
  {name: 'actions', align: 'left', label: t('grid.action'), field: 'actions'},
  {
    name: 'EIGEWICHT',
    align: 'left',
    label: t('auto.ei_gewicht_g_col'),
    field: (row: GewichtTabelle) => getNum(row.EIGEWICHT),
    format: (val: number) => val.toFixed(1),
    sortable: true
  },
  {
    name: 'KLASSE1',
    align: 'right',
    label: t('auto.xl'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE1),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE2',
    align: 'right',
    label: t('auto.l1'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE2),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE3',
    align: 'right',
    label: t('auto.l2'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE3),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE4',
    align: 'right',
    label: t('auto.m1'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE4),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE5',
    align: 'right',
    label: t('auto.m2'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE5),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE6',
    align: 'right',
    label: t('auto.s1'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE6),
    format: (val: number) => val.toFixed(1)
  },
  {
    name: 'KLASSE7',
    align: 'right',
    label: t('auto.s2'),
    field: (row: GewichtTabelle) => getNum(row.KLASSE7),
    format: (val: number) => val.toFixed(1)
  }
]);

// MwSt Logic
const showMwstDialog = ref(false);
const isEditingMwst = ref(false);
const editMwstId = ref<number | null>(null);
const mwstForm = reactive({
  MWSTKZ: '',
  PROZENT: null as number | null,
  KONTO: ''
});

// Alterstabelle Logic
const alterForm = reactive({
  ID: null as number | null,
  TABELLENTYP: 'A',
  TABELLENNUMMER: null as number | null,
  BEZEICHNUNG: '',
  ANLAGEDATUM: '',
  DATUM: ''
});
const tabellenkopfOptions = ref<Tabellenkopf[]>([]);
const selectedHeaderId = ref<number | null>(null);

// Gewichtstabelle Logic
const gewichtForm = reactive({
  ID: null as number | null,
  TABELLENTYP: 'G',
  TABELLENNUMMER: null as number | null,
  BEZEICHNUNG: '',
  ANLAGEDATUM: '',
  DATUM: ''
});
const isEditingAlter = ref(false);
const isEditingGewicht = ref(false);
const tabellenkopfGewichtOptions = ref<Tabellenkopf[]>([]);
const selectedGewichtHeaderId = ref<number | null>(null);

// Alter Row Logic
const showAlterDialog = ref(false);
const isEditingAlterRow = ref(false);
const editAlterRowId = ref<number | null>(null);
const alterRowForm = reactive({
  ALTERINWOCHEN: 0,
  EIGEWICHTWO: 0,
  LEGERATEAH: 0,
  LEGERATEDH: 0,
  EIGEWICHTKUM: 0,
  EIMASSEWO: 0,
  EIMASSEKUM: 0,
  EIZAHLKUM: 0
});

// Gewicht Row Logic
const showGewichtDialog = ref(false);
const isEditingGewichtRow = ref(false);
const editGewichtRowId = ref<number | null>(null);
const gewichtRowForm = reactive({
  EIGEWICHT: 0,
  KLASSE1: 0,
  KLASSE2: 0,
  KLASSE3: 0,
  KLASSE4: 0,
  KLASSE5: 0,
  KLASSE6: 0,
  KLASSE7: 0
});

// Futtersorten Logic
const showFuttersortenDialog = ref(false);
const isEditingFuttersorte = ref(false);
const editFuttersorteId = ref<number | null>(null);
const futtersorteForm = reactive({ BEZEICHNUNG: '' });

// Field Config Logic
const showFieldDialog = ref(false);
const fieldForm = reactive({
  ID: 0,
  FELDNAME: '',
  INHALT: '',
  BETREFF: '',
  NAMEINDB: 1
});


async function loadFields() {
  loadingFields.value = true;
  try {
    const res = await api.get('/api/field-configs');
    fieldRows.value = res.data;
  } catch (err) {
    console.error('Error loading fields:', err);
  } finally {
    loadingFields.value = false;
  }
}

function onEditField(row: FieldRow) {
  fieldForm.ID = row.ID;
  fieldForm.FELDNAME = getStr(row.FELDNAME);
  fieldForm.INHALT = getStr(row.INHALT);
  fieldForm.BETREFF = getStr(row.BETREFF);
  fieldForm.NAMEINDB = getNum(row.NAMEINDB);
  showFieldDialog.value = true;
}

async function onSubmitField() {
  try {
    await api.put(`/api/field-configs/${fieldForm.ID}`, {
      INHALT: fieldForm.INHALT,
      BETREFF: fieldForm.BETREFF,
      NAMEINDB: fieldForm.NAMEINDB
    });
    $q.notify({type: 'positive', message: 'Feldname aktualisiert'});
    showFieldDialog.value = false;
    void loadFields();
  } catch (_err: unknown) {
    $q.notify({type: 'negative', message: 'Fehler beim Speichern'});
  }
}

function onDeleteField(row: FieldRow) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: `Soll das technische Feld '${row.FELDNAME}' wirklich gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/field-configs/${row.ID}`);
      $q.notify({type: 'positive', message: 'Feld gelöscht'});
      void loadFields();
    } catch (err) {
      $q.notify({type: 'negative', message: 'Fehler beim Löschen'});
    }
  });
}

async function syncFieldNames() {
  loadingFields.value = true;
  try {
    const res = await api.post('/api/field-configs/sync');
    const {message, new_count} = res.data as { message: string, new_count: number };
    $q.notify({
      type: new_count > 0 ? 'positive' : 'info',
      message: message,
      icon: 'sync'
    });
    void loadFields();
  } catch (err: unknown) {
    console.error('Sync error:', err);
    $q.notify({type: 'negative', message: 'Fehler bei der Synchronisation'});
  } finally {
    loadingFields.value = false;
  }
}

const originalMwstState = ref('');
const mwstCancelBtn = ref<{ $el: HTMLElement } | null>(null);
const mwstSaveBtn = ref<{ $el: HTMLElement } | null>(null);

const haltungstypOptions = ref<{ label: string, value: string }[]>([]);

function onMwstDialogShow() {
  originalMwstState.value = JSON.stringify(mwstForm);
  setTimeout(() => { (mwstCancelBtn.value)?.$el?.focus(); }, 50);
}

async function loadMwst() {
  loadingMwst.value = true;
  try {
    const res = await api.get('/api/mwst');
    mwstRows.value = (res.data as Mwst[]) || [];
  } catch (err) {
    console.error('loadMwst error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (MwSt)' });
  } finally {
    loadingMwst.value = false;
  }
}

async function loadTabellenkopfLookups() {
  loadingLookups.value = true;
  try {
    const resA = await api.get('/api/tabellenkopf/typ/A');
    if (resA.data && Array.isArray(resA.data)) {
      tabellenkopfOptions.value = resA.data.map((o: Tabellenkopf) => ({
        ...o,
        TABELLENNUMMER_VAL: getNum(o.TABELLENNUMMER),
        BEZEICHNUNG_VAL: getStr(o.BEZEICHNUNG)
      }));
      // Auto-select 1st entry for Type A
      if (!selectedHeaderId.value && tabellenkopfOptions.value.length > 0) {
        const firstA = tabellenkopfOptions.value[0];
        if (firstA) {
          onHeaderSelect(firstA.ID);
          selectedHeaderId.value = firstA.ID;
        }
      }
    } else {
      tabellenkopfOptions.value = [];
    }

    const resG = await api.get('/api/tabellenkopf/typ/G');
    if (resG.data && Array.isArray(resG.data)) {
      tabellenkopfGewichtOptions.value = resG.data.map((o: Tabellenkopf) => ({
        ...o,
        TABELLENNUMMER_VAL: getNum(o.TABELLENNUMMER),
        BEZEICHNUNG_VAL: getStr(o.BEZEICHNUNG)
      }));
      // Auto-select 1st entry for Type G
      if (!selectedGewichtHeaderId.value && tabellenkopfGewichtOptions.value.length > 0) {
        const firstG = tabellenkopfGewichtOptions.value[0];
        if (firstG) {
          onGewichtHeaderSelect(firstG.ID);
          selectedGewichtHeaderId.value = firstG.ID;
        }
      }
    } else {
      tabellenkopfGewichtOptions.value = [];
    }
  } catch (err: unknown) {
    console.error('Fehler beim Laden der Tabellenkopf-Daten:', err);
    const e = err as Record<string, unknown>;
    const eResp = e.response as Record<string, unknown>;
    const eData = eResp?.data as Record<string, string>;
    const msg = eData?.error || (e.message as string) || '';
    $q.notify({ type: 'negative', message: `Fehler beim Laden der Tabellen-Liste: ${msg}` });
  } finally {
    loadingLookups.value = false;
  }
}

async function loadAlterDetails(tabNum: number) {
  loadingAlter.value = true;
  try {
    const res = await api.get(`/api/lsl_klassik/tabnum/${tabNum}`);
    alterRows.value = res.data || [];
  } catch (err: unknown) {
    console.error('loadAlterDetails error:', err);
    const e = err as Record<string, unknown>;
    const eResp = e.response as Record<string, unknown>;
    const eData = eResp?.data as Record<string, string>;
    const msg = eData?.error || (e.message as string) || 'Fehler beim Laden';
    $q.notify({ type: 'negative', message: `Fehler beim Laden der Detaildaten: ${msg}` });
  } finally {
    loadingAlter.value = false;
  }
}

function onHeaderSelect(id: number | null) {
  if (!id) return;
  const header = tabellenkopfOptions.value.find(o => o.ID === id);
  if (header) {
    alterForm.ID = header.ID;
    alterForm.TABELLENNUMMER = header.TABELLENNUMMER_VAL || null;
    alterForm.BEZEICHNUNG = header.BEZEICHNUNG_VAL || '';
    alterForm.ANLAGEDATUM = getStr(header.ANLAGEDATUM);
    alterForm.DATUM = getStr(header.DATUM);
    isEditingAlter.value = false;

    if (alterForm.TABELLENNUMMER) {
      void loadAlterDetails(alterForm.TABELLENNUMMER);
    }
  }
}

function onNewHeader() {
  selectedHeaderId.value = null;
  alterForm.ID = null;
  alterForm.TABELLENNUMMER = null;
  alterForm.BEZEICHNUNG = '';
  alterForm.ANLAGEDATUM = (new Date().toISOString().split('T')[0]) || '';
  alterForm.DATUM = '';
  alterRows.value = [];
  isEditingAlter.value = true;
}

function onCancelHeader() {
  if (selectedHeaderId.value) {
    onHeaderSelect(selectedHeaderId.value);
  } else {
    onNewHeader();
  }
  isEditingAlter.value = false;
}

async function onSubmitAlterHeader() {
  try {
    const payload = {
      TABELLENTYP: 'A',
      TABELLENNUMMER: Number(alterForm.TABELLENNUMMER),
      BEZEICHNUNG: alterForm.BEZEICHNUNG,
      ANLAGEDATUM: alterForm.ANLAGEDATUM,
      DATUM: alterForm.DATUM
    };

    if (alterForm.ID) {
      await api.put(`/api/tabellenkopf/${alterForm.ID}`, payload);
      $q.notify({ type: 'positive', message: 'Kopfdaten aktualisiert' });
    } else {
      const res = await api.post('/api/tabellenkopf', payload);
      alterForm.ID = (res.data as { ID: number }).ID;
      $q.notify({ type: 'positive', message: 'Neuer Tabellenkopf angelegt' });
    }

    void loadTabellenkopfLookups();
    // Nach dem Speichern den Eintrag in der Auswahl suchen und selektieren
    const updated = tabellenkopfOptions.value.find((o: Tabellenkopf) => o.TABELLENNUMMER_VAL === Number(alterForm.TABELLENNUMMER));
    if (updated) selectedHeaderId.value = updated.ID;
    isEditingAlter.value = false;

  } catch (err) {
    console.error('onSubmitAlterHeader error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern der Kopfdaten (Alterstabelle)' });
  }
}

async function onDeleteHeader() {
  if (!alterForm.ID) return;
  $q.dialog({
    title: 'Tabelle löschen',
    message: 'Möchten Sie die gesamte Tabelle (Kopfdaten und alle Zeilen) unwiderruflich löschen?',
    cancel: true,
    persistent: true,
    ok: { label: 'Löschen', color: 'negative', rounded: true, unelevated: true },
    cancel: { label: 'Abbrechen', flat: true }
  }).onOk(async () => {
    try {
      await api.delete(`/api/tabellenkopf/${alterForm.ID}`);
      $q.notify({ type: 'positive', message: 'Tabelle wurde gelöscht' });
      selectedHeaderId.value = null;
      onNewHeader();
      void loadTabellenkopfLookups();
    } catch (err) {
      console.error('onDeleteHeader error:', err);
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen der Tabelle' });
    }
  });
}

// Gewichtstabelle Methoden
async function loadGewichtDetails(tabNum: number) {
  loadingGewicht.value = true;
  try {
    const res = await api.get(`/api/gewichttabelle/tabnum/${tabNum}`);
    gewichtRows.value = res.data || [];
  } catch (err: unknown) {
    console.error('loadGewichtDetails error:', err);
    const e = err as Record<string, unknown>;
    const eResp = e.response as Record<string, unknown>;
    const eData = eResp?.data as Record<string, string>;
    const msg = eData?.error || (e.message as string) || 'Fehler beim Laden';
    $q.notify({ type: 'negative', message: `Fehler beim Laden der Gewicht-Details: ${msg}` });
  } finally {
    loadingGewicht.value = false;
  }
}

function onGewichtHeaderSelect(id: number | null) {
  if (!id) return;
  const header = tabellenkopfGewichtOptions.value.find(o => o.ID === id);
  if (header) {
    gewichtForm.ID = header.ID;
    gewichtForm.TABELLENNUMMER = header.TABELLENNUMMER_VAL || null;
    gewichtForm.BEZEICHNUNG = header.BEZEICHNUNG_VAL || '';
    gewichtForm.ANLAGEDATUM = getStr(header.ANLAGEDATUM);
    gewichtForm.DATUM = getStr(header.DATUM);
    isEditingGewicht.value = false;

    if (gewichtForm.TABELLENNUMMER) {
      void loadGewichtDetails(gewichtForm.TABELLENNUMMER);
    }
  }
}

function onNewGewichtHeader() {
  selectedGewichtHeaderId.value = null;
  gewichtForm.ID = null;
  gewichtForm.TABELLENNUMMER = null;
  gewichtForm.BEZEICHNUNG = '';
  gewichtForm.ANLAGEDATUM = (new Date().toISOString().split('T')[0]) || '';
  gewichtForm.DATUM = '';
  gewichtRows.value = [];
  isEditingGewicht.value = true;
}

function onCancelGewichtHeader() {
  if (selectedGewichtHeaderId.value) {
    onGewichtHeaderSelect(selectedGewichtHeaderId.value);
  } else {
    onNewGewichtHeader();
  }
  isEditingGewicht.value = false;
}

async function onSubmitGewichtHeader() {
  try {
    const payload = {
      TABELLENTYP: 'G',
      TABELLENNUMMER: Number(gewichtForm.TABELLENNUMMER),
      BEZEICHNUNG: gewichtForm.BEZEICHNUNG,
      ANLAGEDATUM: gewichtForm.ANLAGEDATUM,
      DATUM: gewichtForm.DATUM
    };

    if (gewichtForm.ID) {
      await api.put(`/api/tabellenkopf/${gewichtForm.ID}`, payload);
      $q.notify({ type: 'positive', message: 'Gewicht-Kopfdaten aktualisiert' });
    } else {
      const res = await api.post('/api/tabellenkopf', payload);
      gewichtForm.ID = (res.data as { ID: number }).ID;
      $q.notify({ type: 'positive', message: 'Neuer Gewicht-Tabellenkopf angelegt' });
    }

    void loadTabellenkopfLookups();
    const updated = tabellenkopfGewichtOptions.value.find((o: Tabellenkopf) => o.TABELLENNUMMER_VAL === Number(gewichtForm.TABELLENNUMMER));
    if (updated) selectedGewichtHeaderId.value = updated.ID;
    isEditingGewicht.value = false;

  } catch (err) {
    console.error('onSubmitGewichtHeader error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern der Kopfdaten (Gewichtstabelle)' });
  }
}

async function onDeleteGewichtHeader() {
  if (!gewichtForm.ID) return;
  $q.dialog({
    title: 'Tabelle löschen',
    message: 'Möchten Sie die gesamte Tabelle (Kopfdaten und alle Zeilen) unwiderruflich löschen?',
    cancel: true,
    persistent: true,
    ok: { label: 'Löschen', color: 'negative', rounded: true, unelevated: true },
    cancel: { label: 'Abbrechen', flat: true }
  }).onOk(async () => {
    try {
      await api.delete(`/api/tabellenkopf/${gewichtForm.ID}`);
      $q.notify({ type: 'positive', message: 'Tabelle wurde gelöscht' });
      selectedGewichtHeaderId.value = null;
      onNewGewichtHeader();
      void loadTabellenkopfLookups();
    } catch (err) {
      console.error('onDeleteGewichtHeader error:', err);
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen der Tabelle' });
    }
  });
}

// Alter Row Methods
function openCreateAlter() {
  isEditingAlterRow.value = false;
  editAlterRowId.value = null;
  const lastRow = alterRows.value.length > 0 ? alterRows.value[alterRows.value.length - 1] : null;
  Object.assign(alterRowForm, {
    ALTERINWOCHEN: lastRow ? getNum((lastRow as LSLKlassik).ALTERINWOCHEN) + 1 : 1,
    EIGEWICHTWO: 0, LEGERATEAH: 0, LEGERATEDH: 0, EIGEWICHTKUM: 0, EIMASSEWO: 0, EIMASSEKUM: 0, EIZAHLKUM: 0
  });
  showAlterDialog.value = true;
}

function onEditAlter(row: LSLKlassik) {
  isEditingAlterRow.value = true;
  editAlterRowId.value = row.ID;
  Object.assign(alterRowForm, {
    ALTERINWOCHEN: getNum(row.ALTERINWOCHEN),
    EIGEWICHTWO: getNum(row.EIGEWICHTWO),
    LEGERATEAH: getNum(row.LEGERATEAH),
    LEGERATEDH: getNum(row.LEGERATEDH),
    EIGEWICHTKUM: getNum(row.EIGEWICHTKUM),
    EIMASSEWO: getNum(row.EIMASSEWO),
    EIMASSEKUM: getNum(row.EIMASSEKUM),
    EIZAHLKUM: getNum(row.EIZAHLKUM)
  });
  showAlterDialog.value = true;
}

async function onSubmitAlterRow() {
  try {
    const payload = {
      ALTERINWOCHEN: Number(alterRowForm.ALTERINWOCHEN),
      EIGEWICHTWO: Number(alterRowForm.EIGEWICHTWO),
      LEGERATEAH: Number(alterRowForm.LEGERATEAH),
      LEGERATEDH: Number(alterRowForm.LEGERATEDH),
      EIGEWICHTKUM: Number(alterRowForm.EIGEWICHTKUM),
      EIMASSEWO: Number(alterRowForm.EIMASSEWO),
      EIMASSEKUM: Number(alterRowForm.EIMASSEKUM),
      EIZAHLKUM: Number(alterRowForm.EIZAHLKUM),
      TABELLENNUMMER: alterForm.TABELLENNUMMER
    };
    if (isEditingAlterRow.value) {
      await api.put(`/api/lsl_klassik/${editAlterRowId.value}`, payload);
    } else {
      await api.post('/api/lsl_klassik', payload);
    }
    showAlterDialog.value = false;
    if (alterForm.TABELLENNUMMER) void loadAlterDetails(alterForm.TABELLENNUMMER);
    $q.notify({type: 'positive', message: 'Gespeichert'});
  } catch (_err: unknown) {
    $q.notify({type: 'negative', message: 'Fehler beim Speichern'});
  }
}

function onDeleteAlter(row: LSLKlassik) {
  $q.dialog({title: 'Löschen', message: 'Zeile löschen?', cancel: true}).onOk(async () => {
    try {
      await api.delete(`/api/lsl_klassik/${row.ID}`);
      if (alterForm.TABELLENNUMMER) void loadAlterDetails(alterForm.TABELLENNUMMER);
    } catch {
      $q.notify({type: 'negative', message: 'Fehler beim Löschen'});
    }
  });
}

// Gewicht Row Methods
function openCreateGewicht() {
  isEditingGewichtRow.value = false;
  editGewichtRowId.value = null;
  Object.assign(gewichtRowForm, {
    EIGEWICHT: 0,
    KLASSE1: 0,
    KLASSE2: 0,
    KLASSE3: 0,
    KLASSE4: 0,
    KLASSE5: 0,
    KLASSE6: 0,
    KLASSE7: 0
  });
  showGewichtDialog.value = true;
}

function onEditGewicht(row: GewichtTabelle) {
  isEditingGewichtRow.value = true;
  editGewichtRowId.value = row.ID;
  Object.assign(gewichtRowForm, {
    EIGEWICHT: getNum(row.EIGEWICHT),
    KLASSE1: getNum(row.KLASSE1),
    KLASSE2: getNum(row.KLASSE2),
    KLASSE3: getNum(row.KLASSE3),
    KLASSE4: getNum(row.KLASSE4),
    KLASSE5: getNum(row.KLASSE5),
    KLASSE6: getNum(row.KLASSE6),
    KLASSE7: getNum(row.KLASSE7)
  });
  showGewichtDialog.value = true;
}

async function onSubmitGewichtRow() {
  try {
    const payload = {
      EIGEWICHT: Number(gewichtRowForm.EIGEWICHT),
      KLASSE1: Number(gewichtRowForm.KLASSE1),
      KLASSE2: Number(gewichtRowForm.KLASSE2),
      KLASSE3: Number(gewichtRowForm.KLASSE3),
      KLASSE4: Number(gewichtRowForm.KLASSE4),
      KLASSE5: Number(gewichtRowForm.KLASSE5),
      KLASSE6: Number(gewichtRowForm.KLASSE6),
      KLASSE7: Number(gewichtRowForm.KLASSE7),
      TABELLENNUMMER: gewichtForm.TABELLENNUMMER
    };
    if (isEditingGewichtRow.value) {
      await api.put(`/api/gewichttabelle/${editGewichtRowId.value}`, payload);
    } else {
      await api.post('/api/gewichttabelle', payload);
    }
    showGewichtDialog.value = false;
    if (gewichtForm.TABELLENNUMMER) void loadGewichtDetails(gewichtForm.TABELLENNUMMER);
    $q.notify({type: 'positive', message: 'Gespeichert'});
  } catch (err) {
    $q.notify({type: 'negative', message: 'Fehler beim Speichern'});
  }
}

function onDeleteGewicht(row: GewichtTabelle) {
  $q.dialog({title: 'Löschen', message: 'Zeile löschen?', cancel: true}).onOk(() => {
    void (async () => {
      try {
        await api.delete(`/api/gewichttabelle/${row.ID}`);
        if (gewichtForm.TABELLENNUMMER) void loadGewichtDetails(gewichtForm.TABELLENNUMMER);
      } catch (_err: unknown) {
        $q.notify({type: 'negative', message: 'Fehler beim Löschen'});
      }
    })();
  });
}

function openCreateMwst() {
  isEditingMwst.value = false;
  editMwstId.value = null;
  mwstForm.MWSTKZ = '';
  mwstForm.PROZENT = null;
  mwstForm.KONTO = '';
  showMwstDialog.value = true;
}

function onEditMwst(row: Mwst) {
  isEditingMwst.value = true;
  editMwstId.value = row.ID;
  mwstForm.MWSTKZ = getStr(row.MWSTKZ);
  mwstForm.PROZENT = getNum(row.PROZENT);
  mwstForm.KONTO = getStr(row.KONTO);
  showMwstDialog.value = true;
}

// Futtersorten Methods
async function loadFuttersorten() {
  loadingFuttersorten.value = true;
  try {
    const res = await api.get('/api/futtersorten');
    futtersortenRows.value = res.data || [];
  } catch (err) {
    console.error('loadFuttersorten error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden (Futtersorten)' });
  } finally {
    loadingFuttersorten.value = false;
  }
}

function openCreateFuttersorte() {
  isEditingFuttersorte.value = false;
  editFuttersorteId.value = null;
  futtersorteForm.BEZEICHNUNG = '';
  showFuttersortenDialog.value = true;
}

function onEditFuttersorte(row: {ID: number, BEZEICHNUNG: string}) {
  isEditingFuttersorte.value = true;
  editFuttersorteId.value = row.ID;
  futtersorteForm.BEZEICHNUNG = row.BEZEICHNUNG;
  showFuttersortenDialog.value = true;
}

async function onSubmitFuttersorte() {
  try {
    if (isEditingFuttersorte.value && editFuttersorteId.value) {
      await api.put(`/api/futtersorten/${editFuttersorteId.value}`, { bezeichnung: futtersorteForm.BEZEICHNUNG });
      $q.notify({ type: 'positive', message: 'Sorte aktualisiert' });
    } else {
      await api.post('/api/futtersorten', { bezeichnung: futtersorteForm.BEZEICHNUNG });
      $q.notify({ type: 'positive', message: 'Neue Sorte angelegt' });
    }
    showFuttersortenDialog.value = false;
    void loadFuttersorten();
  } catch (_err: unknown) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}

async function onDeleteFuttersorte(row: {ID: number, BEZEICHNUNG: string}) {
  $q.dialog({
    title: 'Löschen bestätigen',
    message: `Soll die Futtersorte '${row.BEZEICHNUNG}' wirklich gelöscht werden?`,
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/futtersorten/${row.ID}`);
      $q.notify({ type: 'positive', message: 'Sorte gelöscht' });
      void loadFuttersorten();
    } catch (err) {
      $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
    }
  });
}

function closeMwstDialog() {
  showMwstDialog.value = false;
}

function onDeleteMwst(row: Mwst) {
  $q.dialog({
    title: 'Löschen',
    message: 'Möchten Sie diesen Steuersatz löschen?',
    cancel: true,
    persistent: true
  }).onOk(() => {
    void (async () => {
      try {
        await api.delete(`/api/mwst/${row.ID}`);
        $q.notify({ type: 'positive', message: 'Gelöscht' });
        void loadMwst();
      } catch {
        $q.notify({ type: 'negative', message: 'Fehler' });
      }
    })();
  });
}

async function onSubmitMwst() {
  try {
    const payload = {
      MWSTKZ: mwstForm.MWSTKZ,
      PROZENT: Number(mwstForm.PROZENT),
      KONTO: mwstForm.KONTO
    };
    if (isEditingMwst.value) {
      await api.put(`/api/mwst/${editMwstId.value}`, payload);
    } else {
      await api.post('/api/mwst', payload);
    }
    $q.notify({ type: 'positive', message: 'Gespeichert' });
    showMwstDialog.value = false;
    void loadMwst();
  } catch (err) {
    console.error('onSubmitMwst error:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern (MwSt)' });
  }
}


// Preise Logic
const preiseRows = ref([]);
const loadingPreise = ref(false);
const showPreiseDialog = ref(false);
const isEditingPreis = ref(false);
const editPreisId = ref<number | null>(null);
const preisForm = reactive({
  KZ_HALTUNGSTYP: '',
  EIERKLASSE: '',
  GEWICHT_VON: null as number | null,
  GEWICHT_BIS: null as number | null,
  PREIS_VON: null as number | null,
  PREIS_BIS: null as number | null
});

const preiseColumns = computed<QTableProps['columns']>(() => [
  {name: 'actions', align: 'center', label: t('grid.action'), field: 'actions'},
  {name: 'KZ_HALTUNGSTYP', align: 'left', label: t('auto.h_typ'), field: 'KZ_HALTUNGSTYP', sortable: true},
  {name: 'EIERKLASSE', align: 'left', label: t('auto.klasse'), field: 'EIERKLASSE', sortable: true},
  {
    name: 'gewicht_von',
    align: 'right',
    label: t('auto.gew_von'),
    field: (row: any) => getNum(row.GEWICHT_VON),
    format: (val: number) => val.toFixed(2),
    sortable: true
  },
  {
    name: 'gewicht_bis',
    align: 'right',
    label: t('auto.gew_bis'),
    field: (row: any) => getNum(row.GEWICHT_BIS),
    format: (val: number) => val.toFixed(2),
    sortable: true
  },
  {
    name: 'preis_von',
    align: 'right',
    label: t('auto.preis_von_col'),
    field: (row: any) => getNum(row.PREIS_VON),
    format: (val: number) => val.toFixed(2),
    sortable: true
  },
  {
    name: 'preis_bis',
    align: 'right',
    label: t('auto.preis_bis_col'),
    field: (row: any) => getNum(row.PREIS_BIS),
    format: (val: number) => val.toFixed(2),
    sortable: true
  }
]);

async function loadPreise() {
  loadingPreise.value = true;
  try {
    const res = await api.get('/api/preise');
    preiseRows.value = (res.data as never[]) || [];
  } catch (err) {
    console.error('loadPreise error:', err);
    $q.notify({type: 'negative', message: 'Fehler beim Laden (Preise)'});
  } finally {
    loadingPreise.value = false;
  }
}

async function loadHaltungstypen() {
  try {
    const res = await api.get('/api/texte/H');
    haltungstypOptions.value = ((res.data as { TEXT: string, KZ: string }[]) || []).map((t) => ({
      label: t.TEXT,
      value: t.KZ
    }));
    if (haltungstypOptions.value.length === 0) {
      haltungstypOptions.value = [
        {label: 'Bio', value: '0'},
        {label: 'Freiland', value: '1'},
        {label: 'Boden', value: '2'},
        {label: 'Käfig', value: '3'}
      ];
    }
  } catch (_err: unknown) {
    haltungstypOptions.value = [
      {label: 'Bio', value: '0'},
      {label: 'Freiland', value: '1'},
      {label: 'Boden', value: '2'},
      {label: 'Käfig', value: '3'}
    ];
  }
}

function openCreatePreis() {
  isEditingPreis.value = false;
  editPreisId.value = null;
  Object.assign(preisForm, {
    KZ_HALTUNGSTYP: '2', // Default Boden
    EIERKLASSE: '',
    GEWICHT_VON: 0,
    GEWICHT_BIS: 0,
    PREIS_VON: 0,
    PREIS_BIS: 0
  });
  showPreiseDialog.value = true;
}

function onEditPreis(row: any) {
  isEditingPreis.value = true;
  editPreisId.value = row.ID;
  Object.assign(preisForm, {
    KZ_HALTUNGSTYP: getStr(row.KZ_HALTUNGSTYP),
    EIERKLASSE: getStr(row.EIERKLASSE),
    GEWICHT_VON: getNum(row.GEWICHT_VON),
    GEWICHT_BIS: getNum(row.GEWICHT_BIS),
    PREIS_VON: getNum(row.PREIS_VON),
    PREIS_BIS: getNum(row.PREIS_BIS)
  });
  showPreiseDialog.value = true;
}

async function onSubmitPreis() {
  try {
    const payload = {
      KZ_HALTUNGSTYP: preisForm.KZ_HALTUNGSTYP,
      EIERKLASSE: preisForm.EIERKLASSE,
      GEWICHT_VON: Number(preisForm.GEWICHT_VON),
      GEWICHT_BIS: Number(preisForm.GEWICHT_BIS),
      PREIS_VON: Number(preisForm.PREIS_VON),
      PREIS_BIS: Number(preisForm.PREIS_BIS)
    };
    if (isEditingPreis.value) {
      await api.put(`/api/preise/${editPreisId.value}`, payload);
    } else {
      await api.post('/api/preise', payload);
    }
    $q.notify({type: 'positive', message: 'Gespeichert'});
    showPreiseDialog.value = false;
    void loadPreise();
  } catch (err) {
    console.error('onSubmitPreis error:', err);
    $q.notify({type: 'negative', message: 'Fehler beim Speichern'});
  }
}

async function onDeletePreis(row: any) {
  $q.dialog({
    title: 'Löschen',
    message: 'Eintrag wirklich löschen?',
    cancel: true,
    persistent: true
  }).onOk(async () => {
    try {
      await api.delete(`/api/preise/${row.ID}`);
      $q.notify({type: 'positive', message: 'Gelöscht'});
      void loadPreise();
    } catch {
      $q.notify({type: 'negative', message: 'Fehler'});
    }
  });
}

watch(tab, (newVal) => {
  void router.push({params: {tab: newVal}});
  if (newVal === 'feldnamen') void loadFields();
  if (newVal === 'futtersorten') void loadFuttersorten();
});

onMounted(() => {
  void loadMwst();
  void loadTabellenkopfLookups();
  void loadPreise();
  void loadHaltungstypen();
  void loadFuttersorten();
  if (tab.value === 'feldnamen') void loadFields();
  if (tab.value === 'futtersorten') void loadFuttersorten();
});
</script>

<style scoped>
.rounded-borders {
  border-radius: 8px;
}
</style>

<template>
  <div>
    <q-tabs
      v-model="innerTab"
      dense
      class="text-grey q-mb-md"
      active-color="primary"
      indicator-color="primary"
      align="left"
      narrow-indicator
    >
      <q-tab name="stamm" :label="t('auto.herden_dashboard')" icon="dashboard" />
      <q-tab name="parameter" :label="t('auto.parameter')" icon="settings"/>
      <q-tab name="uebersicht" :label="t('auto.details_statistik')" icon="list" />
      <q-tab name="grafik" :label="t('auto.grafik')" icon="pie_chart" />
    </q-tabs>

    <q-separator class="q-mb-md" />

    <q-tab-panels v-model="innerTab" animated class="bg-transparent">
      <!-- HERDEN CARD VIEW -->
      <q-tab-panel name="stamm" class="q-pa-none">
        <q-table
          :rows="filteredRows"
          :columns="stammColumns"
          row-key="ID"
          grid
          hide-header
          :loading="loading"
          :filter="filter"
          :pagination="{ rowsPerPage: 0 }"
          hide-bottom
        >
          <template v-slot:top>
            <div class="full-width row items-center justify-between q-px-sm">
              <div class="row items-center q-gutter-md">
                <div class="text-h6 text-weight-bold text-grey-8">Übersicht</div>
                <q-checkbox v-model="nurAktive" :label="t('grid.onlyActive')" color="positive" class="q-ml-sm" />
              </div>
              <div class="row q-gutter-sm items-center">
                <q-input
                  v-model="filter"
                  :label="t('grid.searchHerds')"
                  dense
                  filled
                  rounded
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                  :dark="$q.dark.isActive"
                  style="width: 250px"
                  clearable
                >
                  <template v-slot:append>
                    <q-icon name="search" />
                  </template>
                </q-input>
                <q-btn color="primary" icon="add" :label="t('grid.newHerd')" rounded unelevated @click="openCreate" />
              </div>
            </div>
          </template>

          <template v-slot:item="props">
            <div class="q-pa-md col-xs-12 col-sm-6 col-md-4">
              <q-card
                :class="$q.dark.isActive ? 'bg-dark-page text-white' : 'bg-grey-2 text-black'"
                class="shadow-2 overflow-hidden cursor-pointer"
                :style="{ borderRadius: '16px', border: $q.dark.isActive ? '1px solid #424242' : '1px solid #e0e0e0' }"
                flat
                @click="editHerde(props.row)"
              >
                <q-card-section class="row items-center justify-between bg-primary text-white q-py-sm q-px-md">
                  <div class="column">
                    <div class="text-h6 text-weight-bolder" style="line-height: 1.1;">{{ (props.row.bezeichnung || props.row.BEZEICHNUNG) || '-' }}</div>
                    <div class="row items-center q-mt-xs">
                      <div class="text-caption text-uppercase text-grey-2" style="font-size: 10px; letter-spacing: 1px;">Herde {{ (props.row.herdennummer || props.row.HERDENNUMMER) || '-' }}</div>
                      <q-badge v-if="props.row.aktiv === 1 || props.row.AKTIV === 1" color="positive" class="q-ml-sm" :label="t('auto.aktiv')" rounded size="xs" />
                    </div>
                  </div>
                  <div class="row items-center q-gutter-x-sm">
                    <q-btn dense round icon="edit" color="white" flat @click.stop="editHerde(props.row)" size="sm" />
                    <q-btn dense round icon="delete" color="white" flat @click.stop="onDelete(props.row)" size="sm" />
                  </div>
                </q-card-section>

                <q-card-section class="q-px-none q-pt-sm q-pb-sm">
                  <div class="row q-col-gutter-sm">
                    <div class="col-8 q-pl-md">
                      <div class="column q-gutter-y-xs">
                        <div class="column">
                          <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.rasse_bezeichnung') }}</div>
                          <div class="text-weight-medium text-primary">{{ getRasseName(props.row.ID_RASSE) }}</div>
                        </div>
                        <div class="column">
                          <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.zuechter_haendler') }}</div>
                          <div>{{ getPersonName(props.row.ID_ZUECHTER) }}</div>
                        </div>
                        <div class="row q-col-gutter-x-sm">
                          <div class="col-6">
                            <div class="column">
                              <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.einstallung') }}</div>
                              <div>{{ props.row.EINSTALLDATUM || '-' }}</div>
                            </div>
                          </div>
                          <div class="col-6">
                            <div class="column">
                              <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.legedatum') }}</div>
                              <div>{{ props.row.LEGEDATUM || '-' }}</div>
                            </div>
                          </div>
                        </div>
                        <div class="row q-col-gutter-x-sm">
                          <div class="col-6">
                            <div class="column">
                              <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.anfangsbestand') }}</div>
                              <div class="text-weight-bold">{{ formatNumber(props.row.anfangsbestand || props.row.ANFANGSBESTAND || 0) }}</div>
                            </div>
                          </div>
                          <div class="col-6">
                            <div class="column">
                              <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">Stall</div>
                              <div class="text-weight-medium">{{ props.row.STALL_BEZEICHNUNG || props.row.stall_bezeichnung || props.row.STALLNUMMER || props.row.stall_nummer || '-' }}</div>
                            </div>
                          </div>
                        </div>
                        <div class="column">
                          <div class="text-caption text-weight-bold text-uppercase text-grey-7" style="font-size: 10px; line-height: 1;">{{ t('auto.einstandskosten') }}</div>
                          <div>{{ formatCurrency(props.row.EINSTALLKOSTEN || 0) }}</div>
                        </div>
                      </div>
                    </div>

                    <div
                      class="col-4 column items-start justify-center q-pa-xs rounded-borders"
                      :class="$q.dark.isActive ? 'bg-transparent' : 'bg-white'"
                      :style="{ border: $q.dark.isActive ? '1px solid #424242' : '1px solid #eee' }"
                    >
                      <div class="text-caption text-weight-bold text-uppercase q-mb-xs" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-6'" style="font-size: 9px;">{{ t('auto.mix_s_xl') }}</div>
                      <apexchart
                        type="pie"
                        height="90"
                        width="90"
                        :options="smallApexOptions"
                        :series="[
                          props.row.EGGSTATS?.SUM_SMALL || 0,
                          props.row.EGGSTATS?.SUM_MEDIUM || 0,
                          props.row.EGGSTATS?.SUM_LARGE || 0,
                          props.row.EGGSTATS?.SUM_XL || 0,
                          props.row.EGGSTATS?.SUM_VERLUSTE || 0
                        ]"
                      />
                      <div class="column q-mt-xs q-gutter-y-xs q-pl-xs">
                        <div class="row items-center no-wrap"><div style="width: 8px; height: 8px; border-radius: 50%; background: #3f51b5; margin-right: 4px;"></div><div class="text-grey-7" style="font-size: 9px;">{{ t('auto.s_small') }}</div></div>
                        <div class="row items-center no-wrap"><div style="width: 8px; height: 8px; border-radius: 50%; background: #2196f3; margin-right: 4px;"></div><div class="text-grey-7" style="font-size: 9px;">{{ t('auto.m_medium') }}</div></div>
                        <div class="row items-center no-wrap"><div style="width: 8px; height: 8px; border-radius: 50%; background: #4caf50; margin-right: 4px;"></div><div class="text-grey-7" style="font-size: 9px;">{{ t('auto.l_large') }}</div></div>
                        <div class="row items-center no-wrap"><div style="width: 8px; height: 8px; border-radius: 50%; background: #fbc02d; margin-right: 4px;"></div><div class="text-grey-7" style="font-size: 9px;">{{ t('auto.xl_extra_large') }}</div></div>
                        <div class="row items-center no-wrap"><div style="width: 8px; height: 8px; border-radius: 50%; background: #f44336; margin-right: 4px;"></div><div class="text-grey-7" style="font-size: 9px;">{{ t('auto.v_verluste') }}</div></div>
                      </div>
                    </div>

                    <div class="col-12 q-px-md q-pb-sm">
                      <div
                        class="q-pa-sm rounded-borders shadow-1"
                        :class="$q.dark.isActive ? 'bg-grey-8' : 'bg-white'"
                        style="border-left: 4px solid #fbc02d;"
                      >
                        <div class="row items-center justify-between">
                          <div class="text-caption text-weight-bolder text-uppercase" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'" style="font-size: 10px;">{{ t('auto.summe_klasse_a') }}</div>
                          <div class="text-weight-bolder text-warning" style="font-size: 20px;">{{ formatNumber(props.row.EGGSTATS?.SUM_KLASSE_A) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </q-card-section>
              </q-card>
            </div>
          </template>
        </q-table>
      </q-tab-panel>

      <!-- PARAMETER VIEW -->
      <q-tab-panel name="parameter" class="q-pa-none">
        <q-card flat bordered :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'" style="border-radius: 16px;">
          <q-card-section :class="$q.dark.isActive ? '' : 'bg-white'" class="q-ma-md rounded-borders shadow-1">
            <div class="row q-mb-md items-center q-gutter-md">
              <q-checkbox v-model="firmaAktiv" :label="t('auto.globale_herdenparameter')" color="primary" size="lg"/>
              <q-select v-if="!firmaAktiv" v-model="selectedHerdeId" use-input input-debounce="0"
                        :options="filteredHerdeOptions" @filter="filterHerde" option-value="ID" option-label="label"
                        emit-value map-options :label="t('auto.herde_suchen_auswaehlen')"
                        :bg-color="$q.dark.isActive ? 'grey-10' : 'grey-2'" style="min-width: 400px" filled stack-label
                        clearable hide-selected fill-input>
                <template v-slot:no-option>
                  <q-item>
                    <q-item-section class="text-grey">{{ t('auto.keine_herde_gefunden') }}</q-item-section>
                  </q-item>
                </template>
                <template v-slot:option="scope">
                  <q-item v-bind="scope.itemProps">
                    <q-item-section avatar>
                      <q-icon name="pets" color="primary"/>
                    </q-item-section>
                    <q-item-section>
                      <q-item-label class="text-weight-bold">Herde {{ scope.opt.HERDENNUMMER }}</q-item-label>
                      <q-item-label caption>{{ scope.opt.BEZEICHNUNG }}</q-item-label>
                    </q-item-section>
                  </q-item>
                </template>
              </q-select>
              <q-select v-model="copySourceId" :options="copyHerdeOptions" option-value="ID" option-label="label"
                        emit-value map-options :label="t('auto.herde_mit_parametersatz_waehlen')" filled stack-label dense
                        style="min-width: 300px" :bg-color="$q.dark.isActive ? 'grey-10' : 'grey-2'"
                        @update:model-value="onCopySourceSelected" clearable>
                <template v-slot:prepend>
                  <q-icon name="content_copy" color="secondary"/>
                </template>
              </q-select>
              <div v-if="!firmaAktiv && selectedHerde"
                   class="row items-center q-gutter-x-md bg-white q-px-md q-py-sm rounded-borders shadow-2 border-primary"
                   style="height: 56px; border-left: 5px solid #1976D2;">
                <div class="column justify-center">
                  <div class="text-caption text-uppercase text-grey-7" style="font-size: 10px; line-height: 1">{{ t('auto.aktuell_aktiv') }}
                  </div>
                  <div class="text-subtitle1 text-weight-bolder text-primary" style="line-height: 1.2">
                    {{ selectedHerde.HERDENNUMMER }} - {{ selectedHerde.BEZEICHNUNG }}
                  </div>
                </div>
              </div>
            </div>
            <q-separator class="q-mb-lg"/>
            <q-form v-if="paramFormLoaded" @submit="onSubmitParam" class="q-gutter-md">
              <div class="row q-col-gutter-lg">
                <div class="col-12">
                  <div class="row q-col-gutter-md">
                    <div class="col-12 col-sm-3">
                      <q-input v-model.number="paramForm.MASSVOLLEI" type="number" :label="t('auto.mass_vollei')" dense filled
                               stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-2">
                      <q-input v-model.number="paramForm.ANZAHLKONTROLLW" type="number" :label="t('auto.anzahl_kontrolle')" dense
                               filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-2">
                      <q-input v-model.number="paramForm.LAUFZEITWOCHEN" type="number" :label="t('auto.laufzeit_w')" dense filled
                               stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-2">
                      <q-input v-model.number="paramForm.SCHLACHTERLOESHENNE" type="number" step="0.01"
                               :label="t('auto.schlachterloes')" dense filled stack-label
                               :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-3">
                      <q-input v-model.number="paramForm.PRODUKTIONSDAUER" type="number" :label="t('auto.prod_dauer')" dense
                               filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-2">
                      <q-input v-model.number="paramForm.LEGEBEGINN_LW" type="number" :label="t('auto.legebeg_lw')" dense filled
                               stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-1">
                      <q-checkbox v-model="paramForm.BIO" :label="t('auto.bio')" dense color="primary"
                                  class="full-height items-center q-pt-sm"/>
                    </div>
                    <div class="col-12 col-sm-3">
                      <q-input v-model.number="paramForm.BIOAUFSCHLAG" type="number" step="0.01"
                               :label="t('auto.bio_aufschlag')" dense filled stack-label
                               :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" prefix="€"/>
                    </div>
                    <div class="col-12 col-sm-3">
                      <q-select v-model="paramForm.HALTUNGSTYP" :options="haltungstypOptions" label="Haltungstyp" filled
                                stack-label dense emit-value map-options
                                :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                                @update:model-value="val => { if (val === '0') paramForm.BIO = true }"/>
                    </div>
                    <div class="col-12 col-sm-3">
                      <q-input v-model.number="paramForm.VERPACKUNGKG" type="number" step="any" label="Verpackung (kg)"
                               dense filled stack-label :bg-color="$q.dark.isActive ? 'grey-8' : 'white'"/>
                    </div>
                    <div class="col-12 col-sm-3">
                      <q-input v-model.number="paramForm.MAXTAGEVERMITTELN" type="number" :label="t('auto.max_tage_vermittlung')"
                               dense filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-6">
                      <q-select v-model="paramForm.ID_TABELLEALTER" :options="alterTabellenOptions" option-value="ID"
                                option-label="BEZEICHNUNG" emit-value map-options :label="t('auto.referenz_alterstabelle')" filled
                                stack-label dense :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                    <div class="col-12 col-sm-6">
                      <q-select v-model="paramForm.ID_TABELLEGEWICHT" :options="gewichtTabellenOptions"
                                option-value="ID" option-label="BEZEICHNUNG" emit-value map-options
                                :label="t('auto.referenz_gewichtstabelle')" filled stack-label dense
                                :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"/>
                    </div>
                  </div>
                </div>
                <div class="col-12">
                  <q-separator class="q-my-md"/>
                  <div class="text-subtitle2 q-mb-sm text-primary text-weight-bold">Chargen-Einstellungen</div>
                  <div class="row q-col-gutter-md q-mb-md">
                    <div class="col-12 col-sm-4">
                      <q-input v-model="paramForm.CHARGEPREFIXFIRMA" :label="t('auto.praefix_fuer_chargennummer')" filled stack-label
                               dense :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" disable/>
                    </div>
                    <div class="col-12 col-sm-8 row q-col-gutter-sm items-center">
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGEJUMBOS" :label="t('auto.jumbos')" dense color="primary" disable/>
                      </div>
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGEXL" label="XL" dense color="primary" disable/>
                      </div>
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGELARGE" :label="t('auto.large')" dense color="primary" disable/>
                      </div>
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGEMEDIUM" :label="t('auto.medium')" dense color="primary" disable/>
                      </div>
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGESMALL" :label="t('auto.small')" dense color="primary" disable/>
                      </div>
                      <div class="col-4 col-sm-2">
                        <q-checkbox v-model="paramForm.CHARGEVOLLEI" label="Vollei" dense color="primary" disable/>
                      </div>
                    </div>
                    <div class="col-12 row q-col-gutter-sm items-center">
                      <div class="col-12 col-sm-4">
                        <q-checkbox v-model="paramForm.CHARGEPREFIXHERDENNUMMER" :label="t('auto.herdennummer_einbeziehen')" dense
                                    color="primary" disable/>
                      </div>
                      <div class="col-12 col-sm-4">
                        <q-checkbox v-model="paramForm.CHARGEDATUM" :label="t('auto.datum_einbeziehen')" dense color="primary"
                                    disable/>
                      </div>
                      <div class="col-12 col-sm-4">
                        <q-checkbox v-model="paramForm.CHARGELAGERNUMMER" :label="t('auto.lagernummer_einbeziehen')" dense
                                    color="primary" disable/>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="col-12">
                  <q-separator class="q-my-md"/>
                  <div class="text-subtitle2 q-mb-sm text-primary text-weight-bold">{{ t('auto.optionen') }}</div>
                  <div class="row q-col-gutter-sm">
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.JUMBOS" :label="t('auto.jumbos_erfassen')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.KLASSENERFASSEN" :label="t('auto.gewichtsklassen')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.KLASSEAERFASSEN" :label="t('auto.klassea_erfassen')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.KLASSEAERRECHNEN" :label="t('auto.klassea_errechnen')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.KLASSEAVERMITTELN" :label="t('auto.klassea_vermitteln')" dense
                                  color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.ERFASSESCHMUTZEI" :label="t('auto.schmutzeier')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.ERFASSEKNICKEI" :label="t('auto.knickeier')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.ERFASSEBRUCHEI" :label="t('auto.brucheier')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.ERFASSEVOLLEI" :label="t('auto.vollei_stueck')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.ERFASSEVOLLEIKG" label="Vollei (kg)" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.AUFTEILUNGGEWICHT" :label="t('auto.aufteilung_gewicht')" dense
                                  color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.AUFTEILUNGALTER" :label="t('auto.aufteilung_alter')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.KONTROLLWIEGUNG" :label="t('auto.kontrollwiegung')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.VERLUSTEBEIBUCHUNG"
                                  :label="t('auto.verluste_direkt_bei_leistungsbuchung_erf')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.LAGERBUCHUNGBEIBUCHUNG"
                                  :label="t('auto.lagerbuchungen_automatisch_bei_leistung_')" dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.BEIVERMITTELNDATUMAKTUELL" :label="t('auto.aktuelles_datum_vorschlagen')"
                                  dense color="primary"/>
                    </div>
                    <div class="col-6 col-md-4">
                      <q-checkbox v-model="paramForm.PSEUDOLAGER" :label="t('auto.pseudolager_erlauben')" dense color="primary"/>
                    </div>
                  </div>
                </div>
              </div>
              <div class="row justify-end q-mt-lg">
                <q-btn :label="t('auto.parameter_speichern')" type="submit" color="primary" rounded unelevated padding="xs xl"/>
              </div>
            </q-form>
            <div v-else class="text-center q-pa-xl">
              <q-icon name="info" size="xl" color="grey-5"/>
              <div class="text-subtitle1 text-grey-7 q-mt-md">{{ t('auto.bitte_herde_auswaehlen_um_parameter_zu_l') }}</div>
            </div>
          </q-card-section>
        </q-card>
      </q-tab-panel>

      <!-- DETAILS VIEW -->
      <q-tab-panel name="uebersicht" class="q-pa-none">
        <q-table
          :rows="filteredRows"
          :columns="columns"
          row-key="ID"
          separator="cell"
          v-model:expanded="expandedRows"
          :loading="loading"
          :filter="filter"
          :pagination="{ rowsPerPage: 15 }"
          class="huhnlite-grid-standard rounded-borders shadow-2 resizable-table"
          :class="$q.dark.isActive ? 'text-white' : 'text-black'"
          :card-class="$q.dark.isActive ? 'bg-dark-page' : 'bg-grey-2'"
          :dark="$q.dark.isActive"
          @row-dblclick="(evt, row) => editHerde(row)"
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

          <template v-slot:top-right>
              <div class="row q-gutter-sm items-center">
                <q-checkbox v-model="nurAktive" :label="t('auto.nur_aktive_herden_anzeigen')" class="q-mr-md" color="primary" :dark="$q.dark.isActive" />
                <q-input v-model="filter" :label="t('grid.searchHerds')" dense filled rounded stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" style="width: 250px" clearable>
                  <template v-slot:append><q-icon name="search" /></template>
                </q-input>
                <q-btn color="secondary" icon="unfold_more" :label="t('auto.alle_aus_einklappen')" @click="toggleAllSubgrids" rounded unelevated outline />
              </div>
            </template>

            <template v-slot:body="props">
              <q-tr :props="props" :dark="$q.dark.isActive" :class="$q.dark.isActive ? 'bg-dark text-white' : 'bg-grey-2 text-black'">
                <q-td auto-width>
                  <q-btn size="sm" color="primary" round dense @click="props.expand = !props.expand" :icon="props.expand ? 'remove' : 'add'" unelevated />
                </q-td>
                <q-td key="BEZEICHNUNG" :props="props" class="text-weight-bold">{{ props.row.BEZEICHNUNG ? props.row.BEZEICHNUNG : (props.row.BEZEICHNUNG || '-') }}</q-td>
                <q-td key="HERDENNUMMER" :props="props">{{ props.row.HERDENNUMMER || '-' }}</q-td>
                <q-td key="RASSE" :props="props">{{ getRasseName(props.row.ID_RASSE) }}</q-td>
                <q-td key="LETZTES_DATUM" :props="props">{{ props.row.WEEKLYSTATS && props.row.WEEKLYSTATS.length > 0 ? props.row.WEEKLYSTATS[0].LETZTES_DATUM : '-' }}</q-td>
                <q-td key="SUM_KLASSE_A" :props="props" class="text-weight-bold">{{ formatNumber(props.row.EGGSTATS?.SUM_KLASSE_A) }}</q-td>
                <q-td key="SUM_SMALL" :props="props">{{ formatNumber(props.row.EGGSTATS?.SUM_SMALL) }}</q-td>
                <q-td key="SUM_MEDIUM" :props="props">{{ formatNumber(props.row.EGGSTATS?.SUM_MEDIUM) }}</q-td>
                <q-td key="SUM_LARGE" :props="props">{{ formatNumber(props.row.EGGSTATS?.SUM_LARGE) }}</q-td>
                <q-td key="SUM_XL" :props="props">{{ formatNumber(props.row.EGGSTATS?.SUM_XL) }}</q-td>
                <q-td key="SUM_VERLUSTE" :props="props" class="text-negative text-weight-bold">{{ formatNumber(props.row.EGGSTATS?.SUM_VERLUSTE) }}</q-td>
              </q-tr>
            <q-tr v-show="props.expand" :props="props" :dark="$q.dark.isActive" :class="$q.dark.isActive ? 'bg-dark-page text-white' : 'bg-white text-black'">
              <q-td colspan="100%" class="q-pa-md">
                <div class="row no-wrap items-center q-pb-sm"><div class="text-subtitle2 text-weight-bold">Wochenstatistik für Herde {{ props.row.HERDENNUMMER || '-' }}</div><q-space /><q-icon name="trending_up" size="sm" color="grey" /></div>
                <q-table dense separator="cell" :rows="props.row.WEEKLYSTATS || []" :columns="subColumns" row-key="LEBENSWOCHE" hide-bottom :pagination="{ rowsPerPage: 10 }" class="sticky-subgrid-table bg-transparent shadow-0" :dark="$q.dark.isActive">
                  <template v-slot:body-cell-SUM_KLASSE_A="subProps"><q-td :props="subProps" class="text-weight-bold">{{ formatNumber(subProps.row.SUM_KLASSE_A) }}</q-td></template>
                  <template v-slot:body-cell-SUM_SMALL="subProps"><q-td :props="subProps">{{ formatNumber(subProps.row.SUM_SMALL) }}</q-td></template>
                  <template v-slot:body-cell-SUM_MEDIUM="subProps"><q-td :props="subProps">{{ formatNumber(subProps.row.SUM_MEDIUM) }}</q-td></template>
                  <template v-slot:body-cell-SUM_LARGE="subProps"><q-td :props="subProps">{{ formatNumber(subProps.row.SUM_LARGE) }}</q-td></template>
                  <template v-slot:body-cell-SUM_XL="subProps"><q-td :props="subProps">{{ formatNumber(subProps.row.SUM_XL) }}</q-td></template>
                  <template v-slot:body-cell-SUM_VERLUSTE="subProps"><q-td :props="subProps" class="text-negative text-weight-bold">{{ formatNumber(subProps.row.SUM_VERLUSTE) }}</q-td></template>
                </q-table>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </q-tab-panel>

      <!-- GRAFIK VIEW -->
      <q-tab-panel name="grafik" class="q-pa-md">
        <div class="row q-col-gutter-sm q-mb-lg items-end">
          <div class="col-12 col-sm-4">
             <q-select
                v-model="chartHerdeId"
                :options="chartHerdeOptions"
                option-value="ID"
                option-label="label"
                emit-value
                map-options
                :label="t('auto.herde_auswaehlen')"
                filled
                stack-label
                @update:model-value="onChartHerdeChange"
             />
          </div>
          <div class="col-12 col-sm-2">
            <q-select
              v-model="filterYear"
              :options="yearOptions"
              :label="t('auto.jahr')"
              filled
              stack-label
              clearable
              @update:model-value="loadFilteredStats"
            />
          </div>
          <div class="col-12 col-sm-2">
            <q-select
              v-model="filterQuarter"
              :options="[
                {label: 'Alle Quartale', value: 0},
                {label: 'Q1', value: 1},
                {label: 'Q2', value: 2},
                {label: 'Q3', value: 3},
                {label: 'Q4', value: 4}
              ]"
              emit-value
              map-options
              :label="t('auto.quartal')"
              filled
              stack-label
              :disable="!filterYear || filterMonth !== 0"
              @update:model-value="loadFilteredStats"
            />
          </div>
          <div class="col-12 col-sm-2">
            <q-select
              v-model="filterMonth"
              :options="monthOptions"
              emit-value
              map-options
              :label="t('auto.monat')"
              filled
              stack-label
              :disable="!filterYear || filterQuarter !== 0"
              @update:model-value="loadFilteredStats"
            />
          </div>
          <div class="col-12 col-sm-2">
            <q-btn :label="t('auto.reset')" color="grey" flat @click="resetChartFilters" class="full-width" />
          </div>
        </div>

        <div v-if="filteredStats" class="row justify-center q-col-gutter-lg">
          <div class="col-12 col-md-6">
            <q-card flat bordered class="q-pa-lg shadow-2 full-height" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'" style="border-radius: 20px;">
               <div class="text-h6 text-center text-weight-bold q-mb-lg" :class="$q.dark.isActive ? 'text-white' : 'text-primary'">
                 {{ chartHerdeId === -1 ? 'Produktion (Alle)' : 'Produktion (Herde ' + chartHerdeId + ')' }}
               </div>

               <div style="min-height: 300px;">
                 <apexchart
                  type="pie"
                  height="300"
                  :options="largeApexOptions"
                  :series="[
                    filteredStats.SUM_SMALL || 0,
                    filteredStats.SUM_MEDIUM || 0,
                    filteredStats.SUM_LARGE || 0,
                    filteredStats.SUM_XL || 0,
                    filteredStats.SUM_VERLUSTE || 0
                  ]"
                />
               </div>
               <div class="q-mt-md text-center">
                 <div class="text-caption text-grey-7 text-uppercase">{{ t('auto.gesamt_klasse_a') }}</div>
                 <div class="text-h5 text-weight-bolder">{{ formatNumber(filteredStats.SUM_KLASSE_A) }}</div>
               </div>
            </q-card>
          </div>

          <div class="col-12 col-md-6">
            <q-card v-if="filteredStatsActive" flat bordered class="q-pa-lg shadow-2 full-height" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'" style="border-radius: 20px;">
               <div class="text-h6 text-center text-weight-bold q-mb-lg" :class="$q.dark.isActive ? 'text-white' : 'text-positive'">
                 {{ t('auto.produktion_alle_aktiven_herden') }}
               </div>

               <div style="min-height: 300px;">
                 <apexchart
                  type="pie"
                  height="300"
                  :options="largeApexOptions"
                  :series="[
                    filteredStatsActive.SUM_SMALL || 0,
                    filteredStatsActive.SUM_MEDIUM || 0,
                    filteredStatsActive.SUM_LARGE || 0,
                    filteredStatsActive.SUM_XL || 0,
                    filteredStatsActive.SUM_VERLUSTE || 0
                  ]"
                />
               </div>
               <div class="q-mt-md text-center">
                 <div class="text-caption text-grey-7 text-uppercase">{{ t('auto.gesamt_klasse_a_aktiv') }}</div>
                 <div class="text-h5 text-weight-bolder text-positive">{{ formatNumber(filteredStatsActive.SUM_KLASSE_A) }}</div>
               </div>
            </q-card>
          </div>

          <div class="col-12 q-mt-md">
             <div class="row q-col-gutter-sm justify-center">
                <div class="col-auto" v-for="(item, idx) in [
                  {l: 'S', c: '#3f51b5'}, {l: 'M', c: '#2196f3'}, {l: 'L', c: '#4caf50'}, {l: 'XL', c: '#fbc02d'}, {l: 'V', c: '#f44336'}
                ]" :key="idx">
                   <q-badge :style="{background: item.c}" class="q-pa-xs">{{ item.l }}</q-badge>
                </div>
             </div>
          </div>
        </div>
        <div v-else class="column items-center justify-center q-pa-xl" style="min-height: 400px;">
           <q-icon name="analytics" size="6rem" color="grey-3" />
           <div class="text-h6 text-grey-5 q-mt-md">{{ t('auto.waehlen_sie_herde_und_zeitraum_fuer_die_') }}</div>
        </div>
      </q-tab-panel>


    </q-tab-panels>

    <!-- Dialog Form Herde -->
    <q-dialog v-model="showHerdenMaske" persistent @show="onHerdenDialogShow">
      <q-card style="min-width: 500px; max-width: 800px; border-radius: 16px;">
        <q-card-section class="row items-center q-pb-none bg-primary text-white q-pa-md">
          <div class="text-h6 text-weight-bold">{{ isEditing ? t('grid.editHerd') : t('grid.newHerd') }}</div>
          <q-space />
          <q-btn icon="close" round dense v-close-popup @click="closeDialog" unelevated color="white" flat />
        </q-card-section>
        <q-card-section class="q-pa-lg">
          <q-form @submit="onSubmit" class="q-gutter-md">
            <div class="row q-col-gutter-md">
              <div class="col-12 col-md-4"><q-input v-model.number="form.HERDENNUMMER" type="number" :label="t('grid.internalHerdNumber')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => val !== null && val !== '' || 'Erforderlich']" /></div>
              <div class="col-12 col-md-8"><q-input v-model="form.BEZEICHNUNG" :label="t('grid.designationName')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-md-6"><q-select v-model="form.ID_RASSE" :options="rasseOptions" option-value="ID" option-label="RASSE_NAME" emit-value map-options :label="t('grid.breedRequired')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" :rules="[val => !!val || 'Erforderlich']" /></div>
              <div class="col-12 col-md-6"><q-select v-model="form.ID_ZUECHTER" :options="zuechterOptions" option-value="ID" option-label="label" emit-value map-options :label="t('grid.breederMerchant')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-6 col-md-3"><q-input v-model.number="form.ANFANGSBESTAND" type="number" :label="t('grid.stockAnimals')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-6 col-md-3"><q-input v-model="form.EINSTALLDATUM" type="date" :label="t('grid.housingDate')" stack-label filled :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-6 col-md-3"><q-input v-model="form.LEGEDATUM" type="date" :label="t('grid.layStart')" stack-label filled :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-6 col-md-3"><q-input v-model.number="form.EINSTALLKOSTEN" type="number" :label="t('grid.costPrice')" step="0.01" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-4">
                <q-select
                  v-model="form.ID_STALL"
                  :options="filteredStallOptions"
                  option-value="ID"
                  option-label="label"
                  emit-value
                  map-options
                  use-input
                  fill-input
                  hide-selected
                  @filter="filterStall"
                  :label="t('grid.stall')"
                  filled
                  stack-label
                  :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                />
              </div>
              <div class="col-12 col-sm-4"><q-select v-model="form.ID_SILO" :options="siloOptions" option-value="ID" option-label="label" emit-value map-options :label="t('grid.silo')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 col-sm-4"><q-select v-model="form.ID_EILAGER" :options="eilagerOptions" option-value="ID" option-label="label" emit-value map-options :label="t('grid.eggStorage')" filled stack-label :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'" /></div>
              <div class="col-12 flex items-center justify-end q-gutter-x-md"><q-checkbox v-model="form.ALLE_BUCHUNGEN_MIT_DATUM" :label="t('grid.bookingsMustHaveDate')" color="primary" :true-value="1" :false-value="0" /><q-checkbox v-model="form.AKTIV" :label="t('grid.herdIsActive')" color="positive" :true-value="1" :false-value="0" /></div>
            </div>
            <div class="row justify-end q-mt-lg q-gutter-x-sm"><q-btn ref="herdenCancelBtn" :label="t('form.cancel')" color="negative" outline rounded @click="closeDialog" /><q-btn ref="herdenSaveBtn" :label="isEditing ? 'Aktualisieren' : 'Speichern'" type="submit" color="primary" rounded unelevated /></div>
          </q-form>
        </q-card-section>
      </q-card>
    </q-dialog>
  </div>
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

import type { QTableProps } from 'quasar';
import { ref, reactive, onMounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { api } from '../boot/api';
import apexchart from 'vue3-apexcharts';
import { useResizableColumns } from '../composables/useResizableColumns';

/* eslint-disable @typescript-eslint/no-explicit-any */
const $q = useQuasar();
const { columnWidths, startResize, initWidths, isResizing } = useResizableColumns('Herde');

const innerTab = ref('stamm');
const loading = ref(false);
const filter = ref('');
const nurAktive = ref(true);
const rows = ref<Record<string, any>[]>([]);
const rassen = ref<Record<string, any>[]>([]);
const personen = ref<Record<string, any>[]>([]);
const stallOptions = ref<any[]>([]);
const filteredStallOptions = ref<any[]>([]);
const siloOptions = ref<any[]>([]);
const eilagerOptions = ref<any[]>([]);
const alterTabellenOptions = ref<any[]>([]);
const gewichtTabellenOptions = ref<any[]>([]);

const filteredRows = computed(() => {
  return rows.value.filter(row => {
    if (!nurAktive.value) return true;
    const aktiv = extractIntValue(row.AKTIV !== undefined ? row.AKTIV : row.aktiv);
    return aktiv === 1;
  });
});

const stammColumns: QTableProps['columns'] = [
  { name: 'BEZEICHNUNG', label: 'Bezeichnung', field: (row: any) => extractString(row.BEZEICHNUNG || row.bezeichnung) || '', align: 'left' },
  { name: 'HERDENNUMMER', label: 'Herdennummer', field: (row: any) => extractInt(row.HERDENNUMMER || row.herdennummer)?.toString() || '', align: 'left' }
];

const columns: QTableProps['columns'] = [
  { name: 'expand', label: '', field: 'expand', align: 'left' },
  { name: 'BEZEICHNUNG', align: 'left', label: 'Bezeichnung', field: (row: any) => extractString(row.bezeichnung || row.BEZEICHNUNG) || '-', sortable: true },
  { name: 'HERDENNUMMER', align: 'center', label: 'Nr.', field: (row: any) => extractInt(row.herdennummer || row.HERDENNUMMER) || '-', sortable: true },
  { name: 'RASSE', align: 'left', label: 'Rasse', field: (row: any) => getRasseName(extractInt(row.id_rasse || row.ID_RASSE)), sortable: true },
  { name: 'LEGEDATUM', align: 'left', label: 'Legedatum', field: (row: any) => extractStringValue(row.LEGEDATUM || row.legedatum) || '-', sortable: true },
  { name: 'LETZTES_DATUM', align: 'left', label: 'Letzte Buchung', field: (row: any) => row.WEEKLYSTATS && row.WEEKLYSTATS.length > 0 ? row.WEEKLYSTATS[0].LETZTES_DATUM : '-', sortable: true },
  { name: 'SUM_KLASSE_A', align: 'right', label: 'Klasse A', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_KLASSE_A), sortable: true },
  { name: 'SUM_SMALL', align: 'right', label: 'S', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_SMALL) },
  { name: 'SUM_MEDIUM', align: 'right', label: 'M', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_MEDIUM) },
  { name: 'SUM_LARGE', align: 'right', label: 'L', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_LARGE) },
  { name: 'SUM_XL', align: 'right', label: 'XL', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_XL) },
  { name: 'SUM_VERLUSTE', align: 'right', label: 'Verluste', field: (row: any) => formatNumber(row.EGGSTATS?.SUM_VERLUSTE), sortable: true },
];

const subColumns: QTableProps['columns'] = [
  { name: 'LEBENSWOCHE', align: 'left', label: 'LW', field: 'LEBENSWOCHE' },
  { name: 'LETZTES_DATUM', align: 'left', label: 'Bis Datum', field: 'LETZTES_DATUM' },
  { name: 'SUM_KLASSE_A', align: 'right', label: 'Klasse A', field: 'SUM_KLASSE_A' },
  { name: 'SUM_SMALL', align: 'right', label: 'S', field: 'SUM_SMALL' },
  { name: 'SUM_MEDIUM', align: 'right', label: 'M', field: 'SUM_MEDIUM' },
  { name: 'SUM_LARGE', align: 'right', label: 'L', field: 'SUM_LARGE' },
  { name: 'SUM_XL', align: 'right', label: 'XL', field: 'SUM_XL' },
  { name: 'SUM_VERLUSTE', align: 'right', label: 'Verluste', field: 'SUM_VERLUSTE' }
];

const smallApexOptions = {
  chart: { sparkline: { enabled: true } },
  labels: ['Small', 'Medium', 'Large', 'Extra Large', 'Verluste'],
  colors: ['#3f51b5', '#2196f3', '#4caf50', '#fbc02d', '#f44336'],
  legend: { show: false },
  dataLabels: { enabled: false }
};

const largeApexOptions = computed(() => ({
  chart: {
    type: 'pie',
    background: 'transparent'
  },
  labels: ['Small (S)', 'Medium (M)', 'Large (L)', 'Extra Large (XL)', 'Verluste'],
  colors: ['#3f51b5', '#2196f3', '#4caf50', '#fbc02d', '#f44336'],
  theme: {
    mode: $q.dark.isActive ? 'dark' : 'light'
  },
  legend: {
    position: 'bottom',
    fontSize: '14px',
    fontFamily: 'Inter, sans-serif'
  },
  stroke: {
    show: false
  },
  dataLabels: {
    enabled: true,
    formatter: function (val: number) {
      return val.toFixed(1) + '%';
    },
    dropShadow: {
      enabled: false
    }
  },
  tooltip: {
    y: {
      formatter: function (val: number) {
        return val.toLocaleString('de-DE') + ' Eier';
      }
    }
  }
}));

const showHerdenMaske = ref(false);
const isEditing = ref(false);
const editId = ref<number | null>(null);
const expandedRows = ref<number[]>([]);

const form = reactive({
  HERDENNUMMER: 0,
  BEZEICHNUNG: '',
  ID_RASSE: 0,
  ID_ZUECHTER: 0,
  ID_EILAGER: 0,
  ANFANGSBESTAND: 0,
  EINSTALLDATUM: '0001-01-01',
  LEGEDATUM: '0001-01-01',
  EINSTALLKOSTEN: 0,
  ID_STALL: 0,
  ID_SILO: 0,
  AKTIV: 1,
  ALLE_BUCHUNGEN_MIT_DATUM: 0
});

const rasseOptions = computed(() => rassen.value.map(r => ({ ID: r.ID, RASSE_NAME: r.RASSE || '-' })));
const zuechterOptions = computed(() => personen.value.map(p => ({ ID: p.ID, label: p.NAME || p.NAME || `ID ${p.ID}` })));

const extractStringValue = (v: any) => {
  if (v === null || v === undefined) return '';
  if (typeof v === 'object' && 'String' in v) return String(v.String);
  return String(v);
};

const extractIntValue = (v: any) => {
  if (v === null || v === undefined) return 0;
  if (typeof v === 'object' && 'Int64' in v) return Number(v.Int64) || 0;
  if (typeof v === 'object' && 'Int32' in v) return Number(v.Int32) || 0;
  return Math.floor(Number(v)) || 0;
};

const extractFloatValue = (v: any) => {
  if (v === null || v === undefined) return 0;
  if (typeof v === 'object' && 'Float64' in v) return Number(v.Float64) || 0;
  if (typeof v === 'string') {
    return parseFloat(v.replace(',', '.')) || 0;
  }
  return Number(v) || 0;
};

const extractBoolValue = (v: any) => {
  if (v === null || v === undefined) return false;
  if (typeof v === 'boolean') return v;
  if (typeof v === 'object' && 'Bool' in v) return Boolean(v.Bool);
  if (typeof v === 'object' && 'Int32' in v) return v.Int32 === 1;
  return v === 1 || v === '1' || v === 'true';
};

const getRasseName = (id: number) => {
  const r = rassen.value.find(x => x.ID === id);
  if (!r) return '-';
  return r.RASSE || '-';
};
const getPersonName = (id: number) => {
  const p = personen.value.find(x => x.ID === id);
  if (!p) return '-';
  return extractStringValue(p.NAME) || extractStringValue(p.FIRMA) || p.NAME || '-';
};
const formatNumber = (val: any) => extractIntValue(val).toLocaleString('de-DE');
const formatCurrency = (val: any) => extractIntValue(val).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });

// PARAMETERS
const firmaAktiv = ref(true);
const selectedHerdeId = ref<number | null>(null);
const herdeLookupOptions = ref<any[]>([]);
const filteredHerdeOptions = ref<any[]>([]);
const paramFormLoaded = ref(false);
const copyHerdeIds = ref<number[]>([]);
const copySourceId = ref<number | null>(null);

const paramForm = reactive({
  MASSVOLLEI: '', ANZAHLKONTROLLW: 0, LAUFZEITWOCHEN: 0, SCHLACHTERLOESHENNE: 0, PRODUKTIONSDAUER: 0,
  LEGEBEGINN_LW: 18, BIO: false, BIOAUFSCHLAG: 0, HALTUNGSTYP: '3', VERPACKUNGKG: 0, MAXTAGEVERMITTELN: 0,
  ID_TABELLEALTER: 0, ID_TABELLEGEWICHT: 0,
  JUMBOS: false, KLASSENERFASSEN: false, KLASSEAERFASSEN: false, KLASSEAERRECHNEN: false, KLASSEAVERMITTELN: false,
  ERFASSESCHMUTZEI: false, ERFASSEKNICKEI: false, ERFASSEBRUCHEI: false, ERFASSEVOLLEI: false, ERFASSEVOLLEIKG: false,
  AUFTEILUNGGEWICHT: false, AUFTEILUNGALTER: false, KONTROLLWIEGUNG: false, VERLUSTEBEIBUCHUNG: false,
  LAGERBUCHUNGBEIBUCHUNG: false, BEIVERMITTELNDATUMAKTUELL: false, PSEUDOLAGER: false,
  CHARGEPREFIXFIRMA: '', CHARGEPREFIXHERDENNUMMER: false, CHARGEDATUM: false, CHARGELAGERNUMMER: false,
  CHARGEJUMBOS: false, CHARGEXL: false, CHARGELARGE: false, CHARGEMEDIUM: false, CHARGESMALL: false, CHARGEVOLLEI: false
});

// GRAFIK STATE
const chartHerdeId = ref<number | null>(null);
const filterYear = ref<string | null>(null);
const filterQuarter = ref(0);
const filterMonth = ref(0);
const filteredStats = ref<any>(null);
const filteredStatsActive = ref<any>(null);
const yearOptions = ref<string[]>([]);
const monthOptions = [
  {label: 'Alle Monate', value: 0},
  {label: 'Januar', value: 1}, {label: 'Februar', value: 2}, {label: 'März', value: 3},
  {label: 'April', value: 4}, {label: 'Mai', value: 5}, {label: 'Juni', value: 6},
  {label: 'Juli', value: 7}, {label: 'August', value: 8}, {label: 'September', value: 9},
  {label: 'Oktober', value: 10}, {label: 'November', value: 11}, {label: 'Dezember', value: 12}
];

const chartHerdeOptions = computed(() => [
  { ID: -1, label: 'Alle Herden' },
  ...herdeLookupOptions.value
]);

const haltungstypOptions = [
  { label: '0 - Ökologische Erzeugung', value: '0' },
  { label: '1 - Freilandhaltung', value: '1' },
  { label: '2 - Bodenhaltung', value: '2' },
  { label: '3 - Käfighaltung', value: '3' }
];

const selectedHerde = computed(() => herdeLookupOptions.value.find(h => h.ID === selectedHerdeId.value));
const copyHerdeOptions = computed(() => herdeLookupOptions.value.filter(h => copyHerdeIds.value.includes(h.ID) && h.ID !== selectedHerdeId.value));

async function loadCopyIds() {
  try {
    const res = await api.get('/api/firmenparameter-herden-ids');
    copyHerdeIds.value = res.data || [];
  } catch (e) {
    console.error('Fehler beim Laden der Kopier-Optionen', e);
  }
}

function onHerdenDialogShow() {
  // logic if needed
}

async function loadData() {
  loading.value = true;
  try {
    const resHerden = await api.get('/api/herden');
    rows.value = resHerden.data || [];
    console.log('Herden data loaded:', rows.value);

    herdeLookupOptions.value = rows.value.map(h => {
      const id = h.id !== undefined ? h.id : h.ID;
      const nr = h.herdennummer !== undefined ? h.herdennummer : h.HERDENNUMMER;
      const bez = h.bezeichnung !== undefined ? h.bezeichnung : h.BEZEICHNUNG;
      const aktiv = extractIntValue(h.aktiv !== undefined ? h.aktiv : h.AKTIV);
      
      return {
        ID: id,
        HERDENNUMMER: nr,
        BEZEICHNUNG: bez,
        label: `${nr || ''} - ${bez || ''}`,
        AKTIV: aktiv
      };
    });
    filteredHerdeOptions.value = herdeLookupOptions.value.filter(h => h.AKTIV === 1);
  } catch (err) {
    console.error('Failed to load herden:', err);
    $q.notify({ type: 'negative', message: 'Herden konnten nicht geladen werden' });
  }

  // Load secondary data independently
  void api.get('/api/rasse').then(res => { rassen.value = res.data || []; }).catch(() => {});
  void api.get('/api/person/zuechter').then(res => { personen.value = res.data || []; }).catch(() => {});
  void api.get('/api/stall').then(res => { 
    stallOptions.value = (res.data || []).map((s: any) => ({ ID: s.ID, label: s.BEZEICHNUNG || `Stall ${s.STALLNUMMER}` })); 
    filteredStallOptions.value = [...stallOptions.value];
  }).catch(() => {});
  void api.get('/api/silo').then(res => { siloOptions.value = (res.data || []).map((s: any) => ({ ID: s.ID, label: s.BEZEICHNUNG || `Silo ${s.SILONUMMER}` })); }).catch(() => {});
  void api.get('/api/eilager').then(res => { eilagerOptions.value = (res.data || []).map((s: any) => ({ ID: s.ID, label: s.BEZEICHNUNG || `Eilager ${s.LAGERNUMMER || s.ID}` })); }).catch(() => {});
  void api.get('/api/tabellenkopf/typ/A').then(res => {
    alterTabellenOptions.value = (res.data || []).map((t: any) => ({
      ID: extractIntValue(t.ID || t.id),
      BEZEICHNUNG: extractStringValue(t.BEZEICHNUNG || t.bezeichnung)
    }));
  }).catch(() => {});
  void api.get('/api/tabellenkopf/typ/G').then(res => {
    gewichtTabellenOptions.value = (res.data || []).map((t: any) => ({
      ID: extractIntValue(t.ID || t.id),
      BEZEICHNUNG: extractStringValue(t.BEZEICHNUNG || t.bezeichnung)
    }));
  }).catch(() => {});

  loading.value = false;
  
  if (chartHerdeId.value === null) {
    chartHerdeId.value = -1;
    void onChartHerdeChange(-1);
  }
}

function openCreate() {
  isEditing.value = false;
  editId.value = null;
  Object.assign(form, {
    HERDENNUMMER: 0, BEZEICHNUNG: '', ID_RASSE: 0, ID_ZUECHTER: 0, ID_EILAGER: 0,
    ANFANGSBESTAND: 0, EINSTALLDATUM: '0001-01-01', LEGEDATUM: '0001-01-01', EINSTALLKOSTEN: 0, ID_STALL: 0, ID_SILO: 0,
    AKTIV: 1, ALLE_BUCHUNGEN_MIT_DATUM: 0
  });
  showHerdenMaske.value = true;
}

function editHerde(row: any) {
  isEditing.value = true;
  editId.value = row.ID;
  Object.assign(form, {
    HERDENNUMMER: extractIntValue(row.HERDENNUMMER || row.herdennummer),
    BEZEICHNUNG: extractStringValue(row.BEZEICHNUNG || row.bezeichnung),
    ID_RASSE: extractIntValue(row.ID_RASSE || row.id_rasse),
    ID_ZUECHTER: extractIntValue(row.ID_ZUECHTER || row.id_zuechter),
    ID_EILAGER: extractIntValue(row.ID_EILAGER || row.id_eilager),
    ANFANGSBESTAND: extractIntValue(row.ANFANGSBESTAND || row.anfangsbestand),
    EINSTALLDATUM: extractStringValue(row.EINSTALLDATUM || row.einstalldatum) || '0001-01-01',
    LEGEDATUM: extractStringValue(row.LEGEDATUM || row.legedatum) || '0001-01-01',
    EINSTALLKOSTEN: extractIntValue(row.EINSTALLKOSTEN || row.einstallkosten),
    ID_STALL: extractIntValue(row.ID_STALL || row.id_stall),
    ID_SILO: extractIntValue(row.ID_SILO || row.id_silo),
    AKTIV: extractIntValue(row.AKTIV || row.aktiv),
    ALLE_BUCHUNGEN_MIT_DATUM: extractIntValue(row.ALLE_BUCHUNGEN_MIT_DATUM || row.alle_buchungen_mit_datum || row.ALLEBUCHUNGENMITDATUM || row.allebuchungenmitdatum)
  });
  showHerdenMaske.value = true;
}

function closeDialog() { showHerdenMaske.value = false; }

async function onSubmit() {
  try {
    // We must use the keys the backend expects (mix of lower/upper/snake_case)
    const payload = {
      herdennummer: Number(form.HERDENNUMMER) || 0,
      id_rasse: Number(form.ID_RASSE) || 0,
      ID_ZUECHTER: Number(form.ID_ZUECHTER) || 0,
      ID_EILAGER: Number(form.ID_EILAGER) || 0,
      ANFANGSBESTAND: Number(form.ANFANGSBESTAND) || 0,
      EINSTALLDATUM: form.EINSTALLDATUM || '0001-01-01',
      LEGEDATUM: form.LEGEDATUM || '0001-01-01',
      EINSTALLKOSTEN: Number(form.EINSTALLKOSTEN) || 0.0,
      ID_SILO: Number(form.ID_SILO) || 0,
      ID_STALL: Number(form.ID_STALL) || 0,
      AKTIV: Number(form.AKTIV) || 0,
      BEZEICHNUNG: form.BEZEICHNUNG,
      ALLE_BUCHUNGEN_MIT_DATUM: Number(form.ALLE_BUCHUNGEN_MIT_DATUM) || 0
    };

    if (isEditing.value && editId.value) {
      await api.put(`/api/herden/${editId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Herde aktualisiert' });
    } else {
      await api.post('/api/herden', payload);
      $q.notify({ type: 'positive', message: 'Herde erstellt' });
    }
    closeDialog();
    void loadData();
  } catch (err: any) {
    const errorMsg = err.response?.data?.error || err.message || 'Unbekannter Fehler';
    $q.notify({
      type: 'negative',
      message: 'Fehler beim Speichern',
      caption: errorMsg
    });
  }
}

function onDelete(row: any) {
  $q.dialog({ title: 'Löschen', message: 'Herde wirklich löschen?', cancel: true }).onOk(() => {
    loading.value = true;
    api.delete(`/api/herden/${row.ID}`)
      .then(() => {
        void loadData();
      })
      .catch(() => {
        $q.notify({ type: 'negative', message: 'Fehler beim Löschen' });
      })
      .finally(() => {
        loading.value = false;
      });
  });
}

function toggleAllSubgrids() {
  if (expandedRows.value.length === rows.value.length) expandedRows.value = [];
  else expandedRows.value = rows.value.map(r => r.ID);
}

// PARAMETER LOGIC
watch([firmaAktiv, selectedHerdeId], async () => {
  if (firmaAktiv.value) await loadParams('F', 0);
  else if (selectedHerdeId.value) await loadParams('H', selectedHerdeId.value);
  else paramFormLoaded.value = false;
});

async function loadParams(typ: string, id: number) {
  try {
    const res = await api.get(`/api/firmenparameter/${typ}/${id}`);
    const data = res.data?.data || res.data || {};

    Object.assign(paramForm, {
      MASSVOLLEI: extractStringValue(data.MASSVOLLEI || data.massvollei),
      ANZAHLKONTROLLW: extractIntValue(data.ANZAHLKONTROLLW || data.anzahlkontrollw),
      LAUFZEITWOCHEN: extractIntValue(data.LAUFZEITWOCHEN || data.laufzeitwochen),
      SCHLACHTERLOESHENNE: extractFloatValue(data.SCHLACHTERLOESHENNE || data.schlachterloeshenne),
      PRODUKTIONSDAUER: extractIntValue(data.PRODUKTIONSDAUER || data.produktionsdauer),
      LEGEBEGINN_LW: extractIntValue(data.LEGEBEGINN_LW || data.legebeginn_lw),
      BIO: extractBoolValue(data.BIO || data.bio),
      BIOAUFSCHLAG: extractFloatValue(data.BIOAUFSCHLAG || data.bioaufschlag),
      HALTUNGSTYP: extractStringValue(data.HALTUNGSTYP || data.haltungstyp) || '0',
      VERPACKUNGKG: extractFloatValue(data.VERPACKUNGKG || data.verpackungkg),
      MAXTAGEVERMITTELN: extractIntValue(data.MAXTAGEVERMITTELN || data.maxtagevermitteln),
      ID_TABELLEALTER: extractIntValue(data.ID_TABELLEALTER || data.id_tabellealter),
      ID_TABELLEGEWICHT: extractIntValue(data.ID_TABELLEGEWICHT || data.id_tabellegewicht),
      JUMBOS: extractBoolValue(data.JUMBOS || data.jumbos),
      KLASSENERFASSEN: extractBoolValue(data.KLASSENERFASSEN || data.klassenerfassen),
      KLASSEAERFASSEN: extractBoolValue(data.KLASSEAERFASSEN || data.klasseaerfassen),
      KLASSEAERRECHNEN: extractBoolValue(data.KLASSEAERRECHNEN || data.klasseaerrechnen),
      KLASSEAVERMITTELN: extractBoolValue(data.KLASSEAVERMITTELN || data.klasseavermitteln),
      ERFASSESCHMUTZEI: extractBoolValue(data.ERFASSESCHMUTZEI || data.erfasseschmutzei),
      ERFASSEKNICKEI: extractBoolValue(data.ERFASSEKNICKEI || data.erfasseknickei),
      ERFASSEBRUCHEI: extractBoolValue(data.ERFASSEBRUCHEI || data.erfassebruchei),
      ERFASSEVOLLEI: extractBoolValue(data.ERFASSEVOLLEI || data.erfassevollei),
      ERFASSEVOLLEIKG: extractBoolValue(data.ERFASSEVOLLEIKG || data.erfassevolleikg),
      AUFTEILUNGGEWICHT: extractBoolValue(data.AUFTEILUNGGEWICHT || data.aufteilunggewicht),
      AUFTEILUNGALTER: extractBoolValue(data.AUFTEILUNGALTER || data.aufteilungalter),
      KONTROLLWIEGUNG: extractBoolValue(data.KONTROLLWIEGUNG || data.kontrollwiegung),
      VERLUSTEBEIBUCHUNG: extractBoolValue(data.VERLUSTEBEIBUCHUNG || data.verlustebeibuchung),
      LAGERBUCHUNGBEIBUCHUNG: extractBoolValue(data.LAGERBUCHUNGBEIBUCHUNG || data.lagerbuchungbeibuchung),
      BEIVERMITTELNDATUMAKTUELL: extractBoolValue(data.BEIVERMITTELNDATUMAKTUELL || data.beivermittelndatumaktuell),
      PSEUDOLAGER: extractBoolValue(data.PSEUDOLAGER || data.pseudolager),
      CHARGEPREFIXFIRMA: extractStringValue(data.CHARGEPREFIXFIRMA || data.chargeprefixfirma),
      CHARGEPREFIXHERDENNUMMER: extractBoolValue(data.CHARGEPREFIXHERDENNUMMER || data.chargeprefixherdennummer),
      CHARGEDATUM: extractBoolValue(data.CHARGEDATUM || data.chargedatum),
      CHARGELAGERNUMMER: extractBoolValue(data.CHARGELAGERNUMMER || data.chargelagernummer),
      CHARGEJUMBOS: extractBoolValue(data.CHARGEJUMBOS || data.chargejumbos),
      CHARGEXL: extractBoolValue(data.CHARGEXL || data.chargexl),
      CHARGELARGE: extractBoolValue(data.CHARGELARGE || data.chargelarge),
      CHARGEMEDIUM: extractBoolValue(data.CHARGEMEDIUM || data.chargemedium),
      CHARGESMALL: extractBoolValue(data.CHARGESMALL || data.chargesmall),
      CHARGEVOLLEI: extractBoolValue(data.CHARGEVOLLEI || data.chargevollei)
    });
    paramFormLoaded.value = true;
  } catch (err) {
    console.error('Fehler beim Laden der Parameter:', err);
    paramFormLoaded.value = false;
  }
}

async function onSubmitParam() {
  const id = firmaAktiv.value ? -1 : selectedHerdeId.value;
  if (id === null) return;

  try {
    const payload: any = {
      ...paramForm,
      ID_HERDEN: id,
      KZ: firmaAktiv.value ? 'F' : 'H'
    };

    // Exact mapping for the backend struct (ParamInput)
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

    await api.post('/api/firmenparameter', payload);
    $q.notify({ type: 'positive', message: 'Parameter gespeichert' });
    void loadCopyIds();
  } catch (err) {
    console.error('Fehler beim Speichern:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern' });
  }
}

function filterHerde(val: string, update: any) {
  update(() => {
    const needle = val.toLowerCase();
    filteredHerdeOptions.value = herdeLookupOptions.value.filter(v => v.label.toLowerCase().includes(needle) && v.AKTIV === 1);
  });
}

function filterStall(val: string, update: any) {
  update(() => {
    const needle = val.toLowerCase();
    filteredStallOptions.value = stallOptions.value.filter(v => v.label.toLowerCase().includes(needle));
  });
}

function onCopySourceSelected(sourceId: number | null) {
  if (!sourceId) return;
  firmaAktiv.value = false;
  selectedHerdeId.value = sourceId;
  setTimeout(() => { copySourceId.value = null; }, 100);
}

// GRAFIK LOGIC
async function onChartHerdeChange(val: number | null) {
  if (!val) {
    yearOptions.value = [];
    filteredStats.value = null;
    return;
  }

  // Load available years for this herd
  try {
    const res = await api.get(`/api/herden/${val}/years`);
    // Ensure all years are strings for consistent comparison/display
    const years = (res.data || []).map((y: any) => String(y));
    yearOptions.value = years;
    
    if (yearOptions.value.length > 0 && !filterYear.value) {
      filterYear.value = yearOptions.value[0];
    }
    await loadFilteredStats();
  } catch (e) {
    console.error('Failed to load years', e);
  }
}

async function loadFilteredStats() {
  if (!chartHerdeId.value || !filterYear.value) {
    filteredStats.value = null;
    filteredStatsActive.value = null;
    return;
  }

  try {
    const params = {
      year: filterYear.value,
      quarter: filterQuarter.value,
      month: filterMonth.value,
      onlyActive: 0
    };
    const res = await api.get(`/api/herden/${chartHerdeId.value}/eggstats/filtered`, { params });
    filteredStats.value = res.data;

    // Fetch active-only stats for Comparison
    const paramsActive = { ...params, onlyActive: 1 };
    const resActive = await api.get(`/api/herden/-1/eggstats/filtered`, { params: paramsActive });
    filteredStatsActive.value = resActive.data;
  } catch (e) {
    console.error('Failed to load filtered stats', e);
    $q.notify({ type: 'negative', message: 'Grafikdaten konnten nicht geladen werden' });
  }
}

function resetChartFilters() {
  filterYear.value = yearOptions.value.length > 0 ? yearOptions.value[0] : null;
  filterQuarter.value = 0;
  filterMonth.value = 0;
  void loadFilteredStats();
}

onMounted(async () => {
  initWidths(columns);
  await loadData();
});

defineExpose({ loadData });
</script>

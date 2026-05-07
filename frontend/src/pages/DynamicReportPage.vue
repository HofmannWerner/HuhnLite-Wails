<template>
  <q-page padding>
    <div class="row items-center q-mb-md">
      <div class="text-h4 text-weight-bolder text-primary">
        Dynamische Reports
        <q-badge v-if="reportRows.length > 0" color="orange" floating>{{ reportRows.length }}</q-badge>
      </div>
      <q-space />
      <q-tabs v-model="tab" dense class="text-grey-7 bg-grey-2 rounded-borders q-pa-xs" active-color="primary" indicator-color="primary" align="left" narrow-indicator style="border: 1px solid #e0e0e0;">
        <q-tab name="anzeige" label="Anzeige" icon="insights" />
        <q-tab name="konfiguration" label="Konfiguration" icon="settings" v-if="sessionStore.can('sql_struktur_verwalten')" />
      </q-tabs>
    </div>

    <!-- HTML REPORT PREVIEW DIALOG -->
    <q-dialog v-model="showHtmlReport" full-width full-height>
      <q-card class="bg-white">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6 text-primary">Berichts-Vorschau</div>
          <q-space />
          <q-btn flat round icon="print" color="primary" @click="printHtmlReport" class="q-mr-sm">
            <q-tooltip>Drucken</q-tooltip>
          </q-btn>
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-none scroll" style="height: calc(100vh - 150px);">
          <iframe 
            :srcdoc="generatedHtml" 
            style="width: 100%; height: 100%; border: none;"
            title="Bericht-Vorschau"
          ></iframe>
        </q-card-section>

        <q-separator />

        <q-card-actions align="right" class="bg-grey-1 q-pa-md">
          <q-btn label="Vorschau schließen" icon="close" color="primary" v-close-popup unelevated rounded />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <q-tab-panels v-model="tab" animated class="bg-transparent">
      <!-- TAB 1: ANZEIGE -->
      <q-tab-panel name="anzeige" class="q-pa-none">
        <div class="row q-col-gutter-lg">
          <!-- Report Auswahl (Links/Oben) -->
          <div class="col-12 col-md-4">
            <q-card flat bordered class="shadow-2 rounded-borders overflow-hidden">
              <q-card-section class="bg-primary text-white q-pa-sm row items-center">
                <q-icon name="list" size="sm" class="q-mr-sm" />
                <div class="text-subtitle1 text-weight-bold">Report wählen</div>
              </q-card-section>
              
              <q-card-section class="q-pa-sm border-bottom">
                <div class="row items-center no-wrap">
                  <q-input 
                    v-model="filterText" 
                    dense 
                    filled 
                    placeholder="Suchen..." 
                    class="col"
                    clearable
                  >
                    <template v-slot:prepend>
                      <q-icon name="search" size="xs" />
                    </template>
                  </q-input>
                  <q-btn 
                    flat 
                    round 
                    dense 
                    :icon="allExpanded ? 'unfold_less' : 'unfold_more'" 
                    :color="allExpanded ? 'grey-7' : 'primary'"
                    @click="toggleExpandAll"
                    class="q-ml-xs"
                  >
                    <q-tooltip>{{ allExpanded ? 'Alle einklappen' : 'Alle ausklappen' }}</q-tooltip>
                  </q-btn>
                </div>
              </q-card-section>
              
              <q-card-section class="q-pa-none scroll" style="height: 600px">
                <q-tree
                  ref="treeRef"
                  :nodes="treeNodes"
                  node-key="key"
                  label-key="label"
                  selected-color="primary"
                  v-model:selected="selectedKey"
                  v-model:expanded="expandedKeys"
                  :filter="filterText"
                  no-nodes-label="Keine Berichte gefunden"
                  class="q-pa-sm"
                  @update:selected="onTreeNodeSelected"
                >
                  <template v-slot:default-header="prop">
                    <div class="row items-center no-wrap full-width">
                      <q-icon :name="prop.node.icon" :color="prop.node.iconColor" class="q-mr-sm" size="xs" />
                      <div :class="prop.node.children ? 'text-weight-bold' : ''">{{ prop.node.label }}</div>
                      
                      <!-- Tooltip für Beschaffenheit (besonders für Roots) -->
                      <q-tooltip v-if="prop.node.tooltip" anchor="top middle" self="bottom middle" :offset="[10, 10]" class="bg-indigo-9 text-white shadow-4">
                        <div class="row items-center no-wrap">
                          <q-icon name="info" size="xs" class="q-mr-xs" />
                          <div class="text-caption">{{ prop.node.tooltip }}</div>
                        </div>
                      </q-tooltip>
                    </div>
                  </template>
                </q-tree>
              </q-card-section>
            </q-card>
          </div>

          <!-- Anzeige Bereich (Rechts/Unten) -->
          <div class="col-12 col-md-8">
            <q-card flat bordered class="shadow-2 rounded-borders overflow-hidden" style="min-height: 500px">
              <q-card-section class="bg-secondary text-white q-pa-sm row items-center">
                <div class="bg-orange text-black q-pa-xs q-mr-md font-weight-bold">RECHTE SEITE AKTIV</div>
                <q-icon name="insights" size="sm" class="q-mr-sm" />
                <div class="text-subtitle1 text-weight-bold">
                  {{ currentReportLabel || 'Kein Report ausgewählt' }}
                </div>
                

                  <q-btn 
                   v-if="generatedHtml || (currentReportType === 'S' && resultData.length > 0)" 
                   icon="print" 
                   label="Drucken / PDF" 
                   flat 
                   color="white" 
                   size="sm" 
                   @click="handlePrintRequest" 
                   class="q-mr-sm"
                 />

                <q-btn 
                  v-if="resultData.length > 0 || masterData.length > 0" 
                  icon="save_alt" 
                  label="CSV Export" 
                  flat 
                  color="white" 
                  size="sm" 
                  @click="exportCSV" 
                />

                <q-space />
                
                <q-btn 
                  flat 
                  round 
                  dense 
                  icon="close" 
                  color="white" 
                  class="q-ml-md"
                  @click="selectedKey = null; currentReportLabel = ''; masterData = []; detailData = []; resultData = []"
                >
                  <q-tooltip>Vorschau schließen</q-tooltip>
                </q-btn>
              </q-card-section>

              <!-- SQL-Statement Preview -->
              <q-card-section v-if="executedSQL" class="bg-grey-2 q-pa-xs border-bottom">
                <q-expansion-item
                  icon="terminal"
                  label="Generiertes SQL-Statement (Backend)"
                  header-class="text-caption text-weight-medium text-grey-8"
                  dense
                >
                  <q-card class="bg-dark text-amber-3 q-pa-sm rounded-borders font-mono text-caption overflow-auto" style="max-height: 200px; white-space: pre-wrap;">
                    {{ executedSQL }}
                  </q-card>
                </q-expansion-item>
              </q-card-section>

              <q-card-section class="q-pa-none">
                <!-- Loading State -->
                <div v-if="loadingResult" class="flex flex-center q-pa-xl" style="height: 400px">
                  <q-spinner-cube color="primary" size="60px" />
                  <div class="full-width text-center q-mt-md text-grey-7 text-subtitle1">
                    Generiere Daten... Bitte warten.
                  </div>
                </div>

                <!-- Ergebnisanzeige -->
                <div v-if="!loadingResult && (masterData.length > 0 || resultData.length > 0)">
                  
                  <!-- Variante 1: Standard (S, L) -->
                  <div v-if="['S', 'L'].includes(currentReportType)" class="q-pa-sm">
                     <q-table
                      :title="currentReportLabel || 'Bericht'"
                      :rows="resultData"
                      :columns="resultColumns"
                      separator="cell"
                      flat
                      bordered
                      dense
                      :pagination="{ rowsPerPage: 15 }"
                      :rows-per-page-options="[5, 10, 15]"
                      style="height: 640px"
                      class="resizable-table"
                     >
                         <!-- Resizable Header Cells -->
                         <template v-slot:header-cell="props">
                           <q-th :props="props" 
                                 class="resizable-column" 
                                 :style="{ width: (masterWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
                              <div class="ellipsis">{{ props.col.label }}</div>
                              <div class="resizer" 
                                   :class="{ 'is-resizing': masterIsResizing === props.col.name }"
                                   @pointerdown.stop.prevent.capture="masterStartResize($event, props.col.name)">
                              </div>
                           </q-th>
                         </template>
                        <template v-slot:bottom-row v-if="resultSums && resultSums.length > 0">
                           <q-tr v-for="(sumRow, sIdx) in resultSums" :key="'sum-' + sIdx" :class="$q.dark.isActive ? 'bg-grey-9 text-orange-4' : 'bg-blue-grey-1 text-weight-bold shadow-1'">
                             <q-td v-for="col in resultColumns" :key="col.name" :class="col.align === 'right' ? 'text-right' : 'text-left'" :style="{ borderTop: '2px solid #ccc', fontSize: '1.1em', color: $q.dark.isActive ? '#ffb74d' : '#1a237e' }">
                               {{ formatDynamicValue(sumRow[col.name] || sumRow[col.name.toUpperCase()] || sumRow[col.name.toLowerCase()], col.name) }}
                             </q-td>
                           </q-tr>
                         </template>
                     </q-table>
                  </div>

                  <!-- Variante 2: Master/Detail (M, T) -->
                  <div v-if="['M', 'T'].includes(currentReportType)" class="q-pa-md column q-gutter-y-md">
                     <q-table
                      v-if="showMasterGrid"
                      :title="currentReportLabel || 'Master-Daten'"
                      :rows="masterData"
                      :columns="resultColumns"
                      flat
                      bordered
                      dense
                      selection="single"
                      v-model:selected="selectedMasterRows"
                      @row-click="onMasterRowClick"
                      :row-key="row => row.ID || row.id || JSON.stringify(row)"
                      :pagination="{ rowsPerPage: 15 }"
                      :rows-per-page-options="[5, 10, 15]"
                      class="resizable-table"
                     >
                        <!-- Resizable Header Cells -->
                        <template v-slot:header-cell="props">
                          <q-th :props="props" 
                                class="resizable-column" 
                                :style="{ width: (masterWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
                            <div class="ellipsis">{{ props.col.label }}</div>
                            <div class="resizer" 
                                 :class="{ 'is-resizing': masterIsResizing === props.col.name }"
                                 @pointerdown.stop.prevent.capture="masterStartResize($event, props.col.name)">
                            </div>
                          </q-th>
                        </template>
                        <template v-slot:bottom-row>
                          <q-tr v-for="(sumRow, sIdx) in resultSums" :key="'sum-m-' + sIdx" :class="$q.dark.isActive ? 'bg-grey-9 text-orange-4' : 'bg-blue-grey-1 text-weight-bold shadow-1'">
                            <q-td v-for="col in resultColumns" :key="col.name" :class="col.align === 'right' ? 'text-right' : 'text-left'" :style="{ borderTop: '2px solid #ccc', fontSize: '1.1em', color: $q.dark.isActive ? '#ffb74d' : '#1a237e' }">
                              {{ formatDynamicValue(sumRow[col.name] || sumRow[col.name.toUpperCase()] || sumRow[col.name.toLowerCase()], col.name) }}
                            </q-td>
                          </q-tr>
                        </template>
                     </q-table>
                     <q-table
                        v-if="showDetailGrid"
                        :title="`Detail-Daten (${filteredDetailData.length} Zeilen)`"
                        :rows="filteredDetailData"
                        :columns="detailColumns"
                        flat
                        bordered
                        dense
                        style="height: 500px"
                        virtual-scroll
                        :pagination="{ rowsPerPage: 0 }"
                        :rows-per-page-options="[0]"
                        class="sticky-header-table resizable-table"
                      >
                         <!-- Resizable Header Cells -->
                         <template v-slot:header-cell="props">
                           <q-th :props="props" 
                                 class="resizable-column" 
                                 :style="{ width: (detailWidths[props.col.name] || 150) + 'px', overflow: 'visible !important' }">
                               <div class="ellipsis">{{ props.col.label }}</div>
                               <div class="resizer" 
                                    :class="{ 'is-resizing': detailIsResizing === props.col.name }"
                                    @pointerdown.stop.prevent.capture="detailStartResize($event, props.col.name)">
                               </div>
                           </q-th>
                         </template>
                       </q-table>
                      
                      <div v-if="!showMasterGrid && !showDetailGrid && currentReportType !== 'T'" class="q-pa-xl text-center text-grey-6 bg-grey-2 rounded-borders border-dashed">
                        <q-icon name="grid_view" size="lg" class="q-mb-sm opacity-50" />
                        <div>Keine Gitter-Ansicht für diesen Report konfiguriert.</div>
                        <div class="text-caption">Möchten Sie das Master-Grid in der Konfiguration aktivieren?</div>
                        <q-btn label="Master-Grid jetzt anzeigen" flat color="primary" class="q-mt-md" @click="showMasterGrid = true" />
                      </div>
                   </div>
                </div>
              </q-card-section>
            </q-card>
          </div>
        </div>
      </q-tab-panel>

      <q-tab-panel name="konfiguration" class="q-pa-md">
        <div class="row items-center q-mb-md">
          <div class="text-h6 text-primary text-weight-bold">Berichts-Struktur verwalten</div>
          <q-space />
          <div class="q-gutter-x-sm">
            <q-btn flat round dense icon="unfold_more" color="primary" @click="toggleConfigGroups(true)">
              <q-tooltip>Alle Ordner ausklappen</q-tooltip>
            </q-btn>
            <q-btn flat round dense icon="unfold_less" color="grey-7" @click="toggleConfigGroups(false)">
              <q-tooltip>Alle Ordner einklappen</q-tooltip>
            </q-btn>
            <q-btn color="primary" icon="add" label="Neuer Bericht" @click="openCreate" rounded unelevated class="q-ml-md" />
          </div>
        </div>

        <!-- Neue Baumstruktur für Verwaltung -->
        <div :class="$q.dark.isActive ? 'bg-grey-9 text-white' : 'bg-white text-dark'" class="rounded-borders shadow-1 q-pa-md">
          <q-tree
            :nodes="treeNodes"
            node-key="key"
            ref="mgmtTreeRef"
            v-model:expanded="expandedKeysMgmt"
            no-nodes-label="Keine Berichte gefunden"
            class="report-tree-mgmt"
          >
            <template v-slot:default-header="prop">
              <div 
                class="row items-center full-width no-wrap q-pa-sm rounded-borders droppable-node cursor-pointer"
                :class="{ 
                  'bg-blue-1': dropTargetKey === prop.node.key && !$q.dark.isActive,
                  'bg-indigo-9': dropTargetKey === prop.node.key && $q.dark.isActive,
                  'q-my-xs': true
                }"
                draggable="true"
                @click="prop.node.data ? onEdit(prop.node.data) : null"
                @dragstart="onTreeDragStart($event, prop.node)"
                @dragover.prevent="onTreeDragOver($event, prop.node)"
                @dragleave="onTreeDragLeave($event, prop.node)"
                @drop="onTreeDrop($event, prop.node)"
              >
                <q-icon :name="prop.node.icon" :color="prop.node.iconColor || 'primary'" class="q-mr-sm" size="sm" />
                <div class="text-weight-medium" :class="$q.dark.isActive ? 'text-grey-2' : 'text-dark'">{{ prop.node.label }}</div>
                
                <q-space />
                
                <!-- Aktionen (nur für Berichte, nicht für Root-Kategorien oder leere Ordner ohne Data) -->
                <div v-if="prop.node.data" class="row no-wrap q-gutter-x-xs q-ml-md">
                  <q-btn 
                    flat round dense 
                    icon="edit" 
                    color="primary" 
                    size="sm" 
                    @click.stop="onEdit(prop.node.data)"
                  >
                    <q-tooltip>Bearbeiten</q-tooltip>
                  </q-btn>
                  <q-btn 
                    flat round dense 
                    icon="content_copy" 
                    color="secondary" 
                    size="sm" 
                    @click.stop="onCopy(prop.node.data)"
                  >
                    <q-tooltip>Kopieren</q-tooltip>
                  </q-btn>
                  <q-btn 
                    flat round dense 
                    icon="delete" 
                    color="negative" 
                    size="sm" 
                    :disable="getBool(getP(prop.node.data, 'SYSTEM_KZ')) && !isAdmin"
                    @click.stop="onDelete(prop.node.data)"
                  >
                    <q-tooltip>{{ (getBool(getP(prop.node.data, 'SYSTEM_KZ')) && !isAdmin) ? 'System-Eintrag (gesperrt)' : 'Löschen' }}</q-tooltip>
                  </q-btn>
                </div>
              </div>
            </template>
          </q-tree>
        </div>
      </q-tab-panel>
    </q-tab-panels>

    <!-- Filter Dialog (für Anzeige) -->
    <!-- Backtick Definitions Dialog -->
    <q-dialog v-model="showDefDialog" persistent backdrop-filter="blur(4px)">
      <q-card style="min-width: 450px;" :class="$q.dark.isActive ? 'bg-grey-10 text-white' : ''">
        <q-card-section class="row items-center q-pb-none">
          <div class="text-h6">Parameter definieren</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section>
          <div class="text-caption q-mb-md">Bitte legen Sie Anzeige-Labels und Typen für die im SQL gefundenen Begriffe fest:</div>
          
          <q-list bordered separator rounded-borders>
            <q-item v-for="(d, index) in currentDefinitions" :key="d.term" class="q-py-md">
              <q-item-section>
                <div class="row q-col-gutter-sm items-center">
                  <div class="col-12 text-weight-bold font-mono text-primary">`{{ d.term }}`</div>
                  <q-input v-model="d.label" label="Anzeige-Label" dense outlined class="col-7" :autofocus="index === 0" />
                  <q-select 
                    v-model="d.type" 
                    label="Datentyp" 
                    :options="['TEXT', 'NUMBER', 'DATE', 'BOOLEAN']" 
                    dense outlined class="col-5" 
                  />
                </div>
              </q-item-section>
            </q-item>
          </q-list>
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md">
          <q-btn flat label="Abbrechen" color="grey" v-close-popup />
          <q-btn label="Speichern & Weiter" color="primary" @click="saveDefinitions" unelevated rounded />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <q-dialog v-model="showFilterDialog" persistent>
      <q-card style="min-width: 400px; border-radius: 12px;">
        <q-card-section class="bg-purple text-white row items-center">
          <q-icon name="filter_alt" size="sm" class="q-mr-sm" />
          <div class="text-h6">Filter einstellen</div>
          <q-space />
          <q-btn label="Format def." color="white" flat icon="settings" dense class="q-mr-sm" @click="showQuickParamEdit = true">
            <q-tooltip>Parameter & Format definieren</q-tooltip>
          </q-btn>
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-md">
          <div class="text-caption q-mb-md">Bitte Parameter für den Report "{{ currentReportLabel }}" angeben:</div>
          <div class="q-gutter-md">
            <div v-for="(p, index) in filterParams" :key="p ? p.label : index">
              <template v-if="p">
                <!-- STAMMDATEN COMBOBOXEN HABEN VORRANG -->
                <div v-if="isStammParam(p)" class="q-mb-md">
                  <div class="text-caption text-weight-bold q-mb-xs" style="white-space: normal; display: block; width: 100%; color: var(--q-primary); line-height: 1.2;">{{ p.label }} wählen:</div>
                  <q-select
                    v-model="filterValues[p.label]"
                    :options="getStammFiltered(p)"
                    filled
                    dense
                    emit-value
                    map-options
                    use-input
                    input-debounce="0"
                    @filter="(val, update) => filterStamm(val, update, p)"
                    :autofocus="index === 0"
                  >
                    <template v-slot:no-option>
                      <q-item>
                        <q-item-section class="text-grey">Keine Einträge gefunden</q-item-section>
                      </q-item>
                    </template>
                  </q-select>
                </div>

                <div v-else-if="p.type === 'DATE'" class="q-mb-md">
                  <div class="text-caption text-weight-bold q-mb-xs" style="white-space: normal; display: block; width: 100%; color: var(--q-primary); line-height: 1.2;">{{ p.label }}:</div>
                  <q-input 
                    v-model="filterValues[p.label]"
                    type="date"
                    filled
                    dense
                    :autofocus="index === 0"
                  />
                </div>
                <div v-else-if="p.type === 'NUMBER'" class="q-mb-md">
                  <div class="text-caption text-weight-bold q-mb-xs" style="white-space: normal; display: block; width: 100%; color: var(--q-primary); line-height: 1.2;">{{ p.label }}:</div>
                  <q-input 
                    v-model.number="filterValues[p.label]"
                    type="number"
                    filled
                    dense
                    :autofocus="index === 0"
                  />
                </div>
                <q-item v-else-if="isBool(p)" tag="label" v-ripple :class="['rounded-borders', $q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2']">
                  <q-item-section avatar>
                    <q-checkbox v-model="filterValues[p.label]" color="primary" :autofocus="index === 0" />
                  </q-item-section>
                  <q-item-section>
                    <q-item-label class="text-weight-medium" style="white-space: normal; line-height: 1.2;">{{ p.label }}</q-item-label>
                  </q-item-section>
                </q-item>
                
                <div v-else-if="p.type === 'CHOICE'" class="q-pa-sm rounded-borders" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-grey-2'">
                  <div class="text-caption q-mb-xs text-weight-medium" style="white-space: normal;">{{ p.label }} wählen:</div>
                  <q-option-group
                    v-model="filterValues[p.label]"
                    :options="getChoiceOptions(p)"
                    color="primary"
                    inline
                    dense
                    :autofocus="index === 0"
                  />
                </div>
                
                <div v-else class="q-mb-md">
                  <div class="text-caption text-weight-bold q-mb-xs" style="white-space: normal; display: block; width: 100%; color: var(--q-primary); line-height: 1.2;">{{ p.label }}:</div>
                  <q-input 
                    v-model="filterValues[p.label]"
                    filled
                    dense
                    :autofocus="index === 0"
                  />
                </div>
              </template>
            </div>
          </div>
        </q-card-section>

        <q-separator />
        <q-card-section v-if="dialogSQLPreview" class="bg-dark text-amber-3 q-pa-sm font-mono text-caption scroll" style="max-height: 150px; white-space: pre-wrap;">
          <div class="text-weight-bold q-mb-xs text-grey-5 uppercase">SQL-Vorschau (Live):</div>
          {{ dialogSQLPreview }}
        </q-card-section>

        <q-card-actions align="right" class="q-pa-md">
          <q-btn label="Abbrechen" color="negative" flat v-close-popup />
          <q-btn label="Report ausführen" color="primary" unelevated @click="confirmFilters" />
          <q-btn v-if="currentReportType === 'S'" label="Drucken" color="orange-9" unelevated icon="print" @click="printSimpleReport" />
        </q-card-actions>
      </q-card>
    </q-dialog>
    
    <!-- QUICK PARAM EDIT DIALOG -->
    <q-dialog v-model="showQuickParamEdit" persistent backdrop-filter="blur(4px)">
      <q-card style="width: 700px; max-width: 90vw;" :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-white'">
        <q-card-section class="row items-center bg-teal-8 text-white q-py-sm">
          <q-icon name="settings_suggest" size="sm" class="q-mr-md" />
          <div class="text-h6">Parameter & Format definieren</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>
        
        <q-card-section class="q-pa-md">
           <q-banner dense class="bg-blue-1 text-blue-9 rounded-borders q-mb-md">
             Hier können Sie die Bezeichnungen und Eingabe-Typen für diesen Bericht anpassen.
           </q-banner>
           
           <q-list bordered separator class="rounded-borders">
             <q-item v-for="d in currentDefinitions" :key="d.term" class="q-py-sm">
               <q-item-section>
                 <div class="row q-col-gutter-sm items-center">
                   <div class="col-12 col-md-3 font-mono text-weight-bold text-teal-8">
                     {{ d.term }}
                   </div>
                   <div class="col-12 col-md-5">
                     <q-input v-model="d.label" label="Anzeige-Label" dense filled />
                   </div>
                   <div class="col-12 col-md-4">
                     <q-select 
                        v-model="d.type" 
                        label="Eingabe-Typ" 
                        :options="['TEXT', 'NUMBER', 'DATE', 'BOOLEAN']" 
                        dense filled 
                      />
                   </div>
                 </div>
               </q-item-section>
             </q-item>
           </q-list>
        </q-card-section>
        
        <q-card-actions align="right" class="q-pa-md">
          <q-btn label="Abbrechen" color="grey" flat v-close-popup />
          <q-btn label="Speichern" color="teal-8" unelevated icon="save" @click="saveQuickParams" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- Editor Dialog (für Konfiguration) -->
    <q-dialog v-model="showConfigDialog" persistent maximized transition-show="slide-up" transition-hide="slide-down">
      <q-card :class="[$q.dark.isActive ? 'bg-dark' : 'bg-grey-1', 'flex column']">
        <q-card-section class="row items-center bg-purple text-white q-py-sm">
          <q-icon name="code" size="sm" class="q-mr-md" />
          <div class="text-h6">{{ isEditing ? 'Bericht bearbeiten' : 'Neuer Bericht' }}</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-md col scroll">
          <div class="row q-col-gutter-lg full-height">
            <!-- LINKE SEITE: STEPPER -->
            <div class="col-12 col-md-9">
              <q-stepper
                v-model="configStep"
                header-nav
                ref="stepper"
                color="primary"
                animated
                flat
                bordered
                class="rounded-borders no-shadow bg-transparent"
              >
                <!-- SCHRITT 1: BASIS DATEN -->
                <q-step
                  :name="1"
                  title="1. Name & Art"
                  icon="edit"
                  :done="configStep > 1"
                >
                  <div class="row q-col-gutter-lg">
                    <div class="col-12 col-md-6">
                      <q-input 
                        v-model="configForm.BESCHREIBUNG" 
                        label="Bericht Titel (Anzeigename) *" 
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                        :rules="[val => !!val || 'Titel erforderlich']"
                        :disable="isSystemLocked"
                      />
                    </div>
                    <div class="col-12 col-md-6">
                      <q-select 
                        v-model="configForm.TYP_KZ" 
                        label="Art des Berichts *" 
                        :options="[
                          {label: 'L - Einfache Liste', value: 'S'},
                          {label: 'T - Master-Detail (Template)', value: 'T'},
                          {label: 'M - Master-Detail (Gitter)', value: 'M'}
                        ]"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                      />
                    </div>
                    <div class="col-12 col-md-12">
                      <q-select 
                        v-model="configForm.ROOT_KZ" 
                        label="Basis Kategorie (Root) *" 
                        :options="[
                          {label: 'Einfache Listen', value: 'L'},
                          {label: 'Master Detail(Template)', value: 'T'},
                          {label: 'Master Detail (Gitter)', value: 'M'},
                          {label: 'Keine Zuordnung', value: 'x'}
                        ]"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-2'"
                      />
                    </div>
                    <div class="col-12 col-md-12 text-center q-pa-lg rounded-borders" :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-blue-grey-1'">
                      <p class="text-caption">Legen Sie zuerst den Typ fest, danach folgen die SQL-Statements.</p>
                    </div>
                    <div class="col-12 col-md-12">
                      <q-checkbox 
                         v-model="configForm.IST_SUMMENZEILE" 
                         label="Summenzeile am Ende des Berichts ausgeben" 
                         color="orange-9" 
                         icon="add_circle"
                         keep-color
                         :disable="isSystemLocked"
                      />
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 2: MASTER & ÜBERSCHRIFT -->
                <q-step
                  :name="2"
                  title="Master & Header"
                  icon="settings"
                  :done="configStep > 2"
                >
                  <div class="row q-col-gutter-lg">
                    <div class="col-12 col-md-9">
                      <q-input 
                        v-model="configForm.BESCHREIBUNG" 
                        label="Beschreibung / Titel des Berichts *" 
                        filled 
                        stack-label 
                        autofocus
                        ref="firstInput"
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        :rules="[val => !!val || 'Pflichtfeld']"
                        :disable="isSystemLocked"
                      />
                    </div>
                    <div class="col-12 col-md-3 flex items-center">
                      <q-checkbox 
                        v-model="configForm.SYSTEM_KZ" 
                        label="System-Eintrag" 
                        color="red"
                        :disable="!isAdmin"
                      />
                    </div>

                    <div class="col-12 col-md-3">
                      <q-select 
                        v-model="configForm.ROOT_KZ" 
                        label="Haupt-Kategorie (Root) *" 
                        :options="[
                          {label: 'Einfache Listen', value: 'L'},
                          {label: 'Master Detail(Template)', value: 'T'},
                          {label: 'Master Detail (Gitter)', value: 'M'}
                        ]"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        :rules="[val => !!val || 'Pflichtfeld']"
                        :disable="isSystemLocked"
                      />
                    </div>

                    <div class="col-12 col-md-3">
                      <q-select 
                        v-model="configForm.KATEGORIE_KZ" 
                        label="Art des Eintrags *" 
                        :options="[
                          {label: 'K - Kategorie / Ordner', value: 'K'}, 
                          {label: 'R - Report / Liste', value: 'L'}
                        ]"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        :rules="[val => !!val || 'Pflichtfeld']"
                        :disable="isSystemLocked"
                      />
                    </div>

                    <div class="col-12 col-md-3">
                      <q-select 
                        v-if="configForm.KATEGORIE_KZ !== 'K'"
                        v-model="configForm.TYP_KZ" 
                        label="Genaue Berichts-Art *" 
                        :options="[
                          {label: 'L - Einfache Liste', value: 'S'},
                          {label: 'T - Master-Detail (Template)', value: 'T'},
                          {label: 'M - Master-Detail (Gitter)', value: 'M'}
                        ]"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        :rules="[val => !!val || 'Berichts-Art wählen']"
                        :disable="isSystemLocked"
                      />
                    </div>
                    <div class="col-12 col-md-3">
                      <q-select 
                        v-if="configForm.KATEGORIE_KZ !== 'K'"
                        v-model="configForm.GRUPPEN_KZ" 
                        label="Zugeordneter Ordner"
                        :options="filteredFolderOptions"
                        emit-value
                        map-options
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        use-input
                        @filter="filterFolders"
                        @update:model-value="onFolderSelected"
                        :rules="[val => !!val || 'Bitte wählen Sie einen Ordner']"
                        :disable="isSystemLocked"
                      >
                      </q-select>
                      <q-input 
                        v-else
                        v-model="configForm.GRUPPEN_KZ" 
                        label="Gruppen-Code (Eigener Code)" 
                        filled 
                        stack-label 
                        maxlength="1"
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        :rules="[val => !!val || 'Code erforderlich']"
                        :disable="isSystemLocked"
                      />
                    </div>
                    <div class="col-12 col-md-4">
                      <q-input 
                        v-if="['T', 'F'].includes(configForm.TYP_KZ) || configForm.KATEGORIE_KZ === 'F'"
                        v-model="configForm.TEMPLATE_NAME" 
                        label="Template Name" 
                        filled 
                        stack-label 
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                        placeholder="Name der .html Datei"
                        :disable="isSystemLocked"
                      >
                        <template v-slot:append>
                          <q-btn 
                            round 
                            dense 
                            flat 
                            icon="edit_note" 
                            color="secondary" 
                            @click="openTemplateEditor()" 
                            :disable="!configForm.TEMPLATE_NAME"
                          >
                            <q-tooltip>Template im Editor öffnen</q-tooltip>
                          </q-btn>
                        </template>
                      </q-input>
                    </div>

                    <!-- No more SQL here -->
                  </div>
                </q-step>

                <!-- TEIL 3: VERKNÜPFUNG -->
                <!-- TEIL 3: MASTER-SQL -->
                <q-step
                  :name="3"
                  title="3. Master-SQL (Ebene 2)"
                  icon="storage"
                  :done="configStep > 3"
                  :disable="configForm.KATEGORIE_KZ === 'K'"
                >
                  <div class="q-pa-md">
                    <div class="text-subtitle2 q-mb-sm" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'">Technisches SQL (Intern)</div>
                    <q-input 
                      v-model="configForm.SQLSTATEMENT" 
                      type="textarea" 
                      filled 
                      rows="15"
                      readonly
                      input-style="height: 300px;"
                      :input-class="$q.dark.isActive ? 'font-mono text-grey-4' : 'font-mono text-grey-6'"
                      @focus="sqlTarget = 'SQLSTATEMENT'"
                      :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-1'"
                      class="q-mb-md"
                    />
                     <div class="text-subtitle1 text-weight-bold text-primary q-mb-sm">Master SQL (Nativ / Manuell)</div>
                    <q-input 
                      v-model="configForm.SQLSTATEMENT_NATIVE" 
                      type="textarea" 
                      filled 
                      placeholder="Hier Ihr manuelles SQL schreiben..." 
                      rows="15" 
                      input-style="height: 300px;"
                      :input-class="$q.dark.isActive ? 'font-mono text-blue-2' : 'font-mono text-blue-9'" 
                      :bg-color="$q.dark.isActive ? 'grey-10' : 'blue-1'" 
                      @focus="sqlTarget = 'SQLSTATEMENT_NATIVE'"
                      @dragover.prevent
                      @drop="onDrop($event, 'SQLSTATEMENT_NATIVE')"
                    />

                    <!-- Vorschau für das generierte SQL -->
                    <div v-if="configForm.SQLSTATEMENT && configForm.SQLSTATEMENT !== configForm.SQLSTATEMENT_NATIVE" class="q-mt-md">
                      <div class="row items-center justify-between q-mb-xs">
                        <div class="text-caption text-orange-9 weight-bold">Vorschlag vom SQL-Builder:</div>
                        <q-btn size="sm" color="orange-9" label="In Editor übernehmen" icon="content_copy" flat @click="configForm.SQLSTATEMENT_NATIVE = configForm.SQLSTATEMENT" />
                      </div>
                      <q-input 
                        v-model="configForm.SQLSTATEMENT" 
                        type="textarea" 
                        filled 
                        rows="15"
                        readonly
                        input-style="height: 300px;"
                        :input-class="$q.dark.isActive ? 'font-mono text-grey-4' : 'font-mono text-grey-7'"
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-1'"
                        @dragover.prevent 
                        @drop="onDrop($event, 'SQLSTATEMENT')"
                      />
                    </div>
                    
                    <div class="row q-mt-md q-gutter-x-sm justify-between items-center">
                       <div class="row q-gutter-x-sm">
                         <q-btn label="SQL bilden" icon="auto_fix_high" color="orange-9" unelevated @click="magicWandNative('SQLSTATEMENT')" />
                         <q-btn label="SQL bereinigen" icon="text_format" color="teal" unelevated @click="cleanupSqlForExplorer('SQLSTATEMENT_NATIVE')">
                           <q-tooltip>SQL vom QueryBuilder aufbereiten: Alles in Großbuchstaben, Hochkommas entfernen</q-tooltip>
                         </q-btn>
                         <q-btn label="Filter bilden" icon="filter_alt" color="primary" unelevated @click="openFilterBuilder('SQLSTATEMENT')" />
                         <q-btn label="Where löschen" icon="filter_alt_off" flat color="grey-7" @click="clearWhere('SQLSTATEMENT')" />
                       </div>
                       <q-btn label="SQL Testen" icon="play_arrow" color="primary" unelevated @click="runSqlPreview(configForm.SQLSTATEMENT_NATIVE || configForm.SQLSTATEMENT)" />
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 4: DETAIL-SQL -->
                <q-step
                  :name="4"
                  title="4. Detail-SQL (Ebene 3)"
                  icon="list"
                  :done="configStep > 4"
                  :disable="!['T', 'M'].includes(configForm.TYP_KZ)"
                >
                  <div class="q-pa-md">
                    <div class="text-subtitle2 q-mb-sm text-grey-7">Technisches Detail-SQL (Intern)</div>
                    <q-input 
                      v-model="configForm.DETAIL_SQL" 
                      type="textarea" 
                      filled 
                      rows="15"
                      readonly
                      input-style="height: 300px;"
                      :input-class="$q.dark.isActive ? 'font-mono text-grey-4' : 'font-mono text-grey-6'"
                      @focus="sqlTarget = 'DETAIL_SQL'"
                      :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-1'"
                      class="q-mb-md"
                    />
                    <div class="text-subtitle1 text-weight-bold text-secondary q-mb-sm">Detail SQL (Nativ / Manuell)</div>
                    <q-input 
                      v-model="configForm.DETAIL_SQL_NATIVE" 
                      type="textarea" 
                      filled 
                      placeholder="Hier Ihr manuelles Detail-SQL schreiben..." 
                      rows="15" 
                      input-style="height: 300px;"
                      :input-class="$q.dark.isActive ? 'font-mono text-indigo-2' : 'font-mono text-indigo-9'" 
                      :bg-color="$q.dark.isActive ? 'grey-10' : 'indigo-1'" 
                      @focus="sqlTarget = 'DETAIL_SQL_NATIVE'"
                      :disable="isSystemLocked"
                      @dragover.prevent
                      @drop="onDrop($event, 'DETAIL_SQL_NATIVE')"
                    />

                    <!-- Vorschau für das generierte Detail-SQL -->
                    <div v-if="configForm.DETAIL_SQL && configForm.DETAIL_SQL !== configForm.DETAIL_SQL_NATIVE" class="q-mt-md">
                      <div class="row items-center justify-between q-mb-xs">
                        <div class="text-caption text-secondary weight-bold">Vorschlag vom SQL-Builder (Detail):</div>
                        <q-btn size="sm" color="secondary" label="In Editor übernehmen" icon="content_copy" flat @click="configForm.DETAIL_SQL_NATIVE = configForm.DETAIL_SQL" />
                      </div>
                      <q-input 
                        v-model="configForm.DETAIL_SQL" 
                        type="textarea" 
                        filled 
                        rows="15"
                        readonly
                        input-style="height: 300px;"
                        :input-class="$q.dark.isActive ? 'font-mono text-grey-4' : 'font-mono text-grey-7'"
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'grey-1'"
                        @dragover.prevent 
                        @drop="onDrop($event, 'DETAIL_SQL')"
                      />
                    </div>
                    
                    <div class="row q-mt-md q-gutter-x-sm justify-between items-center">
                       <div class="row q-gutter-x-sm">
                         <q-btn label="SQL bilden" icon="auto_fix_high" color="orange-9" unelevated @click="magicWandNative('DETAIL_SQL')" />
                         <q-btn label="SQL bereinigen" icon="text_format" color="teal" unelevated @click="cleanupSqlForExplorer('DETAIL_SQL_NATIVE')">
                           <q-tooltip>SQL vom QueryBuilder aufbereiten: Alles in Großbuchstaben, Hochkommas entfernen</q-tooltip>
                         </q-btn>
                         <q-btn label="Filter bilden" icon="filter_alt" color="secondary" unelevated @click="openFilterBuilder('DETAIL_SQL')" />
                         <q-btn label="Where löschen" icon="filter_alt_off" flat color="grey-7" @click="clearWhere('DETAIL_SQL')" />
                       </div>
                       <q-btn label="SQL Testen" icon="play_arrow" color="secondary" unelevated @click="runSqlPreview(configForm.DETAIL_SQL_NATIVE || configForm.DETAIL_SQL)" />
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 5: GRUPPIERUNG & LAYOUT -->
                <q-step
                  :name="5"
                  title="Gruppierung & Layout"
                  icon="layers"
                  :done="configStep > 5"
                  :disable="!['T', 'M', 'S'].includes(configForm.TYP_KZ)"
                >
                  <div class="row q-col-gutter-lg">
                    <div class="col-12 col-md-6">
                      <q-input 
                        v-model="configForm.GROUP_FIELD" 
                        label="Gruppenwechsel nach Feld" 
                        filled 
                        placeholder="z.B. STALLNAME"
                        hint="Bei Änderung des Wertes in diesem Feld wird ein Gruppen-Header eingefügt."
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                      />
                    </div>
                    <div class="col-12 col-md-6">
                      <q-input 
                        v-model.number="configForm.ROWS_PER_PAGE" 
                        label="Zeilen pro Seite" 
                        type="number" 
                        filled 
                        placeholder="0 = unbegrenzt"
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                      />
                    </div>
                    <div class="col-12 col-md-6">
                      <q-select 
                        v-model="configForm.PAGE_ORIENTATION" 
                        label="Seiten-Ausrichtung" 
                        :options="[{label: 'Portrait (Hochkant)', value: 'P'}, {label: 'Landscape (Quer)', value: 'L'}]"
                        filled 
                        emit-value
                        map-options
                        :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
                      />
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 6: GRID-OPTIONEN -->
                <q-step
                  :name="6"
                  title="Grid & Anzeige"
                  icon="grid_on"
                  :disable="configForm.KATEGORIE_KZ === 'K'"
                >
                   <div class="q-pa-md">
                    <q-banner :class="$q.dark.isActive ? 'bg-indigo-9 text-indigo-1' : 'bg-indigo-1 text-indigo-9'" class="rounded-borders q-mb-lg">
                      <template v-slot:avatar><q-icon name="visibility" /></template>
                      Legen Sie fest, ob die Daten zusätzlich oder alternativ als interaktive Tabelle (Grid) angezeigt werden sollen.
                    </q-banner>
                    <div class="row q-col-gutter-md">
                      <div class="col-12 col-md-6">
                        <q-checkbox v-model="configForm.SHOW_MASTER_GRID" label="Master-Daten im Dashboard Grid anzeigen" color="primary" />
                      </div>
                      <div class="col-12 col-md-6">
                        <q-checkbox v-model="configForm.SHOW_DETAIL_GRID" label="Detail-Daten im Dashboard Grid anzeigen" color="secondary" />
                      </div>
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 7: SUMMENZEILE -->
                <q-step
                  :name="7"
                  title="Summenzeile"
                  icon="calculate"
                >
                  <div class="q-pa-md">
                    <q-banner v-if="!configForm.IST_SUMMENZEILE" :class="$q.dark.isActive ? 'bg-red-9 text-red-1' : 'bg-red-1 text-red-9'" class="rounded-borders q-mb-lg border-red">
                      <template v-slot:avatar><q-icon name="warning" /></template>
                      <strong>Hinweis:</strong> Die Summenzeile ist für diesen Bericht aktuell in Schritt 1 deaktiviert. 
                      Sie können das SQL hier zwar vorbereiten, es wird aber erst ausgeführt, wenn Sie den Schalter "Summenzeile ausgeben" in Schritt 1 umlegen.
                    </q-banner>

                    <q-banner v-else :class="$q.dark.isActive ? 'bg-orange-9 text-orange-1' : 'bg-orange-1 text-orange-9'" class="rounded-borders q-mb-lg">
                      <template v-slot:avatar><q-icon name="functions" /></template>
                      Hier können Sie ein SQL-Statement definieren, das eine einzelne Zeile mit Summen oder Statistiken liefert.
                      Diese Zeile wird am Ende des Berichts (beim Druck/Export) angefügt.
                    </q-banner>
                    
                    <div class="text-subtitle1 text-weight-bold text-orange-9 q-mb-sm">Summen-SQL (Nativ)</div>
                    <q-input 
                      v-model="tempSummenSql" 
                      type="textarea" 
                      filled 
                      placeholder="SELECT SUM(MENGE) as GESAMT, AVG(PREIS) as DURCHSCHNITT FROM ..." 
                      rows="8" 
                      :input-class="$q.dark.isActive ? 'font-mono text-orange-2' : 'font-mono text-orange-9'" 
                      :bg-color="$q.dark.isActive ? 'grey-10' : 'orange-1'" 
                      style="font-family: monospace"
                      @dragover.prevent
                      @drop="onDrop($event, 'SUMMENZEILE')" 
                      @focus="sqlTarget = 'SUMMENZEILE'"
                      :disable="isSystemLocked"
                    />
                    
                    <div class="row q-mt-md q-gutter-x-sm justify-between items-center">
                       <div class="row q-gutter-x-sm">
                         <q-btn label="SQL Summenzeilen erstellen" icon="auto_fix_high" color="orange-9" unelevated @click="magicWandNative('SUMMENZEILE')" />
                         <q-btn label="Filter bilden" icon="filter_alt" color="primary" unelevated @click="openFilterBuilder('SUMMENZEILE')" />
                         <q-btn label="SQL leeren" icon="delete_sweep" flat color="negative" @click="configForm.SUMMENZEILE = ''" />
                       </div>
                       <q-btn label="Summen-Test" icon="play_arrow" color="orange-9" unelevated @click="runSqlPreview(configForm.SUMMENZEILE)" />
                    </div>
                  </div>
                </q-step>

                <!-- TEIL 8: PARAMETER DEFINITIONEN -->
                <q-step
                  :name="8"
                  title="8. Abfrage-Parameter"
                  icon="contact_support"
                >
                  <div class="q-pa-md">
                    <q-banner :class="$q.dark.isActive ? 'bg-blue-9 text-blue-1' : 'bg-blue-1 text-blue-9'" class="rounded-borders q-mb-lg">
                      <template v-slot:avatar><q-icon name="info" /></template>
                      Legen Sie hier fest, wie die im SQL gefundenen Abfragen (z.B. `Datum`) benannt und welcher Typ sie sein sollen.
                    </q-banner>

                    <div v-if="currentDefinitions.length === 0" class="text-center q-pa-xl text-grey-6 border-dashed rounded-borders">
                      <q-icon name="search_off" size="4em" class="q-mb-md" />
                      <div class="text-h6">Keine Parameter erkannt</div>
                      <p>Keine Begriffe in Backticks oder Prozent-Platzhalter im SQL gefunden.</p>
                      <q-btn label="SQL scannen" icon="refresh" color="primary" flat @click="syncParamsFromSql" />
                    </div>

                    <q-list v-else bordered separator class="rounded-borders">
                      <q-item v-for="(d, index) in currentDefinitions" :key="d.term" class="q-py-md">
                        <q-item-section>
                          <div class="row q-col-gutter-sm items-center">
                            <div class="col-12 col-md-3 text-weight-bold font-mono text-primary">
                              <q-badge color="primary" outline>{{ d.term }}</q-badge>
                            </div>
                            <div class="col-12 col-md-5">
                              <q-input v-model="d.label" label="Anzeige-Label (Fragetext)" dense filled />
                            </div>
                            <div class="col-12 col-md-4">
                              <q-select 
                                v-model="d.type" 
                                label="Eingabe-Typ" 
                                :options="['TEXT', 'NUMBER', 'DATE', 'BOOLEAN']" 
                                dense filled 
                              />
                            </div>
                          </div>
                        </q-item-section>
                      </q-item>
                    </q-list>
                  </div>
                </q-step>

                <template v-slot:navigation>
                  <q-stepper-navigation class="q-pt-md row items-center q-gutter-x-sm">
                    <q-btn v-if="configStep === 1" label="Template erstellen" color="orange-9" icon="auto_fix_high" @click="createBoilerplateTemplate" unelevated rounded />
                    <q-btn v-if="configStep === 1" label="Template Anzeigen" color="teal-7" icon="visibility" @click="showTemplatePreview" outline rounded />
                    <q-btn v-if="configStep === 1" label="Vorschau Ausdruck" color="blue-grey-7" icon="print" @click="showPrintPreview" outline rounded />
                    
                    <q-space />
                    
                    <q-btn v-if="configStep > 1" flat color="primary" @click="configStep--" label="Zurück" icon="chevron_left" />
                    <q-btn v-if="configStep < 8 && !(configStep === 2 && configForm.KATEGORIE_KZ === 'K')" color="primary" @click="() => { if (configStep === 7) syncParamsFromSql(); configStep++; }" label="Weiter" icon-right="chevron_right" />
                    
                    <q-btn 
                      v-if="configStep === 8 || (configStep === 2 && configForm.KATEGORIE_KZ === 'K')" 
                      label="Bericht Speichern" 
                      color="positive" 
                      icon="save" 
                      @click="onConfigSubmit" 
                      rounded unelevated
                    />
                  </q-stepper-navigation>
                </template>
              </q-stepper>
            </div>

            <!-- RECHTE SEITE: DB EXPLORER (HILFESTELLUNG) -->
            <div class="col-12 col-md-3 border-left q-pl-md">
              <div class="column full-height">
                <div class="text-subtitle2 q-mb-sm row items-center" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-8'">
                  <q-icon name="storage" class="q-mr-xs" /> DB-Explorer
                </div>
                <q-input v-model="filterTableText" dense filled placeholder="Tabelle suchen..." class="q-mb-sm shadow-1" :bg-color="$q.dark.isActive ? 'grey-9' : 'white'">
                   <template v-slot:append><q-icon name="search" size="xs" /></template>
                 </q-input>
                 
                 <div class="row q-gutter-sm q-mb-sm" v-if="selectedExplorerCols.length > 0">
                   <q-btn 
                     label="Einfügen" 
                     color="primary" 
                     dense 
                     unelevated 
                     icon="add" 
                     class="col"
                     @click="appendSelectedExplorerCols"
                   >
                     <q-badge color="orange" floating>{{ selectedExplorerCols.length }}</q-badge>
                     <q-tooltip>Markierte Felder mit Komma getrennt einfügen</q-tooltip>
                   </q-btn>
                   <q-btn 
                     icon="clear_all" 
                     color="grey-7" 
                     flat 
                     dense 
                     round
                     @click="selectedExplorerCols = []"
                   >
                     <q-tooltip>Auswahl aufheben</q-tooltip>
                   </q-btn>
                 </div>
                
                <div 
                  class="col scroll rounded-borders shadow-1" 
                  :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-white'"
                  style="border: 2px dashed #ccc; min-height: 100px;"
                  @dragover.prevent="(e) => e.dataTransfer.dropEffect = 'move'"
                  @drop="onExplorerDrop"
                >
                  <q-list dense padding>
                    <q-expansion-item
                      v-for="table in dbTables.filter(t => t.toLowerCase().includes(filterTableText.toLowerCase()))"
                      :key="table"
                      header-class="text-weight-bold text-primary"
                      @show="loadColumns(table)"
                      dense
                    >
                      <template v-slot:header>
                        <div class="row no-wrap items-center full-width">
                          <q-icon 
                            name="drag_indicator" 
                            class="cursor-move q-mr-sm text-grey-6" 
                            draggable="true" 
                            @dragstart.stop="onTableDragStart($event, table)" 
                          />
                          <div class="col text-subtitle2">{{ table }}</div>
                        </div>
                      </template>
                      <q-card class="bg-grey-1">
                        <q-card-section class="q-pa-none">
                          <q-item 
                            v-for="col in tableColumns.filter(c => !c.toUpperCase().startsWith('ID_'))" 
                            :key="col" 
                            dense 
                            class="q-py-none"
                          >
                            <q-item-section side>
                              <q-checkbox v-model="selectedExplorerCols" :val="col" size="xs" color="primary" />
                            </q-item-section>
                            <q-item-section class="q-pl-none">
                              <q-btn 
                                flat 
                                dense 
                                class="full-width text-left text-caption text-weight-regular no-caps" 
                                align="left"
                                @click="appendToSql(col)"
                                draggable="true"
                                @dragstart.stop="onColumnDragStart($event, col, table)"
                              >
                                {{ col }}
                              </q-btn>
                            </q-item-section>
                          </q-item>
                          <div v-if="tableColumns.length === 0" class="text-center q-pa-sm text-grey-5">Lade...</div>
                        </q-card-section>
                      </q-card>
                    </q-expansion-item>
                  </q-list>
                </div>
                <div class="text-caption text-grey-6 q-mt-sm italic">
                  Klicken um Feld zum aktuellen SQL-Block hinzuzufügen.
                </div>
              </div>
            </div>
          </div>
        </q-card-section>

        <q-card-actions align="right" :class="[$q.dark.isActive ? 'bg-grey-10 text-white' : 'bg-white', 'q-pa-md border-top']">
          <q-btn label="Abbrechen" color="grey-7" flat v-close-popup icon="cancel" />
          <q-btn 
            :label="isEditing ? 'Aktualisieren' : 'Speichern'" 
            color="primary" 
            unelevated 
            icon="save" 
            @click="onConfigSubmit"
            :loading="loadingConfig"
            :disable="isSystemLocked"
          />
        </q-card-actions>
      </q-card>
    </q-dialog>
    
    <!-- Template Editor Dialog -->
    <q-dialog v-model="showTemplateEditor" persistent maximized transition-show="fade" transition-hide="fade">
      <q-card class="bg-dark text-white flex column">
        <q-card-section class="row items-center bg-secondary text-white q-py-sm">
          <q-icon name="edit_note" size="sm" class="q-mr-md" />
          <div class="column">
            <div class="text-subtitle1 text-weight-bold">Template-Editor</div>
            <div class="text-caption text-grey-4">{{ templateEditorName }}</div>
          </div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="col q-pa-md relative-position" :class="$q.dark.isActive ? 'bg-dark text-white' : 'bg-white text-black'">
          <div v-if="loadingTemplate" class="absolute-full flex flex-center" :class="$q.dark.isActive ? 'bg-dark' : 'bg-white'" style="z-index: 10; opacity: 0.8;">
            <q-spinner-grid color="primary" size="64px" />
          </div>
          
          <div class="row items-center q-mb-sm text-grey-8">
            <q-icon name="info" class="q-mr-xs" />
            <div class="text-caption">Geladene Zeichen: <strong>{{ templateContent.length }}</strong></div>
          </div>

          <q-input
            v-model="templateContent"
            type="textarea"
            label="Template HTML Quellcode"
            filled
            outlined
            square
            rows="30"
            class="full-width font-mono"
            :input-class="$q.dark.isActive ? 'q-pa-md line-height-relaxed text-blue-2' : 'q-pa-md line-height-relaxed text-black'"
            :style="{ fontSize: '14px', background: $q.dark.isActive ? '#1d1d1d' : '#fff', color: $q.dark.isActive ? '#90caf9' : '#000' }"
            placeholder="Hier den HTML Code eingeben..."
          />
        </q-card-section>

        <q-card-actions align="right" class="bg-grey-10 q-pa-md">
          <div class="row items-center q-gutter-x-md full-width">
            <q-icon name="help_outline" color="grey-7" size="sm">
              <q-tooltip anchor="top middle" self="bottom middle">
                Verwenden Sie Platzhalter wie &lt;Master.FELD&gt;, &lt;Detail.FELD&gt; oder &lt;Summe.FELD&gt;.
              </q-tooltip>
            </q-icon>
            <q-space />
            <q-btn label="Abbrechen" color="grey" flat v-close-popup />
            <q-btn label="Template Speichern" color="secondary" icon="save" unelevated @click="saveTemplate" />
          </div>
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- PARAMETER ABFRAGE DIALOG FÜR TEST-SQL -->
    <q-dialog v-model="showTestParamsDialog" persistent>
      <q-card style="min-width: 400px">
        <q-card-section class="bg-amber-8 text-white row items-center">
          <div class="text-h6">SQL-Parameter eingeben</div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="q-pa-md">
          <div v-for="p in testParamsList" :key="p" class="q-mb-md">
            <q-input
              v-if="p.toLowerCase().includes('datum')"
              v-model="testParamsValues[p]"
              :label="testParamsLabels[p] || p"
              type="date"
              filled
              dense
              ref="testParamInputs"
            />
            <q-input
              v-else
              v-model="testParamsValues[p]"
              :label="testParamsLabels[p] || p"
              filled
              dense
              ref="testParamInputs"
            />
          </div>
        </q-card-section>

        <q-card-actions align="right" class="bg-grey-1">
          <q-btn label="Abbrechen" color="grey" flat v-close-popup />
          <q-btn label="SQL jetzt ausführen" color="secondary" unelevated @click="runSqlPreview(currentTestSql, true); showTestParamsDialog = false" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- FILTER BUILDER DIALOG -->
    <q-dialog v-model="showFilterBuilderDialog" persistent>
      <q-card style="min-width: 400px; max-height: 80vh">
        <q-card-section class="row items-center q-pb-none bg-purple text-white q-py-sm">
          <q-icon name="filter_alt" size="sm" class="q-mr-sm" />
          <div class="text-h6">Filter-Builder</div>
          <q-space />
          <q-btn label="Parameter def." color="white" flat icon="settings" dense class="q-mr-sm" @click="() => { showFilterBuilderDialog = false; configStep = 8; }">
            <q-tooltip>Zu den Parameter-Definitionen springen</q-tooltip>
          </q-btn>
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section>
          <div class="text-subtitle2 q-mb-xs text-grey-7">Tabelle auswählen (aus SQL erkannt):</div>
          <q-select
            v-model="filterBuilderSelectedTable"
            :options="filterBuilderTables"
            :option-label="opt => opt.alias ? `${opt.table} (${opt.alias})` : opt.table"
            filled
            dense
            @update:model-value="loadFilterColumns"
            class="q-mb-md"
          >
            <template v-slot:prepend>
              <q-icon name="table_chart" color="primary" />
            </template>
          </q-select>

          <div class="row items-center q-mb-sm">
            <div class="text-subtitle1 text-weight-bold">Spalten wählen</div>
            <q-space />
            <div class="text-caption text-grey-6">{{ availableFilterCols.length }} verfügbar</div>
          </div>
          
          <q-scroll-area style="height: 300px">
                <q-list dense bordered separator class="rounded-borders">
                  <q-item v-for="col in availableFilterCols" :key="col" tag="label" v-ripple>
                    <q-item-section side top>
                      <q-checkbox v-model="filterBuilderSelected" :val="col" />
                    </q-item-section>
                    <q-item-section>
                      <q-item-label class="text-weight-medium" style="white-space: normal; line-height: 1.2;">{{ col }}</q-item-label>
                    </q-item-section>
                    <q-item-section side v-if="filterBuilderSelected.includes(col)">
                      <div class="row q-col-gutter-xs">
                        <q-select
                          v-model="filterBuilderOps[col]"
                          :options="operatorOptions"
                          dense
                          outlined
                          label="Op"
                          style="width: 80px"
                          class="bg-white"
                          options-dense
                        />
                        <q-select
                          v-model="filterBuilderTypes[col]"
                          :options="['TEXT', 'DATE', 'NUMBER', 'BOOLEAN']"
                          dense
                          outlined
                          label="Typ"
                          style="width: 100px"
                          class="bg-white"
                          options-dense
                        />
                      </div>
                    </q-item-section>
                  </q-item>
                  <q-item v-if="availableFilterCols.length === 0" class="text-grey italic">
                    <q-item-section>Alle verfügbaren Spalten sind bereits im Filter.</q-item-section>
                  </q-item>
                </q-list>
          </q-scroll-area>
        </q-card-section>

        <q-card-actions align="right" class="bg-grey-1">
          <q-btn flat label="Abbrechen" color="grey" v-close-popup />
          <q-btn label="Filter einsetzen" color="primary" @click="applyFilterBuilder" icon="check" />
        </q-card-actions>
      </q-card>
    </q-dialog>

    <!-- SQL VORSCHAU DIALOG -->
    <q-dialog v-model="showPreview" full-width full-height>
      <q-card class="column no-wrap shadow-24">
        <q-card-section class="row items-center q-pb-none bg-secondary text-white">
          <div class="text-h6">
            <q-icon name="visibility" class="q-mr-sm" />
            SQL Ergebnis Vorschau (Nativ)
          </div>
          <q-space />
          <q-btn icon="close" flat round dense v-close-popup />
        </q-card-section>

        <q-card-section class="col scroll q-pt-md">
          <q-table
            :rows="previewRows"
            :columns="previewCols"
            row-key="ID"
            dense
            flat
            bordered
            separator="cell"
            :pagination="{ rowsPerPage: 15 }"
            :rows-per-page-options="[5, 10, 15]"
            class="sticky-header-table font-mono"
          >
            <template v-slot:no-data>
              <div class="full-width row flex-center text-grey q-gutter-sm q-pa-lg">
                <q-icon size="2em" name="sentiment_dissatisfied" />
                <span>Keine Daten gefunden. Das SQL-Statement lieferte 0 Zeilen zurück.</span>
              </div>
            </template>
          </q-table>
        </q-card-section>

        <q-card-actions align="right" class="bg-grey-1 text-primary q-pa-md">
          <div class="text-weight-bold q-mr-md">
            {{ previewRows.length }} Datensätze
          </div>
          <q-btn label="Schließen" color="primary" unelevated v-close-popup />
        </q-card-actions>
      </q-card>
    </q-dialog>

  </q-page>
</template>

<script setup lang="ts">
console.log('DynamicReportPage.vue: Script setup running...');
import { ref, reactive, onMounted, watch, computed, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import { api } from 'src/boot/api';
import { exportFile, useQuasar, debounce, copyToClipboard } from 'quasar';
import { useSessionStore } from '../stores/session';
import type { QTableProps, QTree } from 'quasar';
import { useResizableColumns } from '../composables/useResizableColumns';

const sessionStore = useSessionStore();
const route = useRoute();
const $q = useQuasar();

const reportGridId = computed(() => {
  const rid = selectedKey.value;
  return rid ? `Report_${rid}` : 'ReportGeneric';
});

const { 
  columnWidths: masterWidths, 
  startResize: masterStartResize, 
  initWidths: masterInitWidths, 
  isResizing: masterIsResizing 
} = useResizableColumns(computed(() => reportGridId.value + '_Master'));

const { 
  columnWidths: detailWidths, 
  startResize: detailStartResize, 
  initWidths: detailInitWidths, 
  isResizing: detailIsResizing 
} = useResizableColumns(computed(() => reportGridId.value + '_Detail'));
const sqlEditorRef = ref<any>(null);
const tab = ref(String(route.query.tab || 'anzeige'));
if (tab.value === 'konfiguration' && !sessionStore.can('sql_struktur_verwalten')) {
  tab.value = 'anzeige';
}
const configStep = ref(1);
const dbTables = ref<string[]>([]);
const selectedTable = ref('');
const tableColumns = ref<string[]>([]);
const filterTableText = ref('');
const selectedExplorerCols = ref<string[]>([]);
const sqlTarget = ref<'SQLSTATEMENT' | 'DETAIL_SQL' | 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE' | 'SUMMENZEILE'>('SQLSTATEMENT_NATIVE');

const currentReportLabel = ref('');
const currentReportType = ref('S');
const currentTemplateName = ref('');
const complexReportData = ref<any>(null);
const resultData = ref<any[]>([]);
const resultSums = ref<any[]>([]);
const resultColumns = ref<any[]>([]);
const detailServerCols = ref<string[]>([]);
const masterData = ref<any[]>([]);
const detailData = ref<any[]>([]);
const selectedMasterRows = ref<any[]>([]);
const showMasterGrid = ref(false);
const showDetailGrid = ref(false);
const executedSQL = ref('');
const loadingResult = ref(false);
const showHtmlReport = ref(false);
const generatedHtml = ref('');

watch(resultColumns, (newCols) => {
  if (newCols && newCols.length > 0) {
    masterInitWidths(newCols);
  }
}, { deep: true });

const masterColumns = computed(() => getAutoCols(masterData.value));
const detailColumns = computed(() => getAutoCols(detailData.value, detailServerCols.value));

watch(detailColumns, (newCols) => {
  if (newCols && newCols.length > 0) {
    detailInitWidths(newCols);
  }
}, { immediate: true });

// --- Template Editor State ---
const showTemplateEditor = ref(false);
const templateContent = ref('');
const templateEditorName = ref('');
const templateIsNew = ref(false);
const loadingTemplate = ref(false);

async function openTemplateEditor(name?: string) {
  let targetName = name || configForm.TEMPLATE_NAME;
  // Sicherstellen, dass wir einen String haben
  targetName = getVal(targetName).trim();
  
  if (!targetName) {
    $q.notify({ type: 'warning', message: 'Bitte geben Sie zuerst einen Template-Namen ein.' });
    return;
  }
  
  console.log('Versuche Template zu laden:', targetName);
  templateEditorName.value = targetName;
  loadingTemplate.value = true;
  showTemplateEditor.value = true;
  templateContent.value = ''; // Vorher leeren
  
  try {
    const res = await api.get(`/api/reports/template/${targetName}`);
    console.log('Template-Antwort empfangen:', res.data);
    templateContent.value = res.data.content || '';
    templateIsNew.value = !!res.data.is_new;
  } catch (err: any) {
    console.error('Fehler beim Template-Laden:', err);
    const msg = err.response?.data?.error || 'Template konnte nicht geladen werden.';
    $q.notify({ type: 'negative', message: msg });
    showTemplateEditor.value = false;
  } finally {
    loadingTemplate.value = false;
  }
}

async function saveTemplate() {
  if (!templateEditorName.value) return;
  
  try {
    await api.post(`/api/reports/template/${templateEditorName.value}`, {
      content: templateContent.value
    });
    $q.notify({ type: 'positive', message: `Template '${templateEditorName.value}' wurde gespeichert.` });
    showTemplateEditor.value = false;
  } catch (err: any) {
    const msg = err.response?.data?.error || 'Fehler beim Speichern des Templates.';
    $q.notify({ type: 'negative', message: msg });
  }
}

/**
 * Erstellt ein Basis-HTML Template basierend auf den aktuellen SQL-Statements
 */
async function createBoilerplateTemplate() {
  if (!configForm.BESCHREIBUNG && !configForm.TEMPLATE_NAME) {
     $q.notify({ type: 'warning', message: 'Bitte geben Sie zuerst Bericht-Titel oder Template-Namen ein.' });
     // Falls wir in Schritt 1 sind und nichts da ist, Abbruch
     if (!configForm.BESCHREIBUNG) return;
  }

  // Name generieren falls nicht vorhanden
  if (!configForm.TEMPLATE_NAME) {
    const safeName = (configForm.BESCHREIBUNG || 'Neuer_Bericht')
      .replace(/\s+/g, '_')
      .replace(/[^a-zA-Z0-9_]/g, '')
      .toLowerCase();
    configForm.TEMPLATE_NAME = safeName + '.html';
  }

  // Hilfsfunktion zum Extrahieren von Feldern (einfach gehalten)
  const extractCols = (sql: string) => {
    if (!sql) return [];
    // Wir suchen alles zwischen SELECT und FROM
    const m = sql.match(/SELECT\s+(.*?)\s+FROM/is);
    if (!m) return [];
    
    return m[1].split(',').map(c => {
      // Alias suchen (AS ...)
      const parts = c.trim().split(/\s+AS\s+/i);
      let col = parts[parts.length - 1].trim();
      // Tabellen-Prefix entfernen (T.FELD)
      if (col.includes('.')) {
        col = col.split('.').pop() || '';
      }
      return col.replace(/"/g, '').toUpperCase();
    }).filter(c => c && !c.includes('*') && !c.includes('(') && !c.includes(')'));
  };

  // Spalten extrahieren (ohne Modifikation des Originals!)
  const masterCols = extractCols(configForm.SQLSTATEMENT_NATIVE || configForm.SQLSTATEMENT);
  const detailCols = extractCols(configForm.DETAIL_SQL_NATIVE || configForm.DETAIL_SQL);

  let html = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>§MASTER.BESCHREIBUNG§</title>
  <style>
    @page { size: ${configForm.PAGE_ORIENTATION === 'L' ? 'landscape' : 'portrait'}; margin: 1cm; }
    body { font-family: 'Segoe UI', Arial, sans-serif; margin: 0; padding: 0; color: #333; line-height: 1.4; }
    
    /* Dynamischer Header Bereich */
    .report-header-container { margin-bottom: 20px; }
    
    .main-content { padding: 0.5cm; }
    @media print { .no-print { display: none; } }
    
    .report-title-section {
        width: 100% !important;
        margin-bottom: 1.5cm !important;
        border-bottom: 3px solid #1976D2 !important;
        padding-bottom: 20px !important;
        text-align: center !important;
      }
      .report-title {
        width: 100% !important;
        margin: 15px 0 5px 0 !important;
        font-size: 26pt !important;
        font-weight: bold !important;
        color: #1976D2 !important;
        text-align: center !important;
        text-transform: uppercase !important;
        letter-spacing: 3px !important;
      }
      .report-meta {
        width: 100% !important;
        text-align: center !important;
        font-size: 9pt !important;
        color: #666 !important;
        margin-top: 5px !important;
      }
      .company-info {
        font-size: 10pt;
        color: #333;
        font-style: italic;
      }
    
    .master-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; background: #f9f9f9; border: 1px solid #eee; padding: 15px; border-radius: 6px; margin-bottom: 30px; }
    .master-item { display: flex; flex-direction: column; }
    .master-label { font-size: 9px; font-weight: bold; color: #999; text-transform: uppercase; letter-spacing: 0.5px; }
    .master-value { font-size: 14px; font-weight: 500; color: #000; }

    h3 { color: #1976D2; border-left: 5px solid #1976D2; padding-left: 10px; margin: 25px 0 10px 0; font-size: 16px; text-transform: uppercase; }
    
    table { width: 100%; border-collapse: collapse; margin-top: 10px; }
    th { background-color: #1976D2; color: white; padding: 10px 8px; text-align: left; font-size: 11px; border: 1px solid #1565C0; }
    td { padding: 8px; border-bottom: 1px solid #eee; font-size: 13px; }
    tr:nth-child(even) { background-color: #fcfcfc; }
    
    .report-footer { position: fixed; bottom: 0; width: 100%; border-top: 1px solid #ddd; padding: 10px 0; font-size: 9px; color: #999; display: flex; justify-content: space-between; }
    
    @media print {
      .no-print { display: none; }
    }
  </style>
</head>
<body>
  <!-- DYNAMISCHER HEADER (PLATZHALTER FÜR TRANSLATION ODER SNIPPET) -->
  <div class="report-header-container">
    §REPORT_HEADER§
  </div>

  <div class="main-content">
    <div class="report-title-section">
      <div class="company-info">
        <strong><Firma.NAME></strong> | <Firma.ORT>
      </div>
      <h1 class="report-title">§REPORT_TITLE§</h1>
      <div class="report-meta">Ausgedruckt am: §#DATE§ um §#TIME§</div>
    </div>

    <!-- MASTER DATEN (GENERIERT AUS FELD 1) -->
    <div class="master-grid">
`;

  masterCols.forEach(c => {
    html += `      <div class="master-item">
        <span class="master-label">${c}</span>
        <span class="master-value">§MASTER.${c}§</span>
      </div>\n`;
  });

  html += `    </div>

    <h3>Details / Positionen</h3>
    <table>
      <thead>
        <tr>\n`;

  detailCols.forEach(c => {
    html += `          <th>${c}</th>\n`;
  });

  html += `        </tr>
      </thead>
      <tbody>
        <!-- BEGIN Detail -->
        <tr>\n`;

  detailCols.forEach(c => {
    html += `          <td><Detail.${c}></td>\n`;
  });

  html += `        </tr>
        <!-- END Detail -->
      </tbody>
    </table>
  </div>

  <!-- FOULTER / FOOTER -->
  <div class="report-footer">
    <div>HuhnLite Reporting | §MASTER.BEZEICHNUNG§</div>
    <div>Seite 1</div>
    <div>Firma: <Firma.NAME> | Ort: <Firma.ORT></div>
  </div>
</body>
</html>`;

  templateContent.value = html;
  templateEditorName.value = configForm.TEMPLATE_NAME;
  templateIsNew.value = true;
  showTemplateEditor.value = true;
  $q.notify({ type: 'info', message: 'Standard Template basierend auf SQL generiert.', icon: 'auto_fix_high' });
}

function showTemplatePreview() {
  if (!configForm.TEMPLATE_NAME) {
    $q.notify({ type: 'warning', message: 'Bitte zuerst im Schritt "Master & Header" einen Template-Namen festlegen.' });
    return;
  }
  openTemplateEditor(configForm.TEMPLATE_NAME);
}

async function handlePrintRequest() {
  if (currentReportType.value === 'S') {
    await printSimpleReport();
  } else {
    await showPrintPreview();
  }
}

async function printSimpleReport() {
  if (!currentReport.value) return;
  $q.loading.show({ message: 'Druck-Dokument wird erstellt...' });
  try {
    const rId = getP(currentReport.value, 'ID');
    const res = await api.post(`/api/reports/execute/${rId}`, { 
      PARAMS: { ...filterValues, _PRINT_: true } 
    });
    if (res.data && res.data.html) {
      generatedHtml.value = res.data.html;
      showPrintPreview();
    } else {
      $q.notify({ type: 'warning', message: 'Konnte kein Druck-Format generieren.' });
    }
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Abrufen des Druck-Layouts' });
  } finally {
    $q.loading.hide();
  }
}

async function showPrintPreview() {
  // WICHTIG: Wenn wir ein vom Backend generiertes HTML haben (aus der Ausführung), nehmen wir das!
  let content = generatedHtml.value || templateContent.value;
  
  if (!content && configForm.TEMPLATE_NAME) {
    $q.loading.show({ message: 'Lade Template für Vorschau...' });
    try {
      const res = await api.get(`/api/reports/template/${configForm.TEMPLATE_NAME}`);
      content = res.data.content;
    } catch (_err) {
      $q.notify({ type: 'negative', message: 'Konnte Template nicht laden.' });
    } finally {
      $q.loading.hide();
    }
  }

  if (!content) {
    $q.notify({ type: 'warning', message: 'Kein Berichtsziel vorhanden. Bitte erst den Report ausführen oder ein Template wählen.' });
    return;
  }

  // Wenn es bereits ein fertiger Report ist (enthält <html>), dann direkt anzeigen
  if (content.includes('<html') || content.includes('<HTML')) {
    const win = window.open('', '_blank');
    if (win) {
      win.document.write(content);
      win.document.close();
      // Optional: direkt den Druckdialog öffnen
      // setTimeout(() => win.print(), 500);
    }
    return;
  }

  // --- LIVE VORSCHAU / MOCK ERSETZUNG FÜR DEN BROWSER ---
  let previewHtml = content;
  const now = new Date();
  
  // 1. Bericht & System
  previewHtml = previewHtml.replace(/§REPORT_TITLE§/g, configForm.BESCHREIBUNG || 'Unbenannter Bericht');
  previewHtml = previewHtml.replace(/§REPORT_BESCHREIBUNG§/g, configForm.BESCHREIBUNG || 'Unbenannter Bericht');
  previewHtml = previewHtml.replace(/§MASTER_HEADER§/g, configForm.BESCHREIBUNG || 'Unbenannter Bericht');
  previewHtml = previewHtml.replace(/§#DATE§/g, now.toLocaleDateString('de-DE'));
  previewHtml = previewHtml.replace(/§#TIME§/g, now.getHours() + ':' + String(now.getMinutes()).padStart(2, '0'));
  
  // 2. Master (Mocking der ersten Zeile, falls möglich)
  previewHtml = previewHtml.replace(/§MASTER\.(.*?)§/g, (match, field) => {
    return '[' + field + ']'; // Zeigt Feldname in Klammern an
  });
  
  // 3. Firma (Statische Mocks für die Vorschau)
  previewHtml = previewHtml.replace(/<Firma\.NAME>/g, 'Musterfarm GmbH');
  previewHtml = previewHtml.replace(/<Firma\.ORT>/g, 'Musterhausen');
  
  // 4. Detail-Bereich (Lösche Kommentar-Tags für die Vorschau)
  previewHtml = previewHtml.replace(/<!-- BEGIN Detail -->|<!-- END Detail -->/g, '');

  // 5. Schließen-Button injizieren (nur Screen, nicht Print)
  const closeBtnHtml = `
    <div class="no-print" style="position: fixed; top: 20px; right: 20px; z-index: 9999;">
      <button onclick="window.close()" style="padding: 10px 20px; background: #333; color: white; border: none; border-radius: 5px; cursor: pointer; font-family: sans-serif; box-shadow: 0 2px 10px rgba(0,0,0,0.3);">
        Vorschau schließen
      </button>
    </div>
  `;
  previewHtml = previewHtml.replace('</body>', closeBtnHtml + '</body>');

  const win = window.open('', '_blank');
  if (win) {
    win.document.write(previewHtml);
    win.document.close();
  }
}

async function loadTables() {
  try {
    const res = await api.get('/api/schema/tables');
    dbTables.value = res.data || [];
  } catch (_err) {}
}

async function loadColumns(table: string) {
  if (!table) return;
  try {
    const res = await api.get(`/api/schema/columns/${table}`);
    tableColumns.value = res.data || [];
    selectedTable.value = table;
  } catch (_err) {}
}

/**
 * Sucht nach dem Index eines Keywords in der Root-Ebene (außerhalb von Klammern).
 * Hilft dabei, das Haupt-FROM, WHERE etc. von Subqueries zu unterscheiden.
 */
function findRootKeywordIndex(sql: string, keyword: string, last = false): number {
  if (!sql) return -1;
  const upSql = sql.toUpperCase();
  const upKw = keyword.toUpperCase().trim();
  const kwLen = upKw.length;
  let depth = 0;
  let inSingleQuote = false;
  let inDoubleQuote = false;
  let lastFoundIndex = -1;
  
  for (let i = 0; i < sql.length; i++) {
    const char = sql[i];
    
    // Trivialer Quote-Check (reicht für die meisten SQLs hier aus)
    if (char === "'" && !inDoubleQuote) {
      inSingleQuote = !inSingleQuote;
      continue;
    }
    if (char === '"' && !inSingleQuote) {
      inDoubleQuote = !inDoubleQuote;
      continue;
    }
    
    if (inSingleQuote || inDoubleQuote) continue;

    if (char === '(') depth++;
    else if (char === ')') depth--;
    else if (depth === 0 && i <= sql.length - kwLen) {
      const sub = upSql.substring(i, i + kwLen);
      if (sub === upKw) {
        // Wortgrenzen prüfen: Keine Buchstaben, Zahlen oder Unterstrich davor/danach
        const before = i === 0 || !/[A-Z0-9_]/.test(upSql[i - 1]);
        const after = i + kwLen === sql.length || !/[A-Z0-9_]/.test(upSql[i + kwLen]);
        if (before && after) {
          if (!last) return i;
          lastFoundIndex = i;
        }
      }
    }
  }
  return lastFoundIndex;
}

function appendToSql(text: string) {
  if (!sqlTarget.value || configForm[sqlTarget.value] === undefined) {
    sqlTarget.value = 'SQLSTATEMENT';
  }
  
  let currentVal = configForm[sqlTarget.value] || '';
  const upSql = currentVal.trim().toUpperCase();
  
  if (upSql.length === 0) {
    configForm[sqlTarget.value] = 'SELECT\n    ' + text.toUpperCase();
    return;
  }

  // Haupt-FROM suchen um davor einzufügen
  const fromIndex = findRootKeywordIndex(currentVal, 'FROM');

  if (fromIndex !== -1) {
    const beforeFrom = currentVal.substring(0, fromIndex).trimEnd();
    const afterFrom = currentVal.substring(fromIndex);
    
    let needsComma = false;
    // Prüfen ob vor dem FROM bereits Spalten stehen
    const selectIdx = findRootKeywordIndex(beforeFrom, 'SELECT', true);
    
    if (selectIdx !== -1) {
      const selectContent = beforeFrom.substring(selectIdx + 6).trim();
      if (selectContent.length > 0 && !selectContent.endsWith(',') && !selectContent.endsWith('*')) {
        needsComma = true;
      }
    }
    
    const sep = needsComma ? ',\n    ' : '\n    ';
    configForm[sqlTarget.value] = beforeFrom + sep + text.toUpperCase() + '\n' + afterFrom;
  } else {
    // Kein FROM gefunden -> Fallback auf Append
    const isStartedBySelect = upSql.startsWith('SELECT');
    const needsSeparator = isStartedBySelect && 
                           currentVal.trim().length > 6 && 
                           !currentVal.trim().endsWith(',') &&
                           !currentVal.trim().endsWith('*');

    const sep = needsSeparator ? ',\n    ' : ' ';
    configForm[sqlTarget.value] = currentVal.trimEnd() + sep + text.toUpperCase();
  }
}

function appendSelectedExplorerCols() {
  if (selectedExplorerCols.value.length === 0) return;
  
  const target = sqlTarget.value || 'SQLSTATEMENT_NATIVE';
  const isStartedEmpty = !configForm[target] || configForm[target].trim().length === 0;

  // Kommagetrennte Liste bauen, schön formatiert mit Umbrüchen
  let textToInsert = selectedExplorerCols.value.join(',\n    ');
  
  // Wenn wir leer anfangen, bauen wir ein komplettes Skelett inkl. FROM
  if (isStartedEmpty && selectedTable.value) {
    const table = selectedTable.value.toUpperCase();
    const alias = table.charAt(0);
    
    // Falls Aliase erwünscht sind (optional)
    const aliasedCols = selectedExplorerCols.value.map(c => `${alias}.${c.toUpperCase()}`).join(',\n    ');
    textToInsert = `SELECT\n    ${aliasedCols}\nFROM ${table} ${alias}`;
    
    // Den Text direkt setzen statt appendToSql zu nutzen (da wir das Skelett schon haben)
    configForm[target] = textToInsert;
  } else {
    // Einfach an die aktuelle Position anhängen
    appendToSql(textToInsert);
  }
  
  // Auswahl leeren
  selectedExplorerCols.value = [];
  $q.notify({ type: 'info', message: 'Felder wurden eingefügt', position: 'bottom-right', timeout: 1000 });
}

function onTableDragStart(event: DragEvent, tableName: string) {
  if (event.dataTransfer) {
    const payload = { text: tableName, type: 'table' };
    event.dataTransfer.setData('text/plain', tableName);
    event.dataTransfer.setData('payload', JSON.stringify(payload));
    event.dataTransfer.setData('source', 'db-explorer');
    event.dataTransfer.dropEffect = 'copy';
  }
}

function onColumnDragStart(event: DragEvent, colName: string, tableName: string) {
  if (event.dataTransfer) {
    const payload = { text: colName, table: tableName, type: 'column' };
    event.dataTransfer.setData('text/plain', colName);
    event.dataTransfer.setData('payload', JSON.stringify(payload));
    event.dataTransfer.setData('source', 'db-explorer');
    event.dataTransfer.dropEffect = 'copy';
  }
}

function onDragStart(event: DragEvent, text: string, tableSource?: string) {
  // Legacy-Funktion, falls noch irgendwo referenziert
  if (tableSource) onColumnDragStart(event, text, tableSource);
  else onTableDragStart(event, text);
}

  async function onDrop(event: DragEvent, field: 'SQLSTATEMENT' | 'DETAIL_SQL' | 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE' | 'SUMMENZEILE') {
  const source = event.dataTransfer?.getData('source');
  if (source !== 'db-explorer') return;

  const payloadStr = event.dataTransfer?.getData('payload');
  if (!payloadStr) return;
  
  const payload = JSON.parse(payloadStr);
  const text = (payload.text || '').trim();
  const isColumn = payload.type === 'column';
  const sourceTable = (payload.table || '').trim();

  const currentValue = configForm[field] || '';
  const isEmpty = !currentValue || currentValue.trim() === '';
  const upSql = currentValue.toUpperCase();

  // Hilfsfunktion für Alias-Generierung (sprechend)
  const generatePrefix = (t: string) => t.charAt(0).toUpperCase();

  // Hilfsfunktion: Findet Alias einer Tabelle im SQL
  const findAliasInSql = (sql: string, table: string): string | null => {
    const regex = new RegExp(`(?:FROM|JOIN)\\s+${table.toUpperCase()}(?:\\s+AS)?\\s+([A-Z0-9_]+)`, 'i');
    const m = sql.match(regex);
    return m ? m[1].toUpperCase() : null;
  };

  // Hilfsfunktion: Extrahiert alle Aliase für FK-Suche
  const getAllTablesInSql = (sql: string): {table: string, alias: string}[] => {
    const list: {table: string, alias: string}[] = [];
    const regex = /(?:FROM|JOIN)\s+([A-Z0-9_]+)(?:\s+AS)?\s+([A-Z0-9_]+)/gi;
    let m;
    while ((m = regex.exec(sql)) !== null) {
      list.push({ table: m[1].toUpperCase(), alias: m[2].toUpperCase() });
    }
    return list;
  };

  // Zentrale Funktion für die SQL-Injection
  const insertJoinAndCols = (baseSql: string, newCols: string, joinLine: string, tableForFrom?: string, aliasForFrom?: string) => {
    const fromIndex = findRootKeywordIndex(baseSql, 'FROM');
    
    // Falls kein FROM da ist -> Wir erstellen eines (statt JOIN)
    if (fromIndex === -1) {
      if (isEmpty || baseSql.trim().toUpperCase() === 'SELECT') {
        return `SELECT\n  ${newCols}\nFROM ${tableForFrom || '...'} ${aliasForFrom || ''}`;
      }
      const selectIdx = findRootKeywordIndex(baseSql, 'SELECT');
      if (selectIdx !== -1) {
        // Wir setzen das SELECT neu zusammen um sicherzugehen dass es vollständig ist
        const afterSelect = baseSql.substring(selectIdx + 6).trim();
        const sep = '\n  ';
        const afterSep = (afterSelect && !afterSelect.startsWith(',')) ? ',\n  ' : ' ';
        return `SELECT${sep}${newCols}${afterSelect ? afterSep + afterSelect : ''}\nFROM ${tableForFrom || '...'} ${aliasForFrom || ''}`;
      }
      // Weder FROM noch SELECT -> Wir bauen alles neu
      return `SELECT\n  ${newCols}${baseSql ? ',\n  ' + baseSql.trim() : ''}\nFROM ${tableForFrom || '...'} ${aliasForFrom || ''}\n${joinLine ? joinLine : ''}`;
    }

    const beforeFrom = baseSql.substring(0, fromIndex).trimEnd();
    const afterFrom = baseSql.substring(fromIndex);

    // Ende des FROM/JOIN Bereichs finden
    let endOfFrom = afterFrom.length;
    const keywords = ['WHERE', 'GROUP BY', 'ORDER BY', 'LIMIT'];
    keywords.forEach(kw => {
      const idx = findRootKeywordIndex(afterFrom, kw);
      if (idx !== -1 && idx < endOfFrom) endOfFrom = idx;
    });

    const fromJoinsPart = afterFrom.substring(0, endOfFrom).trimEnd();
    const restPart = afterFrom.substring(endOfFrom);
    
    const selectIdx = findRootKeywordIndex(beforeFrom, 'SELECT', true);
    let needsComma = selectIdx !== -1;
    if (selectIdx !== -1) {
       const selectContent = beforeFrom.substring(selectIdx + 6).trim();
       if (!selectContent || selectContent.endsWith(',') || selectContent.endsWith('*')) {
          needsComma = false;
       }
    }
    
    const sep = needsComma ? ',\n  ' : '\n  ';
    return `${beforeFrom}${sep}${newCols}\n${fromJoinsPart}\n${joinLine}${restPart}`;
  };

  let sourceTableFinal = sourceTable;
  if (isColumn && !sourceTableFinal) {
    // Versuchen die Tabelle anhand des Spaltennamens zu finden
    for (const tbl of dbTables.value) {
       try {
         const cr = await api.get(`/api/schema/columns/${tbl}`);
         const cList = (cr.data || []) as (string | {COLUMN_NAME: string})[];
         const cMatch = cList.find(c => (typeof c === 'string' ? c : (c as any).COLUMN_NAME).toUpperCase() === text.toUpperCase());
         if (cMatch) {
            sourceTableFinal = tbl;
            break;
         }
       } catch (e) { /* skip */ }
    }
  }

  const tableUpper = (isColumn ? sourceTableFinal : text).toUpperCase();
  if (!tableUpper) {
    if (text) configForm[field] += (isEmpty ? '' : ' ') + text.toUpperCase();
    return;
  }

  let prefix = findAliasInSql(currentValue, tableUpper);
  const isNewTable = !prefix;
  if (!prefix) prefix = generatePrefix(tableUpper);

  // Hilfsfunktion für die JOIN-Generierung
  const findBestJoin = async (newTable: string, newAlias: string, droppedCols: string[], sql: string) => {
    const existing = getAllTablesInSql(sql);
    let join = `JOIN ${newTable} ${newAlias} ON ${newAlias}.ID = ...`;
    
    for (const master of existing) {
      try {
        const resMaster = await api.get(`/api/schema/columns/${master.table}`);
        const masterCols = (resMaster.data || []) as string[] | {COLUMN_NAME: string}[];
        const masterColNames = masterCols.map(c => (typeof c === 'string' ? c : (c as any).COLUMN_NAME).toUpperCase());
        
        const fkToMaster = droppedCols.find(c => c.toUpperCase() === `ID_${master.table}`);
        const fkFromMaster = masterColNames.find(c => c.toUpperCase() === `ID_${newTable}`);

        if (fkToMaster) {
          join = `JOIN ${newTable} ${newAlias} ON ${newAlias}.${fkToMaster.toUpperCase()} = ${master.alias}.ID`;
          return join;
        } else if (fkFromMaster) {
          join = `JOIN ${newTable} ${newAlias} ON ${master.alias}.${fkFromMaster.toUpperCase()} = ${newAlias}.ID`;
          return join;
        }
      } catch (e) { /* ignore */ }
    }
    return join;
  };

  try {
    const res = await api.get(`/api/schema/columns/${tableUpper}`);
    const droppedTableColsRaw = (res.data || []) as (string | {COLUMN_NAME: string})[];
    const droppedTableCols = droppedTableColsRaw.map(c => typeof c === 'string' ? c : (c as any).COLUMN_NAME);

    if (!isColumn) {
      // GANZE TABELLE GEDROPPT
      const filteredCols = droppedTableCols.filter(c => !c.toUpperCase().startsWith('ID_') && c.toUpperCase() !== 'ID');
      const hasId = droppedTableCols.some(c => c.toUpperCase() === 'ID');

      const seen = new Set<string>();
      const aliasedColsList = filteredCols.map(c => {
         let name = c.toUpperCase();
         // Bereinigung von :1 Suffixen (oft bei Views)
         if (name.includes(':')) name = name.replace(/:/g, '_');
         
         if (seen.has(name)) {
            let i = 2;
            while (seen.has(`${name}_${i}`)) i++;
            const uniqueName = `${name}_${i}`;
            seen.add(uniqueName);
            return `${prefix}."${c.toUpperCase()}" AS "${uniqueName}"`;
         }
         seen.add(name);
         // Wenn der Name Sonderzeichen enthält, aliassieren und quotiert ausgeben
         if (c.includes(':')) return `${prefix}."${c.toUpperCase()}" AS "${name}"`;
         return `${prefix}.${c.toUpperCase()}`;
      });
      const aliasedCols = aliasedColsList.join(',\n  ');

      if (isEmpty) {
        configForm[field] = `SELECT\n  ${hasId ? prefix + '.ID,\n  ' : ''}${aliasedCols}\nFROM ${tableUpper} ${prefix}`;
      } else if (isNewTable) {
        const join = await findBestJoin(tableUpper, prefix, droppedTableCols, currentValue);
        configForm[field] = insertJoinAndCols(currentValue, aliasedCols, join, tableUpper, prefix);
      } else {
        configForm[field] = insertJoinAndCols(currentValue, aliasedCols, '', tableUpper, prefix);
      }
    } else {
      // SPALTE(N) GEDROPPT
      const colsArray = text.split(',').map(c => c.trim()).filter(c => c !== '');
      const aliasedCols = colsArray.map(c => {
         const upC = c.toUpperCase();
         return upC.includes('.') ? upC : `${prefix}.${upC}`;
      }).join(',\n  ');

      if (isEmpty) {
        configForm[field] = `SELECT\n  ${aliasedCols}\nFROM ${tableUpper} ${prefix}`;
      } else if (isNewTable) {
        const join = await findBestJoin(tableUpper, prefix, droppedTableCols, currentValue);
        configForm[field] = insertJoinAndCols(currentValue, aliasedCols, join, tableUpper, prefix);
        if (join.includes('=')) {
          $q.notify({ type: 'positive', message: `JOIN zu ${tableUpper} erstellt.`, icon: 'link' });
        } else {
          $q.notify({ type: 'info', message: 'Tabelle via JOIN hinzugefügt. Bitte ON-Bedingung prüfen!', icon: 'link' });
        }
      } else {
        // Tabelle ist schon da -> Wir suchen wo wir die Spalten einfügen
        const fromIndex = findRootKeywordIndex(currentValue, 'FROM');
        if (fromIndex !== -1) {
          const beforeFrom = currentValue.substring(0, fromIndex).trimEnd();
          const afterFrom = currentValue.substring(fromIndex);
          const selectIdx = findRootKeywordIndex(beforeFrom, 'SELECT', true);
          let needsComma = selectIdx !== -1;
          if (selectIdx !== -1) {
            const content = beforeFrom.substring(selectIdx + 6).trim();
            if (!content || content.endsWith(',') || content.endsWith('*')) needsComma = false;
          }
          const sep = needsComma ? ',\n  ' : '  ';
          configForm[field] = `${beforeFrom}${sep}${aliasedCols}\n${afterFrom}`;
        } else {
          configForm[field] += (isEmpty ? '' : ' ') + aliasedCols;
        }
      }
    }
    if (field === 'SUMMENZEILE') {
       tempSummenSql.value = configForm[field];
    }
  } catch (e) {
    console.error('Drop error:', e);
    $q.notify({ type: 'negative', message: 'Fehler beim SQL-Update.' });
  }
}

// SQL VORSCHAU LOGIK
const showPreview = ref(false);
const previewRows = ref<any[]>([]);
const previewCols = ref<any[]>([]);

const showTestParamsDialog = ref(false);
const testParamsList = ref<string[]>([]);
const testParamsLabels = reactive<Record<string, string>>({});
const testParamsValues = reactive<Record<string, any>>({});
const currentTestSql = ref('');

// FILTER BUILDER LOGIK
const showFilterBuilderDialog = ref(false);
const filterBuilderCols = ref<string[]>([]);
const filterBuilderSelected = ref<string[]>([]);
const filterBuilderTarget = ref<'SQLSTATEMENT' | 'DETAIL_SQL' | 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE'>('SQLSTATEMENT');
const filterBuilderAlias = ref('');
const filterBuilderOps = reactive<Record<string, string>>({});
const filterBuilderTypes = reactive<Record<string, string>>({});
const operatorOptions = ['=', '>=', '<=', '<>', 'BETWEEN', 'LIKE', 'IN'];
const filterBuilderTables = ref<{table: string, alias: string}[]>([]);
const filterBuilderSelectedTable = ref<{table: string, alias: string} | null>(null);

async function openFilterBuilder(field: 'SQLSTATEMENT' | 'DETAIL_SQL' | 'SUMMENZEILE') {
  try {
    const nativeField = (field === 'SUMMENZEILE' ? 'SUMMENZEILE' : field + '_NATIVE') as any;
    const sql = configForm[nativeField] || configForm[field];
    filterBuilderTarget.value = field as any; 

    if (!sql || sql.trim().length < 5) {
      $q.notify({ type: 'warning', message: 'Bitte erst ein SQL-Statement erstellen oder Tabelle wählen.' });
      return;
    }

    // 1. Alle Tabellen und Aliase aus dem SQL extrahieren (FROM und JOINs)
    const tables: {table: string, alias: string}[] = [];
    
    // Normalisierung: Zeilenumbrüche zu Leerzeichen
    const cleanSql = sql.replace(/\n/g, ' ').replace(/\s+/g, ' ');

    // A: Suche nach FROM
    const fromPartMatch = cleanSql.match(/FROM\s+([A-Z0-9_,.\s]+?)(?:\s+WHERE|\s+JOIN|\s+GROUP|\s+ORDER|\s+LIMIT|$)/i);
    if (fromPartMatch) {
      const rawTables = fromPartMatch[1].split(',');
      rawTables.forEach(tStr => {
        const parts = tStr.trim().split(/\s+/);
        if (parts[0]) {
          let table = parts[0].toUpperCase();
          let alias = '';
          if (parts.length > 1) {
            alias = parts[parts.length - 1].toUpperCase();
            if (alias === 'AS' && parts.length > 2) alias = parts[parts.length - 2].toUpperCase();
            if (alias === 'AS') alias = '';
          }
          tables.push({ table, alias });
        }
      });
    }
    
    // B: Suche nach JOINs
    const joinMatches = cleanSql.matchAll(/(?:LEFT\s+|RIGHT\s+|INNER\s+|CROSS\s+)?JOIN\s+([A-Z0-9_.]+)(?:\s+(?:AS\s+)?([A-Z0-9_]+))?/gi);
    for (const m of joinMatches) {
      const tbl = m[1].toUpperCase();
      const als = (m[2] || '').toUpperCase();
      const finalAls = als || tbl;
      if (!tables.find(t => t.table === tbl && t.alias === finalAls)) {
        tables.push({ table: tbl, alias: finalAls });
      }
    }

    if (tables.length === 0) {
      $q.notify({ type: 'warning', message: 'Keine Tabelle im SQL gefunden (FROM fehlt).' });
      return;
    }

    filterBuilderTables.value = tables;
    filterBuilderSelectedTable.value = tables[0];
    filterBuilderTarget.value = field;
    
    showFilterBuilderDialog.value = true;
    
    // Wir versuchen die Spalten im Hintergrund zu laden
    try {
      await loadFilterColumns();
    } catch (e) {
      console.error('Filter columns load failed:', e);
    }
  } catch (err) {
    console.error('openFilterBuilder FEHLER:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Vorbereiten des Filters.' });
  }
}

async function loadFilterColumns() {
  if (!filterBuilderSelectedTable.value) return;
  const { table, alias } = filterBuilderSelectedTable.value;
  
  try {
    $q.loading.show({ message: `Spalten für ${table} werden geladen...` });
    const res = await api.get(`/api/schema/columns/${table}`);
    const cols = (res.data || []) as string[];
    filterBuilderCols.value = cols;
    filterBuilderAlias.value = alias;
    
    // Operatoren und Typen vorbefüllen
    cols.forEach(col => {
      const ucCol = col.toUpperCase();
      if (!filterBuilderOps[col]) {
        if (ucCol.includes('DATUM')) {
          filterBuilderOps[col] = 'BETWEEN';
        } else {
          filterBuilderOps[col] = '=';
        }
      }
      if (!filterBuilderTypes[col]) {
        if (ucCol.includes('DATUM')) {
          filterBuilderTypes[col] = 'DATE';
        } else if (ucCol.includes('AKTIV') || ucCol.endsWith('_KZ')) {
          filterBuilderTypes[col] = 'BOOLEAN';
        } else if (ucCol.endsWith('_ID') || ucCol.includes('NUMMER') || ucCol.includes('KOSTEN') || ucCol.includes('BETRAG')) {
          filterBuilderTypes[col] = 'NUMBER';
        } else {
          filterBuilderTypes[col] = 'TEXT';
        }
      }
    });
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Spalten.' });
  } finally {
    $q.loading.hide();
  }
}

const testParamInputs = ref<any[]>([]);

watch(showTestParamsDialog, async (val) => {
  if (val) {
    await nextTick();
    setTimeout(() => {
      if (testParamInputs.value && testParamInputs.value.length > 0) {
        const el = testParamInputs.value[0].$el || testParamInputs.value[0];
        const input = el.querySelector('input') || el;
        input.focus();
      }
    }, 200);
  }
});

const availableFilterCols = computed(() => {
  const target = filterBuilderTarget.value;
  const sql = (configForm[target] || '').toUpperCase();
  return filterBuilderCols.value.filter(col => {
    const ucCol = col.toUpperCase();
    const regex = new RegExp('([A-Z0-9_]+\\.)?' + ucCol + '\\s*(=|BETWEEN|IN|>|<|LIKE)', 'i');
    return !regex.test(sql);
  });
});

function applyFilterBuilder() {
  // Filter-Builder wirkt auf das Feld, von dem aus er gestartet wurde (meist Native)
  const targetField = filterBuilderTarget.value;
  let sql = configForm[targetField] || "";
  
  if (!sql && filterBuilderAlias.value) {
    sql = `SELECT * FROM ${filterBuilderAlias.value}`;
  }
  
  const alias = filterBuilderAlias.value ? filterBuilderAlias.value + '.' : '';

  // Filter-Klausel bauen (und Duplikate im SQL vermeiden)
  const filters = filterBuilderSelected.value.map(col => {
    const ucCol = col.toUpperCase();
    const op = filterBuilderOps[col] || '=';
    const field = `${alias}${ucCol}`;

    const pType = filterBuilderTypes[col] || (ucCol.includes('DATUM') ? 'DATE' : 'TEXT');
    const paramName = `${ucCol};${pType}`;

    if (op === 'BETWEEN') {
      if (ucCol.includes('DATUM')) {
        return `${field} BETWEEN \`VON_${ucCol};DATE\` AND \`BIS_${ucCol};DATE\``;
      }
      return `${field} BETWEEN \`VON_${paramName}\` AND \`BIS_${paramName}\``;
    }
    
    if (op === 'IN') {
      return `${field} IN (\`${paramName}\`)`;
    }
    
    if (op === 'LIKE') {
      return `${field} LIKE \`${paramName}\``;
    }

    return `${field} ${op} \`${paramName}\``;
  }).filter(f => {
    // Robusterer Check: Prüfen ob GENAU diese Bedingung schon drin steht
    // (Grobe Prüfung um Doppelungen zu verhindern)
    const normalizedF = f.replace(/\s+/g, ' ').toUpperCase();
    const normalizedSql = sql.replace(/\s+/g, ' ').toUpperCase();
    return !normalizedSql.includes(normalizedF);
  });

  if (filters.length === 0) {
    showFilterBuilderDialog.value = false;
    $q.notify({ type: 'info', message: 'Gewählte Filter sind bereits im SQL vorhanden.' });
    return;
  }

  const whereClause = filters.join('\n  AND ');
  const upperSql = sql.toUpperCase();
  
  // Wir bestimmen das Ende der SELECT/FROM/JOIN Sektion (vor GROUP/ORDER)
  const groupIndex = findRootKeywordIndex(sql, 'GROUP BY');
  const orderIndex = findRootKeywordIndex(sql, 'ORDER BY');
  const limitIndex = findRootKeywordIndex(sql, 'LIMIT');
  
  let cutPos = sql.length;
  if (groupIndex !== -1) cutPos = groupIndex;
  if (orderIndex !== -1 && orderIndex < cutPos) cutPos = orderIndex;
  if (limitIndex !== -1 && limitIndex < cutPos) cutPos = limitIndex;

  let prefix = sql.substring(0, cutPos).trimEnd();
  const suffix = sql.substring(cutPos);
  
  // Prüfen ob im Prefix (wirklich nur im SQL Teil auf Root-Ebene) ein WHERE vorkommt
  if (findRootKeywordIndex(prefix, 'WHERE') !== -1) {
    // Einfach am Ende des Prefix ein AND anhängen
    sql = prefix + '\n  AND ' + whereClause + (suffix ? '\n' + suffix : '');
  } else {
    // WHERE neu anlegen
    sql = prefix + '\nWHERE ' + whereClause + (suffix ? '\n' + suffix : '');
  }
  
  configForm[targetField] = sql;
  $q.notify({ type: 'positive', message: 'Filter erfolgreich eingefügt.' });
  showFilterBuilderDialog.value = false;
}

function clearSqlFilters(field: string) {
  let sql = configForm[field] || '';
  if (!sql) return;

  // Root-Level WHERE finden
  const whereIdx = findRootKeywordIndex(sql, 'WHERE');
  if (whereIdx === -1) {
    $q.notify({ type: 'info', message: 'Keine WHERE-Klausel auf Root-Ebene gefunden.' });
    return;
  }

  // Suche Ende des WHERE-Blocks (vor GROUP, ORDER, LIMIT)
  const groupIdx = findRootKeywordIndex(sql, 'GROUP BY');
  const orderIdx = findRootKeywordIndex(sql, 'ORDER BY');
  const limitIdx = findRootKeywordIndex(sql, 'LIMIT');

  let cutEnd = sql.length;
  if (groupIdx !== -1) cutEnd = groupIdx;
  if (orderIdx !== -1 && (orderIdx < cutEnd || cutEnd === sql.length)) cutEnd = orderIdx;
  if (limitIdx !== -1 && (limitIdx < cutEnd || cutEnd === sql.length)) cutEnd = limitIdx;

  const prefix = sql.substring(0, whereIdx).trimEnd();
  const suffix = sql.substring(cutEnd).trimStart();

  configForm[field] = prefix + (suffix ? '\n' + suffix : '');
  $q.notify({ type: 'positive', message: 'WHERE-Klausel entfernt.' });
}

/**
 * SQL vom externen QueryBuilder (SqlManager) aufbereiten für den DB-Explorer:
 * - Backticks (`) und einfache Hochkommas (') entfernen
 * - Alles in GROSSBUCHSTABEN umwandeln
 */
function cleanupSqlForExplorer(field: 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE') {
  const sql = configForm[field];
  if (!sql || !sql.trim()) {
    $q.notify({ type: 'warning', message: 'Kein SQL zum Bereinigen vorhanden.' });
    return;
  }

  // Backticks und einfache Hochkommas entfernen, alles UPPERCASE
  configForm[field] = sql.replace(/[`']/g, '').toUpperCase();

  $q.notify({ 
    type: 'positive', 
    message: 'SQL bereinigt: Hochkommas entfernt, Großbuchstaben.',
    icon: 'text_format'
  });
}

async function magicWandNative(field: 'SQLSTATEMENT' | 'DETAIL_SQL' | 'SUMMENZEILE') {
  // Für Summenzeile generieren wir Aggregate (SUM und AVG) basierend auf dem Detail-SQL
  if (field === 'SUMMENZEILE') {
    const baseSql = configForm.DETAIL_SQL_NATIVE || configForm.SQLSTATEMENT_NATIVE || configForm.DETAIL_SQL || configForm.SQLSTATEMENT;
    if (!baseSql || baseSql.trim().length < 10) {
      $q.notify({ type: 'warning', message: 'Kein Basis-SQL (Detail oder Master) vorhanden.' });
      return;
    }

    const fromIdx = findRootKeywordIndex(baseSql, 'FROM');
    if (fromIdx === -1) {
      $q.notify({ type: 'warning', message: 'Basis-SQL konnte nicht analysiert werden (FROM nicht gefunden).' });
      return;
    }

    const selectPart = baseSql.substring(baseSql.toUpperCase().indexOf('SELECT') + 6, fromIdx);
    const columns = selectPart.split(',').map(c => c.trim()).filter(c => c.length > 0);

    const columnInfos = columns.map((col, idx) => {
      // Suche Alias: ... AS "NAME" oder ... "NAME"
      const aliasMatch = col.match(/\s+AS\s+["']?([A-Z0-9_]+)["']?$/i) || col.match(/\s+["']?([A-Z0-9_]+)["']?$/i);
      let alias = aliasMatch ? aliasMatch[1] : col.split('.').pop()?.replace(/[^A-Z0-9_]/gi, '');
      if (!alias || alias.toUpperCase() === 'SELECT') alias = `COL${idx+1}`;
      
      // Erweiterte Liste für numerische Felder (inkl. Eier-Größen)
      const isNumeric = /BETRAG|MENGE|PREIS|STUECK|ANZAHL|WERT|EURO|KG|GEWICHT|SUMME|TOTAL|NETTO|BRUTTO|MWST|VAL|VALUE|SMALL|MEDIUM|LARGE|XL|GEWICHT|ANZ/i.test(alias);
      
      // Falls nicht im Keyword-Match, prüfen wir ob es ein bekanntes Text/ID-Feld ist (Skip-Liste)
      const isSkip = /ID|DATUM|KZ|TEXT|NAME|BEZEICHNUNG|INFO|BEMERKUNG|TYPE|TYP|ART|TIMESTAMP|USER|BENUTZER/i.test(alias);
      
      // Wir stufen es als numerisch ein, wenn es ein Match hat ODER wenn es kein Skip-Feld und nicht die erste Spalte ist
      return { original: col, alias, isNumeric: isNumeric || (!isSkip && idx > 0), isFirst: idx === 0 };
    });

    const buildQuery = (label: string, func: string) => {
      const colSelects = columnInfos.map(ci => {
        if (ci.isFirst) return `'${label}' AS "${ci.alias}"`;
        if (ci.isNumeric) return `${func}("${ci.alias}") AS "${ci.alias}"`;
        return `NULL AS "${ci.alias}"`;
      });
      // Wir nutzen das Original-SQL als Subquery, damit Aliase bereits aufgelöst sind
      return `SELECT ${colSelects.join(', ')} FROM (${baseSql})`;
    };

    const sumQuery = buildQuery('SUMME', 'SUM');
    const avgQuery = buildQuery('DURCHSCHNITT', 'AVG');

    const result = `-- Automatisch generierte Summenzeilen\n${sumQuery}\nUNION ALL\n${avgQuery}`;
    configForm.SUMMENZEILE = result;
    tempSummenSql.value = result; // WICHTIG: Synchronisation mit dem Editor-Feld
    
    $q.notify({ type: 'positive', message: 'Summen- und Durchschnittszeilen generiert.', icon: 'auto_fix_high' });
    return;
  }

  // Ziel ist immer das interne Feld (ohne _NATIVE)
  const targetField = field as 'SQLSTATEMENT' | 'DETAIL_SQL';
  // Quelle ist bevorzugt das Native Feld, falls vorhanden
  const sourceField = (field + '_NATIVE') as 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE';
  const sourceSql = configForm[sourceField] || configForm[targetField];

  if (!sourceSql || sourceSql.trim().length < 10) {
    $q.notify({ type: 'warning', message: 'Das SQL ist zu kurz oder leer zum Konvertieren.' });
    return;
  }

  let sql = sourceSql.trim();
  
  // 1. Alle Tabellen und Aliase für die Spaltenerkennung sammeln
  const tables: Record<string, string> = {}; // Alias -> Table
  const fromMatch = sql.match(/FROM\s+([A-Z0-9_]+)(?:\s+AS\s+([A-Z0-9_]+)|\s+([A-Z0-9_]+))?/i);
  if (fromMatch) {
    const tbl = fromMatch[1].toUpperCase();
    const als = (fromMatch[2] || fromMatch[3] || tbl).toUpperCase();
    tables[als] = tbl;
  }
  const joinMatches = sql.matchAll(/(?:LEFT\s+|RIGHT\s+|INNER\s+)?JOIN\s+([A-Z0-9_]+)(?:\s+(?:AS\s+)?([A-Z0-9_]+))?/gi);
  for (const m of joinMatches) {
    const tbl = m[1].toUpperCase();
    const als = (m[2] || tbl).toUpperCase();
    tables[als] = tbl;
  }

  // 2. SELECT-Spalten mit Aliasen versehen (PRO-Modus)
  const fromIdx = findRootKeywordIndex(sql, 'FROM');
  if (fromIdx !== -1 && sql.toUpperCase().trim().startsWith('SELECT')) {
    const selectContent = sql.substring(sql.toUpperCase().indexOf('SELECT') + 6, fromIdx);
    const columns = selectContent.split(',');
    const transformedCols = columns.map(col => {
      let c = col.trim();
      if (!c || c === '*' || c.includes(' AS ') || c.includes('"')) return c;
      const parts = c.split('.');
      const pureField = parts[parts.length - 1].toUpperCase();
      return `${c} AS "${pureField}"`;
    });
    const newSelectContent = '\n    ' + transformedCols.join(',\n    ') + '\n';
    sql = sql.substring(0, sql.toUpperCase().indexOf('SELECT') + 6) + newSelectContent + sql.substring(fromIdx);
  }

  // 3. Filter-Ersetzungen (Smarte Parameter-Namen)
  // Datumserkennung (BETWEEN bevorzugt)
  sql = sql.replace(/([A-Z0-9_\.]+)\s+BETWEEN\s+'(\d{4}-\d{2}-\d{2})'\s+AND\s+'(\d{4}-\d{2}-\d{2})'/gi, (match, field) => {
    const pBase = field.split('.').pop().toUpperCase();
    return `${field} BETWEEN :VON_${pBase} AND :BIS_${pBase}`;
  });

  // Datumserkennung (Einzelwerte)
  sql = sql.replace(/([A-Z0-9_\.]+)\s*=\s*'(\d{4}-\d{2}-\d{2})'/gi, (match, field) => {
    const pBase = field.split('.').pop().toUpperCase();
    return `${field} = :${pBase}`;
  });

  // Zahlen (versuchen, ID-Werte zu verschonen, aber Rest zu Parametrisieren)
  sql = sql.replace(/([A-Z0-9_\.]+)\s*=\s*(\d+(\.\d+)?)/gi, (match, field, val) => {
    if (val.length === 4 && parseInt(val) > 1900 && parseInt(val) < 2100) return match; // Wahrscheinlich Jahreszahl
    const pBase = field.split('.').pop().toUpperCase();
    return `${field} = :${pBase}`;
  });

  // Texte in Anführungszeichen (außer wir haben sie gerade erst zu :DATUM gemacht)
  sql = sql.replace(/([A-Z0-9_\.]+)\s*=\s*'([^']+)'/gi, (match, field, val) => {
    if (val.startsWith(':')) return match;
    // 'de' und 'en' als feste Sprach-Konstanten verschonen
    if (val.toLowerCase() === 'de' || val.toLowerCase() === 'en') return match;
    
    const pBase = field.split('.').pop().toUpperCase();
    return `${field} = :${pBase}`;
  });

  // 4. Ergebnis ins Ziel (Vorschlag) schreiben
  configForm[targetField] = sql;
  $q.notify({ type: 'positive', message: `SQL-Vorschlag generiert. Sie können ihn nun prüfen und übernehmen.`, icon: 'auto_fix_high' });
}

async function runSqlPreview(sql: string, skipParamCheck = false) {
  if (!sql) return;
  
  // Parameter-Check
  if (!skipParamCheck) {
    const matches = sql.match(/§[^§]+§|`[^`]+`|:[A-Z0-9_]+/g) || [];
    const params = [...new Set(matches.map(m => {
      if (m.startsWith(':')) return m.substring(1).toUpperCase();
      return m.replace(/[§`]/g, '').toUpperCase();
    }))];
    
    if (params.length > 0) {
      testParamsList.value = params;
      currentTestSql.value = sql;
      
      // Intelligent Labels finden (z.B. H.DATUM = :PARAM oder H.DATUM BETWEEN :START AND :END)
      params.forEach(p => {
        // 1. Suche nach "FELD = :P"
        const eqRegex = new RegExp(`([A-Z0-9_\\.]+)\\s*=\\s*:${p}`, 'i');
        const eqMatch = sql.match(eqRegex);
        
        // 2. Suche nach "FELD BETWEEN :P AND ..."
        const btwStartRegex = new RegExp(`([A-Z0-9_\\.]+)\\s+BETWEEN\\s+:${p}`, 'i');
        const btwStartMatch = sql.match(btwStartRegex);
        
        // 3. Suche nach "FELD BETWEEN ... AND :P"
        const btwEndRegex = new RegExp(`([A-Z0-9_\\.]+)\\s+BETWEEN\\s+:[A-Z0-9_]+\\s+AND\\s+:${p}`, 'i');
        const btwEndMatch = sql.match(btwEndRegex);

        if (eqMatch && eqMatch[1]) {
           testParamsLabels[p] = eqMatch[1].toUpperCase();
        } else if (btwStartMatch && btwStartMatch[1]) {
           testParamsLabels[p] = btwStartMatch[1].toUpperCase() + ' (VON)';
        } else if (btwEndMatch && btwEndMatch[1]) {
           testParamsLabels[p] = btwEndMatch[1].toUpperCase() + ' (BIS)';
        } else {
           testParamsLabels[p] = p; // Fallback
        }

        if (testParamsValues[p] === undefined || testParamsValues[p] === '') {
           if (p.toUpperCase() === 'SPRACHE_KZ') {
              testParamsValues[p] = 'de';
           } else {
              testParamsValues[p] = p.toLowerCase().includes('datum') ? new Date().toISOString().split('T')[0] : '';
           }
        }
      });
      showTestParamsDialog.value = true;
      return;
    }
  }

  try {
    $q.loading.show({ message: 'SQL wird ausgeführt...' });
    const res = await api.post('/api/reports/preview', { 
      sql,
      params: testParamsValues 
    });
    const data = res.data.data || res.data || [];
    const serverCols = res.data.columns || [];
    
    previewRows.value = data;
    
    // Spalten dynamisch generieren (Priorität auf Server-Reihenfolge)
    if (data.length > 0) {
      let keys = serverCols.length > 0 ? [...serverCols] : Object.keys(data[0]);
      if (serverCols.length > 0) {
        Object.keys(data[0]).forEach(k => { if (!keys.includes(k)) keys.push(k); });
      }

      previewCols.value = keys.map(key => {
        const sampleVal = data[0][key];
        const isNum = typeof sampleVal === 'number';
        return {
          name: key,
          label: key.toUpperCase(),
          field: key,
          sortable: true,
          align: isNum ? 'right' : 'left',
          format: (val: any) => formatDynamicValue(val, key)
        };
      });
    } else {
      previewCols.value = [];
      $q.notify({ type: 'warning', message: 'Das SQL lieferte keine Daten zurÃ¼ck.' });
    }
    
    showPreview.value = true;
  } catch (err: any) {
    console.error('SQL Preview Error:', err);
    $q.notify({ 
      type: 'negative', 
      message: 'SQL Fehler', 
      caption: err.response?.data?.error || err.message 
    });
  } finally {
    $q.loading.hide();
  }
}

function onExplorerDrop(event: DragEvent) {
  event.preventDefault();
  const text = event.dataTransfer?.getData('text/plain')?.trim();
  if (!text) return;

  // Nur reagieren wenn es KEIN Drag aus dem Explorer selbst ist
  const source = event.dataTransfer?.getData('source');
  if (source === 'db-explorer') return;

  const fields: ('SQLSTATEMENT' | 'DETAIL_SQL' | 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE')[] = 
    ['SQLSTATEMENT', 'DETAIL_SQL', 'SQLSTATEMENT_NATIVE', 'DETAIL_SQL_NATIVE'];

  let found = false;
  fields.forEach(f => {
    let sql = configForm[f];
    if (!sql || !sql.includes(text)) return;

    // 1. Das Wort selbst entfernen (inklusive potenziellem Alias davor)
    const escapedText = text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    // Regex sucht nach optionalem Alias + dem Text
    const regex = new RegExp(`([A-Z0-9_]+\\.)?${escapedText}`, 'gi');
    sql = sql.replace(regex, '');

    // 2. Syntax-Bereinigung (hässliche Kommas und Leerzeichen entfernen)
    sql = sql.replace(/,\s*,/g, ','); // Doppelte Kommas
    sql = sql.replace(/SELECT\s*,/i, 'SELECT\n  '); // Komma direkt nach SELECT
    sql = sql.replace(/,\s+(FROM|WHERE|GROUP|ORDER|LIMIT|JOIN)/i, '\n$1'); // Komma vor Keywords
    sql = sql.replace(/\s\s+/g, ' '); // Mehrfache Leerzeichen
    
    // 3. Verwaiste Aliase am Ende von Zeilen oder vor Kommas (Sicherheitshalber)
    sql = sql.replace(/[A-Z0-9_]+\.\s*(?=,|\n|FROM|WHERE|JOIN|$)/gi, '');
    
    configForm[f] = sql.trim();
    found = true;
  });

  if (found) {
    $q.notify({ type: 'info', message: `Entfernt: ${text}`, icon: 'delete_sweep', position: 'top' });
  }
}

function onSqlDragStart(event: DragEvent) {
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move';
  }
}

interface Report {
  ID: number;
  id?: number;
  BESCHREIBUNG: string;
  beschreibung?: string;
  SQLSTATEMENT: string;
  sqlstatement?: string;
  KATEGORIE_KZ: string | { String: string };
  kategorie_kz?: string | { String: string };
  GRUPPEN_KZ: string | { String: string };
  gruppen_kz?: string | { String: string };
  ROOT_KZ: string | { String: string };
  root_kz?: string | { String: string };
  TYP_KZ: string | { String: string };
  typ_kz?: string | { String: string };
  TEMPLATE_NAME?: string | { String: string };
  PARAM_DEF?: string | { String: string; Valid: boolean };
  DETAIL_SQL?: string | { String: string };
  LINK_LOGIC?: string | { String: string };
  GROUP_FIELD?: string | { String: string };
  ROWS_PER_PAGE?: number | { Int64: number };
  PAGE_ORIENTATION?: string | { String: string };
  SHOW_DETAIL_GRID?: number | { Int64: number };
  SYSTEM_KZ: boolean | { Bool: boolean };
  SUMMENZEILE?: string | { String: string };
  IST_SUMMENZEILE?: number | { Int64: number };
}

// --- Shared Data ---
const reportRows = ref<Report[]>([]);

// --- Global Form State ---
const configForm = reactive({
  ID: 0,
  BESCHREIBUNG: '',
  SQLSTATEMENT: '',
  KATEGORIE_KZ: '',
  GRUPPEN_KZ: '',
  TYP_KZ: '',
  ROOT_KZ: 'x',
  TEMPLATE_NAME: '',
  PARAM_DEF: '',
  DETAIL_SQL: '',
  LINK_LOGIC: '',
  GROUP_FIELD: '',
  ROWS_PER_PAGE: 0,
  PAGE_ORIENTATION: 'P',
  SHOW_MASTER_GRID: false,
  SHOW_DETAIL_GRID: false,
  SYSTEM_KZ: false,
  SQLSTATEMENT_NATIVE: '',
  DETAIL_SQL_NATIVE: '',
  SUMMENZEILE: '',
  IST_SUMMENZEILE: false
});

// --- Gruppierung der Reports für die Konfigurations-Ansicht ---
const configExpandedStates = reactive<Record<string, boolean>>({});

const mgmtTreeRef = ref<any>(null);
const tempSummenSql = ref('');
watch(() => configForm.TYP_KZ, (newTyp) => {
  if (newTyp === 'T' || newTyp === 'S') {
    configForm.IST_SUMMENZEILE = true;
  }
});

watch(() => configForm.SUMMENZEILE, (val) => { tempSummenSql.value = val || ''; }, { immediate: true });
watch(tempSummenSql, (val) => { configForm.SUMMENZEILE = val; });

// Automatisches Targeting des SQL-Editors je nach Wizard-Schritt
watch(configStep, (newStep) => {
  if (newStep === 3) sqlTarget.value = 'SQLSTATEMENT_NATIVE';
  if (newStep === 4) sqlTarget.value = 'DETAIL_SQL_NATIVE';
  if (newStep === 7) sqlTarget.value = 'SUMMENZEILE';
});

function toggleConfigGroups(expand: boolean) {
  if (expand) {
    mgmtTreeRef.value?.expandAll();
  } else {
    mgmtTreeRef.value?.collapseAll();
  }
}

const groupedReports = computed(() => {
  const folders = reportRows.value.filter(r => getVal(getP(r, 'KATEGORIE_KZ')) === 'K');
  return folders.map(f => {
    const fGrp = getVal(getP(f, 'GRUPPEN_KZ'));
    const reportsInFolder = reportRows.value.filter(r => getVal(getP(r, 'KATEGORIE_KZ')) !== 'K' && getVal(getP(r, 'GRUPPEN_KZ')) === fGrp);
    if (configExpandedStates[fGrp] === undefined) {
      configExpandedStates[fGrp] = false;
    }
    return {
      label: getVal(getP(f, 'BESCHREIBUNG')),
      code: fGrp,
      reports: reportsInFolder
    };
  }).sort((a, b) => a.label.localeCompare(b.label));
});

const orphanReports = computed(() => {
  const folders = reportRows.value.filter(r => getVal(getP(r, 'KATEGORIE_KZ')) === 'K');
  const folderCodes = folders.map(r => getVal(getP(r, 'GRUPPEN_KZ')));
  return reportRows.value.filter(r => {
    if (getVal(getP(r, 'KATEGORIE_KZ')) === 'K') return true;
    return !folderCodes.includes(getVal(getP(r, 'GRUPPEN_KZ')));
  });
});

function onMasterRowClick(_evt: any, row: any) {
  selectedMasterRows.value = [row];
}

const isAdmin = computed(() => sessionStore.profile_kz === 'A');

// Removed duplicate computed/watch
const isSystemLocked = computed(() => {
  const val = typeof configForm.SYSTEM_KZ === 'boolean' ? configForm.SYSTEM_KZ : (configForm.SYSTEM_KZ as any)?.Bool || false;
  return val && !isAdmin.value;
});

// --- Backtick Definition State ---
const showDefDialog = ref(false);
const currentDefinitions = ref<any[]>([]);
const currentReportId = ref<number | null>(null);

function syncParamsFromSql() {
  const sql = (configForm.SQLSTATEMENT_NATIVE || configForm.SQLSTATEMENT || '') + ' ' + 
              (configForm.DETAIL_SQL_NATIVE || configForm.DETAIL_SQL || '') + ' ' + 
              (configForm.SUMMENZEILE || '');
              
  const btMatches = sql.match(/`([^`]+)`/g) || [];
  const percentMatches = sql.match(/%([^%]+)%/g) || [];
  
  const terms = [...new Set([
    ...btMatches.map(m => m.replace(/`/g, '').split(';')[0]),
    ...percentMatches.map(m => m.replace(/%/g, '').split(';')[0])
  ])];
  
  if (terms.length === 0) {
    currentDefinitions.value = [];
    return;
  }

  let existingDef: any = {};
  if (configForm.PARAM_DEF) {
    try { 
      const parsed = JSON.parse(configForm.PARAM_DEF); 
      existingDef = typeof parsed === 'object' ? parsed : {};
    } catch(e) {}
  }

  // Automatischer Scan nach Typen (sowohl Backticks als auch Legacy-Format)
  const allMatches = [...btMatches.map(m => m.replace(/`/g, '')), ...percentMatches.map(m => m.replace(/%/g, ''))];
  allMatches.forEach(raw => {
    const parts = raw.split(';');
    const term = parts[0];
    if (parts.length > 1 && !existingDef[term]) {
       existingDef[term] = { label: term, type: parts[1].toUpperCase() };
    }
  });

  currentDefinitions.value = terms.map(t => ({
    term: t,
    label: existingDef[t]?.label || t,
    type: existingDef[t]?.type || (t.toUpperCase().includes('DATUM') ? 'DATE' : (t.toUpperCase().includes('FILTER') || t.toUpperCase().includes('MENGE') ? 'NUMBER' : 'TEXT'))
  }));
}

async function saveDefinitions() {
  if (!currentReportId.value) return;
  
  const defMap: Record<string, any> = {};
  currentDefinitions.value.forEach(d => {
    defMap[d.term] = { label: d.label, type: d.type };
  });
  
  const report = reportRows.value.find(r => getP(r, 'ID') === currentReportId.value);
  if (!report) return;
  
  try {
    const payload = {
      BESCHREIBUNG: getVal(getP(report, 'BESCHREIBUNG')),
      SQLSTATEMENT: getVal(getP(report, 'SQLSTATEMENT')),
      KATEGORIE_KZ: getVal(getP(report, 'KATEGORIE_KZ')),
      GRUPPEN_KZ: getVal(getP(report, 'GRUPPEN_KZ')),
      TYP_KZ: getVal(getP(report, 'TYP_KZ')),
      TEMPLATE_NAME: getVal(getP(report, 'TEMPLATE_NAME')),
      PARAM_DEF: JSON.stringify(defMap)
    };
    
    await api.put(`/api/reports/${currentReportId.value}`, payload);
    $q.notify({ type: 'positive', message: 'Parameter-Definitionen gespeichert.' });
    
    // Update local data
    (report as any).PARAM_DEF = { String: payload.PARAM_DEF, Valid: true };
    showDefDialog.value = false;
    
    // Trigger the report again
    void onReportSelected([report]);
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern der Definitionen.' });
  }
}

function expandAll() {
  if (!masterData.value) return;
  const keyField = currentReport.value?.GROUP_FIELD || 'ID';
  expandedRowKeys.value = masterData.value.map((r: any) => {
    return r[keyField] || r[keyField.toUpperCase()] || r.ID || r.id;
  });
}

function collapseAll() {
  expandedRowKeys.value = [];
}

function filterDetailsForMaster(masterRow: any) {
  if (!masterRow || !detailData.value) return [];
  
  // ID suchen (aus ID, id, Id oder dem Gruppen-Feld - Case-Insensitive)
  let masterId = getP(masterRow, 'ID');
  if (masterId === undefined) masterId = masterRow.Id;
  
  if (masterId === undefined) masterId = masterRow._MASTER_ID_ || masterRow._master_id_;
  
  if (masterId === undefined && currentReport.value?.GROUP_FIELD) {
    const gf = currentReport.value?.GROUP_FIELD;
    if (gf) {
      masterId = masterRow[gf] || masterRow[gf.toUpperCase()] || masterRow[gf.toLowerCase()];
    }
  }
  
  if (masterId === undefined) return [];

  return detailData.value.filter((d: any) => {
    // Vorrangig nach unserem neuen sicheren Stempel suchen
    if (d._MASTER_ID_ !== undefined) {
      return d._MASTER_ID_ == masterId;
    }
    // Fallback für alte Daten/andere Typen
    return Object.values(d).some(val => val == masterId);
  });
}

async function loadReports() {
  console.log('Starte API-Request zu /api/reports...');
  try {
    // Cache-Buster um sicherzugehen dass wir die aktuellsten DB-Stände erhalten
    const res = await api.get(`/api/reports?t=${Date.now()}`);
    console.log('API-Antwort /api/reports:', res.data);
    reportRows.value = res.data || [];
    $q.notify({ type: 'info', message: `${reportRows.value.length} Reports geladen`, timeout: 1000 });
    // Hier wurde früher expandedKeys.value = [] aufgerufen - entfernt um Zustand zu behalten
  } catch (err) {
    console.error('Fehler beim Laden der Reports:', err);
    $q.notify({ type: 'negative', message: 'Fehler beim Laden der Report-Liste' });
  }
}

// --- TAB 1: ANZEIGE LOGIC ---
// --- Tree / Selection Logic ---
const filterText = ref('');
const selectedKey = ref<string | null>(null);
const expandedKeys = ref<string[]>([]);
const expandedKeysMgmt = ref<string[]>([]);
const treeRef = ref<InstanceType<typeof QTree> | null>(null);
const allExpanded = ref(false);

function clearWhere(field: 'SQLSTATEMENT' | 'DETAIL_SQL') {
  const nativeField = (field + '_NATIVE') as 'SQLSTATEMENT_NATIVE' | 'DETAIL_SQL_NATIVE';
  const isNative = !!configForm[nativeField];
  const targetKey = isNative ? nativeField : field;
  const sql = configForm[targetKey] || '';

  if (!sql) return;
  
  // Regex sucht nach WHERE (case-insensitive) mit beliebigem Whitespace davor
  const whereMatch = sql.match(/\s+WHERE\s+/i);
  if (!whereMatch || whereMatch.index === undefined) {
    // Spezialfall: WHERE am absoluten Anfang des Strings (unwahrscheinlich bei SELECT)
    if (!sql.trim().toUpperCase().startsWith('WHERE')) return;
  }
  
  const whereIdx = whereMatch ? whereMatch.index : 0;
  const upper = sql.toUpperCase();
  
  // Wir suchen nach dem Ende der WHERE-Klausel (nächster Keyword-Block)
  const markers = [' ORDER BY ', ' GROUP BY ', ' LIMIT ', ' OFFSET '];
  let nextIdx = sql.length;
  
  for (const m of markers) {
    // Suche ab der Position nach dem WHERE
    const idx = upper.indexOf(m, whereIdx + 5);
    if (idx !== -1 && idx < nextIdx) {
      nextIdx = idx;
    }
  }
  
  const before = sql.substring(0, whereIdx);
  const after = sql.substring(nextIdx);
  
  configForm[targetKey] = (before + ' ' + after).trim();
  $q.notify({ type: 'info', message: 'WHERE-Klausel wurde entfernt.' });
}

const getVal = (v: any): string => {
  if (v === null || v === undefined) return '';
  if (typeof v === 'object' && 'String' in v) return String(v.String).trim();
  if (v instanceof Uint8Array || (typeof v === 'object' && v.constructor && v.constructor.name === 'Uint8Array')) {
    return new TextDecoder().decode(v).trim();
  }
  return String(v).trim();
};

const getP = (obj: any, key: string): any => {
  if (!obj) return undefined;
  if (obj[key] !== undefined) return obj[key];
  const lk = key.toLowerCase();
  if (obj[lk] !== undefined) return obj[lk];
  const uk = key.toUpperCase();
  if (obj[uk] !== undefined) return obj[uk];
  return undefined;
};

const getBool = (v: any): boolean => {
  if (v === null || v === undefined) return false;
  if (typeof v === 'boolean') return v;
  if (typeof v === 'object' && 'Bool' in v) return !!v.Bool;
  return !!v;
};

const folderOptions = computed(() => {
  return reportRows.value
    .filter(r => getVal(getP(r, 'KATEGORIE_KZ')) === 'K')
    .map(f => {
      const rootKz = getVal(getP(f, 'ROOT_KZ'));
      const rootDef = [
        { label: 'Listen', kz: 'L' },
        { label: 'Temp', kz: 'T' },
        { label: 'Grid', kz: 'M' }
      ].find(d => d.kz === rootKz);
      
      const rootLabel = rootDef ? `[${rootDef.label}] ` : '';
      
      return {
        label: rootLabel + getVal(getP(f, 'BESCHREIBUNG')),
        value: getVal(getP(f, 'GRUPPEN_KZ')),
        root: rootKz
      };
    });
});

const filteredFolderOptions = ref<{label: string, value: string}[]>([]);

// Helfer zum sauberen Speichern eines Berichts (löscht Kleinteile und Duplikate für das Backend)
async function saveReportEntry(report: any, updates: any = {}) {
  const p: any = {};
  const keys = [
    'BESCHREIBUNG', 'SQLSTATEMENT', 'KATEGORIE_KZ', 'GRUPPEN_KZ', 'TYP_KZ',
    'TEMPLATE_NAME', 'PARAM_DEF', 'DETAIL_SQL', 'LINK_LOGIC', 'GROUP_FIELD',
    'ROWS_PER_PAGE', 'PAGE_ORIENTATION', 'SHOW_MASTER_GRID', 'SHOW_DETAIL_GRID',
    'SYSTEM_KZ', 'SQLSTATEMENT_NATIVE', 'DETAIL_SQL_NATIVE', 'ROOT_KZ',
    'SUMMENZEILE', 'IST_SUMMENZEILE'
  ];
  
  keys.forEach(k => {
    // Wenn in updates vorhanden, nehmen wir das, sonst aus dem report
    const val = updates[k] !== undefined ? updates[k] : getP(report, k);
    p[k] = getVal(val);
    
    // Normalisierung für Schlüssel-Felder auf Großschreibung
    if (['GRUPPEN_KZ', 'ROOT_KZ', 'KATEGORIE_KZ', 'TYP_KZ'].includes(k)) {
       p[k] = p[k].toUpperCase();
    }
    
    // Konvertierung für numerische Felder
    if (['ROWS_PER_PAGE', 'SHOW_MASTER_GRID', 'SHOW_DETAIL_GRID', 'SYSTEM_KZ', 'IST_SUMMENZEILE'].includes(k)) {
       p[k] = Number(p[k] || 0);
    }
  });

  const reportId = getP(report, 'ID');
  await api.put(`/api/reports/${reportId}`, p);
}

function filterFolders(val: string, update: (callback: () => void) => void) {
  update(() => {
    // Filtern nach der aktuell gewählten Root im Formular
    const currentRoot = configForm.ROOT_KZ;
    let options = folderOptions.value;
    
    if (currentRoot && currentRoot !== 'x') {
       options = options.filter(o => o.root === currentRoot);
    }

    if (val === '') {
      filteredFolderOptions.value = options;
    } else {
      const needle = val.toLowerCase();
      filteredFolderOptions.value = options.filter(
        v => v.label.toLowerCase().indexOf(needle) > -1
      );
    }
  });
}

function onFolderSelected(val: string) {
  if (!val) return;
  // Suche den Ordner in der Liste
  const folder = folderOptions.value.find(o => o.value === val);
  if (folder && folder.root) {
    // Synchronisiere die Root des Berichts mit der Root des Ordners
    configForm.ROOT_KZ = folder.root;
  }
}

// UI-Synchronisation: Felder leeren wenn nicht relevant
watch(() => [configForm.KATEGORIE_KZ, configForm.TYP_KZ], ([newKat, newTyp]) => {
  if (newKat !== 'F' && newTyp !== 'T') {
    configForm.TEMPLATE_NAME = '';
  }
  if (newKat === 'K') {
    configForm.SQLSTATEMENT = 'FOLDER'; // Dummy für NOT NULL
    configForm.TEMPLATE_NAME = '';
  } else if (configForm.SQLSTATEMENT === 'FOLDER') {
    configForm.SQLSTATEMENT = '';
  }
});


const treeNodes = computed(() => {
  const allReports = reportRows.value;
  console.log(`treeNodes: Building tree for ${allReports.length} reports...`);
  
  const rootDefinitions = [
    { 
      label: 'Einfache Listen', 
      kz: 'L', 
      icon: 'list_alt', 
      color: 'blue', 
      tooltip: 'Standard Listenansichten' 
    },
    { 
      label: 'Master Detail(Template)', 
      kz: 'T', 
      icon: 'description', 
      color: 'orange', 
      tooltip: 'Berichte mit HTML-Template (Master-Detail)' 
    },
    { 
      label: 'Master Detail (Gitter)', 
      kz: 'M', 
      icon: 'grid_view', 
      color: 'green', 
      tooltip: 'Berichte mit Unter-Gitter (Master-Detail)' 
    },
    { 
      label: 'Gruppierte Reports', 
      kz: 'G', 
      icon: 'reorder', 
      color: 'purple', 
      tooltip: 'Reports mit Gruppierungs-Logik' 
    },
    { 
      label: 'Sonstige / System', 
      kz: 'x', 
      icon: 'settings_suggest', 
      color: 'grey-7', 
      tooltip: 'System- oder Testberichte' 
    }
  ];

  const createReportNode = (r: Report) => {
    const type = getVal(getP(r, 'TYP_KZ'));
    let typeDesc = 'Bericht';
    if (type === 'S') typeDesc = 'Einfache Liste';
    else if (type === 'T') typeDesc = 'Master-Detail (Template)';
    else if (type === 'M') typeDesc = 'Master-Detail (Gitter)';

    return {
      label: getVal(getP(r, 'BESCHREIBUNG')),
      key: String(getP(r, 'ID') || (r as any).id),
      icon: 'description',
      iconColor: 'primary',
      tooltip: `Typ ${type}: ${typeDesc}`,
      data: r
    };
  };

  const buildTreeForRoot = (rootKz: string) => {
    const rkzUpper = rootKz.toUpperCase();
    const folders = allReports.filter(r => 
      getVal(getP(r, 'KATEGORIE_KZ')).toUpperCase() === 'K' && getVal(getP(r, 'ROOT_KZ')).toUpperCase() === rkzUpper
    );
    const reports = allReports.filter(r => 
      getVal(getP(r, 'KATEGORIE_KZ')).toUpperCase() !== 'K' && getVal(getP(r, 'ROOT_KZ')).toUpperCase() === rkzUpper
    );

    const folderNodes = folders.map(f => {
      const fGrp = getVal(getP(f, 'GRUPPEN_KZ')).toUpperCase();
      const children = reports
        .filter(r => getVal(getP(r, 'GRUPPEN_KZ')).toUpperCase() === fGrp && fGrp !== '')
        .map(createReportNode);

      return {
        label: getVal(getP(f, 'BESCHREIBUNG')),
        key: `f-${getP(f, 'ID') || (f as any).id}`,
        icon: 'folder',
        iconColor: 'amber-8',
        children: children.length > 0 ? children : undefined,
        data: f
      };
    });

    const orphanedReports = reports.filter(r => {
      const rGrp = getVal(getP(r, 'GRUPPEN_KZ')).toUpperCase();
      return !folders.some(f => getVal(getP(f, 'GRUPPEN_KZ')).toUpperCase() === rGrp && rGrp !== '');
    });

    return [...folderNodes, ...orphanedReports.map(createReportNode)];
  };

  const finalNodes = rootDefinitions.map(root => {
    const children = buildTreeForRoot(root.kz);
    return {
      label: root.label,
      key: `root-${root.kz}`,
      icon: root.icon,
      iconColor: root.color,
      tooltip: root.tooltip,
      children: children.length > 0 ? children : undefined,
      selectable: false,
      data: { ROOT_KZ: root.kz, KATEGORIE_KZ: 'ROOT' }
    };
  });

  const falloutReports = allReports.filter(r => {
    const rkz = getVal(getP(r, 'ROOT_KZ')).toUpperCase();
    return !rootDefinitions.some(rd => rd.kz.toUpperCase() === rkz);
  });

  if (falloutReports.length > 0) {
    const folders = falloutReports.filter(r => getVal(getP(r, 'KATEGORIE_KZ')) === 'K');
    const reports = falloutReports.filter(r => getVal(getP(r, 'KATEGORIE_KZ')) !== 'K');
    
    const falloutNodes = folders.map(f => {
      const fGrp = getVal(getP(f, 'GRUPPEN_KZ')).toUpperCase();
      const children = reports
        .filter(r => getVal(getP(r, 'GRUPPEN_KZ')).toUpperCase() === fGrp && fGrp !== '')
        .map(createReportNode);

      return {
        label: getVal(getP(f, 'BESCHREIBUNG')),
        key: `f-${getP(f, 'ID') || (f as any).id}`,
        icon: 'folder',
        iconColor: 'grey-5',
        children: children.length > 0 ? children : undefined,
        data: f
      };
    });

    const orphans = reports.filter(r => {
      const rGrp = getVal(getP(r, 'GRUPPEN_KZ')).toUpperCase();
      return !folders.some(f => getVal(getP(f, 'GRUPPEN_KZ')).toUpperCase() === rGrp && rGrp !== '');
    });

    finalNodes.push({
      label: 'Nicht zugeordnete Reports',
      key: 'fallout',
      icon: 'category',
      iconColor: 'grey-4',
      children: [...falloutNodes, ...orphans.map(createReportNode)],
      selectable: false
    });
  }

  return finalNodes;
});

function toggleExpandAll() {
  allExpanded.value = !allExpanded.value;
  if (allExpanded.value) {
    treeRef.value?.expandAll();
  } else {
    treeRef.value?.collapseAll();
  }
}



const filteredDetailData = computed(() => {
  // Wenn nichts ausgewählt ist, zeigen wir zur Sicherheit gar nichts an 
  // (verhindert das Laden von 9000 Zeilen Kraut und Rüben)
  if (!selectedMasterRows.value || selectedMasterRows.value.length === 0) return [];
  
  return filterDetailsForMaster(selectedMasterRows.value[0]);
});
// Removed duplicate refs

const reportPagination = computed(() => {
  const rpp = resData.value?.rows_per_page;
  return (rpp && rpp > 0) ? rpp : 15;
});

const selectedReport = computed(() => {
  if (!selectedKey.value) return [];
  const r = reportRows.value.find(row => String(getP(row, 'ID')) === selectedKey.value);
  return r ? [r] : [];
});

const currentReport = computed(() => selectedReport.value[0] || null);

function isBool(p: { type?: string; label?: string } | null) {
  if (!p) return false;
  const t = String(p.type || '').toUpperCase();
  const l = String(p.label || '').toLowerCase();
  return t === 'BOOL' || t === 'CHECKBOX' || t === 'BOOLEAN' || l.startsWith('b_') || l.endsWith('_kz') || l.endsWith('_bool');
}

function isChoice(p: { type?: string } | null) {
  if (!p || !p.type) return false;
  return p.type.toUpperCase().startsWith('CHOICE');
}

function isStammParam(p: any) {
  if (!p || !p.label) return false;
  const lbl = p.label.toUpperCase();
  return ['HERDENNUMMER', 'SILONUMMER', 'STALLNUMMER', 'LAGERNUMMER', 'ID_HERDE', 'ID_SILO', 'ID_STALL', 'ID_LAGER'].includes(lbl);
}

function getStammKey(p: any) {
  const lbl = p.label.toUpperCase();
  if (lbl.includes('HERDE')) return 'HERDEN';
  if (lbl.includes('SILO')) return 'SILO';
  if (lbl.includes('STALL')) return 'STALL';
  if (lbl.includes('LAGER')) return 'LAGER';
  return 'HERDEN';
}

function getStammFiltered(p: any) {
  const key = getStammKey(p) as keyof typeof stammFilter;
  return stammFilter[key];
}

function filterStamm(val: string, update: Function, p: any) {
  const key = getStammKey(p) as keyof typeof stammOptions;
  if (val === '') {
    update(() => {
      stammFilter[key] = stammOptions[key];
    });
    return;
  }

  update(() => {
    const needle = val.toLowerCase();
    stammFilter[key] = stammOptions[key].filter(
      v => v.label.toLowerCase().indexOf(needle) > -1
    );
  });
}

async function loadAllStammData() {
  try {
    // 1. Herden
    const hRes = await api.get('/api/herden');
    stammOptions.HERDEN = hRes.data.map((h: any) => ({
      label: `${h.HERDENNUMMER} - ${h.BEZEICHNUNG || h.INTERNE_BEZEICHNUNG || ''}`,
      value: h.HERDENNUMMER
    }));

    // 2. Silo
    const sRes = await api.get('/api/silo');
    stammOptions.SILO = sRes.data.map((s: any) => ({
      label: `${s.SILONUMMER} - ${s.BEZEICHNUNG || ''}`,
      value: s.SILONUMMER
    }));

    // 3. Stall
    const stRes = await api.get('/api/stall');
    stammOptions.STALL = stRes.data.map((s: any) => ({
      label: `${s.STALLNUMMER} - ${s.BEZEICHNUNG || ''}`,
      value: s.STALLNUMMER
    }));

    // 4. Lager
    const lRes = await api.get('/api/eilager');
    stammOptions.LAGER = lRes.data.map((l: any) => ({
      label: `${l.LAGERNUMMER} - ${l.BEZEICHNUNG || ''}`,
      value: l.LAGERNUMMER
    }));

    // Filter initial befüllen
    Object.keys(stammOptions).forEach(k => {
      const key = k as keyof typeof stammOptions;
      stammFilter[key] = stammOptions[key];
    });
  } catch (err) {
    console.error('Fehler beim Laden der Stammdaten:', err);
  }
}

function getChoiceOptions(p: { type?: string } | null) {
  if (!p) return [];
  const t = String(p.type || '');
  const match = t.match(/CHOICE\((.*)\)/i);
  if (!match) return [];
  return match[1].split(',').map(pair => {
    const parts = pair.split(':');
    const label = parts[0]?.trim() || '';
    const value = parts[1]?.trim() || label;
    return { label, value };
  });
}

function formatDynamicValue(val: any, key: string) {
  if (val === null || val === undefined) return '';
  const upperKey = key.toUpperCase();
  const lang = sessionStore.selectedLanguage || 'de';
  const locale = lang === 'de' ? 'de-DE' : 'en-US';

  // 1. Datumserkennung (ISO Strings: YYYY-MM-DD...)
  if (typeof val === 'string' && val.length >= 8) {
    if (/^\d{4}-\d{1,2}-\d{1,2}/.test(val)) {
      // Wenn es bereits ein ISO-Datum ist, belassen wir es für die Grid-Anzeige so (wie gewünscht)
      return val;
    }
  }

  // 2. Zahlenformatierung
  if (typeof val === 'number') {
    const isIdOrCode = upperKey.includes('ID') || upperKey.includes('CODE') || upperKey.includes('NR');
    if (isIdOrCode) return val.toString();

    const isCurrency = upperKey.includes('EURO') || upperKey.includes('NETTO') || upperKey.includes('PREIS') || upperKey.includes('KOSTEN') || upperKey.includes('ERTRAG') || upperKey.includes('BETRAG') || upperKey.includes('SUMME');
    if (isCurrency) {
      return new Intl.NumberFormat(locale, { style: 'currency', currency: 'EUR' }).format(val);
    }
    
    // Ganzzahlen (z.B. Herdennummer, Stück) ohne Dezimalstellen
    if (Number.isInteger(val)) {
      return new Intl.NumberFormat(locale).format(val);
    }
    
    return new Intl.NumberFormat(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val);
  }

  return val;
}

function getAutoCols(data: any[], forcedCols?: string[]) {
  if (!data || data.length === 0) {
    if (forcedCols && forcedCols.length > 0) {
      return forcedCols.map(c => ({ 
        name: c, 
        label: c.replace(/_/g, ' ').toUpperCase(), 
        field: c, 
        sortable: true, 
        align: 'left' 
      }));
    }
    return [];
  }
  
  const first = data[0];
  if (!first) return [];
  
  let keys: string[] = [];
  
  // Wenn forcedCols (vom Server in SQL-Reihenfolge) da sind, nehmen wir diese als Basis
  if (forcedCols && forcedCols.length > 0) {
    keys = [...forcedCols];
    // Falls im Datensatz noch Felder sind, die NICHT in forcedCols stehen, hängen wir sie an
    Object.keys(first).forEach(k => {
      if (!keys.includes(k)) keys.push(k);
    });
  } else {
    // Fallback: Einfach alle Keys
    keys = Object.keys(first);
  }
  
  return keys
    .filter((key: string) => !key.startsWith('_') && key.toUpperCase() !== 'ID') 
    .map(key => ({
      name: key,
      label: key.replace(/_/g, ' ').toUpperCase(),
      field: key,
      align: typeof first[key] === 'number' ? 'right' : 'left',
      sortable: true,
      format: (v: any) => formatDynamicValue(v, key)
    }));
}

function onTreeNodeSelected(key: string | null) {

  if (!key || key.startsWith('f-') || key === 'other') {
    // If folder or dummy root selected, ignore or reset
    if (key?.startsWith('f-') || key === 'other') {
       resultData.value = [];
       resultColumns.value = [];
       currentReportLabel.value = 'Gruppe gewählt';
    }
    return;
  }
  
  // The computed selectedReport will update automatically.
  // Trigger loading:
  const report = reportRows.value.find(r => String(getP(r, 'ID')) === key);
  if (report) {
    void onReportSelected([report]);
  }
}


const showFilterDialog = ref(false);
const showQuickParamEdit = ref(false);
const filterParams = ref<{ label: string; type?: string }[]>([]);
const filterValues = reactive<Record<string, any>>({});
const dialogSQLPreview = ref('');

// STAMMDATEN OPTIONEN
const stammOptions = reactive({
  HERDEN: [] as any[],
  SILO: [] as any[],
  STALL: [] as any[],
  LAGER: [] as any[]
});

const stammFilter = reactive({
  HERDEN: [] as any[],
  SILO: [] as any[],
  STALL: [] as any[],
  LAGER: [] as any[]
});

const updatePreview = debounce(async () => {
  if (selectedReport.value.length === 0) return;
  try {
    const sanitizedParams: Record<string, any> = {};
    Object.keys(filterValues).forEach(key => {
      sanitizedParams[key.split(';')[0]] = filterValues[key];
    });
    
    const rId = getP(selectedReport.value[0], 'ID');
    const res = await api.post(`/api/reports/preview/${rId}`, { params: sanitizedParams });
    dialogSQLPreview.value = (res.data as { sql: string }).sql;
  } catch (_err) {
    console.warn('Vorschau-Fehler', _err);
  }
}, 300);

watch(filterValues, () => {
  if (showFilterDialog.value) void updatePreview();
}, { deep: true });

async function saveQuickParams() {
  if (!currentReportId.value) return;
  
  // Aktuellen Bericht finden
  const report = reportRows.value.find(r => getP(r, 'ID') === currentReportId.value);
  if (!report) return;

  // Definitionen aus dem Dialog in ein JSON-Objekt mappen
  const defMap: Record<string, any> = {};
  currentDefinitions.value.forEach(d => {
    defMap[d.term] = { label: d.label, type: d.type };
  });

  // Wir nutzen die existierende Save-Logik, füllen aber nur den Teil der uns interessiert
  // configForm muss dafür kurz "geliehen" werden
  Object.keys(configForm).forEach(key => (configForm as any)[key] = (report as any)[key]);
  configForm.PARAM_DEF = JSON.stringify(defMap);

  try {
    $q.loading.show({ message: 'Speichere Parameter...' });
    await api.post('/api/reports/save', configForm);
    
    // Lokale Daten aktualisieren
    report.PARAM_DEF = configForm.PARAM_DEF;
    
    // Die Definitionen für den aktuellen Filter-Dialog neu berechnen
    await onReportSelected([report], { ...filterValues });
    
    $q.notify({ color: 'positive', message: 'Konfiguration erfolgreich aktualisiert', icon: 'check' });
    showQuickParamEdit.value = false;
  } catch(e) {
    console.error('Save error', e);
    $q.notify({ color: 'negative', message: 'Fehler beim Speichern' });
  } finally {
    $q.loading.hide();
  }
}

// selectRow removed as it was only for q-table cell clicks

async function onReportSelected(selection: readonly Report[], forceParams: Record<string, unknown> | null = null) {
  if (selection.length === 0) {
    resultData.value = [];
    resultColumns.value = [];
    currentReportLabel.value = '';
    return;
  }

  const report = selection[0];
  const sql = report.SQLSTATEMENT || '';
  
  // 0. Parameter Scan (Backticks and Legacy %...% format)
  const btMatches = sql.match(/`([^`]+)`/g) || [];
  const percentMatches = sql.match(/%([^%]+)%/g) || [];
  
  // Map für automatische Typ-Erkennung aus dem Legacy-Format
  const autoTypes: Record<string, string> = {};
  percentMatches.forEach(m => {
    const parts = m.replace(/%/g, '').split(';');
    if (parts.length > 1) {
      autoTypes[parts[0]] = parts[1].toUpperCase();
    }
  });

  const btTerms = [...new Set([
    ...btMatches.map(m => m.replace(/`/g, '').split(';')[0]),
    ...percentMatches.map(m => m.replace(/%/g, '').split(';')[0])
  ])];
  
  if (btTerms.length > 0) {
    let existingDef: any = {};
    if (report.PARAM_DEF) {
      const jsonStr = typeof report.PARAM_DEF === 'string' ? report.PARAM_DEF : (report.PARAM_DEF || '');
      if (jsonStr) {
        try { existingDef = JSON.parse(jsonStr); } catch(_e) { /* ignore */ }
      }
    }
    
    // Automatisch erkannte Typen in existingDef mergen (wenn dort noch nichts steht)
    Object.keys(autoTypes).forEach(key => {
      if (!existingDef[key]) {
        existingDef[key] = { label: key, type: autoTypes[key] };
      }
    });

    // Wir stoppen NICHT mehr hier, sondern nutzen die vorhandenen Definitionen oder Standardwerte
    const missing = btTerms.filter(t => !existingDef[t]);
    if (missing.length > 0) {
      // Fehlende Definitionen temporär ergänzen
      missing.forEach(t => {
        existingDef[t] = { label: t, type: (t.toUpperCase().includes('DATUM') ? 'DATE' : 'TEXT') };
      });
    }
    
    // Definitionen für den Filter-Dialog vorbereiten
    currentDefinitions.value = btTerms.map(t => ({
      term: t,
      label: existingDef[t]?.label || t,
      type: existingDef[t]?.type || 'TEXT'
    }));
  }
  const isFolder = getVal(getP(report, 'KATEGORIE_KZ')) === 'K';
  
  if (isFolder) {
    currentReportLabel.value = `${report.BESCHREIBUNG} (Gruppe)`;
    resultData.value = [];
    resultColumns.value = [];
    executedSQL.value = '';
    loadingResult.value = false;
    return;
  }

  currentReportLabel.value = report.BESCHREIBUNG;
  loadingResult.value = true;
  
  resultData.value = [];
  resultSums.value = [];
  resultColumns.value = [];
  masterData.value = [];
  detailData.value = [];
  filterParams.value = [];
  Object.keys(filterValues).forEach(key => delete filterValues[key]);
  executedSQL.value = '';

  try {
    if (forceParams && dialogSQLPreview.value) {
      executedSQL.value = dialogSQLPreview.value;
    }

    const rId = getP(report, 'ID');
    
    // Parameter-Keys säubern (Semikolon entfernen)
    const sanitizedParams: Record<string, any> = {};
    if (forceParams) {
      Object.keys(forceParams).forEach(key => {
        sanitizedParams[key.split(';')[0]] = forceParams[key];
      });
    }
    
    const res = await api.post(`/api/reports/execute/${rId}`, { PARAMS: sanitizedParams });
    
    if (res.data.needs_params) {
      filterParams.value = res.data.params || [];
      
      // STAMMDATEN LADEN FALLS NÖTIG
      if (filterParams.value.some(p => isStammParam(p))) {
        await loadAllStammData();
      }

      filterParams.value.forEach(p => {
        if (filterValues[p.label] === undefined) {
          if (isBool(p)) {
            filterValues[p.label] = false;
          } else if (isStammParam(p)) {
            filterValues[p.label] = null;
          } else {
            filterValues[p.label] = p.type === 'NUMBER' ? 0 : '';
          }
        }
      });

      dialogSQLPreview.value = '';
      void updatePreview();
      showFilterDialog.value = true;
      loadingResult.value = false;
      return;
    }

    const { sql, typ, html, sums } = res.data;
    executedSQL.value = sql || '';
    
    // SICHERHEIT: Wenn Details da sind, MUSS es ein Master-Detail Typ sein
    const rawDetails = res.data.details || [];
    currentReportType.value = (rawDetails.length > 0) ? 'M' : (typ || 'S');
    
    resultSums.value = sums || [];
    
    const serverCols = res.data.columns || [];
    const dServerCols = res.data.detail_columns || [];
    detailServerCols.value = dServerCols;
    
    // Daten für Grids (Universal für T, M, G)
    const rawData = res.data.masterRows || (res.data.master ? [res.data.master] : (res.data.data || []));
    
    masterData.value = rawData || [];
    detailData.value = rawDetails;
    
    // Spalten: Immer Ordnung wahren
    resultColumns.value = getAutoCols(rawData || [], serverCols);
    resultData.value = masterData.value;
    
    if (rawData && rawData.length > 0) {
       selectedMasterRows.value = [rawData[0]];
       
       // SICHERHEIT: Wenn Daten da sind, MÜSSEN wir sie anzeigen
       showMasterGrid.value = true;
       tab.value = 'anzeige'; // Automatisch zum Anzeige-Tab wechseln
       
       $q.notify({ 
         type: 'positive', 
         message: `${rawData.length} Master-Datensätze geladen`, 
         caption: `Typ: ${currentReportType.value} | Details: ${rawDetails.length}`,
         timeout: 2500,
         icon: 'table_view'
       });
    } else {
       $q.notify({ 
         type: 'warning', 
         message: 'Keine Daten gefunden', 
         caption: 'Die Abfrage lieferte 0 Zeilen zurück.',
         timeout: 5000 
       });
    }
    
    // Falls explizit konfiguriert, Detail-Grid zusätzlich an (Master haben wir oben schon erzwungen)
    if (res.data.show_detail_grid) {
       showDetailGrid.value = true;
    }
    
    // Fallback: Bei Master-Detail immer das Detail-Grid erlauben wenn Daten da sind
    if (['M', 'T'].includes(currentReportType.value) && rawDetails.length > 0) {
       showDetailGrid.value = true;
    }
    
    if (html) {
      generatedHtml.value = html;
    }
  } catch (err: any) {
    console.error('Report Execute Error:', err);
    const errorMsg = err?.response?.data?.error || err.message || 'Fehler bei der Ausführung des SQL-Statements.';
    $q.notify({ 
      type: 'negative', 
      message: 'SQL Fehler', 
      caption: errorMsg,
      timeout: 10000 
    });
  } finally {
    loadingResult.value = false;
    $q.loading.hide();
  }
}

let confirmFilters = async () => {
  showFilterDialog.value = false;
  const selection = selectedReport.value;
  if (selection.length > 0) {
    await onReportSelected(selection, { ...filterValues });
  }
}

function wrapCsvValue(val: unknown) {
  let formatted = val === void 0 || val === null ? '' : String(val);
  formatted = formatted.split('"').join('""');
  return `"${formatted}"`;
}

function exportCSV() {
  const content = [resultColumns.value.map(col => wrapCsvValue(col.label)).join(',')].concat(
    resultData.value.map(row => resultColumns.value.map(col => wrapCsvValue(row[col.field])).join(','))
  ).join('\r\n');

  const name = currentReportLabel.value.replace(/[^a-z0-9]/gi, '_').toLowerCase();
  const status = exportFile(`${name || 'report'}.csv`, content, 'text/csv');
  if (status !== true) {
    $q.notify({ message: 'Export fehlgeschlagen', color: 'negative', icon: 'warning' });
  }
}

function printHtmlReport() {
  const printWindow = window.open('', '_blank');
  if (printWindow) {
    printWindow.document.write(generatedHtml.value);
    printWindow.document.close();
    printWindow.focus();
    setTimeout(() => {
      printWindow.print();
      printWindow.close();
    }, 250);
  }
}

// --- TAB 2: KONFIGURATION LOGIC ---
const loadingConfig = ref(false);
const showConfigDialog = ref(false);
const isEditing = ref(false);
const currentEditId = ref<number | null>(null);
const nativeSql = ref('');

const configCols: QTableProps['columns'] = [
  { name: 'actions', label: 'Aktionen', align: 'center', field: (row: Report) => getP(row, 'ID'), style: 'width: 100px' },
  { name: 'BESCHREIBUNG', label: 'Berichtsbezeichnung', field: (row: Report) => getVal(getP(row, 'BESCHREIBUNG')), align: 'left', sortable: true },
  { 
    name: 'KATEGORIE_KZ', 
    label: 'Kat', 
    field: (row: Report) => getVal(getP(row, 'KATEGORIE_KZ')), 
    align: 'center', 
    sortable: true,
    style: 'width: 60px'
  },
  { 
    name: 'GRUPPEN_KZ', 
    label: 'Ordner / Gruppe', 
    field: (row: Report) => {
      const grp = getVal(getP(row, 'GRUPPEN_KZ'));
      if (!grp) return '';
      const opt = folderOptions.value.find(o => o.value === grp);
      return opt ? opt.label : grp;
    },
    align: 'left', 
    sortable: true
  },
  { 
    name: 'TYP_KZ', 
    label: 'Typ', 
    field: (row: Report) => {
      const val = getVal(getP(row, 'TYP_KZ'));
      if (val === 'S') return 'Single SQL';
      if (val === 'M') return 'Multiple SQL';
      if (val === 'H') return 'Kategorie Record';
      if (val === 'G') return 'Gitter (Gruppiert)';
      if (val === 'T') return 'Template (HTML)';
      return val;
    },
    align: 'left',
    sortable: true
  },
  {
    name: 'ROOT_KZ',
    label: 'Root',
    field: (row: Report) => getVal(getP(row, 'ROOT_KZ')),
    align: 'center',
    sortable: true,
    style: 'width: 60px'
  },
  { 
    name: 'TEMPLATE_NAME', 
    label: 'Template', 
    field: (row: Report) => getVal(getP(row, 'TEMPLATE_NAME')), 
    align: 'left', 
    sortable: true 
  }
];

// Form state moved up to Shared Data section


function openCreate() {
  isEditing.value = false;
  currentEditId.value = null;
  configForm.BESCHREIBUNG = '';
  configForm.SQLSTATEMENT = '';
  configForm.KATEGORIE_KZ = 'L';
  configForm.GRUPPEN_KZ = '';
  configForm.TYP_KZ = 'S';
  configForm.TEMPLATE_NAME = '';
  configForm.PARAM_DEF = '';
  configForm.DETAIL_SQL = '';
  configForm.LINK_LOGIC = '';
  configForm.DETAIL_SQL_NATIVE = '';
  configForm.ROOT_KZ = 'x';
  configForm.SUMMENZEILE = '';
  tempSummenSql.value = '';
  configForm.IST_SUMMENZEILE = false;
  
  configStep.value = 1; // Reset Wizard
  showConfigDialog.value = true;
}

async function onEdit(row: Report) {
  loadingConfig.value = true;
  try {
    const rowId = getP(row, 'ID');
    console.log('onEdit: row =', row, 'rowId =', rowId);

    if (!rowId && rowId !== 0) {
      console.error('onEdit: Ungültige ID für Report:', row);
      $q.notify({ type: 'negative', message: 'Konnte Details nicht laden: Ungültige ID' });
      return;
    }

    const res = await api.get(`/api/reports/${rowId}`);
    console.log('onEdit: API response =', res.data);
    if (res.data) {
      isEditing.value = true;
      currentEditId.value = rowId;
      mapReportToForm(res.data);
      configStep.value = 1; // Reset Wizard
      showConfigDialog.value = true;
    }
  } catch (_err: any) {
    console.error('onEdit FEHLER:', _err, 'row =', row);
    const msg = _err?.response?.data?.error || 'Konnte Details nicht laden';
    $q.notify({ type: 'negative', message: msg });
  } finally {
    loadingConfig.value = false;
  }
}

async function onCopy(row: Report) {
  loadingConfig.value = true;
  try {
    const rowId = getP(row, 'ID');
    const res = await api.get(`/api/reports/${rowId}`);
    if (res.data) {
      isEditing.value = false; // Wir erstellen einen NEUEN Eintrag
      currentEditId.value = 0;
      mapReportToForm(res.data);
      configForm.ID = 0; // WICHTIG: Die ID muss für eine Kopie genullt werden
      configForm.BESCHREIBUNG += ' (Kopie)';
      configForm.SYSTEM_KZ = false; // Kopie ist kein System-Eintrag
      configStep.value = 1; // Reset Wizard
      showConfigDialog.value = true;
      
      $q.notify({
        message: 'Bericht kopiert. Bitte Bezeichnung anpassen und speichern.',
        color: 'secondary',
        icon: 'content_copy',
        position: 'top'
      });
    }
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Konnte Bericht nicht kopieren' });
  } finally {
    loadingConfig.value = false;
  }
}

function mapReportToForm(data: any) {
  configForm.ID = getP(data, 'ID') || 0;
  configForm.BESCHREIBUNG = getVal(getP(data, 'BESCHREIBUNG'));
  configForm.SQLSTATEMENT = getVal(getP(data, 'SQLSTATEMENT'));
  configForm.KATEGORIE_KZ = getVal(getP(data, 'KATEGORIE_KZ'));
  configForm.GRUPPEN_KZ = getVal(getP(data, 'GRUPPEN_KZ'));
  configForm.TYP_KZ = getVal(getP(data, 'TYP_KZ'));
  configForm.TEMPLATE_NAME = getVal(getP(data, 'TEMPLATE_NAME'));
  configForm.PARAM_DEF = getVal(getP(data, 'PARAM_DEF'));
  configForm.DETAIL_SQL = getVal(getP(data, 'DETAIL_SQL'));
  configForm.LINK_LOGIC = getVal(getP(data, 'LINK_LOGIC'));
  configForm.GROUP_FIELD = getVal(getP(data, 'GROUP_FIELD'));
  configForm.ROWS_PER_PAGE = Number(getP(data, 'ROWS_PER_PAGE') || 0);
  configForm.PAGE_ORIENTATION = getVal(getP(data, 'PAGE_ORIENTATION')) || 'P';
  configForm.SHOW_MASTER_GRID = getBool(getP(data, 'SHOW_MASTER_GRID'));
  configForm.SHOW_DETAIL_GRID = getBool(getP(data, 'SHOW_DETAIL_GRID'));
  configForm.SYSTEM_KZ = getBool(getP(data, 'SYSTEM_KZ'));
  configForm.DETAIL_SQL_NATIVE = getVal(getP(data, 'DETAIL_SQL_NATIVE'));
  configForm.SQLSTATEMENT_NATIVE = getVal(getP(data, 'SQLSTATEMENT_NATIVE'));
  configForm.ROOT_KZ = getVal(getP(data, 'ROOT_KZ'));
  configForm.SUMMENZEILE = getVal(getP(data, 'SUMMENZEILE'));
  tempSummenSql.value = configForm.SUMMENZEILE;
  configForm.IST_SUMMENZEILE = getBool(getP(data, 'IST_SUMMENZEILE'));
}

function onExplorerAppend(content: string) {
  if (configStep.value === 3) onDrop({ dataTransfer: { getData: (k: string) => k === 'source' ? 'db-explorer' : (k === 'payload' ? JSON.stringify({type: 'column', text: content}) : '') } } as any, 'SQLSTATEMENT');
  if (configStep.value === 4) onDrop({ dataTransfer: { getData: (k: string) => k === 'source' ? 'db-explorer' : (k === 'payload' ? JSON.stringify({type: 'column', text: content}) : '') } } as any, 'DETAIL_SQL');
  if (configStep.value === 7) {
    const space = tempSummenSql.value && !tempSummenSql.value.endsWith(' ') ? ' ' : '';
    tempSummenSql.value += space + content.toUpperCase();
  }
}

function transformSqlDirect() {
  let sql = configForm.SQLSTATEMENT.trim();
  if (!sql) return;

  // SELECT-Header fixen
  if (!sql.toUpperCase().startsWith('SELECT')) {
    sql = 'SELECT ' + sql;
  }

  // Spalten-Aliasing
  const selectRegex = /(SELECT\s+)([\s\S]*?)(\s+FROM)/i;
  const match = sql.match(selectRegex);
  
  if (match && match[2]) {
    const selectContent = match[2];
    const columns = selectContent.split(/,(?![^(]*\))/);
    
    const transformedCols = columns.map(col => {
      const c = col.trim();
      if (!c || c === '*' || c.endsWith('.*') || c.toUpperCase().includes(' AS ') || c.includes('"') || c.includes('§')) return c;
      const fieldNameMatch = c.match(/([a-zA-Z0-9_]+)(?=\s*\)|$)/);
      const label = (fieldNameMatch && fieldNameMatch[1]) ? fieldNameMatch[1].toUpperCase() : c.toUpperCase().replace(/[^A-Z0-9_]/g, '');
      return `${c} AS "§${label}§"`;
    });
    
    const newSelectContent = '\n    ' + transformedCols.filter(c => c).join(',\n    ') + '\n';
    sql = sql.replace(selectRegex, (m, p1, p2, p3) => p1 + newSelectContent + p3);
  }

  // Filter-Ersetzungen & Auto-Registration
  let currentDef: Record<string, any> = {};
  if (configForm.PARAM_DEF) {
    try { currentDef = JSON.parse(configForm.PARAM_DEF); } catch(e) { /* ignore */ }
  }

  // Datum
  sql = sql.replace(/BETWEEN\s+'\d{4}-\d{2}-\d{2}'\s+AND\s+'\d{4}-\d{2}-\d{2}'/gi, (m) => {
    currentDef['Startdatum'] = { label: 'Startdatum', type: 'DATE' };
    currentDef['Enddatum'] = { label: 'Enddatum', type: 'DATE' };
    return "BETWEEN `Startdatum` AND `Enddatum` ";
  });
  sql = sql.replace(/=\s*'\d{4}-\d{2}-\d{2}'/gi, (m) => {
    currentDef['Datum'] = { label: 'Datum', type: 'DATE' };
    return "= `Datum` ";
  });
  
  // Zahlen
  sql = sql.replace(/=\s*(\d+(\.\d+)?)/gi, (match, p1) => {
    const num = parseFloat(p1);
    if (p1.length === 4 && num >= 1900 && num <= 2100) return match; 
    currentDef['Filter'] = { label: 'Filter', type: 'NUMBER' };
    return "= `Filter` ";
  });

  // Bestehende Backticks synchronisieren
  const btMatches = sql.match(/`([^`]+)`/g) || [];
  btMatches.forEach(m => {
    const term = m.replace(/`/g, '');
    if (!currentDef[term]) {
      let type: string = 'TEXT';
      if (term.toLowerCase().includes('datum')) type = 'DATE';
      if (term.toLowerCase().includes('filter') || term.toLowerCase().includes('menge')) type = 'NUMBER';
      currentDef[term] = { label: term, type };
    }
  });

  configForm.PARAM_DEF = JSON.stringify(currentDef);
  configForm.SQLSTATEMENT = sql;
  $q.notify({ type: 'positive', message: 'SQL aufbereitet (Aliase & Filter)', icon: 'auto_fix_high' });
}

function exportNativeSql() {
  let sql = configForm.SQLSTATEMENT;
  // Aliase entfernen: AS "§NAME§" oder AS §NAME§
  sql = sql.replace(/\s+AS\s+"?§[^§]+§"?/gi, '');
  
  // Parameter-Definitionen laden
  let currentDef: Record<string, any> = {};
  if (configForm.PARAM_DEF) {
    try {
      currentDef = JSON.parse(configForm.PARAM_DEF);
    } catch (_e) { /* ignore */ }
  }

  // Backticks in %Label;TYP% umwandeln (für spätere Wiederverwendung)
  sql = sql.replace(/`([^`]+)`/g, (match, term) => {
    const def = currentDef[term];
    const type = def?.type || 'TEXT';
    return `%${term};${type}%`;
  });
  
  copyToClipboard(sql)
    .then(() => {
      $q.notify({ type: 'info', message: 'SQL (Nativ mit Platzhaltern) in Zwischenablage kopiert' });
    })
    .catch(() => {
      $q.notify({ type: 'negative', message: 'Kopieren fehlgeschlagen' });
    });
}

async function onConfigSubmit() {
  if (!configForm.BESCHREIBUNG) {
    $q.notify({ type: 'warning', message: 'Bitte eine Beschreibung angeben' });
    return;
  }

  // Das Summen-SQL muss synchronisiert werden, da es hierfür keinen separaten "Nativ"-Editor gibt
  configForm.SUMMENZEILE = tempSummenSql.value;

  // Definitionen aus Wizard-Schritt 8 synchronisieren
  if (currentDefinitions.value.length > 0) {
    const defMap: Record<string, any> = {};
    currentDefinitions.value.forEach(d => {
      defMap[d.term] = { label: d.label, type: d.type };
    });
    configForm.PARAM_DEF = JSON.stringify(defMap);
  }

  // Nur bei Ordnern (K) erzwingen wir den Typ H (Hierarchy)
  if (configForm.KATEGORIE_KZ === 'K') {
    configForm.TYP_KZ = 'H';
    configForm.SQLSTATEMENT = 'FOLDER';
    configForm.TEMPLATE_NAME = '';
  }

  // Sicherstellen dass Template_Name N/A oder leer ist wenn nicht Formular
  if (configForm.KATEGORIE_KZ !== 'F') {
    configForm.TEMPLATE_NAME = '';
  }

  if (configForm.KATEGORIE_KZ !== 'K' && !configForm.SQLSTATEMENT) {
    $q.notify({ type: 'warning', message: 'Bitte SQL-Statement angeben' });
    return;
  }
  
  loadingConfig.value = true;
  try {
    const payload: any = {};
    const keys = [
      'BESCHREIBUNG', 'SQLSTATEMENT', 'KATEGORIE_KZ', 'GRUPPEN_KZ', 'TYP_KZ',
      'TEMPLATE_NAME', 'PARAM_DEF', 'DETAIL_SQL', 'LINK_LOGIC', 'GROUP_FIELD',
      'ROWS_PER_PAGE', 'PAGE_ORIENTATION', 'SHOW_MASTER_GRID', 'SHOW_DETAIL_GRID',
      'SYSTEM_KZ', 'SQLSTATEMENT_NATIVE', 'DETAIL_SQL_NATIVE', 'ROOT_KZ',
      'SUMMENZEILE', 'IST_SUMMENZEILE'
    ];
    
    keys.forEach(k => {
      const val = configForm[k];
      if (typeof val === 'boolean') {
        payload[k] = val ? 1 : 0;
      } else {
        payload[k] = getVal(val);
      }
    });

    // SICHERHEIT: Nativ hat Vorrang beim Speichern
    if (payload.SQLSTATEMENT_NATIVE && payload.SQLSTATEMENT_NATIVE.trim()) {
      payload.SQLSTATEMENT = payload.SQLSTATEMENT_NATIVE;
    }
    if (payload.DETAIL_SQL_NATIVE && payload.DETAIL_SQL_NATIVE.trim()) {
      payload.DETAIL_SQL = payload.DETAIL_SQL_NATIVE;
    }

    // Weitere Felder nachbearbeiten
    keys.forEach(k => {
      // SICHERHEIT: Ordner (Kategorie K) MÜSSEN ein Gruppen_KZ haben (ihren eigenen Code)
      if (k === 'GRUPPEN_KZ' && getVal(configForm['KATEGORIE_KZ']).toUpperCase() === 'K') {
        if (!payload[k]) {
          const desc = getVal(configForm['BESCHREIBUNG']);
          payload[k] = desc ? desc.charAt(0).toUpperCase() : 'X';
        }
      }
      if (['ROWS_PER_PAGE', 'SHOW_MASTER_GRID', 'SHOW_DETAIL_GRID', 'SYSTEM_KZ', 'IST_SUMMENZEILE'].includes(k)) {
         payload[k] = Number(payload[k] || 0);
      }
    });

    if (payload.KATEGORIE_KZ === 'K') {
      payload.TYP_KZ = 'H';
      payload.SQLSTATEMENT = 'FOLDER';
      payload.TEMPLATE_NAME = '';
    }

    if (isEditing.value && currentEditId.value) {
      await api.put(`/api/reports/${currentEditId.value}`, payload);
      $q.notify({ type: 'positive', message: 'Report erfolgreich aktualisiert', icon: 'check_circle' });
    } else {
      await api.post('/api/reports', payload);
      $q.notify({ type: 'positive', message: 'Neuer Report wurde gespeichert', icon: 'check_circle' });
    }
    showConfigDialog.value = false;
    setTimeout(loadReports, 150);
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Speichern des Berichts' });
  } finally {
    loadingConfig.value = false;
  }
}

async function onDelete(row: Report) {
  const rowId = getP(row, 'ID');
  const systemKz = getP(row, 'SYSTEM_KZ');
  const beschreibung = getVal(getP(row, 'BESCHREIBUNG'));

  if (getBool(systemKz) && !isAdmin.value) {
    $q.notify({ type: 'warning', message: 'System-Einträge können nur von Administratoren gelöscht werden.' });
    return;
  }

  $q.dialog({
    title: '<div class="text-negative">Bericht löschen?</div>',
    message: `Möchten Sie den Bericht <strong>"${beschreibung}"</strong> wirklich unwiderruflich löschen?`,
    html: true,
    cancel: { label: 'Abbrechen', flat: true },
    ok: { label: 'Löschen', color: 'negative', unelevated: true },
    persistent: true
  }).onOk(async () => {
    loadingConfig.value = true;
    try {
      await api.delete(`/api/reports/${rowId}`);
      $q.notify({ type: 'positive', message: 'Bericht wurde gelöscht' });
      void loadReports();
    } catch (_err) {
      $q.notify({ type: 'negative', message: 'Löschen fehlgeschlagen' });
    } finally {
      loadingConfig.value = false;
    }
  });
}


// --- Drag & Drop Logik ---
const dragNode = ref<any>(null);
const dropTargetKey = ref<string | null>(null);

function onTreeDragStart(event: DragEvent, node: any) {
  if (!node.data) {
    event.preventDefault(); // Keine Roots draggen
    return;
  }
  dragNode.value = node;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', node.key);
  }
}

function onTreeDragOver(event: DragEvent, node: any) {
  // Check obs ein gültiges Drop-Target ist
  const isTargetFolder = node.data && getVal(getP(node.data, 'KATEGORIE_KZ')).toUpperCase() === 'K';
  const isTargetReport = node.data && getVal(getP(node.data, 'KATEGORIE_KZ')).toUpperCase() === 'L';
  const isTargetRoot = node.key.startsWith('root-');
  
  if (isTargetFolder || isTargetReport || isTargetRoot) {
    dropTargetKey.value = node.key;
    event.preventDefault();
  }
}

function onTreeDragLeave(_event: DragEvent, _node: any) {
  dropTargetKey.value = null;
}

async function onTreeDrop(_event: DragEvent, targetNode: any) {
  const sourceNode = dragNode.value;
  dropTargetKey.value = null;
  dragNode.value = null;

  if (!sourceNode || !targetNode || sourceNode.key === targetNode.key) return;

  const sourceData = sourceNode.data;
  const targetData = targetNode.data;
  const isSourceFolder = getVal(getP(sourceData, 'KATEGORIE_KZ')).toUpperCase() === 'K';
  const isTargetFolder = targetData && getVal(getP(targetData, 'KATEGORIE_KZ')).toUpperCase() === 'K';
  const isTargetReport = targetData && getVal(getP(targetData, 'KATEGORIE_KZ')).toUpperCase() === 'L';
  const isTargetRoot = targetNode.key.startsWith('root-');

  // Regel 1: Einzelne Reports dürfen in Ordner, auf andere Reports oder auf Roots verschoben werden
  if (!isSourceFolder) {
    if (isTargetFolder || isTargetReport || isTargetRoot) {
      await moveReportToFolder(sourceData, targetData || { ROOT_KZ: targetNode.key.replace('root-', ''), KATEGORIE_KZ: 'ROOT' });
    } else {
      $q.notify({ 
        type: 'warning', 
        message: 'Einzelne Berichte dürfen nur in Ordner, auf Berichte oder auf Haupt-Kategorien verschoben werden.',
        position: 'top',
        timeout: 2500
      });
    }
    return;
  }

  // Regel 2: Ordner selbst dürfen nur in ROOT verschoben werden
  if (isSourceFolder) {
    if (isTargetRoot) {
      const rootKz = targetNode.key.replace('root-', '');
      await moveFolderToRoot(sourceData, rootKz);
    } else {
      $q.notify({ 
        type: 'warning', 
        message: 'Ordner können nur direkt auf die Haupt-Kategorien (ROOT) verschoben werden.',
        position: 'top',
        timeout: 2500
      });
    }
    return;
  }
}

async function moveReportToFolder(report: any, target: any) {
  const targetKat = getVal(getP(target, 'KATEGORIE_KZ')).toUpperCase();
  const isTargetFolder = targetKat === 'K';
  const isTargetReport = targetKat === 'L';
  const isRootDrop = targetKat === 'ROOT';
  
  let newGrp = '';
  if (isTargetFolder || isTargetReport) {
    newGrp = getVal(getP(target, 'GRUPPEN_KZ'));
    
    // Robuster Fallback für "kaputte" Ordner ohne Code
    if (!newGrp && isTargetFolder) {
      const targetId = getP(target, 'ID');
      const targetDesc = getVal(getP(target, 'BESCHREIBUNG'));
      newGrp = targetDesc ? targetDesc.charAt(0).toUpperCase() : String(targetId);
      console.warn(`Ordner ID ${targetId} hat keinen Code! Nutze Fallback: ${newGrp}`);
    }
  }
  
  const newRoot = getVal(getP(target, 'ROOT_KZ'));
  
  if (!newRoot) {
    console.warn('moveReportToFolder: Kein Root_KZ gefunden im Ziel', target);
    return;
  }

  $q.loading.show({ message: 'Verschiebe Bericht...' });
  try {
    const reportDesc = getVal(getP(report, 'BESCHREIBUNG'));
    
    await saveReportEntry(report, {
      GRUPPEN_KZ: newGrp,
      ROOT_KZ: newRoot
    });
    
    let targetDesc = isRootDrop ? 'Haupt-Kategorie' : (isTargetFolder ? 'Ordner' : 'Position');
    $q.notify({ 
      type: 'positive', 
      message: `"${reportDesc}" verschoben auf ${targetDesc}`,
      icon: 'check_circle'
    });
    // Kurzer Delay um DB-Locks/Sync-Themen vorzubeugen
    setTimeout(loadReports, 200);
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Verschieben' });
  } finally {
    $q.loading.hide();
  }
}

async function moveFolderToRoot(folder: any, rootKz: string) {
  $q.loading.show({ message: 'Verschiebe Ordner...' });
  try {
    const folderGrp = getVal(getP(folder, 'GRUPPEN_KZ'));
    
    // 1. Den Ordner selbst aktualisieren
    await saveReportEntry(folder, { ROOT_KZ: rootKz });
    
    // 2. Alle Berichte im Folder mitverschieben
    const reportsToMove = reportRows.value.filter(r => 
      getVal(getP(r, 'GRUPPEN_KZ')).toUpperCase() === folderGrp.toUpperCase() && 
      getVal(getP(r, 'KATEGORIE_KZ')).toUpperCase() !== 'K'
    );
    
    for (const r of reportsToMove) {
      await saveReportEntry(r, { ROOT_KZ: rootKz });
    }

    $q.notify({ type: 'positive', message: 'Ordner und Inhalte verschoben' });
    setTimeout(loadReports, 150);
  } catch (err) {
    $q.notify({ type: 'negative', message: 'Fehler beim Verschieben des Ordners' });
  } finally {
    $q.loading.hide();
  }
}



watch(() => route.query.tab, (newTab) => {
  if (newTab) {
    tab.value = String(newTab);
  }
});

onMounted(() => {
  // Initialer Tab aus Query-Parameter
  if (route.query.tab) {
    tab.value = String(route.query.tab);
  }
  
  loadReports();
  loadTables();
});
</script>

<style scoped>
.sticky-header-table :deep(.q-table__middle) {
  max-height: 560px;
}
.sticky-header-table :deep(thead tr:first-child th) {
  position: sticky;
  top: 0;
  opacity: 1;
  z-index: 10;
}
.font-mono {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
}
.sql-editor :deep(textarea) {
  line-height: 1.5;
  tab-size: 4;
}
.border-top {
  border-top: 1px solid rgba(0,0,0,0.12);
}
.droppable-node {
  transition: all 0.2s;
  cursor: grab;
  border-bottom: 1px solid rgba(0,0,0,0.05);
  margin-bottom: 2px;
}
.body--light .droppable-node:hover {
  background-color: rgba(0,0,0,0.03);
}
.body--dark .droppable-node {
  border-bottom: 1px solid rgba(255,255,255,0.05);
}
.body--dark .droppable-node:hover {
  background-color: rgba(255,255,255,0.05);
}
.droppable-node:active {
  cursor: grabbing;
}
</style>

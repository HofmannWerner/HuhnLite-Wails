<template>
  <div class="q-pa-md">
    <!-- Filter Row -->
    <div class="row q-gutter-md q-mb-md items-center shadow-1 q-pa-sm rounded-borders"
         :class="$q.dark.isActive ? 'bg-grey-9 text-white' : 'bg-grey-1 text-black'">
      <div class="col-12 col-sm-auto">
        <div class="text-caption q-mb-xs" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'">1. Filter:
          Lager-KZ
        </div>
        <q-select
          v-model="filterLagertyp"
          :options="lagertypOptions"
          emit-value
          map-options
          option-value="KZ"
          option-label="BETREFF"
          style="min-width: 200px"
          outlined
          dense
          :bg-color="$q.dark.isActive ? 'grey-8' : 'white'"
          :dark="$q.dark.isActive"
          @update:model-value="onTypeChange"
        />
      </div>

      <div class="col-12 col-sm-auto">
        <div class="text-caption q-mb-xs" :class="$q.dark.isActive ? 'text-grey-4' : 'text-grey-7'">2. Filter:
          Eierlager
        </div>
        <q-select
          v-model="filterEilager"
          :options="filteredEilagerOptions"
          emit-value
          map-options
          option-value="ID"
          option-label="label"
          style="min-width: 250px"
          outlined
          dense
          :bg-color="$q.dark.isActive ? 'grey-8' : 'white'"
          :dark="$q.dark.isActive"
          @update:model-value="loadData"
        />
      </div>

      <q-space />

      <div class="row q-gutter-sm">
        <q-btn
          :color="$q.dark.isActive ? 'blue-2' : 'primary'"
          :label="expanded.length > 0 ? 'Alle einklappen' : 'Alle ausklappen'"
          :icon="expanded.length > 0 ? 'unfold_less' : 'unfold_more'"
          @click="toggleExpand"
          flat
          dense
        />
        <q-btn icon="refresh" @click="loadData" flat round :color="$q.dark.isActive ? 'white' : 'primary'"/>
      </div>
    </div>

    <!-- Tree View -->
    <q-tree
      :nodes="treeNodes"
      node-key="ID"
      v-model:expanded="expanded"
      no-nodes-label="Keine Bestandsdaten für diese Auswahl gefunden"
      class="bestands-tree q-pa-md rounded-borders shadow-2"
      :class="$q.dark.isActive ? 'bg-grey-10 text-white' : 'bg-white text-black'"
      default-expand-all
      :dark="$q.dark.isActive"
    >
      <template v-slot:default-header="prop">
        <div class="row items-center full-width no-wrap tree-header-content">
          <div class="text-weight-bold"
               :class="prop.node.type === 'lager' ? ($q.dark.isActive ? 'text-blue-2 text-h6' : 'text-primary text-h6') :
                       prop.node.type === 'charge' ? ($q.dark.isActive ? 'text-blue-3 text-subtitle1' : 'text-secondary text-subtitle1') :
                       ($q.dark.isActive ? 'text-grey-3 q-pl-md' : 'text-grey-8 q-pl-md')">
            {{ prop.node.label }}
          </div>
          <q-space />
          <!-- Info Columns -->
          <div class="row q-gutter-sm no-wrap text-center columns-container">
            <div v-for="col in eggColumns" :key="col.name" class="egg-col">
              <div class="text-caption" :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-6'">{{ col.label }}</div>
              <div class="text-weight-medium"
                   :class="prop.node.sums[col.field] > 0 ? ($q.dark.isActive ? 'text-white' : 'text-black') : 'text-grey-4'">
                {{ prop.node.sums[col.field] > 0 ? prop.node.sums[col.field].toLocaleString('de-DE') : '-' }}
              </div>
            </div>
            <div class="egg-col total-col">
              <div class="text-caption text-grey-7">Summe</div>
              <div class="text-weight-bolder"
                   :class="prop.node.total > 0 ? ($q.dark.isActive ? 'text-amber-5' : 'text-deep-orange-9') : 'text-grey-4'">
                {{ prop.node.total > 0 ? prop.node.total.toLocaleString('de-DE') : '-' }}
              </div>
            </div>
          </div>
        </div>
      </template>
    </q-tree>

    <q-inner-loading :showing="loading">
      <q-spinner-tail size="50px" color="primary" />
    </q-inner-loading>
  </div>
</template>

<script setup lang="ts">
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

import { ref, onMounted, computed } from 'vue';
import {api} from 'src/boot/api';
import { useQuasar } from 'quasar';

interface StockRow {
  CHARGE?: any;
  LAGERPLATZ_BEZEICHNUNG?: any;
  JUMBOS?: any;
  XL?: any;
  LARGE?: any;
  MEDIUM?: any;
  SMALL?: any;
  VOLLEIKG?: any;
  EILAGER_KZ?: any;
  EILAGER_ID?: any;
  EILAGER_BEZEICHNUNG?: any;
}

interface EilagerOption {
  ID: number;
  label: string;
  KZ?: string;
}

interface TreeNode {
  ID: string;
  label: string;
  type: 'lager' | 'charge' | 'lagerplatz';
  sums: {
    JUMBOS: number;
    XL: number;
    LARGE: number;
    MEDIUM: number;
    SMALL: number;
    VOLLEIKG: number;
  };
  children?: TreeNode[];
  total: number;
}

const $q = useQuasar();
const loading = ref(false);
const rows = ref<StockRow[]>([]);
const filterLagertyp = ref('0');
const filterEilager = ref(0);
const eilagerOptions = ref<EilagerOption[]>([]);
const lagertypOptions = ref<any[]>([]);
const expanded = ref<string[]>([]);

const eggColumns = [
  {name: 'JUMBOS', label: 'Jumbos', field: 'JUMBOS'},
  {name: 'XL', label: 'XL', field: 'XL'},
  {name: 'LARGE', label: 'L', field: 'LARGE'},
  {name: 'MEDIUM', label: 'M', field: 'MEDIUM'},
  {name: 'SMALL', label: 'S', field: 'SMALL'},
  {name: 'VOLLEIKG', label: 'Vollei (kg)', field: 'VOLLEIKG'}
];

const filteredEilagerOptions = computed(() => {
  let list = eilagerOptions.value;
  if (filterLagertyp.value !== '0') {
    list = list.filter(o => o.ID === 0 || o.KZ === filterLagertyp.value);
  }
  return list;
});

const treeNodes = computed<TreeNode[]>(() => {
  const lagerGroups: Record<string, TreeNode> = {};

  // Apply frontend filtering on top of backend results
  const filteredRows = rows.value.filter(row => {
    // If we want ALL (0), don't filter at all in frontend
    const filterAllTypes = filterLagertyp.value === '0';
    const filterAllLagers = filterEilager.value === 0;

    if (filterAllTypes && filterAllLagers) return true;

    const rKz = row.EILAGER_KZ;
    const rId = Number(row.EILAGER_ID ?? row.EILAGER_ID);

    if (!filterAllLagers && rId !== filterEilager.value) return false;
    if (!filterAllTypes && rKz !== filterLagertyp.value) return false;

    return true;
  });

  filteredRows.forEach(row => {
    const lName = row.EILAGER_BEZEICHNUNG || 'Unbekanntes Lager';
    const lIdVal = Number(row.EILAGER_ID ?? row.EILAGER_ID);

    if (!lagerGroups[lName]) {
      lagerGroups[lName] = {
        ID: `lager-${lIdVal}-${lName}`,
        label: `Lager: ${lName}`,
        type: 'lager',
        sums: {JUMBOS: 0, XL: 0, LARGE: 0, MEDIUM: 0, SMALL: 0, VOLLEIKG: 0},
        children: [],
        total: 0
      };
    }
    const lagerNode = lagerGroups[lName];
    const chargeName = row.CHARGE || 'Keine Charge';
    let chargeNode = (lagerNode.children || []).find(c => c.label === `Charge: ${chargeName}`);
    if (!chargeNode) {
      chargeNode = {
        ID: `lager-${lIdVal}-charge-${chargeName}`,
        label: `Charge: ${chargeName}`,
        type: 'charge',
        sums: {JUMBOS: 0, XL: 0, LARGE: 0, MEDIUM: 0, SMALL: 0, VOLLEIKG: 0},
        children: [],
        total: 0
      };
      lagerNode.children?.push(chargeNode);
    }

    const lpLabel = row.LAGERPLATZ_BEZEICHNUNG || 'Ohne Lagerplatz';

    const extractVal = (v: any) => {
      if (v === null || v === undefined) return 0;
      if (typeof v === 'object' && 'Int64' in v) return Number(v.Int64) || 0;
      if (typeof v === 'object' && 'Int32' in v) return Number(v.Int32) || 0;
      if (typeof v === 'object' && 'Float64' in v) return Number(v.Float64) || 0;
      return Number(v) || 0;
    };

    const lpSums = {
      JUMBOS: extractVal(row.JUMBOS),
      XL: extractVal(row.XL),
      LARGE: extractVal(row.LARGE),
      MEDIUM: extractVal(row.MEDIUM),
      SMALL: extractVal(row.SMALL),
      VOLLEIKG: extractVal(row.VOLLEIKG)
    };

    const lpTotal = Object.values(lpSums).reduce((a, b) => a + b, 0);

    let lpNode = (chargeNode.children || []).find(c => c.label === `Lagerplatz: ${lpLabel}`);
    if (lpNode) {
      lpNode.sums.JUMBOS += lpSums.JUMBOS;
      lpNode.sums.XL += lpSums.XL;
      lpNode.sums.LARGE += lpSums.LARGE;
      lpNode.sums.MEDIUM += lpSums.MEDIUM;
      lpNode.sums.SMALL += lpSums.SMALL;
      lpNode.sums.VOLLEIKG += lpSums.VOLLEIKG;
      lpNode.total += lpTotal;
    } else {
      chargeNode.children?.push({
        ID: `lager-${lIdVal}-charge-${chargeName}-lp-${lpLabel}`,
        label: `Lagerplatz: ${lpLabel}`,
        type: 'lagerplatz',
        sums: lpSums,
        total: lpTotal
      });
    }

    chargeNode.sums.JUMBOS += lpSums.JUMBOS;
    chargeNode.sums.XL += lpSums.XL;
    chargeNode.sums.LARGE += lpSums.LARGE;
    chargeNode.sums.MEDIUM += lpSums.MEDIUM;
    chargeNode.sums.SMALL += lpSums.SMALL;
    chargeNode.sums.VOLLEIKG += lpSums.VOLLEIKG;
    chargeNode.total += lpTotal;

    lagerNode.sums.JUMBOS += lpSums.JUMBOS;
    lagerNode.sums.XL += lpSums.XL;
    lagerNode.sums.LARGE += lpSums.LARGE;
    lagerNode.sums.MEDIUM += lpSums.MEDIUM;
    lagerNode.sums.SMALL += lpSums.SMALL;
    lagerNode.sums.VOLLEIKG += lpSums.VOLLEIKG;
    lagerNode.total += lpTotal;
  });

  return Object.values(lagerGroups);
});

async function loadData() {
  loading.value = true;
  try {
    const id = filterEilager.value || 0;
    const url = id > 0 ? `/api/eilager/bestandsuebersicht?id_eilager=${id}` : '/api/eilager/bestandsuebersicht?id_eilager=0';
    const res = await api.get(url);
    rows.value = (res.data as StockRow[]) || [];
  } catch (_err) {
    $q.notify({ type: 'negative', message: 'Bestandsdaten konnten nicht geladen werden' });
  } finally {
    loading.value = false;
  }
}

async function loadEilager() {
  try {
    const [eRes, tRes] = await Promise.all([
      api.get('/api/eilager'),
      api.get('/api/texte/typ/L') // Eilager Typen (Kategorie L)
    ]);

    const mappedE = eRes.data.map((e: any) => ({
      ID: e.ID,
      label: `${e.LAGERNUMMER} - ${e.BEZEICHNUNG}`,
      KZ: e.KZ
    }));
    eilagerOptions.value = [{ID: 0, label: 'ALLE Läger', KZ: '0'}, ...mappedE];

    const mappedT = (tRes.data || []).map((t: any) => ({
      KZ: t.KZ,
      BETREFF: t.BETREFF
    }));
    lagertypOptions.value = [{KZ: '0', BETREFF: 'ALLE Kennzeichen'}, ...mappedT];
  } catch (err) {
    console.error(err);
  }
}

function onTypeChange() {
  // If type changes, reset specific lager filter to 'ALLE' for that type
  filterEilager.value = 0;
  void loadData();
}

function toggleExpand() {
  if (expanded.value.length > 0) {
    collapseAll();
  } else {
    expandAll();
  }
}

function expandAll() {
  const allIds: string[] = [];
  const collectIds = (nodes: any[]) => {
    nodes.forEach(n => {
      allIds.push(n.ID);
      if (n.children && n.children.length > 0) {
        collectIds(n.children);
      }
    });
  };
  collectIds(treeNodes.value);
  expanded.value = allIds;
}

function collapseAll() {
  expanded.value = [];
}

onMounted(() => {
  void loadEilager();
  void loadData();
});
</script>

<style scoped>
.tree-header-content {
  padding: 8px 16px;
}

.columns-container {
  display: flex;
  align-items: center;
}

.egg-col {
  min-width: 80px;
  border-left: 1px solid rgba(0,0,0,0.05);
}

.total-col {
  min-width: 100px;
  background: rgba(var(--q-secondary), 0.05);
}

.text-h6 {
  font-size: 1.1rem;
}
</style>

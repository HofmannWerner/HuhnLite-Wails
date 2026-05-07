<template>
  <div 
    class="q-pa-md q-mb-lg rounded-borders shadow-1" 
    :class="$q.dark.isActive ? 'bg-grey-10' : 'bg-grey-1'"
    style="border-left: 6px solid var(--q-primary);"
  >
    <div class="row items-center q-col-gutter-md">
      <div class="col-12 col-md-4">
        <div class="text-subtitle1 text-weight-bold row items-center">
          <q-icon name="schedule" color="primary" class="q-mr-sm" size="sm" />
          Erfassungs-Zeitpunkt
        </div>
        <div class="text-caption text-grey-7">Datum und Uhrzeit der Buchung</div>
      </div>
      
      <div class="col-6 col-md-4">
        <q-input
          v-model="date"
          type="date"
          dense
          filled
          stack-label
          label="Datum"
          :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
          @update:model-value="onUpdate"
        >
          <template v-slot:prepend>
            <q-icon name="event" size="xs" />
          </template>
        </q-input>
      </div>

      <div class="col-6 col-md-4">
        <q-input
          v-model="time"
          type="time"
          dense
          filled
          stack-label
          label="Uhrzeit"
          :bg-color="$q.dark.isActive ? 'grey-9' : 'white'"
          @update:model-value="onUpdate"
        >
          <template v-slot:prepend>
            <q-icon name="access_time" size="xs" />
          </template>
        </q-input>
      </div>
    </div>
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

import { ref, onMounted, watch } from 'vue';
import { useQuasar } from 'quasar';

const $q = useQuasar();

const props = defineProps({
  modelValue: {
    type: String, // JJJJ-MM-TT HH:mm
    default: ''
  }
});

const emit = defineEmits(['update:modelValue']);

const date = ref('');
const time = ref('');

function initialize() {
  const source = props.modelValue || '';
  if (source.includes(' ')) {
    const parts = source.split(' ');
    date.value = parts[0] || '';
    time.value = parts[1] ? parts[1].substring(0, 5) : '';
  } else {
    // Current time as default
    const now = new Date();
    date.value = now.toISOString().split('T')[0] || '';
    time.value = (now.toTimeString() || '').substring(0, 5);
    onUpdate(); // emit the initial value
  }
}

function onUpdate() {
  if (date.value && time.value) {
    emit('update:modelValue', `${date.value} ${time.value}`);
  }
}

onMounted(() => {
  initialize();
});

// Update local state if parent changes modelValue
watch(() => props.modelValue, (newVal) => {
  const source = newVal || '';
  if (source.includes(' ')) {
    const [newDate, newTime] = source.split(' ');
    if (newDate && newDate !== date.value) date.value = newDate;
    if (newTime && newTime.substring(0, 5) !== time.value) {
       time.value = newTime.substring(0, 5);
    }
  }
});
</script>

<style scoped>
.rounded-borders {
  border-radius: 12px;
}
</style>

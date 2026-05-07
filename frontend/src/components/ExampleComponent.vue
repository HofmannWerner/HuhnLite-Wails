<template>
  <div>
    <p>{{ title }}</p>
    <ul>
      <li v-for="todo in todos" :key="todo.id" @click="increment">
        {{ todo.id }} - {{ todo.content }}
      </li>
    </ul>
    <p>Count: {{ todoCount }} / {{ meta.totalCount }}</p>
    <p>Active: {{ active ? 'yes' : 'no' }}</p>
    <p>Clicks on todos: {{ clickCount }}</p>
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

import { computed, ref } from 'vue';
import type { Todo, Meta } from './models';

interface Props {
  title: string;
  todos?: Todo[];
  meta: Meta;
  active: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  todos: () => [],
});

const clickCount = ref(0);
function increment() {
  clickCount.value += 1;
  return clickCount.value;
}

const todoCount = computed(() => props.todos.length);
</script>

import { ref, onUnmounted, Ref, watch } from 'vue';
import { api } from '../boot/api';
import { useSessionStore } from '../stores/session';
import { Notify } from 'quasar';

export function useResizableColumns(gridId?: string | Ref<string | undefined>) {
  const sessionStore = useSessionStore();
  const columnWidths = ref<Record<string, number>>({});
  const isResizing = ref<string | null>(null);

  const getGridId = () => {
    if (!gridId) return undefined;
    return typeof gridId === 'string' ? gridId : gridId.value;
  };

  // Reload widths if user logs in/out
  watch(() => sessionStore.username, () => {
    if (getGridId()) {
      loadWidths();
    }
  });

  let startX = 0;
  let startWidth = 0;
  let saveTimeout: any = null;

  const startResize = (event: PointerEvent, colName: string) => {
    isResizing.value = colName;
    startX = event.pageX;
    startWidth = columnWidths.value[colName] || 150;

    document.addEventListener('pointermove', onPointerMove);
    document.addEventListener('pointerup', onPointerUp);
    
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  };

  const onPointerMove = (event: PointerEvent) => {
    if (!isResizing.value) return;

    const diff = event.pageX - startX;
    const newWidth = Math.max(30, startWidth + diff);
    
    columnWidths.value = {
      ...columnWidths.value,
      [isResizing.value]: newWidth
    };
  };

  const onPointerUp = () => {
    if (!isResizing.value) return;
    
    isResizing.value = null;
    document.removeEventListener('pointermove', onPointerMove);
    document.removeEventListener('pointerup', onPointerUp);
    
    document.body.style.cursor = '';
    document.body.style.userSelect = '';

    if (getGridId()) {
      debouncedSave();
    }
  };

  const debouncedSave = () => {
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(saveWidths, 1000);
  };

  const getUsername = () => sessionStore.username || 'default_user';

  const saveWidths = async () => {
    const id = getGridId();
    const username = getUsername();
    if (!id) return;
    try {
      await api.post('/api/user-status', {
        USERNAME: username,
        KEY: `grid_widths_${id}`,
        VALUE: JSON.stringify(columnWidths.value)
      });
    } catch (err) {
      console.error('Failed to save column widths:', err);
    }
  };

  const loadWidths = async () => {
    const id = getGridId();
    const username = getUsername();
    if (!id) return;
    try {
      const res = await api.get(`/api/user-status/grid_widths_${id}?username=${username}`);
      if (res.data?.value) {
        const saved = JSON.parse(res.data.value);
        Object.assign(columnWidths.value, saved);
      }
    } catch (err) {
      console.error('Failed to load column widths:', err);
    }
  };

  const initWidths = async (columns: any[]) => {
    // 1. Set defaults
    columns.forEach(col => {
      if (col.name && !columnWidths.value[col.name]) {
        let defaultWidth = 150;
        if (col.style && typeof col.style === 'string') {
          const match = col.style.match(/width:\s*(\d+)px/);
          if (match) defaultWidth = parseInt(match[1]);
        }
        columnWidths.value[col.name] = defaultWidth;
      }
    });

    // 2. Load from backend if gridId is provided
    if (getGridId()) {
      await loadWidths();
    }
  };

  onUnmounted(() => {
    document.removeEventListener('pointermove', onPointerMove);
    document.removeEventListener('pointerup', onPointerUp);
    if (saveTimeout) clearTimeout(saveTimeout);
  });

  return {
    columnWidths,
    isResizing,
    startResize,
    initWidths,
    saveWidths,
    loadWidths
  };
}

import { defineStore } from 'pinia';
import { ref, computed, watch } from 'vue';
import { api } from 'src/boot/api';
import { Dark } from 'quasar';

export const useSessionStore = defineStore('session', () => {
  // --- State ---
  const now = new Date();
  const initialDate = now.toISOString().split('T')[0];
  const initialTime = now.toTimeString().substring(0, 5);

  const workingTimestamp = ref(`${initialDate} ${initialTime}`);
  const selectedLanguage = ref(localStorage.getItem('selectedLanguage') || 'de');
  const darkMode = ref(localStorage.getItem('darkMode') === 'true');

  // Synchronisiere Quasar Dark Mode reaktiv mit dem Store-State
  watch(darkMode, (val) => {
    Dark.set(val);
    localStorage.setItem('darkMode', val ? 'true' : 'false');
  }, { immediate: true });

  const username = ref<string | null>(null);
  const klarname = ref<string | null>(null);
  const profile_kz = ref<string | null>(null);
  const isLoggedIn = ref(false);
  const authEnabled = ref(true);
  const loginDismissed = ref(false);

  const permissions = ref({
    dashboard: false,
    herden_verwalten: false,
    einrichtungen_verwalten: false,
    personen_verwalten: false,
    buchungen_erfassen: false,
    auswertungen_anzeigen: false,
    sql_struktur_verwalten: false,
    benutzer_profile: false,
    parameter_editieren: false,
    kosten_verwalten: false,
    tabellen_anzeigen: false,
    texte_verwalten: false,
    system_verwaltung: false,
    backup_erstellen: false
  });

  // --- Getters ---
  const can = computed(() => (permission: keyof typeof permissions.value): boolean => {
    if (!authEnabled.value) return true;
    return isLoggedIn.value && !!permissions.value[permission];
  });

  // --- Helper ---
  async function savePreference(key: string, value: string) {
    if (!isLoggedIn.value || !username.value) return;
    try {
      await api.post('/api/user-state', {
        USERNAME: username.value,
        KEY: key,
        VALUE: value
      });
    } catch (err) {
      console.error(`Failed to save preference ${key}:`, err);
    }
  }

  // --- Actions ---
  function setTimestamp(val: string) {
    workingTimestamp.value = val;
  }

  function setLanguage(lang: string) {
    selectedLanguage.value = lang;
    localStorage.setItem('selectedLanguage', lang);
    savePreference('selectedLanguage', lang);
  }

  function setDarkMode(val: boolean) {
    darkMode.value = val;
    // Dark.set(val) und localStorage werden jetzt über den watch erledigt
    savePreference('darkMode', val ? 'true' : 'false');
  }

  async function loadPreferences() {
    if (!isLoggedIn.value || !username.value) return;
    try {
      const [langRes, darkRes] = await Promise.all([
        api.get(`/api/user-state/selectedLanguage?username=${username.value}`),
        api.get(`/api/user-state/darkMode?username=${username.value}`)
      ]);

      if (langRes.data?.value) {
        selectedLanguage.value = langRes.data.value;
        localStorage.setItem('selectedLanguage', langRes.data.value);
      }
      if (darkRes.data?.value) {
        const isDark = darkRes.data.value === 'true';
        darkMode.value = isDark;
        Dark.set(isDark);
        localStorage.setItem('darkMode', darkRes.data.value);
      }
    } catch (err) {
      console.error('Failed to load user preferences:', err);
    }
  }

  function setSession(data: any) {
    username.value = data.username || data.USERNAME;
    klarname.value = data.klarname || data.KLARNAME;
    profile_kz.value = data.profile_kz || data.PROFILE_KZ;
    permissions.value = data.permissions || data.PERMISSIONS || permissions.value;
    isLoggedIn.value = true;
    loginDismissed.value = false;
    
    // Preferences nach Login laden
    loadPreferences();
  }

  function logout() {
    username.value = null;
    klarname.value = null;
    profile_kz.value = null;
    isLoggedIn.value = false;
    loginDismissed.value = false;

    if (!authEnabled.value) {
      setAdminSession();
    }
  }

  function setAdminSession() {
    username.value = "Test-Admin";
    klarname.value = "Administrator";
    profile_kz.value = "A";
    Object.keys(permissions.value).forEach(key => {
      (permissions.value as any)[key] = true;
    });
    isLoggedIn.value = true;
    loginDismissed.value = false;
    loadPreferences();
  }

  function triggerLogin() {
    loginDismissed.value = false;
  }

  function dismissLogin() {
    loginDismissed.value = true;
  }

  return {
    workingTimestamp,
    selectedLanguage,
    darkMode,
    username,
    klarname,
    profile_kz,
    permissions,
    isLoggedIn,
    authEnabled,
    loginDismissed,
    can,
    setTimestamp,
    setLanguage,
    setDarkMode,
    setSession,
    logout,
    setAdminSession,
    triggerLogin,
    dismissLogin
  };
});

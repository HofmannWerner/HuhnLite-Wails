import { boot } from 'quasar/wrappers';
import axios from 'axios';
import type { AxiosInstance } from 'axios';

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $api: AxiosInstance;
  }
}

// Dynamically determine the base URL for the API
let baseURL = 'http://localhost:8080';
if (typeof window !== 'undefined' && window.location) {
  const host = window.location.host;
  const protocol = window.location.protocol;
  // If we are running inside the Wails desktop environment, use localhost:8080 (or proxy)
  if (protocol === 'wails:' || host === 'wails.localhost') {
    baseURL = 'http://localhost:8080';
  } else {
    // If accessed via web browser, use the serving server's origin (IP and port)
    baseURL = window.location.origin;
  }
}

const clientId = Math.random().toString(36).substring(2, 11);
export const getClientId = () => clientId;

const api = axios.create({ baseURL });

export default boot(async ({app}) => {
  // Lazy import to avoid circular dependency with store
  const {useSessionStore} = await import('../stores/session');
  const sessionStore = useSessionStore();

  // Add an interceptor to include language and client ID in every request
  api.interceptors.request.use((config) => {
    // Add language as a query parameter 'lang'
    if (config.params === undefined) {
      config.params = {};
    }
    config.params.lang = localStorage.getItem('selectedLanguage') || sessionStore.selectedLanguage || 'de';

    if (config.headers) {
      config.headers['X-Client-ID'] = clientId;
    }

    return config;
  });

  // Global Response Interceptor: Ensure all keys are available in UPPERCASE
  api.interceptors.response.use((response) => {
    if (response.data && typeof response.data === 'object') {
      const transform = (obj: any) => {
        if (!obj || typeof obj !== 'object') return;

        if (Array.isArray(obj)) {
          obj.forEach(transform);
        } else {
          const keys = Object.keys(obj);
          keys.forEach(key => {
            const upperKey = key.toUpperCase();
            const val = obj[key];

            if (!(upperKey in obj)) {
              obj[upperKey] = val;
            }

            if (val && typeof val === 'object') {
              transform(val);
            }
          });
        }
      };
      transform(response.data);
    }
    return response;
  });

  // for use inside Vue files (Options API) through this.$api
  app.config.globalProperties.$api = api;
  // for use inside Vue files (Composition API)
  app.provide('api', api);
});

export { api };

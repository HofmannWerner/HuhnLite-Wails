import { boot } from 'quasar/wrappers';
import axios from 'axios';
import type { AxiosInstance } from 'axios';

declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $api: AxiosInstance;
  }
}

// Create an axios instance with a base URL
const api = axios.create({baseURL: 'http://localhost:8080'});

export default boot(async ({app}) => {
  // Lazy import to avoid circular dependency with store
  const {useSessionStore} = await import('../stores/session');
  const sessionStore = useSessionStore();

  // Add an interceptor to include the language in every request
  api.interceptors.request.use((config) => {
    // Add language as a query parameter 'lang'
    if (config.params === undefined) {
      config.params = {};
    }
    config.params.lang = localStorage.getItem('selectedLanguage') || sessionStore.selectedLanguage || 'de';

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

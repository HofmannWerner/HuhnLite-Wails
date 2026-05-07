import { defineRouter } from '#q-app/wrappers';
import {
  createMemoryHistory,
  createRouter,
  createWebHashHistory,
  createWebHistory,
} from 'vue-router';
import routes from './routes';
import {useSessionStore} from '../stores/session';


/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */

export default defineRouter(function (/* { store, ssrContext } */) {
  const sessionStore = useSessionStore();
  const createHistory = process.env.SERVER
    ? createMemoryHistory
    : process.env.VUE_ROUTER_MODE === 'history'
      ? createWebHistory
      : createWebHashHistory;

  const Router = createRouter({
    scrollBehavior: () => ({ left: 0, top: 0 }),
    routes,
    history: createHistory(process.env.VUE_ROUTER_BASE),
  });

  // Navigation Guard: Berechtigungsprüfung
  Router.beforeEach((to) => {
    // Falls Testmodus (authEnabled=false) oder eingeloggt, Berechtigungen prüfen
    const requiredPermission = to.meta.permission as string;

    console.log(`Router: Navigiere zu ${to.path}, Benötigtes Recht: ${requiredPermission || 'keines'}`);
    console.log(`Router: Status - Eingeloggt: ${sessionStore.isLoggedIn}, Auth aktiv: ${sessionStore.authEnabled}`);

    // Wenn kein spezielles Recht erforderlich ist (z.B. 404), weitergehen
    if (!requiredPermission) {
      console.log('Router: Kein Recht benötigt, erlaube Navigation.');
      return true;
    }

    // Wenn eingeloggt oder Test-Admin aktiviert ist im sessionStore
    const permissions = (sessionStore.permissions as any);

    if (sessionStore.isLoggedIn) {
      if (permissions[requiredPermission]) {
        console.log(`Router: Recht ${requiredPermission} vorhanden. OK.`);
        return true;
      } else {
        console.warn(`Router: Recht ${requiredPermission} FEHLT! Zurück zum Dashboard.`);
        // Zugriff verweigert -> zurück zum Dashboard
        return {path: '/'};
      }
    } else {
      // Nicht eingeloggt (und auth ist aktiv)
      console.log('Router: Nicht eingeloggt.');
      if (to.path === '/') {
        console.log('Router: Erlaube Dashboard (für Login-Modal).');
        return true;
      } else {
        console.warn('Router: Navigation blockiert - Login erforderlich.');
        return {path: '/'};
      }
    }
  });

  return Router;
});


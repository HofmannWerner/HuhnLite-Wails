import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      {path: '', component: () => import('pages/IndexPage.vue'), meta: {permission: 'dashboard'}},
      {path: 'mist/:tab?', component: () => import('pages/MistPage.vue'), meta: {permission: 'tabellen_anzeigen'}},
      {path: 'herden', component: () => import('pages/HerdenPage.vue'), meta: {permission: 'herden_verwalten'}},
      {
        path: 'einrichtungen',
        component: () => import('pages/EinrichtungenPage.vue'),
        meta: {permission: 'einrichtungen_verwalten'}
      },
      {path: 'person', component: () => import('pages/PersonPage.vue'), meta: {permission: 'personen_verwalten'}},
      {path: 'buchungen', component: () => import('pages/BuchungenPage.vue'), meta: {permission: 'buchungen_erfassen'}},
      {path: 'kosten', component: () => import('pages/KostenPage.vue'), meta: {permission: 'kosten_verwalten'}},
      {
        path: 'textverwaltung',
        component: () => import('pages/Textverwaltung.vue'),
        meta: {permission: 'texte_verwalten'}
      },
      {path: 'settings', component: () => import('pages/FirmenPage.vue'), meta: {permission: 'parameter_editieren'}},
      {
        path: 'reports',
        component: () => import('pages/DynamicReportPage.vue'),
        meta: {permission: 'auswertungen_anzeigen'}
      },
      {
        path: 'management',
        component: () => import('pages/AdminDirect.vue'),
        meta: {permission: 'sql_struktur_verwalten'}
      },
      {path: 'admin', component: () => import('pages/AdminPage.vue'), meta: {permission: 'sql_struktur_verwalten'}},
      {path: 'showtv', component: () => import('pages/ShowTVPage.vue'), meta: {permission: 'parameter_editieren'}},
      {path: 'benutzer', component: () => import('pages/BenutzerPage.vue'), meta: {permission: 'benutzer_profile'}},
      {path: 'profile', component: () => import('pages/ProfilePage.vue'), meta: {permission: 'benutzer_profile'}},
      {
        path: 'tierumstallung',
        component: () => import('pages/TierUmstallungPage.vue'),
        meta: {permission: 'buchungen_erfassen'}
      },
      {path: 'restore', component: () => import('pages/RestorePage.vue'), meta: {permission: 'system_verwaltung'}},
      {path: 'mandanten', component: () => import('pages/MandantenPage.vue'), meta: {permission: 'system_verwaltung'}}



    ]
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
];

export default routes;

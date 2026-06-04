import { boot } from 'quasar/wrappers';
import { createI18n } from 'vue-i18n';
import deMessages from '../i18n/de.json';
import enMessages from '../i18n/en.json';
import itMessages from '../i18n/it.json';

export default boot(({ app }) => {
  // Get preferred language from localStorage or default to German
  const locale = localStorage.getItem('selectedLanguage') || 'de';

  const i18n = createI18n({
    locale: locale,
    fallbackLocale: 'de',
    legacy: false, // Set to false to use Composition API
    globalInjection: true,
    messages: {
      de: deMessages,
      en: enMessages,
      it: itMessages
    }
  });

  // Set i18n instance on app
  app.use(i18n);
});

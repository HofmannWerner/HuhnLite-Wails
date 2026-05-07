import { createPinia } from 'pinia';
import type { Router } from 'vue-router';

/*
 * When adding new properties to our store, we should also
 * extend the `PiniaCustomProperties` interface.
 * @see https://pinia.vuejs.org/core-concepts/plugins.html#typing-new-store-properties
 */
declare module 'pinia' {
  export interface PiniaCustomProperties {
    readonly router: Router;
  }
}

export default () => {
  const pinia = createPinia();

  // You can add Pinia plugins here
  // pinia.use(SomePiniaPlugin)

  return pinia;
};

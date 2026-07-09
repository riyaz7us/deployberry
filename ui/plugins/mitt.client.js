export default defineNuxtPlugin({
  setup(nuxtApp) {
    const handlers = {};
    const bus = {
      $on(event, handler) {
        if (!handlers[event]) handlers[event] = [];
        handlers[event].push(handler);
      },
      $emit(event, payload) {
        if (handlers[event]) {
          handlers[event].forEach(handler => handler(payload));
        }
      },
      $off(event, handler) {
        if (handlers[event]) {
          handlers[event] = handlers[event].filter(h => h !== handler);
        }
      }
    };
    return {
      provide: { bus },
    };
  },
});

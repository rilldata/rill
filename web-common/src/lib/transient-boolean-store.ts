import { writable } from "svelte/store";

export default function transientBooleanStore(duration = 400) {
  const { subscribe, set } = writable(false);
  let timer;
  return {
    subscribe,
    flip() {
      clearTimeout(timer);
      set(true);
      setTimeout(() => {
        set(false);
      }, duration);
    },
  };
}

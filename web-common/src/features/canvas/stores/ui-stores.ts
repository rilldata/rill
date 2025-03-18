import { derived, get, writable } from "svelte/store";

const MIN_DELTA = 12;

export const dropZone = (() => {
  const zoneId = writable<string | null>(null);

  const mouseDelta = writable(0);

  return {
    subscribe: derived([zoneId, mouseDelta], ([$dropZone, $mouseDelta]) => {
      if ($mouseDelta < MIN_DELTA) {
        return null;
      } else {
        return $dropZone;
      }
    }).subscribe,
    set: (id: string) => {
      zoneId.set(id);
    },
    setMouseDelta: (amount: number) => {
      mouseDelta.set(amount);
    },
    clear: () => {
      zoneId.set(null);
    },
  };
})();

export const hoveredDivider = (() => {
  const idStore = writable<string | null>(null);
  let timeout: ReturnType<typeof setTimeout> | null = null;
  const active = writable(false);

  return {
    id: idStore,
    isActive: active,
    set: (id: string) => {
      if (timeout) clearTimeout(timeout);
      idStore.set(id);
    },
    reset: (timeoutLength = 50) => {
      console.log("reset called");
      // if (!get(active)) return;
      timeout = setTimeout(() => {
        active.set(false);
        idStore.set(null);
      }, timeoutLength);
    },
    setActive: (dividerId: string, value: boolean) => {
      if (dividerId !== get(idStore)) return;
      active.set(value);
    },
  };
})();

export const activeDivider = (() => {
  const { subscribe, set } = writable<string | null>(null);

  return {
    subscribe,
    set,
    reset: () => set(null),
  };
})();

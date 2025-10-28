import { getCanvasQueryOptions } from "@rilldata/web-common/features/canvas/selector.ts";
import type { SearchParamsStore } from "@rilldata/web-common/features/canvas/stores/canvas-entity.ts";
import { TimeControls } from "@rilldata/web-common/features/canvas/stores/time-control.ts";
import { createQuery } from "@tanstack/svelte-query";
import { get, type Readable, writable } from "svelte/store";

export function getCanvasDefaultUrlParams(canvasNameStore: Readable<string>) {
  const urlSearchParamsStore = writable(new URLSearchParams());
  const searchParamsStore: SearchParamsStore = (() => {
    return {
      subscribe: urlSearchParamsStore.subscribe,
      set: (key: string, value: string | undefined, checkIfSet = false) => {
        const urlSearchParams = get(urlSearchParamsStore);
        if (checkIfSet && urlSearchParams.has(key)) return false;

        if (value === undefined || value === null || value === "") {
          urlSearchParams.delete(key);
        } else {
          urlSearchParams.set(key, value);
        }
        urlSearchParamsStore.set(urlSearchParams);
        return true;
      },
      clearAll: () => {
        urlSearchParamsStore.set(new URLSearchParams());
      },
    };
  })();

  const specStore = createQuery(getCanvasQueryOptions(canvasNameStore));

  canvasNameStore.subscribe((canvasName) => {
    new TimeControls(specStore, searchParamsStore, undefined, canvasName);
  });

  return urlSearchParamsStore;
}

/**
 * The `errorStore` holds the state of any runtime errors, which
 * the `ErrorBoundary` component catches and routes to the `ErrorPage`.
 */

import { derived, writable, Writable } from "svelte/store";

export interface ErrorStoreState {
  statusCode: number | null;
  header: string;
  body: string;
}

export interface ErrorStore extends Writable<ErrorStoreState> {
  reset: () => void;
}

const createErrorStore = (): ErrorStore => {
  const { subscribe, set, update } = writable({
    statusCode: null,
    header: "",
    body: "",
  });

  const reset = (): void => {
    set({ statusCode: null, header: "", body: "" });
  };

  return { subscribe, set, update, reset };
};

export const errorStore = createErrorStore();

export const isErrorStoreEmpty = derived(errorStore, ($errorStore) => {
  const { statusCode, header, body } = $errorStore;
  return statusCode === null && header === "" && body === "";
});

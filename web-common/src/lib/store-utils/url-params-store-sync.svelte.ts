import { page } from "$app/state";
import { untrack } from "svelte";

export interface UrlParamsStore {
  setUrlParams(urlParams: URLSearchParams): void;
  applyFilterToParams(urlParams: URLSearchParams): void;
  specLoaded: boolean;
  dataLoaded: boolean;
}

export function syncStoreWithSource(
  store: UrlParamsStore,
  sync: (newUrlParams: URLSearchParams) => Promise<void>,
  syncFromUrl = true,
  log = false,
) {
  let lock = false;

  if (syncFromUrl) {
    $effect(() => {
      // Read all dependencies first so the subscription survives the guard.
      const currentUrl = page.url;

      if (!store.specLoaded || lock) return;
      lock = true;

      const newUrlParams = new URLSearchParams(currentUrl.searchParams);

      if (log) console.log("sync:fromUrl", newUrlParams.toString());
      // No need to safeguard against unchanged url.
      // It should already happen in setUrlParams since it will have other callers.
      untrack(() => store.setUrlParams(newUrlParams));

      lock = false;
    });
  }

  let prevStateParams = new URLSearchParams();
  $effect(() => {
    // Read all dependencies first so the subscription survives the guard.
    const curStateParams = new URLSearchParams();
    store.applyFilterToParams(curStateParams);

    if (
      !store.dataLoaded ||
      lock ||
      curStateParams.toString() === prevStateParams.toString()
    )
      return;
    lock = true;

    const currentUrlParams = untrack(() =>
      syncFromUrl
        ? page.url.searchParams
        : new URLSearchParams(prevStateParams),
    );
    prevStateParams = curStateParams;

    const newUrlParams = new URLSearchParams(currentUrlParams);
    untrack(() => {
      store.applyFilterToParams(newUrlParams);
    });

    if (log) {
      console.log(
        "sync:toUrl",
        newUrlParams.toString() === currentUrlParams.toString(),
        newUrlParams.toString(),
      );
    }
    if (newUrlParams.toString() === currentUrlParams.toString()) {
      lock = false;
      return;
    }
    try {
      // Do not react to `sync` method changes
      const syncPromise = untrack(() => sync(newUrlParams));
      if (!syncPromise.then) {
        lock = false;
        return;
      }

      void syncPromise.then(
        () => (lock = false),
        () => (lock = false),
      );
    } catch {
      lock = false;
    }
  });
}

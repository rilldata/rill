import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params.ts";
import { page } from "$app/state";
import { untrack } from "svelte";

export interface UrlParamsStore {
  curStateParams: URLSearchParams;
  curSetParams: URLSearchParams;
  setUrlParams(urlParams: URLSearchParams): void;
  applyFilterToParams(urlParams: URLSearchParams): void;
}

export function syncStoreWithSource(
  store: UrlParamsStore,
  sync: (newUrlParams: URLSearchParams) => Promise<void>,
  readyGetter: () => boolean,
  defaultUrlParamsGetter?: () => URLSearchParams | undefined,
  skipUrlSync = false,
) {
  let lock = false;
  let prevUrlSearch = "";

  if (!skipUrlSync) {
    $effect(() => {
      // Read all dependencies first so the subscription survives the guard.
      const currentUrl = page.url;
      const defaultUrlParams = untrack(() =>
        defaultUrlParamsGetter ? defaultUrlParamsGetter() : undefined,
      );
      const ready = readyGetter();

      console.log(`syncStoreWithSource:fromUrl ready=${ready} lock=${lock}`);
      if (!ready || lock) return;
      lock = true;

      const newUrlParams = new URLSearchParams(currentUrl.searchParams);
      if (defaultUrlParams) {
        defaultUrlParams.forEach((value, key) => {
          if (newUrlParams.has(key)) return;
          newUrlParams.set(key, value);
        });
      }

      console.log(
        "syncStoreWithSource:fromUrl",
        newUrlParams.toString() === prevUrlSearch,
        newUrlParams.toString(),
      );
      if (newUrlParams.toString() === prevUrlSearch) {
        lock = false;
        return;
      }
      prevUrlSearch = newUrlParams.toString();

      untrack(() => store.setUrlParams(newUrlParams));

      lock = false;
    });
  }

  let prevStateParams = "";
  $effect(() => {
    // Read all dependencies first so the subscription survives the guard.
    const curStateParams = store.curStateParams;
    const defaultUrlParams = untrack(() =>
      defaultUrlParamsGetter ? defaultUrlParamsGetter() : undefined,
    );
    const ready = readyGetter();

    console.log(`syncStoreWithSource:toUrl ready=${ready} lock=${lock}`);
    if (!ready || lock || curStateParams.toString() === prevStateParams) return;
    lock = true;
    prevStateParams = curStateParams.toString();

    const currentUrlParams = untrack(() => page.url.searchParams);
    let newUrlParams = new URLSearchParams(currentUrlParams);
    if (defaultUrlParams) {
      newUrlParams = cleanUrlParams(newUrlParams, defaultUrlParams);
    }
    untrack(() => {
      store.applyFilterToParams(newUrlParams);
    });

    console.log(
      "syncStoreWithSource:toUrl",
      newUrlParams.toString() === currentUrlParams.toString(),
      newUrlParams.toString(),
    );
    if (newUrlParams.toString() === currentUrlParams.toString()) {
      lock = false;
      return;
    }

    try {
      const syncPromise = sync(newUrlParams);
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

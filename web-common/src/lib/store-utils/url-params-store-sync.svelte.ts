import { copyParamsToTarget } from "@rilldata/web-common/lib/url-utils.ts";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params.ts";
import { page } from "$app/state";
import { untrack } from "svelte";

interface UrlParamsStore {
  curStateParams: URLSearchParams;
  curSetParams: URLSearchParams;
  setUrlParams(urlParams: URLSearchParams): void;
}

export function syncStoreWithSource(
  store: UrlParamsStore,
  sync: (newUrlParams: URLSearchParams) => Promise<void>,
  readyGetter: () => boolean,
  defaultUrlParamsGetter: () => URLSearchParams | undefined,
) {
  let lock = false;
  let prevUrlSearch = "";

  $effect(() => {
    // Read all dependencies first so the subscription survives the guard.
    const currentUrl = page.url;
    const defaultUrlParams = untrack(() =>
      defaultUrlParamsGetter ? defaultUrlParamsGetter() : undefined,
    );
    const ready = readyGetter();

    if (!ready || lock) return;
    lock = true;

    const newUrlParams = new URLSearchParams(currentUrl.searchParams);
    if (defaultUrlParams) {
      defaultUrlParams.forEach((value, key) => {
        if (newUrlParams.has(key)) return;
        newUrlParams.set(key, value);
      });
    }

    if (newUrlParams.toString() === prevUrlSearch) {
      lock = false;
      return;
    }
    prevUrlSearch = newUrlParams.toString();

    console.log("syncStoreWithSource:fromUrl", newUrlParams.toString());
    untrack(() => store.setUrlParams(newUrlParams));

    lock = false;
  });

  $effect(() => {
    // Read all dependencies first so the subscription survives the guard.
    const curStateParams = store.curStateParams;
    const curSetParams = untrack(() => store.curSetParams);
    const defaultUrlParams = untrack(() =>
      defaultUrlParamsGetter ? defaultUrlParamsGetter() : undefined,
    );
    const ready = readyGetter();

    if (!ready || lock || curStateParams.toString() === curSetParams.toString())
      return;
    lock = true;

    const currentUrlParams = untrack(() => page.url.searchParams);
    let newUrlParams = new URLSearchParams(currentUrlParams);
    copyParamsToTarget(curStateParams, newUrlParams);
    if (defaultUrlParams) {
      newUrlParams = cleanUrlParams(newUrlParams, defaultUrlParams);
    }

    if (newUrlParams.toString() === currentUrlParams.toString()) {
      lock = false;
      return;
    }

    console.log(
      "syncStoreWithSource:toUrl",
      newUrlParams.toString(),
      currentUrlParams.toString(),
    );
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

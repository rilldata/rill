import { navigating, page, updated } from "$app/state";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";

/**
 * Recovers tabs that are still running a previous build after a deploy.
 *
 * A deploy replaces the content-hashed files under `_app/immutable/`, so a tab running the old build fails
 * as soon as it lazy-loads code that is not in memory yet: "Failed to fetch dynamically imported module",
 * "Importing a module script failed." (Safari) or "Unable to preload CSS for ...".
 * SvelteKit recovers from this on its own only when the failed import is a route node.
 * Dynamic `import()` calls in app and library code (lazily loaded dialogs, Prism, jsPDF, ...) reject as
 * unhandled errors and leave the page half-broken until the user reloads it.
 *
 * Vite dispatches `vite:preloadError` on `window` for every failed dynamic import and CSS preload.
 * This listens for that event, confirms through SvelteKit's version check that a new build was deployed,
 * and does a single full-page navigation to pick it up.
 *
 * Embeds are left as they are: they keep their access token in memory and strip it from the page URL,
 * so a full-page navigation would lose it. SvelteKit's own recovery has the same limitation there.
 */

/** Minimum time between two automatic reloads of a tab, so a build that keeps failing cannot cause a reload loop. */
export const STALE_BUILD_RELOAD_COOLDOWN_MS = 60_000;

const STALE_BUILD_RELOAD_AT_KEY = "rill:stale-build-reload-at";

export type StaleBuildReloadDeps = {
  /** Resolves to true when a newer build than the running one has been deployed. */
  isUpdated: () => Promise<boolean>;
  /** The page to load, or null when the page must not be reloaded. */
  getReloadUrl: () => URL | null;
  /** Does a full-page navigation to the URL. */
  navigate: (url: URL) => void;
  /** Keeps the time of the last reload across reloads of the tab. */
  storage: Pick<Storage, "getItem" | "setItem">;
  now: () => number;
};

export function createStaleBuildReloadHandler(deps: StaleBuildReloadDeps) {
  let checking = false;
  // Fallback for when storage is unavailable, for example in an iframe that is denied storage access.
  let lastReloadInMemory: number | null = null;

  function readLastReload(): number | null {
    try {
      const value = Number(deps.storage.getItem(STALE_BUILD_RELOAD_AT_KEY));
      if (value > 0) return value;
    } catch {
      // Storage access is denied; fall through to the in-memory value.
    }
    return lastReloadInMemory;
  }

  function writeLastReload(at: number) {
    lastReloadInMemory = at;
    try {
      deps.storage.setItem(STALE_BUILD_RELOAD_AT_KEY, String(at));
    } catch {
      // Storage access is denied; the in-memory value still guards this page.
    }
  }

  return async function handlePreloadError(): Promise<void> {
    // Several imports can fail at once; one check and one navigation cover all of them.
    if (checking) return;
    const url = deps.getReloadUrl();
    if (url === null) return;

    checking = true;
    try {
      // The import may have failed for another reason (offline, blocked by an extension); a reload would not help then.
      if (!(await deps.isUpdated())) return;

      const now = deps.now();
      const lastReload = readLastReload();
      if (
        lastReload !== null &&
        now - lastReload < STALE_BUILD_RELOAD_COOLDOWN_MS
      ) {
        return;
      }

      writeLastReload(now);
      deps.navigate(url);
    } finally {
      checking = false;
    }
  };
}

/** Registers the `vite:preloadError` listener on `window`. Returns a function that removes it. */
export function registerStaleBuildReload(): () => void {
  const handlePreloadError = createStaleBuildReloadHandler({
    // Fetches `_app/version.json` and compares it with the running build. Always false in dev.
    isUpdated: () => updated.check(),
    getReloadUrl,
    navigate: (url) => {
      // Like SvelteKit's own fallback navigation: the current URL replaces the history entry, a new URL pushes one.
      location.href = url.href;
    },
    // Accessed lazily: reading `sessionStorage` itself throws in iframes that are denied storage access.
    storage: {
      getItem: (key) => sessionStorage.getItem(key),
      setItem: (key, value) => sessionStorage.setItem(key, value),
    },
    now: Date.now,
  });

  // The event's default is not prevented: the error keeps propagating,
  // so the code that awaited the import fails instead of continuing with an undefined module.
  const listener = () => void handlePreloadError();
  window.addEventListener("vite:preloadError", listener);
  return () => window.removeEventListener("vite:preloadError", listener);
}

/** The destination of an in-flight navigation, otherwise the current page; null inside an embed. */
export function getReloadUrl(): URL | null {
  // Embeds keep their access token in memory and strip it from the page URL, so a reload would lose it.
  if (EmbedStore.isEmbedded()) return null;
  return navigating.to?.url ?? page.url;
}

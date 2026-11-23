import { navigating, updated } from "$app/state";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import type { BeforeNavigate } from "@sveltejs/kit";

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
 * Two parts cover this:
 * - `navigateToNewBuild` is SvelteKit's documented pattern: once polling (`kit.version.pollInterval`)
 *   has found a newer build, the next client-side navigation becomes a full page load into it.
 * - `registerStaleBuildReload` is the backstop for imports that fail before polling noticed the deploy,
 *   most of them inside a page, with no navigation to hook into.
 *
 * Embeds are left as they are: they keep their access token in memory and strip it from the page URL,
 * so a full page load would lose it. SvelteKit's own recovery has the same limitation there.
 */

/** Minimum time between two automatic reloads of a tab, so a build that keeps failing cannot cause a reload loop. */
export const STALE_BUILD_RELOAD_COOLDOWN_MS = 60_000;

const STALE_BUILD_RELOAD_AT_KEY = "rill:stale-build-reload-at";

/** For `beforeNavigate` in the root layout: turns the navigation into a full page load once a newer build is deployed. */
export function navigateToNewBuild({ to, willUnload }: BeforeNavigate) {
  if (!updated.current || willUnload || !to || EmbedStore.isEmbedded()) {
    return;
  }
  loadPage(to.url);
}

/**
 * Listens for Vite's `vite:preloadError`, which it dispatches on `window` for every failed dynamic import and CSS preload,
 * confirms through SvelteKit's version check that a newer build is deployed,
 * and does one full page load to the destination of an in-flight navigation, otherwise of the current page.
 * Returns a function that removes the listener.
 */
export function registerStaleBuildReload(): () => void {
  let checking = false;
  // Fallback for when storage access is denied, for example when the user blocks all site data.
  let lastReloadInMemory: number | null = null;

  async function handlePreloadError() {
    // Several imports can fail at once; one check and one navigation cover all of them.
    if (checking || EmbedStore.isEmbedded()) return;
    const destination = navigating.to?.url;

    checking = true;
    try {
      // The import may have failed for another reason (offline, blocked by an extension); a reload would not help then.
      // Fetches `_app/version.json` and compares it with the running build. Always false in dev.
      if (!(await updated.check())) return;

      const now = Date.now();
      let lastReload = lastReloadInMemory;
      try {
        lastReload =
          Number(sessionStorage.getItem(STALE_BUILD_RELOAD_AT_KEY)) ||
          lastReload;
      } catch {
        // Storage access is denied; the in-memory value still guards this page.
      }
      if (
        lastReload !== null &&
        now - lastReload < STALE_BUILD_RELOAD_COOLDOWN_MS
      ) {
        return;
      }

      lastReloadInMemory = now;
      try {
        sessionStorage.setItem(STALE_BUILD_RELOAD_AT_KEY, String(now));
      } catch {
        // Storage access is denied; see above.
      }

      if (destination) loadPage(destination);
      else location.reload();
    } finally {
      checking = false;
    }
  }

  // The event's default is not prevented: the error keeps propagating,
  // so the code that awaited the import fails instead of continuing with an undefined module.
  const listener = () => void handlePreloadError();
  window.addEventListener("vite:preloadError", listener);
  return () => window.removeEventListener("vite:preloadError", listener);
}

/** Does a full page load of the URL, also when it is the current page. */
function loadPage(url: URL) {
  // Assigning a URL that is the current document apart from its fragment only scrolls to the fragment.
  // That is also the case for a back or forward navigation, where `location` already is the destination.
  if (url.href.split("#")[0] === location.href.split("#")[0]) {
    location.reload();
  } else {
    location.href = url.href;
  }
}

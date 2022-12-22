/*global  RILL_RUNTIME_URL */
import { browser } from "$app/environment";
import { createApplicationBuildMetadataStore } from "@rilldata/web-local/lib/application-state-stores/build-metadata";
import { createQueryHighlightStore } from "@rilldata/web-local/lib/application-state-stores/query-highlight-store";
import { setContext } from "svelte";

/** determined by Vite's define option. */
declare global {
  const RILL_RUNTIME_URL: string;
}

export const RuntimeUrl = RILL_RUNTIME_URL; // constant defined in svelte.config.js

/** This function will initialize the existing node stores and will connect them
 * to the Node server. It is best used in various application layouts to ensure that all children of the layout
 * get the same stores available to them. Once we have moved away from node & the resulting state model, we should
 * deprecate this function.
 */
export function initializeNodeStoreContexts() {
  const queryHighlight = createQueryHighlightStore();

  /** set build-specific metadata as a context.  */
  const buildMetadataStore = createApplicationBuildMetadataStore();
  setContext("rill:app:metadata", buildMetadataStore);

  /** Set the existing node stores, which are consumed through getContext within routes. */
  if (browser) {
    setContext("rill:app:query-highlight", queryHighlight);
  }
}

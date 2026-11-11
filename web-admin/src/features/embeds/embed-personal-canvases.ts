import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { derived } from "svelte/store";

/**
 * Whether embed users can manage personal canvases.
 * It requires an admin token (only issued for embeds with an external user ID, which identifies the owner),
 * navigation (to reach the personal canvas pages), and the `personal_canvases` feature flag.
 */
export const embedPersonalCanvasesEnabled = derived(
  [featureFlags.personalCanvases, featureFlags.embedAdminActions],
  ([personalCanvases, embedAdminActions]) =>
    personalCanvases &&
    embedAdminActions &&
    !!EmbedStore.getInstance()?.navigationEnabled,
);

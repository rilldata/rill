import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { derived } from "svelte/store";

/**
 * Whether embed users can manage their personal bookmarks.
 * It requires an admin token (only issued for embeds with an external user ID, which identifies the owner),
 * and the `embedded_bookmarks` feature flag.
 */
export const embedBookmarksEnabled = derived(
  [featureFlags.embeddedBookmarks, featureFlags.embedAdminActions],
  ([embeddedBookmarks, embedAdminActions]) =>
    embeddedBookmarks && embedAdminActions,
);

import {
  type DashboardInfo,
  getDashboardFromEmbedRoute,
} from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
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

/**
 * Returns the dashboard that bookmarks apply to for an embed route, or null if bookmarks are not supported there.
 * Bookmarks are only supported on project dashboards, not on personal canvases.
 */
export function getEmbedBookmarkResource(
  routeId: string | null,
  params: Record<string, string> | undefined,
): DashboardInfo | null {
  const dashboard = getDashboardFromEmbedRoute(routeId, params);
  return dashboard?.personal ? null : dashboard;
}

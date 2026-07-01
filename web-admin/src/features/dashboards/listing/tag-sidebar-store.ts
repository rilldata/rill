import { localStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage";

// Width bounds (px) for the dashboards tag sidebar, resizable via the divider
// between the tag list and the dashboards list.
export const DEFAULT_TAG_SIDEBAR_WIDTH = 200;
export const MIN_TAG_SIDEBAR_WIDTH = 160;
export const MAX_TAG_SIDEBAR_WIDTH = 400;

/**
 * Width of the tag sidebar on the project dashboards listing, controlled by the
 * resizable divider between the tags and the dashboards list. Persisted so the
 * split is preserved as the user navigates back and forth.
 */
export const dashboardsTagSidebarWidth = localStorageStore<number>(
  "dashboards-tag-sidebar-width",
  DEFAULT_TAG_SIDEBAR_WIDTH,
);

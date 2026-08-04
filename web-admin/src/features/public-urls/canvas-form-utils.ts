export {
  getCanvasFilters,
  hasCanvasFilters,
} from "@rilldata/web-common/features/canvas/filters/canvas-filter-expressions";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

/**
 * Returns the sanitized canvas state from the URL.
 * Removes filter parameters (f and f.*) so locked filters don't appear in the shared URL.
 * This ensures we do not leak hidden filter information to the URL recipient.
 */
export function getSanitizedCanvasStateUrl(currentUrl: URL): string {
  const searchParams = new URLSearchParams(currentUrl.search);
  const filterPrefix: string = ExploreStateURLParams.Filters; // "f"

  // Remove all filter-related parameters (f, f.metricsViewName, etc.)
  const keysToDelete: string[] = [];
  searchParams.forEach((_, key) => {
    if (key === filterPrefix || key.startsWith(`${filterPrefix}.`)) {
      keysToDelete.push(key);
    }
  });
  keysToDelete.forEach((key) => searchParams.delete(key));

  return searchParams.toString();
}

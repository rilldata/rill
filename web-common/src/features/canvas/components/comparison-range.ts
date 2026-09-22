import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { ComponentFilterProperties } from "./types";

// Special values of a component's `comparison_range`.
// Any other value is a comparison range such as `rill-PW` or a custom `<start>,<end>` pair.
export const COMPARISON_RANGE_INHERIT = "inherit";
export const COMPARISON_RANGE_NONE = "none";

export type ResolvedComparisonRange =
  | { mode: "inherit" }
  | { mode: "none" }
  | { mode: "local"; range: string };

/**
 * Resolves how a component compares against a previous period.
 * An absent `comparison_range` inherits the canvas comparison, except for components with a local time range:
 * those keep the legacy behaviour where `compare_tr` inside `time_filters` decides, and its absence means no comparison.
 */
export function resolveComparisonRange(
  spec: Partial<ComponentFilterProperties> | undefined,
): ResolvedComparisonRange {
  const explicit =
    typeof spec?.comparison_range === "string"
      ? spec.comparison_range.trim()
      : "";
  if (explicit) {
    if (explicit === COMPARISON_RANGE_INHERIT) return { mode: "inherit" };
    if (explicit === COMPARISON_RANGE_NONE) return { mode: "none" };
    return { mode: "local", range: explicit };
  }

  const timeFilters = new URLSearchParams(spec?.time_filters ?? "");
  if (timeFilters.has(ExploreStateURLParams.TimeRange)) {
    const legacyRange = timeFilters.get(
      ExploreStateURLParams.ComparisonTimeRange,
    );
    return legacyRange
      ? { mode: "local", range: legacyRange }
      : { mode: "none" };
  }

  return { mode: "inherit" };
}

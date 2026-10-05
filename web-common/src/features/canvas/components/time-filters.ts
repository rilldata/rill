import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

// A component's `time_filters` uses the explore URL vocabulary (`tr`, `compare_tr`, `grain`, `tz`).
// `inherit` is a special value for `tr` and `compare_tr` that follows the canvas:
//   (absent)                   inherit the canvas time range and comparison
//   tr=inherit                 canvas time range, comparison off
//   tr=inherit&compare_tr=X    canvas time range, compare against X
//   tr=7D                      local time range, comparison off
//   tr=7D&compare_tr=inherit   local time range, canvas comparison
//   tr=7D&compare_tr=X         local time range, compare against X
export const TIME_FILTER_INHERIT = "inherit";

export type ResolvedComparison =
  | { mode: "inherit" }
  | { mode: "none" }
  | { mode: "local"; range: string };

export type ResolvedTimeFilters = {
  // True when `tr` is a real range rather than absent or `inherit`.
  hasLocalTimeRange: boolean;
  comparison: ResolvedComparison;
};

export function resolveTimeFilters(
  timeFilters: string | undefined,
): ResolvedTimeFilters {
  const params = new URLSearchParams(timeFilters ?? "");
  const range = params.get(ExploreStateURLParams.TimeRange);
  const comparisonRange = params.get(ExploreStateURLParams.ComparisonTimeRange);

  const hasLocalTimeRange = Boolean(range) && range !== TIME_FILTER_INHERIT;

  let comparison: ResolvedComparison;
  if (comparisonRange === TIME_FILTER_INHERIT) {
    comparison = { mode: "inherit" };
  } else if (comparisonRange) {
    comparison = { mode: "local", range: comparisonRange };
  } else if (range) {
    // Once `tr` is set, a missing `compare_tr` means no comparison.
    comparison = { mode: "none" };
  } else {
    comparison = { mode: "inherit" };
  }

  return { hasLocalTimeRange, comparison };
}

/**
 * Canonical `time_filters` string for a set of params, or undefined when everything is inherited.
 * A comparison without a time range gets `tr=inherit`, grain and zone only apply to a local time range,
 * and `tr` always comes first so the YAML reads the same however it was produced.
 */
export function normalizeTimeFilters(
  params: URLSearchParams,
): string | undefined {
  const normalized = new URLSearchParams(params);

  if (
    !normalized.has(ExploreStateURLParams.TimeRange) &&
    normalized.has(ExploreStateURLParams.ComparisonTimeRange)
  ) {
    normalized.set(ExploreStateURLParams.TimeRange, TIME_FILTER_INHERIT);
  }

  if (normalized.get(ExploreStateURLParams.TimeRange) === TIME_FILTER_INHERIT) {
    normalized.delete(ExploreStateURLParams.TimeGrain);
    normalized.delete(ExploreStateURLParams.TimeZone);
    if (
      normalized.get(ExploreStateURLParams.ComparisonTimeRange) ===
      TIME_FILTER_INHERIT
    ) {
      normalized.delete(ExploreStateURLParams.TimeRange);
      normalized.delete(ExploreStateURLParams.ComparisonTimeRange);
    }
  }

  if (!normalized.size) return undefined;

  const rangeKey: string = ExploreStateURLParams.TimeRange;
  const ordered = new URLSearchParams();
  const range = normalized.get(rangeKey);
  if (range) ordered.set(rangeKey, range);
  normalized.forEach((value, key) => {
    if (key !== rangeKey) ordered.append(key, value);
  });
  return ordered.toString();
}

/** The params a widget's own time state understands: `inherit` is not a range it can resolve. */
export function stripInheritedTimeFilters(
  params: URLSearchParams,
): URLSearchParams {
  const local = new URLSearchParams(params);
  if (local.get(ExploreStateURLParams.TimeRange) === TIME_FILTER_INHERIT) {
    local.delete(ExploreStateURLParams.TimeRange);
  }
  if (
    local.get(ExploreStateURLParams.ComparisonTimeRange) === TIME_FILTER_INHERIT
  ) {
    local.delete(ExploreStateURLParams.ComparisonTimeRange);
  }
  return local;
}

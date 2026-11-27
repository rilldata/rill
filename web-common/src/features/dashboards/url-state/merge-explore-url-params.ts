import {
  capEphemeralMeasureDefs,
  mergeEphemeralMeasureDefs,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/library";
import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import {
  fromEphemeralMeasuresParam,
  toEphemeralMeasuresParam,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/url-param";
import { MAX_EPHEMERAL_MEASURES } from "@rilldata/web-common/features/dashboards/ephemeral-measures/validation";
import { convertProtoToUrlParams } from "@rilldata/web-common/features/dashboards/proto-state/convert-proto-to-url-params";
import { expandCompressedParams } from "@rilldata/web-common/features/dashboards/url-state/compression";
import {
  ExploreWebViewCommonURLParams,
  ExploreWebViewSpecificURLParams,
  GlobalExploreURLParams,
} from "@rilldata/web-common/features/dashboards/url-state/explore-web-view-specific-url-params";
import { ExploreUrlWebView } from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";

export type ExploreUrlParamsSources = {
  // Url params of the page, cleaned of embed params and normalized.
  urlParams: URLSearchParams;
  // Stored params of the view in the url, when the url maps to a session entry.
  sessionParams: URLSearchParams | undefined;
  mostRecentParams: URLSearchParams | undefined;
  // Normalized params of the home bookmark.
  bookmarkParams: URLSearchParams | null | undefined;
  yamlParams: URLSearchParams;
  rillDefaultParams: URLSearchParams;
  backButtonUsed: boolean;
};

/**
 * Picks the params an explore opens with.
 * The sources are not a cascade: one provides the params and the rest only fill in what it leaves out.
 *
 * 1. Url: when the back button was used, or the url has params and no session entry applies.
 *    Filled from rill defaults only, since the url sync drops params equal to them;
 *    a missing param in a url means the rill default, not the yaml default.
 * 2. Session: when a session entry applies to the url.
 *    Filled from most recent, home bookmark, yaml and rill defaults, for entries of views not visited yet.
 * 3. Fresh visit: the home bookmark with the most recent params written over it.
 *    Filled from yaml and rill defaults.
 */
export function mergeExploreUrlParams({
  urlParams,
  sessionParams,
  mostRecentParams,
  bookmarkParams,
  yamlParams,
  rillDefaultParams,
  backButtonUsed,
}: ExploreUrlParamsSources) {
  if (backButtonUsed || (urlParams.size > 0 && !sessionParams)) {
    const params = new URLSearchParams(urlParams);
    fillExploreUrlParams(params, rillDefaultParams);
    return params;
  }

  if (sessionParams) {
    const params = new URLSearchParams(sessionParams);
    if (mostRecentParams) fillExploreUrlParams(params, mostRecentParams);
    if (bookmarkParams) fillExploreUrlParams(params, bookmarkParams);
    fillExploreUrlParams(params, yamlParams);
    fillExploreUrlParams(params, rillDefaultParams);
    return params;
  }

  const params = new URLSearchParams(bookmarkParams ?? undefined);
  if (mostRecentParams) {
    fillExploreUrlParams(params, mostRecentParams, { overwrite: true });
  }
  fillExploreUrlParams(params, yamlParams);
  fillExploreUrlParams(params, rillDefaultParams);
  return params;
}

/**
 * Expands `gzipped_state` and the legacy proto `state` param into plain params.
 * The url's other params are set over the ones from `state`, matching the order the url parser applies them in.
 * A `state` that does not decode is kept so that the parser reports it.
 */
export function normalizeExploreUrlParams(params: URLSearchParams) {
  const expanded = new URLSearchParams(expandCompressedParams(params));
  const legacyState = expanded.get(ExploreStateURLParams.LegacyProtoState);
  if (!legacyState) return expanded;

  let normalized: URLSearchParams;
  try {
    normalized = convertProtoToUrlParams(legacyState);
  } catch {
    return expanded;
  }
  expanded.forEach((value, key: ExploreStateURLParams) => {
    if (key === ExploreStateURLParams.LegacyProtoState) return;
    normalized.set(key, value);
  });
  return normalized;
}

/**
 * Copies params of `source` into `target`, following the view rules of `copyUrlSearchParamsForView`:
 * shared params are always copied, view specific ones only between the same views,
 * along with the params two different views have in common.
 * This keeps an explore `sort_by` out of a pivot url, where it would be read as the pivot sort.
 *
 * Without `overwrite`, params `target` already has are kept, including empty ones:
 * `compare_tr=` (comparison off) and `f=` (no filter) must hold.
 */
export function fillExploreUrlParams(
  target: URLSearchParams,
  source: URLSearchParams,
  { overwrite }: { overwrite: boolean } = { overwrite: false },
) {
  const targetView = getView(target);
  const sourceView = getView(source);

  const copy = (param: ExploreStateURLParams) => {
    if (!source.has(param) || (!overwrite && target.has(param))) return;
    target.set(param, source.get(param)!);
  };

  GlobalExploreURLParams.forEach(copy);
  if (targetView === sourceView) {
    ExploreWebViewSpecificURLParams[targetView].forEach(copy);
  } else {
    ExploreWebViewCommonURLParams[sourceView]?.[targetView]?.forEach(copy);
  }
}

/**
 * Adds the ad-hoc measure library to the definitions in `adhoc_m`.
 * The url only carries the definitions the page shows, so this brings back hidden definitions
 * and ones created in another explore on the same metrics view.
 * Definitions in the params win over stored copies.
 *
 * This happens before the params are parsed, since names in other params are validated against `adhoc_m`.
 * A param holds a limited number of definitions,
 * so ones the params reference are kept ahead of unused ones when capping.
 */
export function addEphemeralMeasureLibraryToUrlParams(
  params: URLSearchParams,
  library: EphemeralMeasureDef[],
) {
  if (!library.length) return;

  const { ephemeralMeasures } = fromEphemeralMeasuresParam(
    params.get(ExploreStateURLParams.EphemeralMeasures) ?? "",
  );
  const namesInUse = new Set(
    [
      ExploreStateURLParams.VisibleMeasures,
      ExploreStateURLParams.LeaderboardMeasures,
      ExploreStateURLParams.SortBy,
      ExploreStateURLParams.ExpandedMeasure,
      ExploreStateURLParams.PivotRows,
      ExploreStateURLParams.PivotColumns,
    ].flatMap((param) => params.get(param)?.split(",") ?? []),
  );
  params.set(
    ExploreStateURLParams.EphemeralMeasures,
    toEphemeralMeasuresParam(
      capEphemeralMeasureDefs(
        mergeEphemeralMeasureDefs(ephemeralMeasures, library),
        namesInUse,
        MAX_EPHEMERAL_MEASURES,
      ),
    ),
  );
}

function getView(params: URLSearchParams) {
  const view = params.get(ExploreStateURLParams.WebView);
  return view && view in ExploreWebViewSpecificURLParams
    ? (view as ExploreUrlWebView)
    : ExploreUrlWebView.Explore;
}

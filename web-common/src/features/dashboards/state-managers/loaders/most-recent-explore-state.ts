import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { validateAndCleanExploreState } from "@rilldata/web-common/features/dashboards/stores/validate-and-clean-explore-state";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import { ToLegacySortTypeMap } from "@rilldata/web-common/features/dashboards/url-state/legacyMappers";
import {
  ExploreUrlWebView,
  FromURLParamsSortTypeMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type {
  V1ExploreSpec,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

// The subset of the explore view's params that is kept long term.
// Ad-hoc measure definitions are stored per metrics view instead; see ephemeral-measures/library.ts.
const MostRecentExploreURLParams = [
  ExploreStateURLParams.TimeZone,
  ExploreStateURLParams.VisibleMeasures,
  ExploreStateURLParams.VisibleDimensions,
  ExploreStateURLParams.SortBy,
  ExploreStateURLParams.SortType,
  ExploreStateURLParams.SortDirection,
  ExploreStateURLParams.LeaderboardMeasures,
  ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
];

function getKeyForLocalStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  return `rill:app:explore:${storageNamespacePrefix ?? ""}${exploreName}:most_recent_params`.toLowerCase();
}

// Key of the explore state saved as JSON, before the params were saved instead.
function getLegacyKeyForLocalStore(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  return `rill:app:explore:${storageNamespacePrefix ?? ""}${exploreName}`.toLowerCase();
}

/**
 * Returns the most recent explore view params, validated against the specs.
 * Stale names are cleaned the way `validateAndCleanExploreState` does it:
 * visible fields that are all stale are dropped so that other sources fill them,
 * a stale sort measure falls back to the first visible one,
 * and stale leaderboard measures are dropped while the valid ones are kept.
 * The url parser would instead drop a stale sort and the whole leaderboard measure list.
 *
 * `ephemeralMeasures` are the ad-hoc definitions restored from the per-metrics-view library.
 * The saved params do not hold definitions, only names that may refer to them,
 * so they must be known before validation or the ad-hoc names are dropped.
 */
export function getMostRecentExploreUrlParams(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  ephemeralMeasures: EphemeralMeasureDef[] = [],
) {
  let partialExploreState: Partial<ExploreState> | undefined;
  try {
    const rawParams = localStorage.getItem(
      getKeyForLocalStore(exploreName, storageNamespacePrefix),
    );
    if (rawParams !== null) {
      partialExploreState = decodeMostRecentUrlParams(
        new URLSearchParams(rawParams),
        exploreSpec,
      );
    } else {
      const rawExploreState = localStorage.getItem(
        getLegacyKeyForLocalStore(exploreName, storageNamespacePrefix),
      );
      if (rawExploreState) {
        partialExploreState = JSON.parse(
          rawExploreState,
        ) as Partial<ExploreState>;
      }
    }
  } catch {
    // Reading storage throws when the browser blocks it, and parsing throws on a corrupt entry.
  }
  if (!partialExploreState) return undefined;

  if (ephemeralMeasures.length) {
    partialExploreState.ephemeralMeasures = ephemeralMeasures;
  }
  validateAndCleanExploreState(
    metricsViewSpec,
    exploreSpec,
    partialExploreState,
  );

  const params = convertPartialExploreStateToUrlParams(
    exploreSpec,
    metricsViewSpec,
    partialExploreState,
    undefined,
  );
  // The time zone is only written along with a time range, which the most recent state does not have.
  if (partialExploreState.selectedTimezone) {
    params.set(
      ExploreStateURLParams.TimeZone,
      partialExploreState.selectedTimezone,
    );
  }
  return pickMostRecentUrlParams(params);
}

/**
 * Saves the long term subset of the explore view's params.
 * Other views' params are not saved.
 */
export function saveMostRecentExploreUrlParams(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  urlParams: URLSearchParams,
) {
  const view = urlParams.get(
    ExploreStateURLParams.WebView,
  ) as ExploreUrlWebView | null;
  if (view && view !== ExploreUrlWebView.Explore) return;

  setMostRecentExploreUrlParamsInLocalStorage(
    exploreName,
    storageNamespacePrefix,
    pickMostRecentUrlParams(urlParams),
  );
}

export function clearMostRecentExploreState(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
) {
  localStorage.removeItem(
    getKeyForLocalStore(exploreName, storageNamespacePrefix),
  );
  localStorage.removeItem(
    getLegacyKeyForLocalStore(exploreName, storageNamespacePrefix),
  );
}

export function setMostRecentExploreUrlParamsInLocalStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  urlParams: URLSearchParams,
) {
  try {
    localStorage.setItem(
      getKeyForLocalStore(exploreName, storageNamespacePrefix),
      urlParams.toString(),
    );
    // The params replace the legacy state.
    localStorage.removeItem(
      getLegacyKeyForLocalStore(exploreName, storageNamespacePrefix),
    );
  } catch {
    // Writing storage throws when the quota is exceeded or the browser blocks it, e.g. in an embed.
  }
}

/**
 * Writes the most recent state the way it was saved before params were saved instead.
 * Used to test that the legacy state still loads.
 */
export function setLegacyMostRecentExploreStateInLocalStorage(
  exploreName: string,
  storageNamespacePrefix: string | undefined,
  exploreState: Partial<ExploreState>,
) {
  localStorage.setItem(
    getLegacyKeyForLocalStore(exploreName, storageNamespacePrefix),
    JSON.stringify(exploreState),
  );
}

/**
 * Decodes the saved params into the state fields they map to, without validating names.
 * The url parser is not used since it validates while parsing,
 * which leaves nothing for `validateAndCleanExploreState` to clean.
 */
function decodeMostRecentUrlParams(
  params: URLSearchParams,
  exploreSpec: V1ExploreSpec,
) {
  const partialExploreState: Partial<ExploreState> = {};

  const timeZone = params.get(ExploreStateURLParams.TimeZone);
  if (timeZone) partialExploreState.selectedTimezone = timeZone;

  const measures = params.get(ExploreStateURLParams.VisibleMeasures);
  if (measures) {
    partialExploreState.visibleMeasures =
      measures === "*"
        ? [...(exploreSpec.measures ?? [])]
        : measures.split(",");
  }

  const dimensions = params.get(ExploreStateURLParams.VisibleDimensions);
  if (dimensions) {
    partialExploreState.visibleDimensions =
      dimensions === "*"
        ? [...(exploreSpec.dimensions ?? [])]
        : dimensions.split(",");
  }

  const sortBy = params.get(ExploreStateURLParams.SortBy);
  if (sortBy) partialExploreState.leaderboardSortByMeasureName = sortBy;

  const sortType = params.get(ExploreStateURLParams.SortType);
  if (sortType && sortType in FromURLParamsSortTypeMap) {
    partialExploreState.dashboardSortType = Number(
      ToLegacySortTypeMap[FromURLParamsSortTypeMap[sortType]],
    );
  }

  const sortDirection = params.get(ExploreStateURLParams.SortDirection);
  if (sortDirection) {
    partialExploreState.sortDirection =
      sortDirection === "ASC"
        ? SortDirection.ASCENDING
        : SortDirection.DESCENDING;
  }

  const leaderboardMeasures = params.get(
    ExploreStateURLParams.LeaderboardMeasures,
  );
  if (leaderboardMeasures) {
    partialExploreState.leaderboardMeasureNames =
      leaderboardMeasures.split(",");
  }

  const showContext = params
    .get(ExploreStateURLParams.LeaderboardShowContextForAllMeasures)
    ?.toLowerCase();
  if (showContext) {
    partialExploreState.leaderboardShowContextForAllMeasures =
      showContext === "true" || showContext === "1";
  }

  return partialExploreState;
}

function pickMostRecentUrlParams(urlParams: URLSearchParams) {
  const params = new URLSearchParams();
  MostRecentExploreURLParams.forEach((param) => {
    const value = urlParams.get(param);
    if (value !== null) params.set(param, value);
  });
  return params;
}

import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  compressUrlParams,
  shouldCompressParams,
} from "@rilldata/web-common/features/dashboards/url-state/compression";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import { convertPartialExploreStateToUrlParams } from "./convert-partial-explore-state-to-url-params";

export function getCleanedUrlParamsForGoto(
  exploreSpec: V1ExploreSpec,
  partialExploreState: Partial<MetricsExplorerEntity>,
  timeControlsState: TimeControlState | undefined,
  defaultExploreUrlParams: URLSearchParams,
  urlForCompressionCheck?: URL,
) {
  // Create params from the explore state
  const stateParams = convertPartialExploreStateToUrlParams(
    exploreSpec,
    partialExploreState,
    timeControlsState,
  );

  // Remove params with default values
  [...stateParams.entries()].forEach(([key, value]) => {
    const defaultValue = defaultExploreUrlParams.get(key);
    if (
      (defaultValue === null && value !== "") ||
      (defaultValue !== null && value !== defaultValue)
    ) {
      return;
    }
    stateParams.delete(key);
  });

  if (!urlForCompressionCheck) return stateParams;

  // compression
  const urlCopy = new URL(urlForCompressionCheck);
  urlCopy.search = stateParams.toString();
  const shouldCompress = shouldCompressParams(urlCopy);
  if (!shouldCompress) return stateParams;

  const compressedUrlParams = new URLSearchParams();
  compressedUrlParams.set(
    ExploreStateURLParams.GzippedParams,
    compressUrlParams(stateParams.toString()),
  );
  return compressedUrlParams;
}

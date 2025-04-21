import { getCompoundQuery } from "@rilldata/web-common/features/compound-query-result";
import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params";
import {
  ExploreUrlWebView,
  FromURLParamViewMap,
  ToActivePageViewMap,
} from "@rilldata/web-common/features/dashboards/url-state/mappers";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

export function getRillDefaultExploreUrlParams(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const rillDefaultExploreState = getRillDefaultExploreState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    rillDefaultExploreState,
  );
  return convertPartialExploreStateToUrlParams(
    exploreSpec,
    rillDefaultExploreState,
    timeControlState,
  );
}

export function getRillDefaultExploreUrlParamsByView(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRangeSummary: V1TimeRangeSummary | undefined,
) {
  const rillDefaultExploreState = getRillDefaultExploreState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
  );
  const timeControlState = getTimeControlState(
    metricsViewSpec,
    exploreSpec,
    timeRangeSummary,
    rillDefaultExploreState,
  );

  const rillDefaultExploreURLParamsByView = {} as Record<
    ExploreUrlWebView,
    URLSearchParams
  >;
  for (const webView in FromURLParamViewMap) {
    rillDefaultExploreState.activePage = Number(
      ToActivePageViewMap[FromURLParamViewMap[webView]],
    );
    rillDefaultExploreURLParamsByView[webView] =
      convertPartialExploreStateToUrlParams(
        exploreSpec,
        rillDefaultExploreState,
        timeControlState,
      );
  }

  return rillDefaultExploreURLParamsByView;
}

export function createRillDefaultExploreUrlParamsByView(
  validSpecQuery: ReturnType<typeof useExploreValidSpec>,
  fullTimeRangeQuery: ReturnType<typeof useMetricsViewTimeRange>,
) {
  return getCompoundQuery(
    [validSpecQuery, fullTimeRangeQuery],
    ([validSpecResp, metricsViewTimeRangeResp]) => {
      const metricsViewSpec = validSpecResp?.metricsView;
      const exploreSpec = validSpecResp?.explore;

      if (
        !metricsViewSpec ||
        !exploreSpec ||
        // safeguard to make sure time range summary is loaded for metrics view with time dimension
        (metricsViewSpec.timeDimension &&
          !metricsViewTimeRangeResp?.timeRangeSummary)
      ) {
        return undefined;
      }

      return getRillDefaultExploreUrlParamsByView(
        metricsViewSpec,
        exploreSpec,
        metricsViewTimeRangeResp?.timeRangeSummary,
      );
    },
  );
}

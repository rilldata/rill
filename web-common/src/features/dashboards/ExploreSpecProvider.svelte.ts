import {
  createQueryServiceMetricsViewTimeRange,
  createRuntimeServiceListResources,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";

/**
 * Provides explore & metrics view specs.
 * Bridges svelte4 stores to svelte5 stores.
 */
export class ExploreSpecProvider {
  public exploreSpec = $state<V1ExploreSpec | undefined>(undefined);
  public metricsViewSpec = $state<V1MetricsViewSpec | undefined>(undefined);
  public timeRangeSummary = $state<V1TimeRangeSummary | undefined>(undefined);
  public metricsViewNames: string[] = $state([]);

  public cleanup: () => void;

  public constructor(runtimeClient: RuntimeClient, exploreName: string) {
    const allResourcesQuery = createRuntimeServiceListResources(
      runtimeClient,
      {},
    );

    let timeRangeSummaryUnsub: (() => void) | undefined = undefined;
    const allResourcesUnsub = allResourcesQuery.subscribe(
      (allResourcesResp) => {
        this.exploreSpec = allResourcesResp.data?.resources?.find(
          (r) => r.meta?.name?.name === exploreName,
        )?.explore?.state?.validSpec;
        if (!this.exploreSpec?.metricsView) return;
        this.metricsViewNames = [this.exploreSpec.metricsView];

        this.metricsViewSpec = allResourcesResp.data?.resources?.find(
          (r) => r.meta?.name?.name === this.exploreSpec?.metricsView,
        )?.metricsView?.state?.validSpec;

        if (!timeRangeSummaryUnsub) {
          const timeRangeSummaryQuery = createQueryServiceMetricsViewTimeRange(
            runtimeClient,
            { metricsViewName: this.exploreSpec.metricsView },
          );
          timeRangeSummaryUnsub = timeRangeSummaryQuery.subscribe(
            (timeRangeSummaryResp) => {
              this.timeRangeSummary =
                timeRangeSummaryResp.data?.timeRangeSummary;
            },
          );
        }
      },
    );

    this.cleanup = () => {
      allResourcesUnsub();
      timeRangeSummaryUnsub?.();
    };
  }

  public getDefaultUrlParams() {
    return getRillDefaultExploreUrlParams(
      this.metricsViewSpec ?? {},
      this.exploreSpec ?? {},
      this.timeRangeSummary,
    );
  }
}

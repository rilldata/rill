import type {
  V1ExploreSpec,
  V1GetExploreResponse,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { V1ReconcileStatus } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { createCachedExploreValidSpecSelector } from "./selectors";

function makeExploreSpec(metricsViewName: string): V1ExploreSpec {
  return {
    metricsView: metricsViewName,
    measures: ["impressions"],
  } as V1ExploreSpec;
}

function makeMetricsViewSpec(timeDimensionName: string): V1MetricsViewSpec {
  return {
    timeDimension: timeDimensionName,
  } as V1MetricsViewSpec;
}

function makeGetExploreResponse({
  exploreName = "ad_bids_explore",
  metricsViewName = "ad_bids_metrics",
  exploreSpec,
  metricsViewSpec,
  exploreReconcileStatus = V1ReconcileStatus.RECONCILE_STATUS_IDLE,
  metricsViewReconcileStatus = V1ReconcileStatus.RECONCILE_STATUS_IDLE,
}: {
  exploreName?: string;
  metricsViewName?: string;
  exploreSpec: V1ExploreSpec | undefined;
  metricsViewSpec: V1MetricsViewSpec | undefined;
  exploreReconcileStatus?: V1ReconcileStatus;
  metricsViewReconcileStatus?: V1ReconcileStatus;
}): V1GetExploreResponse {
  return {
    explore: {
      meta: {
        name: {
          name: exploreName,
        },
        reconcileStatus: exploreReconcileStatus,
      },
      explore: {
        state: {
          validSpec: exploreSpec,
        },
      },
    },
    metricsView: {
      meta: {
        name: {
          name: metricsViewName,
        },
        reconcileStatus: metricsViewReconcileStatus,
      },
      metricsView: {
        state: {
          validSpec: metricsViewSpec,
        },
      },
    },
  } as V1GetExploreResponse;
}

describe("createCachedExploreValidSpecSelector", () => {
  it("keeps the last complete valid specs while resources reconcile", () => {
    const selector = createCachedExploreValidSpecSelector();

    const initialResponse = makeGetExploreResponse({
      exploreSpec: makeExploreSpec("ad_bids_metrics"),
      metricsViewSpec: makeMetricsViewSpec("timestamp"),
    });
    const initialSpecs = selector(initialResponse);

    const reconcilingResponse = makeGetExploreResponse({
      exploreSpec: undefined,
      metricsViewSpec: undefined,
      exploreReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      metricsViewReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    });
    const reconcilingSpecs = selector(reconcilingResponse);

    expect(reconcilingSpecs).toEqual(initialSpecs);
  });

  it("does not keep stale specs once reconciliation is idle", () => {
    const selector = createCachedExploreValidSpecSelector();

    selector(
      makeGetExploreResponse({
        exploreSpec: makeExploreSpec("ad_bids_metrics"),
        metricsViewSpec: makeMetricsViewSpec("timestamp"),
      }),
    );

    const idleResponse = makeGetExploreResponse({
      exploreSpec: undefined,
      metricsViewSpec: undefined,
      exploreReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
      metricsViewReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
    });
    const idleSpecs = selector(idleResponse);

    expect(idleSpecs).toEqual({
      explore: undefined,
      metricsView: undefined,
    });
  });

  it("updates the cache when new complete specs arrive", () => {
    const selector = createCachedExploreValidSpecSelector();

    selector(
      makeGetExploreResponse({
        exploreSpec: makeExploreSpec("ad_bids_metrics"),
        metricsViewSpec: makeMetricsViewSpec("timestamp"),
      }),
    );

    selector(
      makeGetExploreResponse({
        exploreSpec: undefined,
        metricsViewSpec: undefined,
        exploreReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
    );

    const updatedSpecs = selector(
      makeGetExploreResponse({
        exploreSpec: makeExploreSpec("ad_impressions_metrics"),
        metricsViewSpec: makeMetricsViewSpec("created_at"),
      }),
    );

    expect(updatedSpecs.explore?.metricsView).toBe("ad_impressions_metrics");
    expect(updatedSpecs.metricsView?.timeDimension).toBe("created_at");
  });

  it("keeps the cached pair when only one spec is missing during reconciliation", () => {
    const selector = createCachedExploreValidSpecSelector();

    const initialSpecs = selector(
      makeGetExploreResponse({
        exploreSpec: makeExploreSpec("ad_bids_metrics"),
        metricsViewSpec: makeMetricsViewSpec("timestamp"),
      }),
    );

    const partialSpecs = selector(
      makeGetExploreResponse({
        exploreSpec: makeExploreSpec("ad_bids_metrics"),
        metricsViewSpec: undefined,
        metricsViewReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
    );

    expect(partialSpecs).toEqual(initialSpecs);
  });

  it("does not bleed cached specs across explores", () => {
    const selector = createCachedExploreValidSpecSelector();

    selector(
      makeGetExploreResponse({
        exploreName: "explore_a",
        metricsViewName: "metrics_a",
        exploreSpec: makeExploreSpec("metrics_a"),
        metricsViewSpec: makeMetricsViewSpec("timestamp"),
      }),
    );

    const otherExploreSpecs = selector(
      makeGetExploreResponse({
        exploreName: "explore_b",
        metricsViewName: "metrics_b",
        exploreSpec: undefined,
        metricsViewSpec: undefined,
        exploreReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
        metricsViewReconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
    );

    expect(otherExploreSpecs).toEqual({
      explore: undefined,
      metricsView: undefined,
    });
  });
});

import type { CreateQueryResult } from "@rilldata/svelte-query";
import {
  AD_BIDS_INIT_DIMENSIONS,
  AD_BIDS_INIT_MEASURES,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_NAME,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { ExploreValidSpecResponse } from "@rilldata/web-common/features/explores/selectors";
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  RpcStatus,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/query-core";
import { writable } from "svelte/store";

export function createMetricsMetaQueryMock(
  shouldInit = true,
): CreateQueryResult<V1MetricsViewSpec, RpcStatus> & {
  setMeasures: (measures: Array<MetricsViewSpecMeasureV2>) => void;
  setDimensions: (dimensions: Array<MetricsViewSpecDimensionV2>) => void;
} {
  const { update, subscribe } = writable<
    QueryObserverResult<V1MetricsViewSpec, RpcStatus>
  >({
    data: undefined,
    isSuccess: false,
    isRefetching: false,
  } as any);

  const mock = {
    subscribe,
    setMeasures: (measures: MetricsViewSpecMeasureV2[]) =>
      update((value) => {
        value.isSuccess = true;
        value.data ??= {
          measures: [],
          dimensions: [],
        };
        value.data.measures = measures;
        return value;
      }),
    setDimensions: (dimensions: MetricsViewSpecDimensionV2[]) =>
      update((value) => {
        value.isSuccess = true;
        value.data ??= {
          measures: [],
          dimensions: [],
        };
        value.data.dimensions = dimensions;
        return value;
      }),
  };

  if (shouldInit) {
    mock.setMeasures(AD_BIDS_INIT_MEASURES);
    mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
  }

  return mock;
}

export function createValidSpecQueryMock(
  metricsView = AD_BIDS_NAME,
  shouldInit = true,
  initMetrics: V1MetricsViewSpec = AD_BIDS_METRICS_INIT,
): CreateQueryResult<ExploreValidSpecResponse, RpcStatus> & {
  setMeasures: (measures: Array<MetricsViewSpecMeasureV2>) => void;
  setDimensions: (dimensions: Array<MetricsViewSpecDimensionV2>) => void;
} {
  const { update, subscribe } = writable<
    QueryObserverResult<ExploreValidSpecResponse, RpcStatus>
  >({
    data: undefined,
    isSuccess: false,
    isRefetching: false,
  } as any);

  function initData() {
    return {
      metricsView: {
        measures: [],
        dimensions: [],
      },
      explore: {
        metricsView,
        measures: [],
        dimensions: [],
      },
    };
  }

  const mock = {
    subscribe,
    setMeasures: (
      measures: MetricsViewSpecMeasureV2[],
      exploreMeasures?: string[],
    ) => {
      exploreMeasures ??= measures.map((m) => m.name!);
      update((value) => {
        value.isSuccess = true;
        value.data ??= initData();
        value.data.metricsView!.measures = measures;
        value.data.explore!.measures = exploreMeasures;
        return value;
      });
    },
    setDimensions: (
      dimensions: MetricsViewSpecDimensionV2[],
      exploreDimensions?: string[],
    ) => {
      exploreDimensions ??= dimensions.map((d) => d.name!);
      update((value) => {
        value.isSuccess = true;
        value.data ??= initData();
        value.data.metricsView!.dimensions = dimensions;
        value.data.explore!.dimensions = exploreDimensions;
        return value;
      });
    },
  };

  if (shouldInit) {
    mock.setMeasures(initMetrics.measures!);
    mock.setDimensions(initMetrics.dimensions!);
  }

  return mock;
}

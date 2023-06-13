import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_INIT_DIMENSIONS,
  AD_BIDS_INIT_MEASURES,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_WITH_DELETED_DIMENSION,
  AD_BIDS_WITH_DELETED_MEASURE,
  assertVisiblePartsOfMetricsView,
  createAdBidsInStore,
  createAdBidsMirrorInStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores-test-data";
import {
  createShowHideDimensionsStore,
  createShowHideMeasuresStore,
} from "@rilldata/web-common/features/dashboards/show-hide-selectors";
import type {
  MetricsViewDimension,
  MetricsViewMeasure,
  RpcStatus,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { get, writable } from "svelte/store";
import { describe, it, expect } from "vitest";

describe("Show/Hide Selectors", () => {
  describe("Show/Hide measures", () => {
    it("Toggle individual visibility", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      expect(get(showHideMeasure).selectedItems).toEqual([]);

      createAdBidsInStore();
      assertVisiblePartsWithMirroring(
        [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
        undefined
      );
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).availableKeys).toEqual([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      // toggle visibility
      showHideMeasure.toggleVisibility(AD_BIDS_BID_PRICE_MEASURE);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring([AD_BIDS_IMPRESSIONS_MEASURE], undefined);
      expect(get(showHideMeasure).selectedItems).toEqual([true, false]);

      // toggle visibility
      showHideMeasure.toggleVisibility(AD_BIDS_IMPRESSIONS_MEASURE);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring([], undefined);
      expect(get(showHideMeasure).selectedItems).toEqual([false, false]);

      // toggle back visibility
      showHideMeasure.toggleVisibility(AD_BIDS_BID_PRICE_MEASURE);
      showHideMeasure.toggleVisibility(AD_BIDS_IMPRESSIONS_MEASURE);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        [AD_BIDS_BID_PRICE_MEASURE, AD_BIDS_IMPRESSIONS_MEASURE],
        undefined
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);
    });

    it("Toggle all visibility", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      createAdBidsInStore();
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      // toggle all to not visible
      showHideMeasure.setAllToNotVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring([], undefined);
      expect(get(showHideMeasure).selectedItems).toEqual([false, false]);

      // toggle all back to visible
      showHideMeasure.setAllToVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        [AD_BIDS_BID_PRICE_MEASURE, AD_BIDS_IMPRESSIONS_MEASURE],
        undefined
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);
    });

    it("Meta query updates", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      createAdBidsInStore();
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      mock.setMeasures([
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE,
          expression: "count(*)",
        },
      ]);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_DELETED_MEASURE);
      expect(get(showHideMeasure).availableKeys).toEqual([
        AD_BIDS_IMPRESSIONS_MEASURE,
      ]);
      expect(get(showHideMeasure).selectedItems).toEqual([true]);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring([AD_BIDS_IMPRESSIONS_MEASURE], undefined);
    });
  });

  describe("Show/Hide dimensions", () => {
    it("Toggle individual visibility", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock
      );
      expect(get(showHideDimensions).selectedItems).toEqual([]);

      createAdBidsInStore();
      assertVisiblePartsWithMirroring(undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
      expect(get(showHideDimensions).availableKeys).toEqual([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);

      // toggle visibility
      showHideDimensions.toggleVisibility(AD_BIDS_PUBLISHER_DIMENSION);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, [AD_BIDS_DOMAIN_DIMENSION]);
      expect(get(showHideDimensions).selectedItems).toEqual([false, true]);

      // toggle visibility
      showHideDimensions.toggleVisibility(AD_BIDS_DOMAIN_DIMENSION);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, []);
      expect(get(showHideDimensions).selectedItems).toEqual([false, false]);

      // toggle back visibility
      showHideDimensions.toggleVisibility(AD_BIDS_PUBLISHER_DIMENSION);
      showHideDimensions.toggleVisibility(AD_BIDS_DOMAIN_DIMENSION);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
    });

    it("Toggle all visibility", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock
      );
      createAdBidsInStore();
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);

      // toggle all to not visible
      showHideDimensions.setAllToNotVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, []);
      expect(get(showHideDimensions).selectedItems).toEqual([false, false]);

      // toggle all back to visible
      showHideDimensions.setAllToVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
    });

    it("Meta query updates", () => {
      const mock = createMetricsMetaQueryMock();
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock
      );
      createAdBidsInStore();
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);

      mock.setDimensions([
        {
          name: AD_BIDS_PUBLISHER_DIMENSION,
        },
      ]);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_DELETED_DIMENSION);
      expect(get(showHideDimensions).availableKeys).toEqual([
        AD_BIDS_PUBLISHER_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true]);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(undefined, [AD_BIDS_PUBLISHER_DIMENSION]);
    });
  });
});

function createMetricsMetaQueryMock(): CreateQueryResult<
  V1MetricsView,
  RpcStatus
> & {
  setMeasures: (measures: Array<MetricsViewMeasure>) => void;
  setDimensions: (dimensions: Array<MetricsViewDimension>) => void;
} {
  const { update, subscribe } = writable<
    QueryObserverResult<V1MetricsView, RpcStatus>
  >({
    data: undefined,
    isSuccess: false,
    isRefetching: false,
  } as any);

  return {
    subscribe,
    setMeasures: (measures) =>
      update((value) => {
        value.isSuccess = true;
        value.data ??= {
          measures: [],
          dimensions: [],
        };
        value.data.measures = measures;
        return value;
      }),
    setDimensions: (dimensions: Array<MetricsViewDimension>) =>
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
}

function assertVisiblePartsWithMirroring(
  measures: Array<string> | undefined,
  dimensions: Array<string> | undefined
) {
  assertVisiblePartsOfMetricsView(AD_BIDS_NAME, measures, dimensions);
  // create a mirror using the proto and assert that the visible selections are persisted
  createAdBidsMirrorInStore();
  assertVisiblePartsOfMetricsView(AD_BIDS_MIRROR_NAME, measures, dimensions);
}

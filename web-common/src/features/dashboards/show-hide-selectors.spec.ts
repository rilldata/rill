import {
  createShowHideDimensionsStore,
  createShowHideMeasuresStore,
} from "@rilldata/web-common/features/dashboards/show-hide-selectors";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_COUNTRY_DIMENSION,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_INIT_DIMENSIONS,
  AD_BIDS_INIT_MEASURES,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_COUNT_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_THREE_DIMENSIONS,
  AD_BIDS_THREE_MEASURES,
  AD_BIDS_WITH_DELETED_DIMENSION,
  AD_BIDS_WITH_DELETED_MEASURE,
  AD_BIDS_WITH_THREE_DIMENSIONS,
  AD_BIDS_WITH_THREE_MEASURES,
  assertVisiblePartsOfMetricsView,
  createAdBidsMirrorInStore,
  createMetricsMetaQueryMock,
  resetDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
import {
  getPersistentDashboardStore,
  initPersistentDashboardStore,
} from "@rilldata/web-common/features/dashboards/stores/persistent-dashboard-state";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

describe("Show/Hide Selectors", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_NAME);
    initPersistentDashboardStore(AD_BIDS_NAME);
  });

  beforeEach(() => {
    getPersistentDashboardStore().reset();
    resetDashboardStore();
  });

  describe("Show/Hide measures", () => {
    it("Toggle individual visibility", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      expect(get(showHideMeasure).selectedItems).toEqual([]);

      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_IMPRESSIONS_MEASURE, AD_BIDS_BID_PRICE_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).availableKeys).toEqual([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
      ]);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      // toggle visibility
      showHideMeasure.toggleVisibility(AD_BIDS_BID_PRICE_MEASURE);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, false]);

      // toggle visibility (hiding all is not supported from UI)
      // showHideMeasure.toggleVisibility(AD_BIDS_IMPRESSIONS_MEASURE);
      // // assert visibility is persisted in AdBids and after mirroring using the url proto state
      // assertVisiblePartsWithMirroring(get(mock).data, [], undefined);
      // expect(get(showHideMeasure).selectedItems).toEqual([false, false]);

      // toggle back visibility
      showHideMeasure.toggleVisibility(AD_BIDS_BID_PRICE_MEASURE);
      // showHideMeasure.toggleVisibility(AD_BIDS_IMPRESSIONS_MEASURE);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_BID_PRICE_MEASURE, AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);
    });

    it("Toggle all visibility", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      // toggle all to not visible
      showHideMeasure.setAllToNotVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, false]);

      // toggle all back to visible
      showHideMeasure.setAllToVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_BID_PRICE_MEASURE, AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);
    });

    it("Meta query updates", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
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
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
    });

    it("Meta query updates with new measure and some measures selected", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      // toggle visibility of one measure
      showHideMeasure.toggleVisibility(AD_BIDS_BID_PRICE_MEASURE);

      mock.setMeasures(AD_BIDS_THREE_MEASURES);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_THREE_MEASURES);
      expect(get(showHideMeasure).availableKeys).toEqual([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
        AD_BIDS_PUBLISHER_COUNT_MEASURE,
      ]);
      // assert that the new measure is not visible since not all measures were selected
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [AD_BIDS_IMPRESSIONS_MEASURE],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, false, false]);
    });

    it("Meta query updates with new measure and all measures selected", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideMeasure = createShowHideMeasuresStore(AD_BIDS_NAME, mock);
      mock.setMeasures(AD_BIDS_INIT_MEASURES);
      expect(get(showHideMeasure).selectedItems).toEqual([true, true]);

      mock.setMeasures(AD_BIDS_THREE_MEASURES);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_THREE_MEASURES);
      expect(get(showHideMeasure).availableKeys).toEqual([
        AD_BIDS_IMPRESSIONS_MEASURE,
        AD_BIDS_BID_PRICE_MEASURE,
        AD_BIDS_PUBLISHER_COUNT_MEASURE,
      ]);
      // assert that the new measure is visible since all measures were selected
      assertVisiblePartsWithMirroring(
        get(mock).data,
        [
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
          AD_BIDS_PUBLISHER_COUNT_MEASURE,
        ],
        undefined,
      );
      expect(get(showHideMeasure).selectedItems).toEqual([true, true, true]);
    });
  });

  describe("Show/Hide dimensions", () => {
    it("Toggle individual visibility", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock,
      );
      expect(get(showHideDimensions).selectedItems).toEqual([]);

      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
      expect(get(showHideDimensions).availableKeys).toEqual([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);

      // toggle visibility
      showHideDimensions.toggleVisibility(AD_BIDS_PUBLISHER_DIMENSION);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([false, true]);

      // toggle visibility (hiding all is not supported from UI)
      // showHideDimensions.toggleVisibility(AD_BIDS_DOMAIN_DIMENSION);
      // // assert visibility is persisted in AdBids and after mirroring using the url proto state
      // assertVisiblePartsWithMirroring(get(mock).data, undefined, [
      //   AD_BIDS_DOMAIN_DIMENSION,
      // ]);
      // expect(get(showHideDimensions).selectedItems).toEqual([false, false]);

      // toggle back visibility
      showHideDimensions.toggleVisibility(AD_BIDS_PUBLISHER_DIMENSION);
      // showHideDimensions.toggleVisibility(AD_BIDS_DOMAIN_DIMENSION);
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
    });

    it("Toggle all visibility", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock,
      );
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);

      // toggle all to not visible
      showHideDimensions.setAllToNotVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, false]);

      // toggle all back to visible
      showHideDimensions.setAllToVisible();
      // assert visibility is persisted in AdBids and after mirroring using the url proto state
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);
    });

    it("Meta query updates", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock,
      );
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
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
      ]);
    });

    it("Meta query updates with new dimension and some dimensions selected", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock,
      );
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);

      // toggle visibility of one dimension
      showHideDimensions.toggleVisibility(AD_BIDS_PUBLISHER_DIMENSION);

      mock.setDimensions(AD_BIDS_THREE_DIMENSIONS);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_THREE_DIMENSIONS);
      expect(get(showHideDimensions).availableKeys).toEqual([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_COUNTRY_DIMENSION,
      ]);
      // assert that the new dimension is not visible since not all dimensions were selected
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_DOMAIN_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([
        false,
        true,
        false,
      ]);
    });

    it("Meta query updates with new dimension and all dimensions selected", () => {
      const mock = createMetricsMetaQueryMock(false);
      const showHideDimensions = createShowHideDimensionsStore(
        AD_BIDS_NAME,
        mock,
      );
      mock.setDimensions(AD_BIDS_INIT_DIMENSIONS);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true]);

      mock.setDimensions(AD_BIDS_THREE_DIMENSIONS);
      // we have to manually call sync since in the app it is handled by a reactive statement
      metricsExplorerStore.sync(AD_BIDS_NAME, AD_BIDS_WITH_THREE_DIMENSIONS);
      expect(get(showHideDimensions).availableKeys).toEqual([
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_COUNTRY_DIMENSION,
      ]);
      // assert that the new dimension is visible since all dimensions were selected
      assertVisiblePartsWithMirroring(get(mock).data, undefined, [
        AD_BIDS_PUBLISHER_DIMENSION,
        AD_BIDS_DOMAIN_DIMENSION,
        AD_BIDS_COUNTRY_DIMENSION,
      ]);
      expect(get(showHideDimensions).selectedItems).toEqual([true, true, true]);
    });
  });
});

function assertVisiblePartsWithMirroring(
  metrics: V1MetricsView,
  measures: Array<string> | undefined,
  dimensions: Array<string> | undefined,
) {
  assertVisiblePartsOfMetricsView(AD_BIDS_NAME, measures, dimensions);
  // create a mirror using the proto and assert that the visible selections are persisted
  createAdBidsMirrorInStore(metrics);
  assertVisiblePartsOfMetricsView(AD_BIDS_MIRROR_NAME, measures, dimensions);
}

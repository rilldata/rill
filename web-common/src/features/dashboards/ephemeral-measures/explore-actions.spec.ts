import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import type { NotificationMessage } from "@rilldata/web-common/lib/event-bus/events";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  deleteExploreEphemeralMeasure,
  saveExploreEphemeralMeasure,
} from "./explore-actions";

const PROFIT = {
  name: "profit",
  displayName: "Profit",
  expression: "impressions * 2",
};

describe("explore ephemeral measure actions", () => {
  let notifications: NotificationMessage[];
  let unsubscribe: () => void;

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    metricsExplorerStore.init(
      AD_BIDS_EXPLORE_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY,
      ),
    );
    notifications = [];
    unsubscribe = eventBus.on("notification", (n) => notifications.push(n));
  });

  afterEach(() => unsubscribe());

  const definitions = () =>
    get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME]
      .ephemeralMeasures ?? [];

  it("confirms creates, updates and deletes", () => {
    saveExploreEphemeralMeasure(AD_BIDS_EXPLORE_NAME, PROFIT, true);
    saveExploreEphemeralMeasure(
      AD_BIDS_EXPLORE_NAME,
      { ...PROFIT, displayName: "Net profit" },
      false,
    );
    expect(definitions()).toEqual([{ ...PROFIT, displayName: "Net profit" }]);

    deleteExploreEphemeralMeasure(
      AD_BIDS_EXPLORE_NAME,
      "profit",
      AD_BIDS_EXPLORE_INIT,
    );
    expect(definitions()).toEqual([]);
    expect(notifications.map((n) => n.message)).toEqual([
      m.dashboard_pivot_ephemeral_created({ name: "Profit" }),
      m.dashboard_pivot_ephemeral_updated({ name: "Net profit" }),
      m.dashboard_pivot_ephemeral_deleted({ name: "Net profit" }),
    ]);
  });

  it("does not confirm deleting a measure that does not exist", () => {
    deleteExploreEphemeralMeasure(
      AD_BIDS_EXPLORE_NAME,
      "profit",
      AD_BIDS_EXPLORE_INIT,
    );
    expect(notifications).toEqual([]);
  });
});

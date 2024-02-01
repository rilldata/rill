import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  AD_BIDS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
  AD_BIDS_SCHEMA,
  AD_BIDS_WITH_BOOL_DIMENSION,
  TestTimeConstants,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
import { createInExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

describe("toProto/fromProto", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_NAME);
  });

  beforeEach(() => {
    getLocalUserPreferences().set({
      timeZone: "UTC",
    });
  });

  it("backwards compatibility for time controls", () => {
    const metricsExplorer = getDefaultMetricsExplorerEntity(
      AD_BIDS_NAME,
      AD_BIDS_INIT_WITH_TIME,
      {
        timeRangeSummary: {
          min: TestTimeConstants.LAST_DAY.toISOString(),
          max: TestTimeConstants.NOW.toISOString(),
          interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
        },
      },
    );
    metricsExplorer.selectedTimeRange = {
      name: "LAST_SIX_HOURS",
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    } as any;
    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(metricsExplorer),
      AD_BIDS_INIT_WITH_TIME,
      AD_BIDS_SCHEMA,
    );
    expect(newState.selectedTimeRange?.name).toEqual("PT6H");
  });

  it("backwards compatibility for dimension values", () => {
    const metricsExplorer = getDefaultMetricsExplorerEntity(
      AD_BIDS_NAME,
      AD_BIDS_WITH_BOOL_DIMENSION,
      {
        timeRangeSummary: {
          min: TestTimeConstants.LAST_DAY.toISOString(),
          max: TestTimeConstants.NOW.toISOString(),
          interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
        },
      },
    );
    metricsExplorer.whereFilter.cond?.exprs?.push(
      createInExpression(AD_BIDS_PUBLISHER_IS_NULL_DOMAIN, ["false"]),
    );
    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(metricsExplorer),
      AD_BIDS_WITH_BOOL_DIMENSION,
      AD_BIDS_SCHEMA,
    );
    expect(newState.whereFilter?.cond?.exprs?.[0]).toEqual(
      createInExpression(AD_BIDS_PUBLISHER_IS_NULL_DOMAIN, [false]),
    );
  });
});

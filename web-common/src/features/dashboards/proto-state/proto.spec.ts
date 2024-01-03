import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  AD_BIDS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  TestTimeConstants,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
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
      }
    );
    metricsExplorer.selectedTimeRange = {
      name: "LAST_SIX_HOURS",
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    } as any;
    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(metricsExplorer),
      AD_BIDS_INIT_WITH_TIME
    );
    expect(newState.selectedTimeRange?.name).toEqual("PT6H");
  });
});

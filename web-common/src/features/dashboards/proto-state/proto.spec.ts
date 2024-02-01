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
      },
    );
    metricsExplorer.selectedTimeRange = {
      name: "LAST_SIX_HOURS",
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    } as any;
    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(metricsExplorer),
      AD_BIDS_INIT_WITH_TIME,
    );
    expect(newState.selectedTimeRange?.name).toEqual("PT6H");
  });

  it("backwards compatibility for dimension values", () => {
    const newState = getDashboardStateFromUrl(
      decodeURIComponent(
        "CgUKA1A0VxgFIgkKB3JpbGwtUFAqEnA5MF9kZXRlY3Rpb25fdGltZTgAQhJwOTBfZGV0ZWN0aW9uX3RpbWVSGGlzX2RldGVjdGVkX3VuZGVyX2Ffd2Vla1IqaGFzX3NlY29uZF9zY3JhcGluZ19zZXNzaW9uX2FmdGVyX29jY3VycmVkUiJldmVudF9vY2N1cnJlZF9kYXRlX2lzX2FwcHJveGltYXRlUgpldmVudF90eXBlUgxldmVudF9zb3VyY2VgBGoHRXRjL1VUQ3gCgAEBmAH%252F%252F%252F%252F%252F%252F%252F%252F%252F%252F%252F8BogFxGm8ICBJrGmkIBxIsGioICRIaChhpc19kZXRlY3RlZF91bmRlcl9hX3dlZWsSChIIGgYidHJ1ZSISNxo1CAkSJAoiZXZlbnRfb2NjdXJyZWRfZGF0ZV9pc19hcHByb3hpbWF0ZRILEgkaByJmYWxzZSI%253D",
      ),
      AD_BIDS_INIT_WITH_TIME,
    );
    console.log(newState);
  });
});

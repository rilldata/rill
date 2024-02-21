import { protoBase64, Value } from "@bufbuild/protobuf";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getDefaultMetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  AD_BIDS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
  AD_BIDS_SCHEMA,
  AD_BIDS_WITH_BOOL_DIMENSION,
  TestTimeConstants,
} from "@rilldata/web-common/features/dashboards/stores/dashboard-stores-test-data";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import { DashboardState } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

describe("toProto/fromProto", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_NAME);
  });

  beforeEach(() => {
    getLocalUserPreferences().updateTimeZone("UTC");
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
    const message = new DashboardState({
      filters: new MetricsViewFilter({
        include: [
          new MetricsViewFilter_Cond({
            name: AD_BIDS_PUBLISHER_DIMENSION,
            in: [
              new Value({
                kind: {
                  case: "stringValue",
                  value: "Yahoo",
                },
              }),
              new Value({
                kind: {
                  case: "stringValue",
                  value: "Google",
                },
              }),
            ],
          }),
          new MetricsViewFilter_Cond({
            name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
            in: [
              new Value({
                kind: {
                  case: "stringValue",
                  value: "false",
                },
              }),
            ],
          }),
        ],
      }),
    });
    const proto = protoBase64.enc(message.toBinary());

    const newState = getDashboardStateFromUrl(
      proto,
      AD_BIDS_WITH_BOOL_DIMENSION,
      AD_BIDS_SCHEMA,
    );
    expect(newState.whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo", "Google"]),
        createInExpression(AD_BIDS_PUBLISHER_IS_NULL_DOMAIN, [false]),
      ]),
    );
  });

  it("test", () => {
    console.log(
      getDashboardStateFromUrl(
        decodeURIComponent(
          "CgUKA1A0VxgGKhBhdmdfcmVmcmVzaF9yYXRlOABIAVgBYARqB0V0Yy9VVEN4AoABAZgB%252F%252F%252F%252F%252F%252F%252F%252F%252F%252F%252F%252FAaIBOBo2CAgSGBoWCAkSDAoKaXNfc3VjY2VzcxIEEgIgARIYGhYICRIJCgdjb250ZXh0EgcSBRoDU1NM",
        ),
        AD_BIDS_WITH_BOOL_DIMENSION,
        AD_BIDS_SCHEMA,
      ),
    );
  });
});

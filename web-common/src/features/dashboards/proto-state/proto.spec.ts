import { protoBase64, Value } from "@bufbuild/protobuf";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { getFullInitExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-store-defaults";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_WITH_BOOL_DIMENSION,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_WITH_BOOL_DIMENSION,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
  TestTimeConstants,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  MetricsViewFilter,
  MetricsViewFilter_Cond,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/queries_pb";
import {
  DashboardState,
  DashboardState_ActivePage,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  type PivotChipData,
  PivotChipType,
} from "@rilldata/web-common/features/dashboards/pivot/types";

describe("toProto/fromProto", () => {
  // A legacy bookmark keeps the sort of a field that has since left the pivot.
  it("drops a pivot sort on a field that is not in the pivot", () => {
    const exploreState = getFullInitExploreState(
      AD_BIDS_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
        {
          timeRangeSummary: {
            min: TestTimeConstants.LAST_DAY.toISOString(),
            max: TestTimeConstants.NOW.toISOString(),
          },
        },
      ),
    );
    exploreState.activePage = DashboardState_ActivePage.PIVOT;
    exploreState.pivot = {
      ...exploreState.pivot,
      rows: [
        {
          id: AD_BIDS_PUBLISHER_DIMENSION,
          title: AD_BIDS_PUBLISHER_DIMENSION,
          type: PivotChipType.Dimension,
        },
      ],
      columns: [
        {
          id: AD_BIDS_IMPRESSIONS_MEASURE,
          title: AD_BIDS_IMPRESSIONS_MEASURE,
          type: PivotChipType.Measure,
        },
      ],
    };

    const roundTrip = (id: string) =>
      getDashboardStateFromUrl(
        getProtoFromDashboardState(
          {
            ...exploreState,
            pivot: { ...exploreState.pivot, sorting: [{ id, desc: true }] },
          },
          AD_BIDS_EXPLORE_INIT,
        ),
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
      ).pivot?.sorting;

    expect(roundTrip(AD_BIDS_IMPRESSIONS_MEASURE)).toEqual([
      { id: AD_BIDS_IMPRESSIONS_MEASURE, desc: true },
    ]);
    expect(roundTrip(AD_BIDS_BID_PRICE_MEASURE)).toEqual([]);
  });

  it("backwards compatibility for time controls", () => {
    const exploreState = getFullInitExploreState(
      AD_BIDS_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
        {
          timeRangeSummary: {
            min: TestTimeConstants.LAST_DAY.toISOString(),
            max: TestTimeConstants.NOW.toISOString(),
          },
        },
      ),
    );
    exploreState.selectedTimeRange = {
      name: "LAST_SIX_HOURS",
      interval: V1TimeGrain.TIME_GRAIN_HOUR,
    } as any;
    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(exploreState, AD_BIDS_EXPLORE_INIT),
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
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
      AD_BIDS_METRICS_WITH_BOOL_DIMENSION,
      AD_BIDS_EXPLORE_WITH_BOOL_DIMENSION,
    );
    expect(newState.whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo", "Google"]),
        createInExpression(AD_BIDS_PUBLISHER_IS_NULL_DOMAIN, [false]),
      ]),
    );
  });

  it("drops fields the metrics view no longer has from a legacy bookmark", () => {
    // The bookmark was saved while the explore still had these fields.
    const staleExploreSpec: V1ExploreSpec = {
      ...AD_BIDS_EXPLORE_INIT,
      measures: [...AD_BIDS_EXPLORE_INIT.measures!, "stale_measure"],
      dimensions: [...AD_BIDS_EXPLORE_INIT.dimensions!, "stale_dimension"],
    };
    const exploreState = getFullInitExploreState(
      AD_BIDS_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_INIT_WITH_TIME,
        staleExploreSpec,
      ),
    );
    exploreState.visibleMeasures = [
      AD_BIDS_IMPRESSIONS_MEASURE,
      "stale_measure",
    ];
    exploreState.allMeasuresVisible = false;
    exploreState.visibleDimensions = [
      "stale_dimension",
      AD_BIDS_PUBLISHER_DIMENSION,
    ];
    exploreState.allDimensionsVisible = false;
    exploreState.leaderboardSortByMeasureName = "stale_measure";
    exploreState.leaderboardMeasureNames = ["stale_measure"];
    exploreState.selectedComparisonDimension = "stale_dimension";
    exploreState.whereFilter = createAndExpression([
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
      createInExpression("stale_dimension", ["Google"]),
    ]);

    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(exploreState, staleExploreSpec),
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    expect(newState).toMatchObject({
      visibleMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      allMeasuresVisible: false,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      allDimensionsVisible: false,
      leaderboardSortByMeasureName: AD_BIDS_IMPRESSIONS_MEASURE,
      leaderboardMeasureNames: [AD_BIDS_IMPRESSIONS_MEASURE],
      whereFilter: createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
      ]),
    });
    expect(newState.selectedComparisonDimension).toBeUndefined();
  });
});

describe("fromProto pivot fields", () => {
  const dim = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Dimension,
  });
  const measure = (id: string): PivotChipData => ({
    id,
    title: id,
    type: PivotChipType.Measure,
  });

  it("keeps a field repeated in legacy state once", () => {
    // Legacy bookmarks store the pivot as proto state, bypassing the url parser.
    const exploreState = getFullInitExploreState(
      AD_BIDS_NAME,
      getInitExploreStateForTest(
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
      ),
    );
    exploreState.activePage = DashboardState_ActivePage.PIVOT;
    exploreState.pivot.rows = [dim("publisher"), dim("publisher")];
    exploreState.pivot.columns = [
      dim("publisher"),
      dim("domain"),
      measure("impressions"),
      measure("impressions"),
    ];

    const newState = getDashboardStateFromUrl(
      getProtoFromDashboardState(exploreState, AD_BIDS_EXPLORE_INIT),
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );

    expect(newState.pivot?.rows).toEqual([dim("publisher")]);
    expect(newState.pivot?.columns).toEqual([
      dim("domain"),
      measure("impressions"),
    ]);
  });
});

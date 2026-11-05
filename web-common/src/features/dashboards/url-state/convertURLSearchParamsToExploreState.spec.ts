import { protoBase64 } from "@bufbuild/protobuf";
import { toExpressionProto } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
  useTestFilterManager,
} from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  DashboardDimensionFilter,
  DashboardState,
  DashboardState_ActivePage,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1ExploreSpec,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { beforeEach, describe, expect, it } from "vitest";

// The filter manager validates the filter param against the metrics view specs, so an invalid
// filter has to be invalid for it as well.
const getFilterManager = useTestFilterManager({
  [AD_BIDS_METRICS_NAME]: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
});

const TestCases: {
  title: string;
  url: string;
  errors: string[];
  entity: Partial<ExploreState>;
}[] = [
  {
    title: "Invalid filter syntax: unknown syntax",
    url: "http://localhost/explore/AdBids_explore?f=abc",
    errors: ["Failed to parse filter: abc"],
    entity: {},
  },
  {
    title: "Invalid filter syntax: incorrect syntax",
    url: "http://localhost/explore/AdBids_explore?f=publisher+I+('ABC')",
    errors: [
      `Selected filter is invalid: Syntax error at line 1 col 12:

1 publisher I ('ABC')
             ^

Unexpected " ".`,
    ],
    entity: {},
  },
  {
    title: "Invalid filter: missing dimension",
    url: "http://localhost/explore/AdBids_explore?f=pub+IN+('ABC')",
    errors: [`Selected filter dimension: "pub" is not valid.`],
    entity: {},
  },
  {
    title: "Invalid filter: missing dimension in measure filter",
    url: "http://localhost/explore/AdBids_explore?f=publisher+IN+('ABC')+AND+pub+having+(impressions+lt+10)",
    errors: [`Selected filter dimension: "pub" is not valid.`],
    entity: {
      // partial valid filter is retained
      whereFilter: createAndExpression([
        createInExpression("publisher", ["ABC"]),
      ]),
    },
  },
  {
    title: "Invalid filter: missing measure in measure filter",
    url: "http://localhost/explore/AdBids_explore?f=publisher+IN+('ABC')+AND+publisher+having+(imp+lt+10)",
    errors: [`Selected filter field: "imp" is not valid.`],
    entity: {
      // partial valid filter is retained
      whereFilter: createAndExpression([
        createInExpression("publisher", ["ABC"]),
      ]),
    },
  },

  {
    title: "Time range without a grain is adaptive",
    url: "http://localhost/explore/AdBids_explore?tr=7D",
    errors: [],
    entity: {
      selectedTimeRange: {
        interval: "TIME_GRAIN_DAY",
        name: "7D",
      } as DashboardTimeControls,
      adaptiveTimeGrain: true,
    },
  },
  {
    title: "Fixed grain",
    url: "http://localhost/explore/AdBids_explore?tr=7D&grain=hour",
    errors: [],
    entity: {
      selectedTimeRange: {
        interval: "TIME_GRAIN_HOUR",
        name: "7D",
      } as DashboardTimeControls,
      adaptiveTimeGrain: false,
    },
  },
  {
    title: "Invalid time ranges",
    url: "http://localhost/explore/AdBids_explore?tr=abc&grain=xyz",
    errors: [
      `Selected time range: "abc" is not valid.`,
      `Selected time grain: "xyz" is not valid.`,
    ],
    entity: {},
  },
  {
    title: "Invalid time ranges: only comparison is invalid",
    url: "http://localhost/explore/AdBids_explore?tr=P4W&grain=week&compare_tr=xyz",
    errors: [`Selected compare time range: "xyz" is not valid.`],
    entity: {
      selectedTimeRange: {
        interval: "TIME_GRAIN_WEEK",
        name: "P4W",
      } as DashboardTimeControls,
      adaptiveTimeGrain: false,
    },
  },

  {
    title: "Invalid measure/dimension visibility selections",
    url: "http://localhost/explore/AdBids_explore?dims=pub,domain&measures=imps,bid_price",
    errors: [
      `Selected measure: "imps" is not valid.`,
      `Selected dimension: "pub" is not valid.`,
    ],
    entity: {
      visibleMeasures: ["bid_price"],
      allMeasuresVisible: false,
      visibleDimensions: ["domain"],
      allDimensionsVisible: false,
    },
  },
  {
    title: "Invalid sort by",
    url: "http://localhost/explore/AdBids_explore?sort_by=bp",
    errors: [`Selected sort by measure: "bp" is not valid.`],
    entity: {
      // defaults to 1st measure
      leaderboardSortByMeasureName: "impressions",
    },
  },
  {
    title: "Invalid expanded dimension",
    url: "http://localhost/explore/AdBids_explore?expand_dim=pub",
    errors: [`Selected expanded dimension: "pub" is not valid.`],
    entity: {
      // active page is set to default
      activePage: DashboardState_ActivePage.DEFAULT,
      selectedDimensionName: "",
    },
  },
];

/**
 * Has tests that call convertURLSearchParamsToExploreState directly.
 * Unlike url-state-variations that tests actions on explore state.
 * TODO: find a good unification of test cases so that a single set of cases run all variations.
 */
describe("direct convertURLSearchParamsToExploreState variations", () => {
  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  for (const { title, url, errors, entity } of TestCases) {
    it(title, () => {
      metricsExplorerStore.init(
        AD_BIDS_EXPLORE_NAME,
        getInitExploreStateForTest(
          AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
          AD_BIDS_EXPLORE_INIT,
          AD_BIDS_TIME_RANGE_SUMMARY,
        ),
      );
      const initState = getCleanMetricsExploreForAssertion();
      const defaultExplorePreset = getDefaultExplorePreset(
        AD_BIDS_EXPLORE_INIT,
        AD_BIDS_METRICS_INIT,
        AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
      );

      const errorsFromUrl = applyURLToExploreState(
        new URL(url),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
        getFilterManager(),
      );
      expect(errorsFromUrl.map((e) => e.message)).toEqual(errors);
      const currentState = getCleanMetricsExploreForAssertion();
      // current state should match the initial state
      expect(currentState).toEqual({
        ...initState,
        ...entity,
      });
    });
  }
});

describe("legacy state url param", () => {
  it("drops filters on fields the metrics view no longer has", () => {
    // The explore's last valid spec still names the stale fields, but its metrics view dropped them.
    const staleExploreSpec: V1ExploreSpec = {
      ...AD_BIDS_EXPLORE_INIT,
      measures: [...AD_BIDS_EXPLORE_INIT.measures!, "stale_measure"],
      dimensions: [...AD_BIDS_EXPLORE_INIT.dimensions!, "stale_dimension"],
    };
    const legacyState = new DashboardState({
      where: toExpressionProto(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
          createInExpression("stale_dimension", ["Google"]),
        ]),
      ),
      // Becomes a measure filter that selects both measures.
      having: [
        new DashboardDimensionFilter({
          name: AD_BIDS_DOMAIN_DIMENSION,
          filter: toExpressionProto(
            createAndExpression([
              createBinaryExpression(
                AD_BIDS_IMPRESSIONS_MEASURE,
                V1Operation.OPERATION_GT,
                10,
              ),
              createBinaryExpression(
                "stale_measure",
                V1Operation.OPERATION_GT,
                5,
              ),
            ]),
          ),
        }),
      ],
    });

    const { partialExploreState, errors } =
      convertURLSearchParamsToExploreState(
        new URLSearchParams({
          state: protoBase64.enc(legacyState.toBinary()),
        }),
        AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
        staleExploreSpec,
        {},
      );

    expect(partialExploreState.whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
        createSubQueryExpression(
          AD_BIDS_DOMAIN_DIMENSION,
          [AD_BIDS_IMPRESSIONS_MEASURE],
          createAndExpression([
            createBinaryExpression(
              AD_BIDS_IMPRESSIONS_MEASURE,
              V1Operation.OPERATION_GT,
              10,
            ),
          ]),
        ),
      ]),
    );
    expect(errors.map((e) => e.message)).toEqual(
      expect.arrayContaining([
        `Selected filter dimension: "stale_dimension" is not valid.`,
        `Selected filter field: "stale_measure" is not valid.`,
      ]),
    );
  });
});

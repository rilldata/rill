import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_TIME_RANGE_SUMMARY,
  AD_BIDS_METRICS_INIT,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import {
  applyURLToExploreState,
  getCleanMetricsExploreForAssertion,
} from "@rilldata/web-common/features/dashboards/url-state/url-state-variations.spec";
import {
  getLocalUserPreferences,
  initLocalUserPreferenceStore,
} from "@rilldata/web-common/features/dashboards/user-preferences";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

const TestCases: {
  title: string;
  url: string;
  errors: string[];
  entity: Partial<MetricsExplorerEntity>;
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
    title: "Invalid filter: missing dimension in measure filter",
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
      visibleMeasureKeys: new Set(["bid_price"]),
      allMeasuresVisible: false,
      visibleDimensionKeys: new Set(["domain"]),
      allDimensionsVisible: false,
    },
  },
  {
    title: "Invalid sort by",
    url: "http://localhost/explore/AdBids_explore?sort_by=bp",
    errors: [`Selected sort by measure: "bp" is not valid.`],
    entity: {
      // defaults to 1st measure
      leaderboardMeasureName: "impressions",
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

describe("Invalid Human readable URL State", () => {
  beforeAll(() => {
    initLocalUserPreferenceStore(AD_BIDS_EXPLORE_NAME);
  });

  beforeEach(() => {
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
    getLocalUserPreferences().updateTimeZone("UTC");
    localStorage.setItem(
      `${AD_BIDS_EXPLORE_NAME}-userPreference`,
      `{"timezone":"UTC"}`,
    );
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
        AD_BIDS_TIME_RANGE_SUMMARY,
      );

      const errorsFromUrl = applyURLToExploreState(
        new URL(url),
        AD_BIDS_EXPLORE_INIT,
        defaultExplorePreset,
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

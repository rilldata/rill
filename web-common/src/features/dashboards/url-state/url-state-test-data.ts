import { createAndExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";

export const URLStateTestMetricsExplorerEntity: Partial<MetricsExplorerEntity> =
  {
    activePage: DashboardState_ActivePage.DEFAULT,

    visibleMeasureKeys: new Set(["impressions", "bid_price"]),
    allMeasuresVisible: true,
    visibleDimensionKeys: new Set(["publisher", "domain"]),
    allDimensionsVisible: true,

    whereFilter: createAndExpression([]),
    dimensionThresholdFilters: [],

    selectedDimensionName: "",
    selectedTimezone: "UTC",
    sortDirection: 2,
    selectedTimeRange: {
      name: "inf",
    } as DashboardTimeControls,
    selectedComparisonDimension: "",
    selectedComparisonTimeRange: undefined,
    showTimeComparison: false,

    leaderboardMeasureName: "impressions",

    pivot: {
      active: false,
      activeCell: null,
      columnPage: 1,
      columns: {
        dimension: [],
        measure: [],
      },
      enableComparison: false,
      expanded: {},
      rowJoinType: "nest",
      rowPage: 1,
      rows: {
        dimension: [],
      },
      sorting: [],
    },

    tdd: {
      chartType: TDDChart.DEFAULT,
      expandedMeasureName: "",
      pinIndex: -1,
    },
  };

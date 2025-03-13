import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";

export function useDashboardFetchMocksForComponentTests() {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  mocks.mockMetricsViewAggregation(
    /BUILTIN_MEASURE_COUNT_DISTINCT.*"val":"%oo%"/,
    {
      data: [{ publisher__distinct_count: 3 }],
    },
  );
  mocks.mockMetricsViewAggregation(/"val":"%oo%"/, {
    data: [
      { publisher: "Facebook" },
      { publisher: "Google" },
      { publisher: "Yahoo" },
    ],
  });
  mocks.mockMetricsViewAggregation(
    /BUILTIN_MEASURE_COUNT_DISTINCT.*{"val":"Facebook"},{"val":"Google"}/,
    {
      data: [{ publisher__distinct_count: 2 }],
    },
  );
  mocks.mockMetricsViewAggregation(/{"val":"Facebook"},{"val":"Google"}/, {
    data: [{ publisher: "Facebook" }, { publisher: "Google" }],
  });
  mocks.mockMetricsViewAggregation(/publisher/, {
    data: [
      { publisher: null },
      { publisher: "Facebook" },
      { publisher: "Google" },
      { publisher: "Yahoo" },
      { publisher: "Microsoft" },
    ],
  });
  return mocks;
}

import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { act, screen, waitFor } from "@testing-library/svelte";
import { expect } from "vitest";

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

export async function addFilter(name: string) {
  await act(() => {
    screen.getByLabelText("Add filter button").click();
  });
  await waitFor(() =>
    expect(screen.getByRole("menuitem", { name })).toBeVisible(),
  );
  await act(() => {
    screen.getByRole("menuitem", { name }).click();
  });
  await waitFor(() =>
    expect(screen.queryByRole("menuitem", { name })).toBeNull(),
  );
}

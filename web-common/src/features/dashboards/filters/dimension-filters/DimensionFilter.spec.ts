import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import {
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/svelte";
import DimensionFilter from "./DimensionFilter.svelte";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";

describe("DimensionFilter", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  mocks.mockMetricsViewAggregation(/publisher/, {
    data: [
      { publisher: null },
      { publisher: "Facebook" },
      { publisher: "Google" },
      { publisher: "Yahoo" },
      { publisher: "Microsoft" },
    ],
  });

  it("Select mode filter", async () => {
    renderDimensionFilter(["Facebook", "Google"], undefined, undefined);
    await asyncWait(10); // TODO: better waiting

    expect(screen.queryByText("Search list")).not.toBeInTheDocument();

    // Open dropdown
    screen.getByLabelText("View filter", { selector: "button" }).click();
    await asyncWait(10); // TODO: better waiting

    expect(screen.getByLabelText("Search list")).toHaveValue("");
    expect(screen.getByRole("combobox")).toHaveTextContent("Select");
  });

  it("Search mode filter", async () => {
    renderDimensionFilter([], "oo", undefined);
    await asyncWait(10); // TODO: better waiting

    expect(screen.queryByText("Search list")).not.toBeInTheDocument();

    // Open dropdown
    screen.getByLabelText("View filter", { selector: "button" }).click();
    await asyncWait(10); // TODO: better waiting

    expect(screen.getByLabelText("Search list")).toHaveValue("oo");
    expect(screen.getByRole("combobox")).toHaveTextContent("Contains");
  });

  it("In List mode filter", async () => {
    renderDimensionFilter(["Facebook", "Google"], undefined, true);
    await asyncWait(10); // TODO: better waiting

    expect(screen.queryByText("Search list")).not.toBeInTheDocument();

    // Open dropdown
    screen.getByLabelText("View filter", { selector: "button" }).click();
    await asyncWait(10); // TODO: better waiting

    expect(screen.getByLabelText("Search list")).toHaveValue("Facebook,Google");
    expect(screen.getByRole("combobox")).toHaveTextContent("In List");
  });
});

function renderDimensionFilter(
  selectedValues: string[],
  searchText: string | undefined,
  isMatchList: boolean | undefined,
) {
  return render(DimensionFilter, {
    props: {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      metricsViewNames: [AD_BIDS_METRICS_NAME],
      label: AD_BIDS_PUBLISHER_DIMENSION,
      selectedValues,
      searchText,
      isMatchList,
      excludeMode: false,
      readOnly: false,
      timeStart: undefined,
      timeEnd: undefined,
      timeControlsReady: true,
      onRemove: () => {},
      onBulkSelect: () => {},
      onSelect: () => {},
      onSearch: () => {},
      onToggleFilterMode: () => {},
    },

    context: new Map([["$$_queryClient", queryClient]]),
  });
}

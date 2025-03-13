import { useDashboardFetchMocksForComponentTests } from "@rilldata/web-common/features/dashboards/filters/test/component-test-data";
import { renderFilterComponent } from "@rilldata/web-common/features/dashboards/filters/test/renderFilterComponent";
import {
  createAndExpression,
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { describe, it, expect } from "vitest";
import { screen, waitFor, fireEvent, act } from "@testing-library/svelte";
import { get } from "svelte/store";

describe("DimensionFilter", () => {
  const mocks = useDashboardFetchMocksForComponentTests();
  mocks.mockMetricsExplore(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_METRICS_INIT,
    AD_BIDS_EXPLORE_INIT,
  );

  it("Select mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    screen.getByLabelText("Add filter button").click();
    await waitFor(() =>
      expect(screen.getByRole("menuitem", { name: "publisher" })).toBeVisible(),
    );
    screen.getByRole("menuitem", { name: "publisher" }).click();

    await waitFor(() => expect(screen.getByText("Facebook")).toBeVisible());

    await act(() => {
      screen.getByText("Facebook").click();
      screen.getByText("Google").click();
    });
    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );

    screen.getByLabelText("publisher view filter").click();
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher Facebook +1 other",
    );
  });

  it("Search mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    screen.getByLabelText("Add filter button").click();
    await waitFor(() =>
      expect(screen.getByRole("menuitem", { name: "publisher" })).toBeVisible(),
    );
    screen.getByRole("menuitem", { name: "publisher" }).click();

    await waitFor(() => expect(screen.getByText("Facebook")).toBeVisible());
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "oo" },
      }),
    );

    await act(() => screen.getByRole("combobox").click());
    await act(() => screen.getByRole("option", { name: /Contains/ }).click());
    await waitFor(() =>
      expect(screen.getByLabelText("publisher results")).toHaveTextContent(
        "3 results",
      ),
    );
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createLikeExpression(AD_BIDS_PUBLISHER_DIMENSION, "%oo%"),
      ]),
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher Contains oo (3)",
    );
  });

  it("Bulk mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    screen.getByLabelText("Add filter button").click();
    await waitFor(() =>
      expect(screen.getByRole("menuitem", { name: "publisher" })).toBeVisible(),
    );
    screen.getByRole("menuitem", { name: "publisher" }).click();

    await waitFor(() => expect(screen.getByText("Facebook")).toBeVisible());
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );

    await waitFor(() =>
      expect(screen.getByLabelText("publisher results")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    const inExpr = createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
      "Facebook",
      "Google",
      "Apple",
    ]);
    (inExpr as any).isMatchList = true;
    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([inExpr]),
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });
});

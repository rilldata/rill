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
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { describe, it, expect } from "vitest";
import { screen, waitFor, fireEvent, act } from "@testing-library/svelte";
import { get } from "svelte/store";

describe("DimensionFilter", () => {
  const mocks = useDashboardFetchMocksForComponentTests();
  mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
  mocks.mockMetricsExplore(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_METRICS_INIT,
    AD_BIDS_EXPLORE_INIT,
  );

  it("Select mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    await addFilter("publisher");

    await act(() => {
      screen.getByText("Facebook").click();
      screen.getByText("Google").click();
    });
    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );

    await act(() => screen.getByRole("combobox").click());
    await act(() => screen.getByRole("option", { name: /Contains/ }).click());
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "oo" },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "3 results",
      ),
    );

    // "Contains" mode does not persist since Apply was not clicked
    await act(() => screen.getByLabelText("publisher view filter").click());
    await waitFor(() =>
      expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    await act(() => screen.getByLabelText("publisher view filter").click());
    await act(() => screen.getByRole("combobox").click());
    await act(() => screen.getByRole("option", { name: /In List/ }).click());
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "Facebook Google",
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // "In List" mode does not persist since Apply was not clicked
    await act(() => screen.getByLabelText("publisher view filter").click());
    await waitFor(() =>
      expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );
  });

  it("Search mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    await addFilter("publisher");

    await act(() => screen.getByRole("combobox").click());
    await act(() => screen.getByRole("option", { name: /Contains/ }).click());
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "0 results",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "no results",
    );

    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "oo" },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "3 results",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "Facebook Google Yahoo",
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher Contains oo (3)",
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

  it("Bulk mode filter using dropdown", async () => {
    const { stateManagers } = renderFilterComponent();

    await addFilter("publisher");

    await act(() => screen.getByRole("combobox").click());
    await act(() => screen.getByRole("option", { name: /In List/ }).click());
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "0 results",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "no results",
    );

    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "Facebook Google",
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // Adding a comma at the end doesnt add an extra element
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple," },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "Facebook Google",
    );

    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
          "Facebook",
          "Google",
          "Apple",
        ]),
      ]),
    );
    expect(get(stateManagers.dashboardStore).metadata).toEqual({
      dimensionInListFilter: { publisher: true },
    });
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });

  it("Bulk mode filter using search text", async () => {
    const { stateManagers } = renderFilterComponent();

    await addFilter("publisher");

    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "Facebook Google",
    );
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
          "Facebook",
          "Google",
          "Apple",
        ]),
      ]),
    );
    expect(get(stateManagers.dashboardStore).metadata).toEqual({
      dimensionInListFilter: { publisher: true },
    });
    expect(screen.getByLabelText("publisher view filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });
});

async function addFilter(name: string) {
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

import {
  addFilter,
  useDashboardFetchMocksForComponentTests,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import { renderFilterComponent } from "@rilldata/web-common/features/dashboards/filters/test/render-filter-component";
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
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { act, fireEvent, screen, waitFor } from "@testing-library/svelte";
import { get } from "svelte/store";
import { describe, expect, it } from "vitest";

describe("DimensionFilter", () => {
  const mocks = useDashboardFetchMocksForComponentTests();
  mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
  mocks.mockMetricsExplore(
    AD_BIDS_EXPLORE_NAME,
    AD_BIDS_METRICS_INIT,
    AD_BIDS_EXPLORE_INIT,
  );

  mockAnimationsForComponentTesting();

  it("Select mode filter", async () => {
    const { stateManagers } = renderFilterComponent();

    await addFilter("publisher");

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
    await act(() => screen.getByLabelText("Open publisher filter").click());
    await waitFor(() =>
      expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    await act(() => screen.getByLabelText("Open publisher filter").click());
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
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // "In List" mode does not persist since Apply was not clicked
    await act(() => screen.getByLabelText("Open publisher filter").click());
    await waitFor(() =>
      expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
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
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher Contains oo (3)",
    );
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createLikeExpression(AD_BIDS_PUBLISHER_DIMENSION, "%oo%"),
      ]),
    );
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
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
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
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
    expect(
      get(stateManagers.dashboardStore).dimensionsWithInlistFilter,
    ).toEqual(["publisher"]);
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
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
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
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
    expect(
      get(stateManagers.dashboardStore).dimensionsWithInlistFilter,
    ).toEqual(["publisher"]);
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });
});

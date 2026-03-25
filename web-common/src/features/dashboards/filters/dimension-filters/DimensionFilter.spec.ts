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
import { beforeAll, describe, expect, it } from "vitest";

// bits-ui 2.x Select uses PointerEvent APIs that jsdom doesn't support.
// Polyfill the missing types and methods so pointer-based interactions work in tests.
if (typeof globalThis.PointerEvent === "undefined") {
  (globalThis as Record<string, unknown>).PointerEvent =
    class PointerEvent extends MouseEvent {
      readonly pointerId: number;
      readonly pointerType: string;
      constructor(
        type: string,
        init?: PointerEventInit & Record<string, unknown>,
      ) {
        super(type, init);
        this.pointerId = (init?.pointerId as number) ?? 0;
        this.pointerType = (init?.pointerType as string) ?? "mouse";
      }
    };
}
if (!HTMLElement.prototype.hasPointerCapture) {
  HTMLElement.prototype.hasPointerCapture = () => false;
}
if (!HTMLElement.prototype.releasePointerCapture) {
  HTMLElement.prototype.releasePointerCapture = () => {};
}
if (!HTMLElement.prototype.scrollIntoView) {
  HTMLElement.prototype.scrollIntoView = () => {};
}

/**
 * bits-ui 2.x Select: open via keyboard (Space) on the trigger, then navigate
 * with ArrowDown and select with Enter. Pointer events don't work reliably
 * in jsdom because bits-ui's item ref tracking requires real layout.
 */
async function selectMode(name: RegExp) {
  const trigger = document.getElementById("dimension-filter-mode-selector")!;
  trigger.focus();
  // Open the select dropdown
  await act(() => {
    fireEvent.keyDown(trigger, { key: " " });
  });
  // Wait for options to appear
  await waitFor(() =>
    expect(screen.getByRole("option", { name })).toBeVisible(),
  );
  // bits-ui auto-highlights the current value on open. Navigate ArrowDown
  // until the target option is highlighted, then press Enter to select.
  const options = screen.getAllByRole("option");
  const targetIndex = options.findIndex((opt) => name.test(opt.textContent!));
  const highlightedOpt = options.findIndex((opt) =>
    opt.hasAttribute("data-highlighted"),
  );
  const startIndex = highlightedOpt >= 0 ? highlightedOpt : -1;
  const steps =
    targetIndex >= startIndex
      ? targetIndex - startIndex
      : options.length - startIndex + targetIndex;
  for (let i = 0; i < steps; i++) {
    await act(() => {
      fireEvent.keyDown(trigger, { key: "ArrowDown" });
    });
  }
  await act(() => {
    fireEvent.keyDown(trigger, { key: "Enter" });
  });
}

/**
 * Returns the text content of the mode selector trigger button.
 */
function getModeSelectorText() {
  return document.getElementById("dimension-filter-mode-selector")!.textContent;
}

/**
 * Returns the text of each item inside a DropdownMenu group, as an array.
 * bits-ui 2.x renders items as adjacent elements without whitespace, so
 * checking individual items is more reliable than toHaveTextContent.
 */
function getGroupItemTexts(groupLabel: string): string[] {
  const group = screen.getByLabelText(groupLabel);
  const items = group.querySelectorAll(
    "[data-dropdown-menu-item], [data-dropdown-menu-checkbox-item]",
  );
  return Array.from(items).map((el) => el.textContent?.trim() ?? "");
}

describe("DimensionFilter", () => {
  mockAnimationsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  beforeAll(() => {
    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT,
      AD_BIDS_EXPLORE_INIT,
    );
  });

  it("Select filter mode", async () => {
    const { stateManagers } = renderFilterComponent();

    // Add a filter pill for publisher
    await addFilter("publisher");

    // Once the pill is added and dropdown is open, select "Facebook" and "Google"
    await waitFor(() => expect(screen.getByText("Facebook")).toBeVisible());
    await act(() => screen.getByText("Facebook").click());
    await act(() => screen.getByText("Google").click());

    // Close the dropdown to apply the selections (Select mode applies on close)
    await act(() => screen.getByLabelText("Open publisher filter").click());

    // Assert that filters are now applied to the dashboard store
    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );

    // Reopen the dropdown
    await act(() => screen.getByLabelText("Open publisher filter").click());

    // Change the mode to "Contains" and enter a search term "oo"
    await selectMode(/Contains/);
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
    // Pill changes to reflect the current state of the dropdown
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher Contains oo (3)",
    );

    // Close the dropdown.
    await act(() => screen.getByLabelText("Open publisher filter").click());
    // "Contains" mode does not persist since Apply was not clicked
    await waitFor(() =>
      expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    // Open the dropdown again
    await act(() => screen.getByLabelText("Open publisher filter").click());
    // Switch to "In List" mode and enter a value "Facebook,Google,Apple"
    await selectMode(/In List/);
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
    expect(getGroupItemTexts("publisher results")).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);
    // Pill changes to reflect the current state of the dropdown
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // Close the dropdown
    await act(() => screen.getByLabelText("Open publisher filter").click());
    // "In List" mode does not persist since Apply was not clicked
    await waitFor(() =>
      expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );
  });

  it("Contains filter mode", async () => {
    const { stateManagers } = renderFilterComponent();

    // Add a filter pill for publisher
    await addFilter("publisher");

    // Change the mode to "Contains"
    await selectMode(/Contains/);
    // No results yet.
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "0 results",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "no results",
    );

    // Enter a search text "oo"
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "oo" },
      }),
    );
    // 3 results based on the mocked response.
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "3 results",
      ),
    );
    expect(getGroupItemTexts("publisher results")).toEqual([
      "Facebook",
      "Google",
      "Yahoo",
    ]);
    // Pill is updated as well.
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher Contains oo (3)",
    );

    // Apply to get the filter to take effect.
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    // Filter is added to the dashboard
    expect(get(stateManagers.dashboardStore).whereFilter).toEqual(
      createAndExpression([
        createLikeExpression(AD_BIDS_PUBLISHER_DIMENSION, "%oo%"),
      ]),
    );
    // Filter pill is persisted
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher Contains oo (3)",
    );
  });

  it("In-List filter mode using dropdown", async () => {
    const { stateManagers } = renderFilterComponent();

    // Add a filter pill for publisher
    await addFilter("publisher");

    // Change the mode to "In List"
    await selectMode(/In List/);
    // No results yet.
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "0 results",
      ),
    );
    expect(screen.getByLabelText("publisher results")).toHaveTextContent(
      "no results",
    );

    // Enter a search term with commas
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );
    // 2 of 3 results matched based on mocked response.
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(getGroupItemTexts("publisher results")).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);
    // Pill is updated as well.
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // Adding a comma at the end doesnt add an extra element
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple," },
      }),
    );
    // Same 2 of 3 matched results as before
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(getGroupItemTexts("publisher results")).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);

    // Apply to get the filter to take effect.
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    // Filter is added to the dashboard
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
    // Filter pill is persisted
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });

  it("In-List filter mode using search text", async () => {
    const { stateManagers } = renderFilterComponent();

    // Add a filter pill for publisher
    await addFilter("publisher");

    // Enter a search term with commas
    await act(() =>
      fireEvent.input(screen.getByLabelText("publisher search list"), {
        target: { value: "Facebook,Google,Apple" },
      }),
    );
    // Mode is automatically changed to In-List
    expect(getModeSelectorText()).toContain("In List");
    // 2 of 3 results matched based on mocked response.
    await waitFor(() =>
      expect(screen.getByLabelText("publisher result count")).toHaveTextContent(
        "2 of 3 matched",
      ),
    );
    expect(getGroupItemTexts("publisher results")).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);
    // Pill is updated as well.
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    // Apply to get the filter to take effect.
    await act(() => screen.getByRole("button", { name: "Apply" }).click());

    // Filter is added to the dashboard
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
    // Filter pill is persisted
    expect(screen.getByLabelText("Open publisher filter")).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });
});

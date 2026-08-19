import ExpressionFiltersTest from "@rilldata/web-common/features/dashboards/filters/test/ExpressionFiltersTest.svelte";
import {
  addFilter,
  applyMeasureFilter,
  closeFilter,
  closeMeasureFilter,
  fillMeasureFilterForm,
  getDimensionFilterModeText,
  getDimensionFilterResults,
  getFilterChip,
  getMeasureFilterChip,
  isMeasureFilterFormOpen,
  removeMeasureFilter,
  mockPointerEventsForComponentTesting,
  selectDimensionFilterMode,
  toggleFilter,
  toggleMeasureFilter,
  typeInDimensionFilterSearch,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
  waitForDimensionFilterResultCount,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import type { ActionResult } from "@sveltejs/kit";
import {
  type HoistedPageForExploreTests,
  PageMockForExploreTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  createAndExpression,
  createBetweenExpression,
  createBinaryExpression,
  createInExpression,
  createLikeExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getCleanMetricsExploreForAssertion } from "@rilldata/web-common/features/dashboards/url-state/test/url-state-test-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";

const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
    // superforms, which the measure filter form uses, registers a navigation guard for tainted
    // forms as soon as it is created.
    beforeNavigate: () => {},
  };
});
vi.mock("$app/forms", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$app/forms")>();
  return {
    ...actual,
    // The real `applyAction` needs the SvelteKit client runtime, which a component test does not
    // boot. The measure filter form reads its validation result back from the page, so hand the
    // result to the page mock instead.
    applyAction: (result: ActionResult) => {
      hoistedPage.applyAction(result);
      return Promise.resolve();
    },
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
    // superforms cancels a submit that navigates to another route, so it subscribes to this while
    // the measure filter form is submitting. Nothing here navigates away from the dashboard.
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});

// Url params for the dashboard before any filter is applied, coming from the yaml preset.
const PageURLForInitialState =
  "measures=impressions&dims=publisher&sort_type=percent&sort_dir=ASC";

// `getMeasureDisplayName` falls back to the measure expression when the measure has no display
// name, so the add filter menu and the measure chips read `count(*)` instead of `impressions`.
const AD_BIDS_IMPRESSIONS_MEASURE_LABEL = "count(*)";

// Tests filter interactions with the filter bar as the only source of state.
// Loading state from other sources is covered by DashboardStateManager.spec.ts.
describe("ExpressionFilters", () => {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();
  let pageMock!: PageMockForExploreTests;

  beforeEach(() => {
    pageMock = new PageMockForExploreTests(hoistedPage);

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
    mocks.mockMetricsExplore(AD_BIDS_EXPLORE_NAME, AD_BIDS_METRICS_INIT, {
      ...AD_BIDS_EXPLORE_INIT,
      defaultPreset: AD_BIDS_PRESET_WITHOUT_TIMESTAMP,
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
    metricsExplorerStore.remove(AD_BIDS_EXPLORE_NAME);
  });

  afterAll(waitForBodyScrollCleanup);

  it("Should apply a dimension filter selected in the filter bar", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectValues(["Facebook", "Google"]);
    // Select mode applies once the dropdown closes.
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );
    const filterUrlSearch = urlSearchWithFilter(
      `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
    );
    pageMock.assertSearchParams(filterUrlSearch);
    // Applying the filter should add a single entry to history.
    expect(pageMock.urlSearchHistory).toEqual([
      PageURLForInitialState,
      filterUrlSearch,
    ]);
  });

  it("Should not persist Contains and In List modes without applying them", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectValues(["Facebook", "Google"]);
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    // Change the mode to "Contains" and enter a search term "oo".
    await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/Contains/);
    await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "3 results",
    );
    // The chip reflects the current state of the dropdown.
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "publisher Contains oo (3)",
    );

    // Contains does not persist since Apply was not clicked.
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    // Same for "In List".
    await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/In List/);
    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook,Google,Apple",
    );
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "2 of 3 matched",
    );
    expect(getDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION)).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "publisher In list (2 of 3)",
    );

    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +1 other",
      ),
    );

    // Neither of the staged modes reached the dashboard.
    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );
  });

  it("Should apply a Contains filter", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/Contains/);

    // No results until there is a search text.
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "0 results",
    );
    await waitFor(() =>
      expect(
        screen.getByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} results`),
      ).toHaveTextContent("no results"),
    );

    await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "3 results",
    );
    expect(getDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION)).toEqual([
      "Facebook",
      "Google",
      "Yahoo",
    ]);

    // Enter applies, same as the Apply button the other modes use below.
    await pressEnter();

    assertWhereFilter(
      createAndExpression([
        createLikeExpression(AD_BIDS_PUBLISHER_DIMENSION, "%oo%"),
      ]),
    );
    pageMock.assertSearchParams(
      urlSearchWithFilter(`${AD_BIDS_PUBLISHER_DIMENSION} LIKE '%oo%'`),
    );
    // The chip keeps the applied filter.
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "publisher Contains oo (3)",
    );
  });

  it("Should apply an In List filter", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/In List/);

    // No results until there are values.
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "0 results",
    );
    await waitFor(() =>
      expect(
        screen.getByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} results`),
      ).toHaveTextContent("no results"),
    );

    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook,Google,Apple",
    );
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "2 of 3 matched",
    );
    expect(getDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION)).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);

    // A trailing comma does not add an extra value.
    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook,Google,Apple,",
    );
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "2 of 3 matched",
    );
    expect(getDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION)).toEqual([
      "Facebook",
      "Google",
      "Apple",
    ]);

    await apply();

    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
          "Facebook",
          "Google",
          "Apple",
        ]),
      ]),
    );
    expect(
      getCleanMetricsExploreForAssertion().dimensionsWithInlistFilter,
    ).toEqual([AD_BIDS_PUBLISHER_DIMENSION]);
    // In List filters get their own url syntax, so that they are restored in the same mode.
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN LIST ('Facebook','Google','Apple')`,
      ),
    );
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "publisher In list (2 of 3)",
    );
  });

  it("Should treat comma search text literally in Select mode", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook,Google,Apple",
    );

    // Select mode should not auto-switch to In List.
    expect(getDimensionFilterModeText()).toContain("Select");
    expect(
      screen.queryByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} result count`),
    ).not.toBeInTheDocument();
    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher",
      ),
    );
  });

  it("Should apply the exclude toggle when the dropdown closes", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectValues(["Facebook", "Google"]);
    await toggleExclude();

    expect(screen.getByLabelText("Include exclude toggle")).toHaveAttribute(
      "data-state",
      "checked",
    );
    // The chip shows the pending exclude before it is applied.
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "Exclude publisher",
    );
    expect(
      screen.getByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} filter`),
    ).toHaveClass("exclude");

    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    assertWhereFilter(
      createAndExpression([
        createInExpression(
          AD_BIDS_PUBLISHER_DIMENSION,
          ["Facebook", "Google"],
          true,
        ),
      ]),
    );
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} NIN ('Facebook','Google')`,
      ),
    );

    // Toggling it back returns the chip and the filter to include.
    await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await toggleExclude();
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).not.toHaveTextContent(
      "Exclude",
    );
    expect(
      screen.getByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} filter`),
    ).not.toHaveClass("exclude");
    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );
  });

  it("Should keep the values when switching from In List to Select", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/In List/);
    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      "Facebook,Google",
    );
    await waitForDimensionFilterResultCount(
      AD_BIDS_PUBLISHER_DIMENSION,
      "2 of 2 matched",
    );
    await apply();

    // Switching back to Select carries the values over and drops the in-list marker.
    await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/Select/);
    await waitFor(() =>
      expect(getDimensionFilterModeText()).toContain("Select"),
    );
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
      ]),
    );
    expect(
      getCleanMetricsExploreForAssertion().dimensionsWithInlistFilter,
    ).toEqual([]);
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
      ),
    );
  });

  it("Should block an In List filter that makes the url too long", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectDimensionFilterMode(/In List/);
    // Far past the ~13KB url limit, even after the params are compressed.
    await typeInDimensionFilterSearch(
      AD_BIDS_PUBLISHER_DIMENSION,
      randomFilterValues(500, 64).join(","),
    );

    await waitFor(() =>
      expect(
        screen.getByText("List is too long. Please remove some values."),
      ).toBeVisible(),
    );
    expect(screen.getByRole("button", { name: "Apply" })).toBeDisabled();

    // Enter is the keyboard equivalent of Apply, so it must be blocked as well.
    await pressEnter();
    expect(
      screen.getByLabelText(`${AD_BIDS_PUBLISHER_DIMENSION} search list`),
    ).toBeVisible();
    assertWhereFilter(createAndExpression([]));

    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);
    assertWhereFilter(createAndExpression([]));
  });

  it("Should apply a measure filter selected in the filter bar", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });
    await applyMeasureFilter();

    assertWhereFilter(
      createAndExpression([
        measureFilterExpression(V1Operation.OPERATION_GT, 10),
      ]),
    );
    const filterUrlSearch = urlSearchWithFilter(
      `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
    );
    pageMock.assertSearchParams(filterUrlSearch);
    // Applying the filter should add a single entry to history.
    expect(pageMock.urlSearchHistory).toEqual([
      PageURLForInitialState,
      filterUrlSearch,
    ]);
    expect(
      getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
    ).toHaveTextContent("count(*) for publisher > 10");
  });

  it("Should not apply a measure filter until Apply is clicked", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });

    // Closing the popover discards the form, unlike a dimension filter in Select mode.
    await closeMeasureFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);

    expect(
      getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
    ).toHaveTextContent(/^count\(\*\)$/);
    assertWhereFilter(createAndExpression([]));
    pageMock.assertSearchParams(PageURLForInitialState);
    expect(pageMock.urlSearchHistory).toEqual([PageURLForInitialState]);
  });

  it("Should require a dimension and a numeric value", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    // A new chip has neither a dimension nor a value, so Apply cannot go through.
    await applyMeasureFilter();

    // The only message is the one for the value. The dimension select is not wired to the form
    // errors, so a missing dimension keeps the popover open without saying why.
    await waitFor(() => expect(screen.getByText("Required")).toBeVisible());
    expect(isMeasureFilterFormOpen()).toBe(true);

    // A value alone is not enough either, since the filter is a subquery on a dimension.
    await fillMeasureFilterForm({ value1: "10" });
    await applyMeasureFilter();
    expect(isMeasureFilterFormOpen()).toBe(true);

    await fillMeasureFilterForm({ dimension: /^publisher$/, value1: "abc" });
    await applyMeasureFilter();

    await waitFor(() =>
      expect(screen.getByText("Value must be a valid number")).toBeVisible(),
    );
    expect(isMeasureFilterFormOpen()).toBe(true);
    assertWhereFilter(createAndExpression([]));
    pageMock.assertSearchParams(PageURLForInitialState);
  });

  it("Should apply a Between filter with two values", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    // The second value is only part of the form for the range operations.
    expect(document.getElementById("value2")).toBeNull();

    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Between$/,
      value1: "10",
    });
    await waitFor(() =>
      expect(document.getElementById("value2")).toBeInTheDocument(),
    );
    await fillMeasureFilterForm({ value2: "20" });
    await applyMeasureFilter();

    assertWhereFilter(
      createAndExpression([
        createSubQueryExpression(
          AD_BIDS_PUBLISHER_DIMENSION,
          [AD_BIDS_IMPRESSIONS_MEASURE],
          createBetweenExpression(AD_BIDS_IMPRESSIONS_MEASURE, 10, 20, false),
        ),
      ]),
    );
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10 AND ${AD_BIDS_IMPRESSIONS_MEASURE} LT 20)`,
      ),
    );
    expect(
      getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
    ).toHaveTextContent("count(*) for publisher (10,20)");
  });

  it("Should edit an applied measure filter", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });
    await applyMeasureFilter();

    // Reopening the chip starts from the applied filter.
    await toggleMeasureFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await waitFor(() => expect(isMeasureFilterFormOpen()).toBe(true));
    await fillMeasureFilterForm({
      operation: /^Less Than Or Equals$/,
      value1: "20",
    });
    await applyMeasureFilter();

    assertWhereFilter(
      createAndExpression([
        measureFilterExpression(V1Operation.OPERATION_LTE, 20),
      ]),
    );
    const editedFilterUrlSearch = urlSearchWithFilter(
      `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} LTE 20)`,
    );
    pageMock.assertSearchParams(editedFilterUrlSearch);
    // The edit is a single entry in history, same as the first apply.
    expect(pageMock.urlSearchHistory).toEqual([
      PageURLForInitialState,
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
      ),
      editedFilterUrlSearch,
    ]);
    expect(
      getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
    ).toHaveTextContent("count(*) for publisher <= 20");
  });

  it("Should remove a measure filter", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });
    await applyMeasureFilter();

    await removeMeasureFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);

    await waitFor(() =>
      expect(
        screen.queryByLabelText(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
      ).toBeNull(),
    );
    assertWhereFilter(createAndExpression([]));
    pageMock.assertSearchParams(PageURLForInitialState);
  });

  it("Should apply a measure filter on Enter", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });

    // The form submits on a window level Enter, which is why the selects in it have to be
    // committed with Space.
    await pressEnter();
    await waitFor(() => expect(isMeasureFilterFormOpen()).toBe(false));

    assertWhereFilter(
      createAndExpression([
        measureFilterExpression(V1Operation.OPERATION_GT, 10),
      ]),
    );
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
      ),
    );
  });

  it("Should keep dimension and measure filters side by side", async () => {
    await renderExpressionFilters();

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectValues(["Facebook", "Google"]);
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
    await fillMeasureFilterForm({
      dimension: /^publisher$/,
      operation: /^Greater Than$/,
      value1: "10",
    });
    await applyMeasureFilter();

    // Dimension filters lead the AND expression, and the single `f` param holds both.
    assertWhereFilter(
      createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
        measureFilterExpression(V1Operation.OPERATION_GT, 10),
      ]),
    );
    pageMock.assertSearchParams(
      urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google') AND ${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
      ),
    );
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
      "publisher Facebook +1 other",
    );
    expect(
      getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
    ).toHaveTextContent("count(*) for publisher > 10");
  });
});

// This needs to be there each file because of how hoisting works with vitest.
// TODO: find if there is a way to share code.
async function renderExpressionFilters() {
  const renderResults = render(ExpressionFiltersTest, {
    props: {
      exploreName: AD_BIDS_EXPLORE_NAME,
    },
    // TODO: we need to make sure every single query uses an explicit queryClient instead of the global one
    //       only then we can use a fresh client here.
    context: new Map<string | symbol, unknown>([
      ["$$_queryClient", queryClient],
      [
        RUNTIME_CONTEXT_KEY,
        new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
      ],
    ]),
  });
  await waitFor(() => expect(screen.getByText("Dashboard loaded!")));

  return { queryClient, renderResults };
}

/** Checks `values` in the open dimension filter dropdown, which is in Select mode. */
async function selectValues(values: string[]) {
  await waitFor(() => expect(screen.getByText(values[0])).toBeVisible());
  for (const value of values) {
    await act(() => screen.getByText(value).click());
  }
}

async function toggleExclude() {
  await act(() =>
    fireEvent.click(screen.getByLabelText("Include exclude toggle")),
  );
}

/**
 * Applies the staged filter, which Contains and In List modes need, and which is the only way a
 * measure filter reaches the dashboard.
 */
async function apply() {
  await act(() => screen.getByRole("button", { name: "Apply" }).click());
}

/** The expression a measure filter on `impressions`, split by `publisher`, produces. */
function measureFilterExpression(operation: V1Operation, value: number) {
  return createSubQueryExpression(
    AD_BIDS_PUBLISHER_DIMENSION,
    [AD_BIDS_IMPRESSIONS_MEASURE],
    createBinaryExpression(AD_BIDS_IMPRESSIONS_MEASURE, operation, value),
  );
}

/** The open dropdown applies the staged filter on a window level Enter. */
async function pressEnter() {
  await act(() => fireEvent.keyDown(window, { key: "Enter" }));
}

function assertWhereFilter(expected: unknown) {
  expect(getCleanMetricsExploreForAssertion().whereFilter).toEqual(expected);
}

/**
 * Deterministic pseudo random values. The url params are compressed, so a list of similar values
 * stays well under the length limit no matter how long it is.
 */
function randomFilterValues(count: number, length: number) {
  let seed = 42;
  const nextChar = () => {
    seed = (seed * 1103515245 + 12345) % 2 ** 31;
    return String.fromCharCode(97 + (seed % 26));
  };
  return Array.from({ length: count }, () =>
    Array.from({ length }, nextChar).join(""),
  );
}

function urlSearchWithFilter(filter: string) {
  return new URLSearchParams(
    `f=${filter}&${PageURLForInitialState}`,
  ).toString();
}

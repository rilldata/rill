import {
  AD_BIDS_IMPRESSIONS_MEASURE_LABEL,
  measureFilterExpression,
  testCommonFilterFlows,
  variantAssertions,
} from "@rilldata/web-common/features/dashboards/filters/test/expression-filters-suite";
import { useExploreFiltersVariant } from "@rilldata/web-common/features/dashboards/filters/test/explore-filters-variant";
import {
  addFilter,
  applyFilter,
  applyMeasureFilter,
  closeFilter,
  closeMeasureFilter,
  fillMeasureFilterForm,
  getDimensionFilterModeText,
  getDimensionFilterResults,
  getFilterChip,
  getMeasureFilterChip,
  getSelectAllButton,
  isMeasureFilterFormOpen,
  pressEnter,
  removeDimensionFilter,
  removeMeasureFilter,
  selectDimensionFilterMode,
  selectValues,
  toggleExclude,
  toggleFilter,
  toggleMeasureFilter,
  toggleSelectAll,
  typeInDimensionFilterSearch,
  waitForDimensionFilterResultCount,
  waitForDimensionFilterResults,
  waitForEmptyFilters,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import type { HoistedPageForExploreTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import {
  createAndExpression,
  createBetweenExpression,
  createInExpression,
  createLikeExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { screen, waitFor } from "@testing-library/svelte";
import type { ActionResult } from "@sveltejs/kit";
import { describe, expect, it, vi } from "vitest";

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
// The rune based url stores read `page` from here, and SvelteKit only populates it through its
// client router, which a component test does not boot. `PageMockForExploreTests` keeps this in sync
// with the `$app/stores` page below.
vi.mock("$app/state", async () => {
  return {
    page: (
      await import(
        "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
      )
    ).pageStateMock,
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

// Tests filter interactions on the explore dashboard, with the filter bar as the only source of
// state. Loading state from other sources is covered by DashboardStateManager.spec.ts.
//
// The flows that hold wherever the filter bar is rendered live in `expression-filters-suite.ts`, so
// that StandaloneExpressionFilters.spec.ts can run them too. What is left here is behaviour that
// does not depend on what consumes the filter, and so only needs to be covered once.
describe("ExploreExpressionFilters", () => {
  const variant = useExploreFiltersVariant(hoistedPage);
  const {
    initialUrlSearch,
    urlSearchWithFilter,
    assertWhereFilter,
    assertInListDimensions,
    assertUrlSearch,
    assertUrlSearchHistory,
  } = variantAssertions(variant);

  testCommonFilterFlows(variant);

  describe("Dimension filters", () => {
    it("Should not persist Contains and In List modes without applying them", async () => {
      await variant.render();

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
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
          ]),
        ]),
      );
    });

    it("Should apply a Contains filter", async () => {
      await variant.render();

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
      assertUrlSearch(
        urlSearchWithFilter(`${AD_BIDS_PUBLISHER_DIMENSION} LIKE '%oo%'`),
      );
      // The chip keeps the applied filter.
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Contains oo (3)",
      );
    });

    it("Should apply an In List filter", async () => {
      await variant.render();

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

      await applyFilter();

      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
            "Apple",
          ]),
        ]),
      );
      assertInListDimensions([AD_BIDS_PUBLISHER_DIMENSION]);
      // In List filters get their own url syntax, so that they are restored in the same mode.
      assertUrlSearch(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} IN LIST ('Facebook','Google','Apple')`,
        ),
      );
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher In list (2 of 3)",
      );
    });

    it("Should treat comma search text literally in Select mode", async () => {
      await variant.render();

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

    it("Should deselect all applied values when searching after a select all", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await waitForDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION, [
        "null",
        "Facebook",
        "Google",
        "Yahoo",
        "Microsoft",
      ]);

      // Without a search, Select all takes every value of the dimension, including the null one.
      expect(getSelectAllButton()).toHaveTextContent("Select all");
      await toggleSelectAll();
      await waitFor(() =>
        expect(getSelectAllButton()).toHaveTextContent("Deselect all"),
      );
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            null,
            "Facebook",
            "Google",
            "Yahoo",
            "Microsoft",
          ]),
        ]),
      );
      const filterUrlSearch = urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN (null,'Facebook','Google','Yahoo','Microsoft')`,
      );
      assertUrlSearch(filterUrlSearch);
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher +4 others",
      );

      // The applied values stay in the result list even when the search does not match them, so the
      // search that follows a select all does not narrow what Deselect all clears.
      await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
      await waitForDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
        "Google",
        "Yahoo",
        "null",
        "Microsoft",
      ]);

      expect(getSelectAllButton()).toHaveTextContent("Deselect all");
      await toggleSelectAll();
      await waitFor(() =>
        expect(getSelectAllButton()).toHaveTextContent("Select all"),
      );
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      // Microsoft and the null value do not match "oo", but they are cleared along with the matches.
      await waitForEmptyFilters();
      assertWhereFilter(createAndExpression([]));
      assertUrlSearch(initialUrlSearch);
      assertUrlSearchHistory(filterUrlSearch, initialUrlSearch);
    });

    it("Should apply the exclude toggle when the dropdown closes", async () => {
      await variant.render();

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
      assertUrlSearch(
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
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
          ]),
        ]),
      );
    });

    it("Should keep the values when switching from In List to Select", async () => {
      await variant.render();

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
      await applyFilter();

      // Switching back to Select carries the values over and drops the in-list marker.
      await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectDimensionFilterMode(/Select/);
      await waitFor(() =>
        expect(getDimensionFilterModeText()).toContain("Select"),
      );
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
          ]),
        ]),
      );
      assertInListDimensions([]);
      assertUrlSearch(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
        ),
      );
    });

    it("Should keep an applied Contains filter when Select is picked and dismissed", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectDimensionFilterMode(/Contains/);
      await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
      await waitForDimensionFilterResultCount(
        AD_BIDS_PUBLISHER_DIMENSION,
        "3 results",
      );
      await applyFilter();

      // Contains carries no values over to Select, so the dropdown has nothing staged. Select mode
      // has no Apply button, so committing that on dismiss would drop the filter with no way back.
      await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectDimensionFilterMode(/Select/);
      await waitFor(() =>
        expect(getDimensionFilterModeText()).toContain("Select"),
      );
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      assertWhereFilter(
        createAndExpression([
          createLikeExpression(AD_BIDS_PUBLISHER_DIMENSION, "%oo%"),
        ]),
      );
      assertUrlSearch(
        urlSearchWithFilter(`${AD_BIDS_PUBLISHER_DIMENSION} LIKE '%oo%'`),
      );
      await waitFor(() =>
        expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
          "publisher Contains oo (3)",
        ),
      );
    });

    it("Should block an In List filter that makes the url too long", async () => {
      await variant.render();

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

    it("Should remove a dimension filter", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectValues(["Facebook", "Google"]);
      // Select mode applies once the dropdown closes.
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      await removeDimensionFilter(AD_BIDS_PUBLISHER_DIMENSION);

      await waitForEmptyFilters();
      assertWhereFilter(createAndExpression([]));
      assertUrlSearch(initialUrlSearch);
      assertUrlSearchHistory(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
        ),
        initialUrlSearch,
      );
    });
  });

  describe("Measure filters", () => {
    it("Should apply a measure filter selected in the filter bar", async () => {
      await variant.render();

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
      assertUrlSearch(filterUrlSearch);
      // Applying the filter should add a single entry to history.
      assertUrlSearchHistory(filterUrlSearch);
      expect(
        getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
      ).toHaveTextContent("count(*) for publisher > 10");
    });

    it("Should not apply a measure filter until Apply is clicked", async () => {
      await variant.render();

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
      assertUrlSearch(initialUrlSearch);
      // Nothing was applied, so the dashboard never left the url it loaded with.
      assertUrlSearchHistory();
    });

    it("Should require a dimension and a numeric value", async () => {
      await variant.render();

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
      assertUrlSearch(initialUrlSearch);
    });

    it("Should apply a Between filter with two values", async () => {
      await variant.render();

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
      assertUrlSearch(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10 AND ${AD_BIDS_IMPRESSIONS_MEASURE} LT 20)`,
        ),
      );
      expect(
        getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
      ).toHaveTextContent("count(*) for publisher (10,20)");
    });

    it("Should edit an applied measure filter", async () => {
      await variant.render();

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
      assertUrlSearch(editedFilterUrlSearch);
      // The edit is a single entry in history, same as the first apply.
      assertUrlSearchHistory(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
        ),
        editedFilterUrlSearch,
      );
      expect(
        getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
      ).toHaveTextContent("count(*) for publisher <= 20");
    });

    it("Should apply a measure filter on Enter", async () => {
      await variant.render();

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
      assertUrlSearch(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
        ),
      );
    });

    it("Should remove a measure filter", async () => {
      await variant.render();

      await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
      await fillMeasureFilterForm({
        dimension: /^publisher$/,
        operation: /^Greater Than$/,
        value1: "10",
      });
      await applyMeasureFilter();

      await removeMeasureFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);

      await waitForEmptyFilters();
      assertWhereFilter(createAndExpression([]));
      assertUrlSearch(initialUrlSearch);
      assertUrlSearchHistory(
        urlSearchWithFilter(
          `${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
        ),
        initialUrlSearch,
      );
    });
  });
});

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

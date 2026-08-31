import {
  addFilter,
  applyMeasureFilter,
  clearFilters,
  closeFilter,
  fillMeasureFilterForm,
  getFilterChip,
  getMeasureFilterChip,
  getSelectAllButton,
  selectValues,
  toggleFilter,
  toggleSelectAll,
  typeInDimensionFilterSearch,
  waitForDimensionFilterResults,
  waitForEmptyFilters,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import type { PageMockForExploreTests } from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForExploreTests";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { waitFor } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";

/**
 * How a test reads the filter state back. Explore takes it from the dashboard store the filter bar
 * writes to, while a standalone filter bar has the manager itself as the only holder of the state.
 */
export interface ExpressionFilterManagerAccessor {
  /** The filter as it reached whatever consumes the filter bar. */
  getWhereFilter(): V1Expression | undefined;
  /** The dimensions whose filter is in In List mode, as they reached the consumer. */
  getInListDimensions(): string[];
}

/**
 * One way of rendering the filter bar. The tests in this file drive the UI identically for every
 * variant, so a variant only covers where the filter state ends up and how it reaches the url.
 */
export interface ExpressionFiltersVariant {
  /** Renders the filter bar and resolves once it is ready for interaction. */
  render(): Promise<void>;
  expressionFilterManager: ExpressionFilterManagerAccessor;
  /** Url search of the dashboard before any filter is applied. */
  initialUrlSearch: string;
  /** Url searches the dashboard goes through before the first filter is applied. */
  initialUrlSearchHistory: string[];
  /** Url search once `filter`, in filter param syntax, is applied. */
  urlSearchWithFilter(filter: string): string;
  /** A method rather than a field, since the page mock is rebuilt for every test. */
  pageMock(): PageMockForExploreTests;
}

/**
 * `getMeasureDisplayName` falls back to the measure expression when the measure has no display
 * name, so the add filter menu and the measure chips read `count(*)` instead of `impressions`.
 */
export const AD_BIDS_IMPRESSIONS_MEASURE_LABEL = "count(*)";

/**
 * The assertions that differ between variants, bound to one of them. Every test asserts through
 * these, so the test bodies themselves stay variant agnostic.
 */
export function variantAssertions(variant: ExpressionFiltersVariant) {
  return {
    initialUrlSearch: variant.initialUrlSearch,
    urlSearchWithFilter: (filter: string) =>
      variant.urlSearchWithFilter(filter),

    assertWhereFilter: (expected: unknown) =>
      expect(variant.expressionFilterManager.getWhereFilter()).toEqual(
        expected,
      ),

    assertInListDimensions: (expected: string[]) =>
      expect(variant.expressionFilterManager.getInListDimensions()).toEqual(
        expected,
      ),

    assertUrlSearch: (expectedSearch: string) =>
      variant.pageMock().assertSearchParams(expectedSearch),

    /**
     * Asserts that the dashboard went through exactly `searches` after it loaded, which is how the
     * tests catch extra history entries.
     */
    assertUrlSearchHistory: (...searches: string[]) =>
      expect(variant.pageMock().urlSearchHistory).toEqual([
        ...variant.initialUrlSearchHistory,
        ...searches,
      ]),
  };
}

/**
 * Flows that have to hold wherever the filter bar is rendered, so every variant runs them.
 *
 * Behaviour that is independent of what consumes the filter, such as the dimension dropdown modes
 * and the measure filter form, only needs to be covered once and stays in the explore spec.
 */
export function testCommonFilterFlows(variant: ExpressionFiltersVariant) {
  const {
    initialUrlSearch,
    urlSearchWithFilter,
    assertWhereFilter,
    assertUrlSearch,
    assertUrlSearchHistory,
  } = variantAssertions(variant);

  describe("Common filter flows", () => {
    it("Should apply a dimension filter selected in the filter bar", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectValues(["Facebook", "Google"]);
      // Select mode applies once the dropdown closes.
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
          ]),
        ]),
      );
      const filterUrlSearch = urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
      );
      assertUrlSearch(filterUrlSearch);
      // Applying the filter should add a single entry to history.
      assertUrlSearchHistory(filterUrlSearch);
    });

    it("Should select and deselect all search results in Select mode", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
      await waitForDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
        "Google",
        "Yahoo",
      ]);

      // Select all covers the values matching the search, not every value in the dimension.
      expect(getSelectAllButton()).toHaveTextContent("Select all");
      await toggleSelectAll();
      await waitFor(() =>
        expect(getSelectAllButton()).toHaveTextContent("Deselect all"),
      );
      // Select mode applies once the dropdown closes.
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
            "Yahoo",
          ]),
        ]),
      );
      const filterUrlSearch = urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google','Yahoo')`,
      );
      assertUrlSearch(filterUrlSearch);
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +2 others",
      );

      // Reopening with the same search shows every match selected, so the button deselects instead.
      await toggleFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await typeInDimensionFilterSearch(AD_BIDS_PUBLISHER_DIMENSION, "oo");
      await waitForDimensionFilterResults(AD_BIDS_PUBLISHER_DIMENSION, [
        "Facebook",
        "Google",
        "Yahoo",
      ]);

      expect(getSelectAllButton()).toHaveTextContent("Deselect all");
      await toggleSelectAll();
      await waitFor(() =>
        expect(getSelectAllButton()).toHaveTextContent("Select all"),
      );
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

      // Deselecting every value leaves an empty filter, so the url goes back to the initial state.
      await waitForEmptyFilters();
      assertWhereFilter(createAndExpression([]));
      assertUrlSearch(initialUrlSearch);
      assertUrlSearchHistory(filterUrlSearch, initialUrlSearch);
    });

    it("Should keep dimension and measure filters side by side", async () => {
      await variant.render();

      await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
      await selectValues(["Facebook", "Google"]);
      await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);
      const urlAfterDimensionFilter = urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google')`,
      );

      await addFilter(AD_BIDS_IMPRESSIONS_MEASURE_LABEL);
      await fillMeasureFilterForm({
        dimension: /^publisher$/,
        operation: /^Greater Than$/,
        value1: "10",
      });
      await applyMeasureFilter();
      const urlAfterBothFilters = urlSearchWithFilter(
        `${AD_BIDS_PUBLISHER_DIMENSION} IN ('Facebook','Google') AND ${AD_BIDS_PUBLISHER_DIMENSION} having (${AD_BIDS_IMPRESSIONS_MEASURE} GT 10)`,
      );

      // Dimension filters lead the AND expression, and the single `f` param holds both.
      assertWhereFilter(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, [
            "Facebook",
            "Google",
          ]),
          measureFilterExpression(V1Operation.OPERATION_GT, 10),
        ]),
      );
      assertUrlSearch(urlAfterBothFilters);
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toHaveTextContent(
        "publisher Facebook +1 other",
      );
      expect(
        getMeasureFilterChip(AD_BIDS_IMPRESSIONS_MEASURE_LABEL),
      ).toHaveTextContent("count(*) for publisher > 10");

      await clearFilters();
      assertWhereFilter(createAndExpression([]));
      assertUrlSearch(initialUrlSearch);
      assertUrlSearchHistory(
        urlAfterDimensionFilter,
        urlAfterBothFilters,
        initialUrlSearch,
      );
    });
  });
}

/** The expression a measure filter on `impressions`, split by `publisher`, produces. */
export function measureFilterExpression(operation: V1Operation, value: number) {
  return createSubQueryExpression(
    AD_BIDS_PUBLISHER_DIMENSION,
    [AD_BIDS_IMPRESSIONS_MEASURE],
    createBinaryExpression(AD_BIDS_IMPRESSIONS_MEASURE, operation, value),
  );
}

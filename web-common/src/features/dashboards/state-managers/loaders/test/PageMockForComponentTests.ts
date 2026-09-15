import type { afterNavigate } from "$app/navigation";
import { AD_BIDS_EXPLORE_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import type { ActionResult, AfterNavigate, Page } from "@sveltejs/kit";
import { writable, get, type Readable, type Updater } from "svelte/store";
import { expect } from "vitest";
import { setPageStateMock } from "./page-state.mock.svelte";

/**
 * To actually mock the page we need to hoist the variable using vi.hoisted and vi.mock.
 * To avoid having to rearrange imports we define an empty object and add methods to it.
 */
export type HoistedPageForComponentTests = Readable<Page> & {
  // A string is resolved against the current url, the way SvelteKit does.
  goto: (url: URL | string, opts?: { replaceState?: boolean }) => void;
  afterNavigate: typeof afterNavigate;
  applyAction: (result: ActionResult) => void;
};

/**
 * Handles mocking of page object and navigation.
 *
 * Usage
 * ```
 * const hoistedPage: HoistedPageForExploreTests = vi.hoisted(() => ({}) as any);
 *
 * vi.mock("$app/navigation", () => {
 *   return {
 *     goto: (url) => hoistedPage.goto(url),
 *     afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
 *   };
 * });
 * vi.mock("$app/stores", () => {
 *   return {
 *     page: hoistedPage,
 *   };
 * });
 * // Needed for components that read the url through the rune based `page`.
 * vi.mock("$app/state", async () => {
 *   return {
 *     page: (await import("./page-state.mock.svelte")).pageStateMock,
 *   };
 * });
 * // Only needed for components with a `sveltekit-superforms` form.
 * vi.mock("$app/forms", async (importOriginal) => {
 *   return {
 *     ...(await importOriginal<typeof import("$app/forms")>()),
 *     applyAction: (result: ActionResult) => {
 *       hoistedPage.applyAction(result);
 *       return Promise.resolve();
 *     },
 *   };
 * });
 *
 * ...
 * beforeEach(() => {
 *   pageMock = new PageMock(hoistedPage);
 * });
 * ...
 * ```
 */
export class PageMockForComponentTests {
  private readonly update: (updater: Updater<Page>) => void;
  private afterNavigateCallback:
    | ((navigation: AfterNavigate) => void)
    | undefined = undefined;
  // Save the url search history to assert that extra entries are not added.
  public urlSearchHistory: string[] = [];

  public constructor(
    private readonly hoistedPage: HoistedPageForComponentTests,
    private readonly exploreName = AD_BIDS_EXPLORE_NAME,
  ) {
    const initialPage = {
      url: new URL(`http://localhost/explore/${this.exploreName}`),
      params: { name: "AdBids_explore" },
      route: { id: "/explore/[name]" },
    } as any as Page;
    const { update, subscribe } = writable<Page>(initialPage);
    // Keep the `$app/state` page, which the rune based stores read, on the same page as the
    // `$app/stores` one.
    setPageStateMock(initialPage);
    this.update = (updater: Updater<Page>) =>
      update((page) => {
        const newPage = updater(page);
        setPageStateMock(newPage);
        return newPage;
      });

    hoistedPage.subscribe = subscribe;

    hoistedPage.goto = (
      url: URL | string,
      opts?: { replaceState?: boolean },
    ) => {
      this.update((page) => {
        page.url = typeof url === "string" ? new URL(url, page.url) : url;

        const search = normalizeSearch(page.url);

        // If replaceState is used then replace the last entry.
        if (opts?.replaceState && this.urlSearchHistory.length) {
          this.urlSearchHistory[this.urlSearchHistory.length - 1] = search;
        } else {
          this.urlSearchHistory.push(search);
        }

        return page;
      });
    };

    hoistedPage.afterNavigate = (
      callback: (navigation: AfterNavigate) => void,
    ) => {
      this.afterNavigateCallback = callback;
    };

    // Stand in for SvelteKit's `applyAction`, which puts the result of a form action on the page.
    // Forms built with `sveltekit-superforms` read their validation result back from there, so
    // without this their errors never reach the form.
    hoistedPage.applyAction = (result: ActionResult) => {
      this.update((page) => {
        page.status = "status" in result ? (result.status ?? 200) : 200;
        page.form = "data" in result ? result.data : undefined;
        return page;
      });
    };
  }

  public assertSearchParams(expectedSearch: string) {
    const actualSearch = normalizeSearch(get(this.hoistedPage).url);
    expect(sortSearchParams(actualSearch)).toEqual(
      sortSearchParams(expectedSearch),
    );
  }

  /**
   * Asserts the full url search history, ignoring the order params were set in.
   */
  public assertSearchHistory(expectedSearches: string[]) {
    expect(this.urlSearchHistory.map(sortSearchParams)).toEqual(
      expectedSearches.map(sortSearchParams),
    );
  }

  public gotoSearch(search: string) {
    const prevUrl = get(this.hoistedPage).url;
    this.update((page) => {
      page.url = new URL(
        `http://localhost/explore/${this.exploreName}?${search}`,
      );
      return page;
    });
    this.urlSearchHistory.push(search);
    this.afterNavigateCallback?.({
      from: { url: prevUrl },
      to: { url: get(this.hoistedPage).url },
      type: "goto",
    } as AfterNavigate);
  }

  public popState(search: string) {
    const prevUrl = get(this.hoistedPage).url;
    this.update((page) => {
      page.url = new URL(
        `http://localhost/explore/${this.exploreName}?${search}`,
      );
      return page;
    });
    this.urlSearchHistory.push(search);
    this.afterNavigateCallback?.({
      from: { url: prevUrl },
      to: { url: get(this.hoistedPage).url },
      type: "popstate",
    } as AfterNavigate);
  }

  public reset() {
    this.gotoSearch("");
    this.urlSearchHistory = [];
  }
}

function normalizeSearch(url: URL) {
  let normalizedSearch = url.searchParams.toString();
  // Correction for canvas that applies clear=true.
  // Instead of handling this at every place in tests, we just remove it.
  if (normalizedSearch === "clear=true") normalizedSearch = "";
  return normalizedSearch;
}

/**
 * Order in which `convertPartialExploreStateToUrlParams` adds params to the url.
 * Params added under more than one web view are ordered by their 1st occurrence.
 */
const UrlParamOrder: string[] = [
  ExploreStateURLParams.WebView,

  // Time controls, added by `toTimeRangesUrl`.
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeDimension,
  ExploreStateURLParams.TimeZone,
  ExploreStateURLParams.ComparisonTimeRange,
  ExploreStateURLParams.TimeGrain,
  ExploreStateURLParams.ComparisonDimension,
  ExploreStateURLParams.HighlightedTimeRange,

  ExploreStateURLParams.Filters,

  // Explore view, added by `toExploreUrlParams`.
  ExploreStateURLParams.VisibleMeasures,
  ExploreStateURLParams.VisibleDimensions,
  ExploreStateURLParams.ExpandedDimension,
  ExploreStateURLParams.SortBy,
  ExploreStateURLParams.SortType,
  ExploreStateURLParams.SortDirection,
  ExploreStateURLParams.LeaderboardMeasures,
  ExploreStateURLParams.LeaderboardShowContextForAllMeasures,
  ExploreStateURLParams.DynamicYAxisScale,
  ExploreStateURLParams.ChartType,

  // Time dimension detail view, added by `toTimeDimensionUrlParams`.
  ExploreStateURLParams.ExpandedMeasure,

  // Pivot view, added by `toPivotUrlParams`.
  ExploreStateURLParams.PivotRows,
  ExploreStateURLParams.PivotColumns,
  ExploreStateURLParams.PivotTableMode,
  ExploreStateURLParams.PivotRowLimit,
  ExploreStateURLParams.PivotShowTotalsColumn,
  ExploreStateURLParams.PivotShowTotalsRow,
  ExploreStateURLParams.PivotFormatting,
];
const UrlParamOrderMap = new Map(UrlParamOrder.map((p, i) => [p, i]));

/**
 * Sorts params in the order `convertPartialExploreStateToUrlParams` adds them,
 * so that the assertion is not sensitive to the order params were set in.
 * Unknown params retain their relative order and are placed at the end.
 */
function sortSearchParams(search: string) {
  const entries = [...new URLSearchParams(search).entries()];
  const sortedEntries = entries
    .map((entry, index) => ({ entry, index }))
    .sort((a, b) => {
      const aOrder = UrlParamOrderMap.get(a.entry[0]) ?? UrlParamOrder.length;
      const bOrder = UrlParamOrderMap.get(b.entry[0]) ?? UrlParamOrder.length;
      return aOrder === bOrder ? a.index - b.index : aOrder - bOrder;
    })
    .map(({ entry }) => entry);

  return new URLSearchParams(sortedEntries).toString();
}

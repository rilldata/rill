import type { afterNavigate } from "$app/navigation";
import { AD_BIDS_EXPLORE_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { AfterNavigate, Page } from "@sveltejs/kit";
import { writable, get, type Readable, type Updater } from "svelte/store";
import { expect } from "vitest";

/**
 * To actually mock the page we need to hoist the variable using vi.hoisted and vi.mock.
 * To avoid having to rearrange imports we define an empty object and add methods to it.
 */
export type HoistedPageForExploreTests = Readable<Page> & {
  // This is specifically for explore tests where goto is called with a url.
  // If we ever need it elsewhere then we need to handle string arguments as well.
  goto: (url: URL, opts?: { replaceState?: boolean }) => void;
  afterNavigate: typeof afterNavigate;
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
 *
 * ...
 * beforeEach(() => {
 *   pageMock = new PageMock(hoistedPage);
 * });
 * ...
 * ```
 */
export class PageMockForExploreTests {
  private readonly update: (updater: Updater<Page>) => void;
  private afterNavigateCallback: (navigation: AfterNavigate) => void;
  // Save the url search history to assert that extra entries are not added.
  public urlSearchHistory: string[] = [
    // This is to simulate user going to url without any search params 1st
    "",
  ];

  public constructor(
    private readonly hoistedPage: HoistedPageForExploreTests,
    private readonly exploreName = AD_BIDS_EXPLORE_NAME,
  ) {
    const { update, subscribe } = writable<Page>({
      url: new URL(`http://localhost/explore/${this.exploreName}`),
      params: { name: "AdBids_explore" },
    } as any);
    this.update = update;

    hoistedPage.subscribe = subscribe;

    hoistedPage.goto = (url: URL, opts?: { replaceState?: boolean }) => {
      update((page) => {
        page.url = url;

        // Trim the leading `?` to make assertions consistent.
        const trimmedSearch = page.url.search.replace(/^\?/, "");
        // If replaceState is used then replace the last entry.
        if (opts?.replaceState && this.urlSearchHistory.length) {
          this.urlSearchHistory[this.urlSearchHistory.length - 1] =
            trimmedSearch;
        } else {
          this.urlSearchHistory.push(trimmedSearch);
        }

        return page;
      });
    };

    hoistedPage.afterNavigate = (
      callback: (navigation: AfterNavigate) => void,
    ) => {
      this.afterNavigateCallback = callback;
    };
  }

  public assertSearchParams(expectedSearch: string) {
    expect(get(this.hoistedPage).url.searchParams.toString()).toEqual(
      new URLSearchParams(expectedSearch).toString(),
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
    this.afterNavigateCallback({
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
    this.afterNavigateCallback({
      from: { url: prevUrl },
      to: { url: get(this.hoistedPage).url },
      type: "popstate",
    } as AfterNavigate);
  }

  public reset() {
    this.urlSearchHistory = [];
    this.gotoSearch("");
  }
}

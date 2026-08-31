import type { afterNavigate } from "$app/navigation";
import { AD_BIDS_EXPLORE_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import type { ActionResult, AfterNavigate, Page } from "@sveltejs/kit";
import { writable, get, type Readable, type Updater } from "svelte/store";
import { expect } from "vitest";
import { setPageStateMock } from "./page-state.mock.svelte";

/**
 * To actually mock the page we need to hoist the variable using vi.hoisted and vi.mock.
 * To avoid having to rearrange imports we define an empty object and add methods to it.
 */
export type HoistedPageForExploreTests = Readable<Page> & {
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
export class PageMockForExploreTests {
  private readonly update: (updater: Updater<Page>) => void;
  private afterNavigateCallback: (navigation: AfterNavigate) => void;
  // Save the url search history to assert that extra entries are not added.
  public urlSearchHistory: string[] = [];

  public constructor(
    private readonly hoistedPage: HoistedPageForExploreTests,
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

    hoistedPage.goto = (url: URL | string, opts?: { replaceState?: boolean }) => {
      this.update((page) => {
        page.url = typeof url === "string" ? new URL(url, page.url) : url;

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
    this.gotoSearch("");
    this.urlSearchHistory = [];
  }
}

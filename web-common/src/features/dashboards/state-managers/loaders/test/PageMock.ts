import type { afterNavigate } from "$app/navigation";
import type { AfterNavigate, Page } from "@sveltejs/kit";
import { writable, get, type Readable, type Updater } from "svelte/store";
import { expect } from "vitest";

export type HoistedPage = Readable<Page> & {
  goto: (path: string, opts?: { replaceState?: boolean }) => void;
  afterNavigate: typeof afterNavigate;
};

export class PageMock {
  private readonly update: (updater: Updater<Page>) => void;
  private afterNavigateCallback: (navigation: AfterNavigate) => void;
  // Save the url search history to assert that extra entries are not added.
  public urlSearchHistory: string[] = [];

  public constructor(private readonly hoistedPage: HoistedPage) {
    const { update, subscribe } = writable<Page>({
      url: new URL("http://localhost/explore/AdBids_explore"),
      params: { name: "AdBids_explore" },
    } as any);
    this.update = update;

    hoistedPage.subscribe = subscribe;

    hoistedPage.goto = (path: string, opts?: { replaceState?: boolean }) => {
      update((page) => {
        page.url = new URL(`http://localhost${path}`);

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
      page.url = new URL("http://localhost/explore/AdBids_explore/?" + search);
      return page;
    });
    this.afterNavigateCallback({
      from: { url: prevUrl },
      to: { url: get(this.hoistedPage).url },
      type: "goto",
    } as AfterNavigate);
  }

  public popState(search: string) {
    const prevUrl = get(this.hoistedPage).url;
    this.update((page) => {
      page.url = new URL("http://localhost/explore/AdBids_explore/?" + search);
      return page;
    });
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

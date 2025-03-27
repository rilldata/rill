import type { afterNavigate } from "$app/navigation";
import type { AfterNavigate, Page } from "@sveltejs/kit";
import { writable, get, type Readable, type Updater } from "svelte/store";
import { expect } from "vitest";

export type HoistedPage = Readable<Page> & {
  goto: (path: string) => void;
  afterNavigate: typeof afterNavigate;
};

export class PageMock {
  private readonly update: (updater: Updater<Page>) => void;
  private afterNavigateCallback: (navigation: AfterNavigate) => void;

  public constructor(private readonly hoistedPage: HoistedPage) {
    const { update, subscribe } = writable<Page>({
      url: new URL("http://localhost/explore/AdBids_explore"),
      params: { name: "AdBids_explore" },
    } as any);
    this.update = update;

    hoistedPage.subscribe = subscribe;

    hoistedPage.goto = (path: string) => {
      update((page) => {
        page.url = new URL(`http://localhost${path}`);
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
    expect(get(this.hoistedPage).url.searchParams.toString()).toMatch(
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
}

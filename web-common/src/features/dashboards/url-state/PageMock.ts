import type { Page } from "@sveltejs/kit";
import { writable, get, type Readable } from "svelte/store";
import { vi, expect, type MockInstance } from "vitest";

export type PageMock = Readable<Page> & {
  updateState: (state: string) => void;
  goto: (path: string) => void;

  gotoSpy: MockInstance;
  assertSearchParams: (expectedSearch: string) => void;
};
export function createPageMock(pageMock: PageMock) {
  const { update, subscribe } = writable<Page>({
    url: new URL("http://localhost/explore/AdBids_explore"),
    params: { name: "AdBids_explore" },
  } as any);

  pageMock.subscribe = subscribe;
  pageMock.updateState = (state: string) => {
    update((page) => {
      if (state) {
        page.url = new URL(
          `http://localhost/explore/AdBids_explore?state=${encodeURIComponent(
            state,
          )}`,
        );
      } else {
        page.url = new URL("http://localhost/explore/AdBids_explore");
      }
      return page;
    });
  };
  pageMock.goto = (path: string) => {
    update((page) => {
      page.url = new URL(`http://localhost${path}`);
      return page;
    });
  };

  pageMock.gotoSpy = vi.spyOn(pageMock, "goto");
  pageMock.assertSearchParams = (expectedSearch: string) => {
    expect(get(pageMock).url.searchParams.toString()).toMatch(
      new URLSearchParams(expectedSearch).toString(),
    );
  };
}

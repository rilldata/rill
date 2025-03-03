import type { Page } from "@sveltejs/kit";
import { type Readable, writable } from "svelte/store";
import { beforeAll, describe, type MockInstance, vi } from "vitest";

const pageMock: PageMock = vi.hoisted(() => ({}) as any);

vi.mock("$app/navigation", () => {
  return {
    goto: (url) => pageMock.goto(url),
  };
});
vi.mock("$app/stores", () => {
  return {
    page: pageMock,
  };
});

describe.skip("url-utils", () => {
  beforeAll(() => {
    createPageMock();
  });

  // DELETED TESTS FOR UNUSED FUNCTIONS
  // HOWEVER, RETAIN THE TEST FOUNDATION FOR FUTURE TESTS
});

type PageMock = Readable<Page> & {
  updateState: (state: string) => void;
  goto: (path: string) => void;
  setUrl: (url: string) => void;
  gotoSpy: MockInstance;
};
function createPageMock() {
  const { update, subscribe } = writable<Page>({
    url: new URL("http://localhost/dashboard/AdBids"),
  } as any);

  pageMock.subscribe = subscribe;
  pageMock.goto = (path: string) => {
    update((page) => {
      page.url = new URL(`http://localhost${path}`);
      return page;
    });
  };
  pageMock.setUrl = (url: string) => {
    update((page) => {
      page.url = new URL(url);
      return page;
    });
  };
}

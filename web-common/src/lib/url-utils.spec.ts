import { getFullUrlForPath } from "@rilldata/web-common/lib/url-utils";
import type { Page } from "@sveltejs/kit";
import { Readable, writable } from "svelte/store";
import { beforeAll, describe, it, SpyInstance, vi, expect } from "vitest";

const pageMock: PageMock = vi.hoisted(() => ({} as any));

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

describe("url-utils", () => {
  beforeAll(() => {
    createPageMock();
  });

  describe("getFullUrlForPath", () => {
    const Variations: Array<[string, string, string]> = [
      ["/path/to/dashboard", "/new/path/dashboard", "/new/path/dashboard"],
      ["/path/to/dashboard", "new/path/dashboard", "/new/path/dashboard"],
      [
        "/path/to/dashboard/?feature=all",
        "new/path/dashboard",
        "/new/path/dashboard",
      ],
      [
        "/path/to/dashboard?feature=all&state=qwerty",
        "new/path/dashboard",
        "/new/path/dashboard?state=qwerty",
      ],
    ];
    for (const [currentPath, newPath, expectedPath] of Variations) {
      it(`${currentPath} => ${expectedPath}`, () => {
        pageMock.goto(currentPath);
        expect(getFullUrlForPath(newPath, ["state"])).toBe(expectedPath);
      });
    }
  });
});

type PageMock = Readable<Page> & {
  updateState: (state: string) => void;
  goto: (path: string) => void;
  gotoSpy: SpyInstance;
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
}

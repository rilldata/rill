import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  clearMetricsExplorerStore,
  createAdBidsInStore,
  createMetricsMetaQueryMock,
} from "@rilldata/web-common/features/dashboards/dashboard-stores-test-data";
import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
import type { Page } from "@sveltejs/kit";
import { get, Readable, writable } from "svelte/store";
import {
  beforeEach,
  beforeAll,
  it,
  describe,
  vi,
  expect,
  SpyInstance,
} from "vitest";

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

describe("useDashboardUrlSync", () => {
  beforeAll(() => {
    createPageMock();
  });

  beforeEach(() => {
    clearMetricsExplorerStore();
  });

  it("Changes from dashboard", () => {
    const metaMock = createMetricsMetaQueryMock();
    createAdBidsInStore();
    const unsubscribe = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    // needed to set the defaults correctly
    metricsExplorerStore.displayComparison(AD_BIDS_NAME, true);
    metricsExplorerStore.allDefaultsSelected(AD_BIDS_NAME);

    expect(pageMock.gotoSpy).toBeCalledTimes(0);

    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
      "Google"
    );
    expect(get(pageMock).url.searchParams.get("state")).toEqual(
      get(metricsExplorerStore).entities[AD_BIDS_NAME].proto
    );
    const protoWithFilter =
      get(metricsExplorerStore).entities[AD_BIDS_NAME].proto;
    expect(pageMock.gotoSpy).toBeCalledTimes(1);

    pageMock.goto("/dashboard/AdBids");
    expect(get(metricsExplorerStore).entities[AD_BIDS_NAME].proto).toEqual(
      get(metricsExplorerStore).entities[AD_BIDS_NAME].defaultProto
    );
    expect(get(metricsExplorerStore).entities[AD_BIDS_NAME].filters).toEqual({
      include: [],
      exclude: [],
    });
    expect(pageMock.gotoSpy).toBeCalledTimes(2);

    pageMock.updateState(protoWithFilter);
    expect(get(pageMock).url.searchParams.get("state")).toEqual(
      get(metricsExplorerStore).entities[AD_BIDS_NAME].proto
    );
    expect(get(metricsExplorerStore).entities[AD_BIDS_NAME].filters).toEqual({
      include: [
        {
          name: AD_BIDS_PUBLISHER_DIMENSION,
          in: ["Google"],
        },
      ],
      exclude: [],
    });
    expect(pageMock.gotoSpy).toBeCalledTimes(2);

    unsubscribe();
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
  pageMock.updateState = (state: string) => {
    update((page) => {
      if (state) {
        page.url = new URL(
          `http://localhost/dashboard/AdBids?state=${encodeURIComponent(state)}`
        );
      } else {
        page.url = new URL("http://localhost/dashboard/AdBids");
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
}

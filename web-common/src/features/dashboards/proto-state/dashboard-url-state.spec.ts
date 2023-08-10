import {
  MetricsExplorerEntity,
  metricsExplorerStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import {
  AD_BIDS_DEFAULT_TIME_RANGE,
  AD_BIDS_DEFAULT_URL_TIME_RANGE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXCLUDE_FILTER,
  AD_BIDS_INIT,
  AD_BIDS_INIT_MEASURES,
  AD_BIDS_MIRROR_NAME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  assertMetricsView,
  createDashboardState,
  createMetricsMetaQueryMock,
  resetDashboardStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores-test-data";
import { useDashboardUrlSync } from "@rilldata/web-common/features/dashboards/proto-state/dashboard-url-state";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
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
    initLocalUserPreferenceStore(AD_BIDS_NAME);
  });

  beforeEach(() => {
    resetDashboardStore();
    pageMock.goto("/dashboard/AdBids");
    pageMock.gotoSpy.mockClear();
  });

  it("Changes from dashboard", async () => {
    const metaMock = createMetricsMetaQueryMock();
    const unsubscribe = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    await wait();

    expect(pageMock.gotoSpy).toBeCalledTimes(0);

    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
      "Google"
    );
    await wait();
    assertUrlState(get(metricsExplorerStore).entities[AD_BIDS_NAME].proto);
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
    await wait();
    assertUrlState(get(metricsExplorerStore).entities[AD_BIDS_NAME].proto);
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

  it("Init load from url", async () => {
    gotoDashboardState(
      createDashboardState(AD_BIDS_NAME, AD_BIDS_INIT, AD_BIDS_EXCLUDE_FILTER)
    );
    const metaMock = createMetricsMetaQueryMock();
    const unsubscribe = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    await wait();

    assertUrlState(get(metricsExplorerStore).entities[AD_BIDS_NAME].proto);
    assertMetricsView(
      AD_BIDS_NAME,
      AD_BIDS_EXCLUDE_FILTER,
      AD_BIDS_DEFAULT_URL_TIME_RANGE
    );

    unsubscribe();
  });

  it("Changing active dashboard", async () => {
    const metaMock = createMetricsMetaQueryMock();
    let unsubscribe1 = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    let unsubscribe2 = useDashboardUrlSync(AD_BIDS_MIRROR_NAME, metaMock);
    await wait();

    metricsExplorerStore.toggleFilter(
      AD_BIDS_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
      "Google"
    );
    await wait();
    assertUrlState(get(metricsExplorerStore).entities[AD_BIDS_NAME].proto);
    assertMetricsView(
      AD_BIDS_NAME,
      {
        include: [
          {
            name: AD_BIDS_PUBLISHER_DIMENSION,
            in: ["Google"],
          },
        ],
        exclude: [],
      },
      AD_BIDS_DEFAULT_TIME_RANGE
    );

    unsubscribe1();
    pageMock.goto("/dashboard/AdBids_mirror");
    await wait();
    metricsExplorerStore.toggleFilter(
      AD_BIDS_MIRROR_NAME,
      AD_BIDS_DOMAIN_DIMENSION,
      "www.google.com"
    );
    await wait();
    assertUrlState(
      get(metricsExplorerStore).entities[AD_BIDS_MIRROR_NAME].proto
    );
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      {
        include: [
          {
            name: AD_BIDS_DOMAIN_DIMENSION,
            in: ["www.google.com"],
          },
        ],
        exclude: [],
      },
      AD_BIDS_DEFAULT_TIME_RANGE
    );

    // Going back to AdBids should retain the selected filters
    unsubscribe2();
    pageMock.goto("/dashboard/AdBids");
    unsubscribe1 = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    await wait();
    assertMetricsView(
      AD_BIDS_NAME,
      {
        include: [
          {
            name: AD_BIDS_PUBLISHER_DIMENSION,
            in: ["Google"],
          },
        ],
        exclude: [],
      },
      AD_BIDS_DEFAULT_TIME_RANGE
    );

    // Going back to AdBids_mirror should retain the selected filters
    pageMock.goto("/dashboard/AdBids_mirror");
    unsubscribe2 = useDashboardUrlSync(AD_BIDS_NAME, metaMock);
    await wait();
    assertMetricsView(
      AD_BIDS_MIRROR_NAME,
      {
        include: [
          {
            name: AD_BIDS_DOMAIN_DIMENSION,
            in: ["www.google.com"],
          },
        ],
        exclude: [],
      },
      AD_BIDS_DEFAULT_TIME_RANGE
    );

    unsubscribe1();
    unsubscribe2();
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

function assertUrlState(expected: string) {
  const actual = decodeURIComponent(
    get(pageMock).url.searchParams.get("state")
  );
  expect(actual).toEqual(expected);
}

function wait() {
  return new Promise((resolve) => setTimeout(resolve, 10));
}

function gotoDashboardState(state: MetricsExplorerEntity) {
  pageMock.goto(
    `/dashboard/${state.name}?state=${encodeURIComponent(
      getProtoFromDashboardState(state)
    )}`
  );
}

import "@testing-library/jest-dom/vitest";
import AlertDialogTest from "@rilldata/web-admin/features/alerts/AlertDialogTest.svelte";
import {
  mockPointerEventsForComponentTesting,
  mockResizeObserverForComponentTesting,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  type V1AlertSpec,
  type V1MetricsViewAggregationRequest,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import type { ActionResult } from "@sveltejs/kit";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import {
  afterAll,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";

// jsdom has no matchMedia, which the dashboard components read when they are imported; web-common's test setup stubs it the same way.
vi.hoisted(() => {
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }),
  });
});

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);
const editAlert = vi.hoisted(() =>
  vi.fn<
    (request: {
      data: {
        options: {
          resolver?: string;
          resolverProperties?: Record<string, unknown>;
          queryName?: string;
          queryArgsJson?: string;
          metricsViewName?: string;
        };
      };
    }) => Promise<unknown>
  >(),
);

vi.stubEnv("TZ", "UTC");

vi.mock("$app/navigation", () => {
  return {
    goto: (url, opts) => hoistedPage.goto(url, opts),
    afterNavigate: (cb) => hoistedPage.afterNavigate(cb),
    onNavigate: () => {},
    // superforms registers a navigation guard for tainted forms as soon as it is created.
    beforeNavigate: () => {},
  };
});
vi.mock("$app/forms", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$app/forms")>();
  return {
    ...actual,
    // superforms reads its validation result back from the page, so hand it to the page mock.
    applyAction: (result: ActionResult) => {
      hoistedPage.applyAction(result);
      return Promise.resolve();
    },
  };
});
vi.mock("$app/state", async () => {
  return {
    page: (
      await import(
        "@rilldata/web-common/features/dashboards/state-managers/loaders/test/page-state.mock.svelte"
      )
    ).pageStateMock,
  };
});
vi.mock("$app/stores", () => {
  return {
    page: hoistedPage,
    navigating: {
      subscribe: (run: (value: null) => void) => {
        run(null);
        return () => {};
      },
    },
  };
});
// The dialog saves through the admin API, which web-common unit tests cannot reach.
vi.mock("@rilldata/web-admin/client", async () => {
  const { readable } = await import("svelte/store");
  return {
    getAdminServiceListBookmarksQueryOptions: () => ({}),
    createAdminServiceGetCurrentUser: () =>
      readable({ data: { user: { email: "user@rilldata.com" } } }),
    createAdminServiceSearchProjectUsers: () => readable({ data: undefined }),
    createAdminServiceCreateAlert: () =>
      readable({ mutateAsync: vi.fn(), error: null }),
    createAdminServiceEditAlert: () =>
      readable({ mutateAsync: editAlert, error: null }),
  };
});

// The properties the ai resolver accepts (aiProps in runtime/resolvers/ai.go).
const AIResolverProperties = [
  "agent",
  "prompt",
  "time_range",
  "comparison_time_range",
  "time_zone",
  "explore",
  "dimensions",
  "measures",
  "where",
];

/** A formula alert created in the dialog, as the parser stores it: the legacy metrics resolver over the dialog's query. */
function getFormulaAlertSpec(): V1AlertSpec {
  const queryArgs: V1MetricsViewAggregationRequest = {
    metricsView: AD_BIDS_METRICS_NAME,
    measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
    dimensions: [],
    timeRange: { isoDuration: "P7D", timeZone: "UTC" },
    having: {
      cond: {
        op: V1Operation.OPERATION_AND,
        exprs: [
          {
            cond: {
              op: V1Operation.OPERATION_GT,
              exprs: [{ ident: AD_BIDS_IMPRESSIONS_MEASURE }, { val: 10 }],
            },
          },
        ],
      },
    },
  };
  const exploreState = getInitExploreStateForTest(
    AD_BIDS_METRICS_INIT_WITH_TIME,
    AD_BIDS_EXPLORE_INIT,
    AD_BIDS_TIME_RANGE_SUMMARY,
  );
  return {
    displayName: "Impressions alert",
    resolver: "legacy_metrics",
    resolverProperties: {
      query_name: "MetricsViewAggregation",
      query_args_json: JSON.stringify(queryArgs),
    },
    notifiers: [
      { connector: "email", properties: { recipients: ["user@rilldata.com"] } },
    ],
    annotations: {
      admin_owner_user_id: "owner-1",
      web_open_path: `/explore/${AD_BIDS_EXPLORE_NAME}`,
      web_open_state: getProtoFromDashboardState(
        exploreState,
        AD_BIDS_EXPLORE_INIT,
      ),
    },
  };
}

/** An AI alert created through the API: the ai resolver's properties, and no captured dashboard state. */
function getAPIAIAlertSpec(): V1AlertSpec {
  return {
    displayName: "Impressions drop",
    resolver: "ai",
    resolverProperties: {
      prompt: "Impressions dropped against the previous weeks",
      explore: AD_BIDS_EXPLORE_NAME,
    },
    notifiers: [
      { connector: "email", properties: { recipients: ["user@rilldata.com"] } },
    ],
    annotations: { admin_owner_user_id: "owner-1" },
  };
}

async function openEditDialog(alertSpec: V1AlertSpec) {
  render(AlertDialogTest, {
    props: { alertSpec },
    context: new Map<string | symbol, unknown>([
      ["$$_queryClient", queryClient],
      [
        RUNTIME_CONTEXT_KEY,
        new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
      ],
    ]),
  });
  await waitFor(() =>
    expect(screen.getByRole("button", { name: "Next" })).toBeTruthy(),
  );
}

async function typeCondition(text: string) {
  const condition = await waitFor(() =>
    screen.getByPlaceholderText(
      "Revenue dropped unusually against the previous weeks",
    ),
  );
  await fireEvent.input(condition, { target: { value: text } });
}

async function getSavedOptions() {
  await waitFor(() => expect(editAlert).toHaveBeenCalledOnce());
  return editAlert.mock.calls[0][0].data.options;
}

async function clickButton(name: string) {
  await act(() => screen.getByRole("button", { name }).click());
}

describe("the alert dialog", () => {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  // jsdom runs a `requestSubmit()` from inside a submit handler, which browsers ignore (see ScheduledReportDialog.spec.ts).
  beforeAll(() => {
    vi.spyOn(HTMLFormElement.prototype, "requestSubmit").mockImplementation(
      () => {},
    );
  });

  beforeEach(() => {
    new PageMockForComponentTests(hoistedPage);
    editAlert.mockReset();
    editAlert.mockResolvedValue({});

    mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT_WITH_TIME);
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_INIT_WITH_TIME,
      AD_BIDS_EXPLORE_INIT,
    );
    mocks.mockTimeRangeSummary(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!,
    );

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  it("converts a formula alert to an AI condition without carrying the formula's properties", async () => {
    await openEditDialog(getFormulaAlertSpec());

    // The mode is chosen on the data tab, since it changes which data fields are needed.
    await clickButton("AI condition");
    await clickButton("Next");
    await typeCondition("Impressions dropped against the previous weeks");
    await clickButton("Next");
    await clickButton("Update");

    const options = await getSavedOptions();
    expect(options.resolver).toBe("ai");
    expect(options.queryName).toBeUndefined();
    expect(options.queryArgsJson).toBeUndefined();
    // Only the ai resolver's own properties: the formula's query_name and query_args_json would fail under strict resolver properties.
    const props = options.resolverProperties ?? {};
    expect(props).toMatchObject({
      prompt: "Impressions dropped against the previous weeks",
      explore: AD_BIDS_EXPLORE_NAME,
    });
    for (const key of Object.keys(props)) {
      expect(AIResolverProperties).toContain(key);
    }
    // The formula's measure was not chosen for the AI, so it does not narrow what the AI looks at.
    expect(props).not.toHaveProperty("measures");
  });

  it("restores the formula's measure after leaving the data tab in AI mode", async () => {
    await openEditDialog(getFormulaAlertSpec());

    // The data tab unmounts while another tab is shown, so what it remembers must outlive it.
    await clickButton("AI condition");
    await clickButton("Next");
    await clickButton("Back");
    await clickButton("Criteria");
    await clickButton("Next");
    await clickButton("Next");
    await clickButton("Update");

    const options = await getSavedOptions();
    expect(options.resolver).toBeUndefined();
    const queryArgs = JSON.parse(
      options.queryArgsJson ?? "{}",
    ) as V1MetricsViewAggregationRequest;
    expect(queryArgs.measures?.[0]?.name).toBe(AD_BIDS_IMPRESSIONS_MEASURE);
  });

  it("edits a formula alert created through the API, which has no captured dashboard state", async () => {
    const alertSpec = getFormulaAlertSpec();
    delete alertSpec.annotations?.web_open_state;
    await openEditDialog(alertSpec);

    await clickButton("Next");
    await clickButton("Next");
    await clickButton("Update");

    const options = await getSavedOptions();
    expect(options.resolver).toBeUndefined();
    const queryArgs = JSON.parse(
      options.queryArgsJson ?? "{}",
    ) as V1MetricsViewAggregationRequest;
    expect(queryArgs.measures?.[0]?.name).toBe(AD_BIDS_IMPRESSIONS_MEASURE);
    expect(queryArgs.having?.cond?.exprs).toEqual([
      {
        cond: {
          op: V1Operation.OPERATION_GT,
          exprs: [{ ident: AD_BIDS_IMPRESSIONS_MEASURE }, { val: 10 }],
        },
      },
    ]);
  });

  it("keeps a comparison with the previous month saved as a rill time expression", async () => {
    const queryArgs: V1MetricsViewAggregationRequest = {
      metricsView: AD_BIDS_METRICS_NAME,
      measures: [
        { name: AD_BIDS_IMPRESSIONS_MEASURE },
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE + "_delta",
          comparisonDelta: { measure: AD_BIDS_IMPRESSIONS_MEASURE },
        },
        {
          name: AD_BIDS_IMPRESSIONS_MEASURE + "_delta_perc",
          comparisonRatio: { measure: AD_BIDS_IMPRESSIONS_MEASURE },
        },
      ],
      dimensions: [],
      timeRange: { expression: "MTD as of latest/D", timeZone: "UTC" },
      comparisonTimeRange: { expression: "MTD as of latest/D offset -1M" },
      having: {
        cond: {
          op: V1Operation.OPERATION_AND,
          exprs: [
            {
              cond: {
                op: V1Operation.OPERATION_LT,
                exprs: [
                  { ident: AD_BIDS_IMPRESSIONS_MEASURE + "_delta_perc" },
                  { val: -0.1 },
                ],
              },
            },
          ],
        },
      },
    };
    const alertSpec = getFormulaAlertSpec();
    alertSpec.resolverProperties = {
      ...alertSpec.resolverProperties,
      query_args_json: JSON.stringify(queryArgs),
    };
    await openEditDialog(alertSpec);

    await clickButton("Next");
    await waitFor(() =>
      expect(screen.getByText(/previous month/i)).toBeInTheDocument(),
    );
    await clickButton("Next");
    await clickButton("Update");

    const options = await getSavedOptions();
    const saved = JSON.parse(
      options.queryArgsJson ?? "{}",
    ) as V1MetricsViewAggregationRequest;
    expect(saved.comparisonTimeRange?.expression).toBe(
      "MTD as of latest/D offset -1M",
    );
  });

  it("waits for the explore before editing an AI alert created through the API", async () => {
    // Such an alert has no captured dashboard state. Hold the explore response until after the current user's,
    // which is when the dialog would otherwise start without knowing the alert's metrics view.
    let releaseExplore = () => {};
    const exploreGate = new Promise<void>((resolve) => {
      releaseExplore = resolve;
    });
    const mockedFetch = globalThis.fetch;
    vi.stubGlobal("fetch", async (url: string, init: RequestInit) => {
      if (String(url).endsWith("/GetExplore")) await exploreGate;
      return mockedFetch(url, init);
    });
    try {
      render(AlertDialogTest, {
        props: { alertSpec: getAPIAIAlertSpec() },
        context: new Map<string | symbol, unknown>([
          ["$$_queryClient", queryClient],
          [
            RUNTIME_CONTEXT_KEY,
            new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
          ],
        ]),
      });
      await new Promise((r) => setTimeout(r, 50));
      releaseExplore();
      await waitFor(() =>
        expect(screen.getByRole("button", { name: "Next" })).toBeTruthy(),
      );

      await clickButton("Next");
      await clickButton("Next");
      await clickButton("Update");

      const options = await getSavedOptions();
      // The managed alert's YAML refs MetricsView/<metricsViewName>, so it must not be empty.
      expect(options.metricsViewName).toBe(AD_BIDS_METRICS_NAME);
      expect(options.resolverProperties).toMatchObject({
        prompt: "Impressions dropped against the previous weeks",
        explore: AD_BIDS_EXPLORE_NAME,
      });
    } finally {
      vi.stubGlobal("fetch", mockedFetch);
    }
  });
});

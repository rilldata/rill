import "@testing-library/jest-dom/vitest";
import AlertDialogTest from "@rilldata/web-admin/features/alerts/AlertDialogTest.svelte";
import {
  getFilterChip,
  mockPointerEventsForComponentTesting,
  mockResizeObserverForComponentTesting,
  useDashboardFetchMocksForComponentTests,
  waitForBodyScrollCleanup,
} from "@rilldata/web-common/features/dashboards/filters/test/filter-test-utils";
import {
  type HoistedPageForComponentTests,
  PageMockForComponentTests,
} from "@rilldata/web-common/features/dashboards/state-managers/loaders/test/PageMockForComponentTests.ts";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  V1ExportFormat,
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
          refreshCron?: string;
          exportAttach?: boolean;
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
// The dialog saves through the admin API, which component tests cannot reach.
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
    createAdminServiceDeleteAlert: () =>
      readable({ mutateAsync: vi.fn(), error: null }),
  };
});

const AIConditionText = "Impressions dropped against the previous weeks";

/** A criteria alert created in the dialog, checked whenever its data refreshes. */
function getCriteriaAlertSpec(annotations: Record<string, string> = {}) {
  const queryArgs: V1MetricsViewAggregationRequest = {
    metricsView: AD_BIDS_METRICS_NAME,
    measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
    dimensions: [],
    where: createAndExpression([
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook"]),
    ]),
    timeRange: { isoDuration: "P7D", timeZone: "UTC" },
    having: createAndExpression([
      {
        cond: {
          op: V1Operation.OPERATION_GT,
          exprs: [{ ident: AD_BIDS_IMPRESSIONS_MEASURE }, { val: 10 }],
        },
      },
    ]),
  };
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
    exportAttach: true,
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
    annotations: {
      admin_owner_user_id: "owner-1",
      web_open_path: `/explore/${AD_BIDS_EXPLORE_NAME}`,
      ...annotations,
    },
  } satisfies V1AlertSpec;
}

/** The resolver properties the dialog writes for the criteria alert's data and the AI condition. */
const AIAlertResolverProperties = {
  prompt: AIConditionText,
  explore: AD_BIDS_EXPLORE_NAME,
  measures: [AD_BIDS_IMPRESSIONS_MEASURE],
  where: {
    cond: {
      op: "and",
      exprs: [
        {
          cond: {
            op: "in",
            exprs: [{ name: AD_BIDS_PUBLISHER_DIMENSION }, { val: "Facebook" }],
          },
        },
      ],
    },
  },
  time_range: { expression: "P7D" },
  time_zone: "UTC",
};

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
  // The filters and the time range are applied once the metrics view loads.
  await waitFor(() => {
    expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toBeVisible();
    expect(
      screen.getByLabelText(m.dashboard_select_time_range()),
    ).toHaveTextContent(getRangeLabel("P7D"));
  });
}

async function clickButton(name: string) {
  await act(() => screen.getByRole("button", { name }).click());
}

async function getSavedOptions() {
  await waitFor(() => expect(editAlert).toHaveBeenCalledOnce());
  return editAlert.mock.calls[0][0].data.options;
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
    mocks.mockResolvedRillTimes(AD_BIDS_METRICS_NAME, {
      P7D: { start: "2022-03-24T00:00:00Z", end: "2022-03-31T00:00:00Z" },
    });

    localStorage.clear();
    sessionStorage.clear();
    queryClient.clear();
  });

  afterAll(waitForBodyScrollCleanup);

  it("turns a criteria alert into an AI alert over the same data, checked on a schedule", async () => {
    await openEditDialog(getCriteriaAlertSpec());

    await clickButton(m.alert_form_condition_type_ai());
    await clickButton(m.alert_form_next());
    await fireEvent.input(
      screen.getByPlaceholderText(m.alert_form_ai_condition_placeholder()),
      { target: { value: AIConditionText } },
    );
    await clickButton(m.alert_form_next());
    // AI alerts can't be checked whenever their data refreshes, and have no rows to attach.
    expect(
      screen.getByRole("button", { name: m.alert_form_trigger_data_refresh() }),
    ).toBeDisabled();
    expect(screen.queryByLabelText(m.alert_form_attach_rows())).toBeNull();
    await clickButton(m.alert_form_update());

    const options = await getSavedOptions();
    expect(options).toMatchObject({
      resolver: "ai",
      resolverProperties: AIAlertResolverProperties,
      metricsViewName: AD_BIDS_METRICS_NAME,
      exportAttach: false,
    });
    expect(options.queryName).toBeUndefined();
    expect(options.queryArgsJson).toBeUndefined();
    expect(options.refreshCron).toMatch(/^\d+ \d+ \* \* \d$/);
  });

  it("saves an AI alert created through the API as it was", async () => {
    // Such an alert has no dashboard annotations, so its explore and metrics view are read from its resolver properties.
    await openEditDialog({
      displayName: "Impressions drop",
      resolver: "ai",
      resolverProperties: AIAlertResolverProperties,
      refreshSchedule: { cron: "0 9 * * 1", timeZone: "UTC" },
      notifiers: [
        {
          connector: "email",
          properties: { recipients: ["user@rilldata.com"] },
        },
      ],
      annotations: { admin_owner_user_id: "owner-1" },
    });

    await clickButton(m.alert_form_next());
    expect(
      screen.getByPlaceholderText(m.alert_form_ai_condition_placeholder()),
    ).toHaveValue(AIConditionText);
    await clickButton(m.alert_form_next());
    await clickButton(m.alert_form_update());

    expect(await getSavedOptions()).toMatchObject({
      resolver: "ai",
      resolverProperties: AIAlertResolverProperties,
      // The managed alert's YAML refs MetricsView/<metricsViewName>, so it must not be empty.
      metricsViewName: AD_BIDS_METRICS_NAME,
      refreshCron: "0 9 * * 1",
    });
  });

  it("does not offer an AI condition for an alert created from an embed", async () => {
    // The admin service only lets such alerts run a query, including when an admin edits one outside of the embed.
    await openEditDialog(
      getCriteriaAlertSpec({ admin_embed_owner_id: "embed-user-1" }),
    );

    expect(
      screen.queryByRole("button", { name: m.alert_form_condition_type_ai() }),
    ).toBeNull();
  });
});

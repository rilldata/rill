import {
  addFilter,
  closeFilter,
  getFilterChip,
  mockPointerEventsForComponentTesting,
  mockResizeObserverForComponentTesting,
  selectValues,
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
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import ScheduledReportDialogTest from "@rilldata/web-common/features/scheduled-reports/test/ScheduledReportDialogTest.svelte";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import {
  V1ExportFormat,
  type V1Expression,
  type V1MetricsViewAggregationRequest,
  type V1ReportSpec,
} from "@rilldata/web-common/runtime-client";
import {
  RUNTIME_CONTEXT_KEY,
  RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import type { ActionResult } from "@sveltejs/kit";
import { act, render, screen, waitFor } from "@testing-library/svelte";
import {
  afterAll,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";

// The SvelteKit mocks have to be declared in the spec file, since `vi.mock` is hoisted per file.
const hoistedPage: HoistedPageForComponentTests = vi.hoisted(() => ({}) as any);
const editReport = vi.hoisted(() =>
  vi.fn<
    (request: {
      data: {
        options: {
          queryArgsJson: string;
          exportAttach?: boolean;
          emailRecipients?: string[];
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
    // The real `applyAction` needs the SvelteKit client runtime, which a component test does not
    // boot. superforms reads its validation result back from the page, so hand it to the page mock.
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
    // The report's recipient is a project member, so the report can also run as the recipient.
    createAdminServiceListProjectMemberUsers: () =>
      readable({ data: { members: [{ userEmail: "user@rilldata.com" }] } }),
    createAdminServiceCreateReport: () =>
      readable({ mutateAsync: vi.fn(), error: null }),
    createAdminServiceEditReport: () =>
      readable({ mutateAsync: editReport, error: null }),
    createAdminServiceDeleteReport: () =>
      readable({ mutateAsync: vi.fn(), error: null }),
  };
});

// In an embed, the dialog saves with the embed's admin token instead.
vi.mock(
  "@rilldata/web-common/features/scheduled-reports/report-client-using-token.ts",
  async () => {
    const { readable } = await import("svelte/store");
    return {
      createAdminServiceEditReportUsingToken: () =>
        readable({ mutateAsync: editReport, error: null }),
      createAdminServiceDeleteReportUsingToken: () =>
        readable({ mutateAsync: vi.fn(), error: null }),
    };
  },
);

const PUBLISHER_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Facebook", "Google"]),
]);

function getReportSpec(
  where: V1Expression | undefined,
  webOpenMode = "creator",
): V1ReportSpec {
  const queryArgs: V1MetricsViewAggregationRequest = {
    metricsView: AD_BIDS_METRICS_NAME,
    dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
    measures: [{ name: AD_BIDS_IMPRESSIONS_MEASURE }],
    timeRange: { isoDuration: "P7D", timeZone: "UTC" },
    where,
  };
  return {
    displayName: "Weekly report",
    refreshSchedule: { cron: "0 9 * * 1", timeZone: "UTC" },
    queryName: "MetricsViewAggregation",
    queryArgsJson: JSON.stringify(queryArgs),
    exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
    notifiers: [
      { connector: "email", properties: { recipients: ["user@rilldata.com"] } },
    ],
    annotations: {
      explore: AD_BIDS_EXPLORE_NAME,
      web_open_mode: webOpenMode,
    },
  };
}

/** Opens the edit dialog for `reportSpec`, the way the report page does. */
async function openEditDialog(reportSpec: V1ReportSpec) {
  const rendered = render(ScheduledReportDialogTest, {
    props: { reportSpec, metricsViewName: AD_BIDS_METRICS_NAME },
    context: new Map<string | symbol, unknown>([
      ["$$_queryClient", queryClient],
      [
        RUNTIME_CONTEXT_KEY,
        new RuntimeClient({ host: "http://localhost", instanceId: "test" }),
      ],
    ]),
  });
  await waitFor(() =>
    expect(screen.getByLabelText("Filters form")).toBeVisible(),
  );
  return rendered;
}

/** Saves the report and returns the options the dialog sent to the admin API. */
async function saveAndGetOptions() {
  await act(() => screen.getByLabelText("Save report").click());
  await waitFor(() => expect(editReport).toHaveBeenCalledOnce());
  return editReport.mock.calls[0][0].data.options;
}

/** Saves the report and returns the `where` of the query the dialog sent to the admin API. */
async function saveAndGetWhere() {
  const options = await saveAndGetOptions();
  const queryArgs = JSON.parse(
    options.queryArgsJson,
  ) as V1MetricsViewAggregationRequest;
  return queryArgs.where;
}

describe("ScheduledReportDialog", () => {
  mockAnimationsForComponentTesting();
  mockPointerEventsForComponentTesting();
  mockResizeObserverForComponentTesting();
  const mocks = useDashboardFetchMocksForComponentTests();

  // In browsers a `requestSubmit()` from inside a submit handler is a no-op, since the form is
  // already firing its submission events. jsdom does not implement that flag, and the report form
  // depends on it: its submit handler calls superforms' `submit()`, which requests another submit.
  // Every submit in this file starts from the Save button, so nothing else reaches `requestSubmit`.
  beforeAll(() => {
    vi.spyOn(HTMLFormElement.prototype, "requestSubmit").mockImplementation(
      () => {},
    );
  });

  beforeEach(() => {
    new PageMockForComponentTests(hoistedPage);
    editReport.mockReset();
    editReport.mockResolvedValue({});

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

  it("saves a filter added in the dialog and shows it when the report is edited again", async () => {
    const firstDialog = await openEditDialog(getReportSpec(undefined));

    await addFilter(AD_BIDS_PUBLISHER_DIMENSION);
    await selectValues(["Facebook", "Google"]);
    // Select mode applies once the dropdown closes.
    await closeFilter(AD_BIDS_PUBLISHER_DIMENSION);

    const savedWhere = await saveAndGetWhere();
    expect(savedWhere).toEqual(PUBLISHER_FILTER);
    firstDialog.unmount();
    editReport.mockClear();

    // The report page refetches the report after a save, so the dialog reopens on the saved filter.
    await openEditDialog(getReportSpec(savedWhere));
    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toBeVisible(),
    );

    expect(await saveAndGetWhere()).toEqual(PUBLISHER_FILTER);
  });

  it("keeps the saved filters when the report is saved without touching them", async () => {
    await openEditDialog(getReportSpec(PUBLISHER_FILTER));

    await waitFor(() =>
      expect(getFilterChip(AD_BIDS_PUBLISHER_DIMENSION)).toBeVisible(),
    );

    expect(await saveAndGetWhere()).toEqual(PUBLISHER_FILTER);
  });

  it("keeps attaching the export when the report runs as the creator", async () => {
    await openEditDialog({ ...getReportSpec(undefined), exportAttach: true });

    const checkbox = screen.getByLabelText("Attach file to emails");
    expect(checkbox).toBeChecked();
    expect(checkbox).toBeEnabled();

    expect((await saveAndGetOptions()).exportAttach).toBe(true);
  });

  it("does not attach the export when the report runs as the recipient", async () => {
    await openEditDialog({
      ...getReportSpec(undefined, "recipient"),
      exportAttach: true,
    });

    const checkbox = screen.getByLabelText("Attach file to emails");
    expect(checkbox).not.toBeChecked();
    expect(checkbox).toBeDisabled();

    expect((await saveAndGetOptions()).exportAttach).toBe(false);
  });

  it("always attaches the export in an embed", async () => {
    const isEmbedded = vi.spyOn(EmbedStore, "isEmbedded").mockReturnValue(true);
    const getInstance = vi.spyOn(EmbedStore, "getInstance").mockReturnValue({
      userEmail: "user@rilldata.com",
      adminToken: "token",
    } as EmbedStore);
    try {
      await openEditDialog(getReportSpec(undefined));

      expect(screen.queryByLabelText("Attach file to emails")).toBeNull();
      expect(
        screen.getByText(
          "The file is attached to emails if it is 10 MB or smaller.",
        ),
      ).toBeVisible();

      expect((await saveAndGetOptions()).exportAttach).toBe(true);
    } finally {
      isEmbedded.mockRestore();
      getInstance.mockRestore();
    }
  });

  it("keeps the existing recipients when editing in an embed", async () => {
    const isEmbedded = vi.spyOn(EmbedStore, "isEmbedded").mockReturnValue(true);
    const getInstance = vi.spyOn(EmbedStore, "getInstance").mockReturnValue({
      userEmail: "embed@rilldata.com",
      adminToken: "token",
    } as EmbedStore);
    try {
      await openEditDialog(getReportSpec(undefined));

      // Recipients can be any address, so the email input is shown with the existing recipients
      expect(screen.getByText("user@rilldata.com")).toBeVisible();

      expect((await saveAndGetOptions()).emailRecipients).toEqual([
        "user@rilldata.com",
      ]);
    } finally {
      isEmbedded.mockRestore();
      getInstance.mockRestore();
    }
  });
});

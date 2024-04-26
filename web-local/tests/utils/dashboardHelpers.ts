import { expect, type Locator } from "@playwright/test";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import type { Page, Response } from "playwright";
import {
  clickMenuButton,
  openFileNavEntryContextMenu,
  updateCodeEditor,
  waitForValidResource,
} from "./commonHelpers";

export async function createDashboardFromSource(
  page: Page,
  sourcePath: string,
) {
  await openFileNavEntryContextMenu(page, sourcePath);
  await clickMenuButton(page, "Generate dashboard");
}

export async function createDashboardFromModel(page: Page, modelPath: string) {
  await openFileNavEntryContextMenu(page, modelPath);
  await clickMenuButton(page, "Generate dashboard");
}

export async function assertLeaderboards(
  page: Page,
  leaderboards: Array<{
    label: string;
    values: Array<string>;
  }>,
) {
  for (const { label, values } of leaderboards) {
    const leaderboardBlock = page.getByRole("grid", {
      name: `${label} leaderboard`,
    });
    await expect(leaderboardBlock).toBeVisible();

    const actualValues = await leaderboardBlock
      .locator(".leaderboard-entry > div:first-child")
      .allInnerTexts();
    expect(actualValues).toEqual(values);
  }
}

export type RequestMatcher = (response: Response) => boolean;

/**
 * Waits for a time series query to end.
 * Optionally takes a filter matcher: {@link metricsViewRequestFilterMatcher}.
 */
export async function waitForTimeSeries(
  page: Page,
  metricsView: string,
  filterMatcher?: RequestMatcher,
) {
  const timeSeriesUrlRegex = new RegExp(
    `/metrics-views/${metricsView}/timeseries`,
  );
  await page.waitForResponse(
    (response) =>
      timeSeriesUrlRegex.test(response.url()) &&
      (filterMatcher ? filterMatcher(response) : true),
  );
}

/**
 * Waits for a set of top list queries to end.
 * Optionally takes a filter matcher: {@link metricsViewRequestFilterMatcher}.
 */
export async function waitForTopLists(
  page: Page,
  metricsView: string,
  dimensions: Array<string>,
  filterMatcher?: RequestMatcher,
) {
  const topListUrlRegex = new RegExp(`/metrics-views/${metricsView}/toplist`);
  await Promise.all(
    dimensions.map((dimension) =>
      page.waitForResponse(
        (response) =>
          topListUrlRegex.test(response.url()) &&
          response.request().postDataJSON().dimensionName === dimension &&
          (filterMatcher ? filterMatcher(response) : true),
      ),
    ),
  );
}

/**
 * Waits for a set of top list queries to end.
 * Optionally takes a filter matcher: {@link metricsViewRequestFilterMatcher}.
 */
export async function waitForComparisonTopLists(
  page: Page,
  metricsView: string,
  dimensions: Array<string>,
  filterMatcher?: RequestMatcher,
) {
  const topListUrlRegex = new RegExp(
    `/metrics-views/${metricsView}/compare-toplist`,
  );
  await Promise.all(
    dimensions.map((dimension) =>
      page.waitForResponse(
        (response) =>
          topListUrlRegex.test(response.url()) &&
          response.request().postDataJSON().dimension.name === dimension &&
          (filterMatcher ? filterMatcher(response) : true),
      ),
    ),
  );
}

export type RequestMatcherFilter = { label: string; values: unknown[] };

/**
 * Helper to add a request matcher to match metrics view queries with certain filter
 */
export function metricsViewRequestFilterMatcher(
  response: Response,
  includeFilters: RequestMatcherFilter[],
  excludeFilters: RequestMatcherFilter[],
) {
  const filterRequest = response.request().postDataJSON().where as V1Expression;
  const includeFilterRequest = new Map<string, string[]>();
  const excludeFilterRequest = new Map<string, string[]>();

  if (filterRequest?.cond?.exprs) {
    for (const expr of filterRequest.cond.exprs) {
      if (!expr.cond?.exprs?.[0]?.ident) continue;
      if (expr.cond.op === "OPERATION_IN") {
        includeFilterRequest.set(
          expr.cond.exprs[0].ident,
          expr.cond.exprs.slice(1).map((e) => e.val as string),
        );
      } else if (expr.cond.op === "OPERATION_NIN") {
        excludeFilterRequest.set(
          expr.cond.exprs[0].ident,
          expr.cond.exprs.slice(1).map((e) => e.val as string),
        );
      }
    }
  }

  return (
    includeFilters.every(
      ({ label, values }) =>
        includeFilterRequest
          .get(label)
          ?.every((val) => values.indexOf(val) >= 0) ?? false,
    ) &&
    excludeFilters.every(
      ({ label, values }) =>
        excludeFilterRequest
          .get(label)
          ?.every((val) => values.indexOf(val) >= 0) ?? false,
    )
  );
}

// Helper that opens the time range menu, calls your interactions, and then waits until the menu closes
export async function interactWithTimeRangeMenu(
  page: Page,
  cb: () => void | Promise<void>,
) {
  // Open the menu
  await page.getByLabel("Select a time range").click();
  // Run the defined interactions
  await cb();
  // Wait for menu to close
  await expect(
    page.getByRole("menu", { name: "Select a time range" }),
  ).not.toBeVisible();
}

export async function interactWithComparisonMenu(
  page: Page,
  curLabel: string,
  cb: (l: Locator) => void | Promise<void>,
) {
  // Open the menu
  await page.getByRole("button", { name: curLabel }).click();
  // Run the defined interactions
  await cb(page.getByLabel("Comparison selector"));
  // Wait for menu to close
  await expect(
    page.getByRole("menu", { name: "Comparison selector" }),
  ).not.toBeVisible();
}

export async function waitForDashboard(page: Page) {
  return waitForValidResource(
    page,
    "AdBids_model_dashboard",
    "rill.runtime.v1.MetricsView",
  );
}

export async function updateAndWaitForDashboard(page: Page, code: string) {
  return Promise.all([
    updateCodeEditor(page, code),
    waitForValidResource(
      page,
      "AdBids_model_dashboard",
      "rill.runtime.v1.MetricsView",
    ),
  ]);
}

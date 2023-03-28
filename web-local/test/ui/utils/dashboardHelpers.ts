import { expect } from "@jest/globals";
import type {
  MetricsViewFilterCond,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import type { Page, Response } from "playwright";
import { expect as playwrightExpect } from "@playwright/test";
import { clickMenuButton, openEntityMenu } from "./helpers";

export async function createDashboardFromSource(page: Page, source: string) {
  await openEntityMenu(page, source);
  await clickMenuButton(page, "Autogenerate Dashboard");
}

export async function createDashboardFromModel(page: Page, model: string) {
  await openEntityMenu(page, model);
  await clickMenuButton(page, "Autogenerate Dashboard");
}

export async function assertLeaderboards(
  page: Page,
  leaderboards: Array<{
    label: string;
    values: Array<string>;
  }>
) {
  for (const { label, values } of leaderboards) {
    const leaderboardBlock = await page.locator("svelte-virtual-list-row", {
      hasText: label,
    });
    await playwrightExpect(leaderboardBlock).toBeVisible();

    const actualValues = await leaderboardBlock
      .locator(".leaderboard-entry")
      .locator("div[slot='title']")
      .allInnerTexts();
    expect(actualValues).toEqual(values);
  }
}

export async function clickOnFilter(
  page: Page,
  dimensionLabel: string,
  value: string
) {
  await page
    .locator("svelte-virtual-list-row", {
      hasText: dimensionLabel,
    })
    .getByText(value)
    .click();
}

export type RequestMatcher = (response: Response) => boolean;

/**
 * Waits for a time series query to end.
 * Optionally takes a filter matcher: {@link metricsViewRequestFilterMatcher}.
 */
export async function waitForTimeSeries(
  page: Page,
  metricsView: string,
  filterMatcher?: RequestMatcher
) {
  const timeSeriesUrlRegex = new RegExp(
    `/metrics-views/${metricsView}/timeseries`
  );
  await page.waitForResponse(
    (response) =>
      timeSeriesUrlRegex.test(response.url()) &&
      (filterMatcher ? filterMatcher(response) : true)
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
  filterMatcher?: RequestMatcher
) {
  const topListUrlRegex = new RegExp(`/metric-views/${metricsView}/toplist`);
  await Promise.all(
    dimensions.map((dimension) =>
      page.waitForResponse(
        (response) =>
          topListUrlRegex.test(response.url()) &&
          response.request().postDataJSON().dimensionName === dimension &&
          (filterMatcher ? filterMatcher(response) : true)
      )
    )
  );
}

export type RequestMatcherFilter = { label: string; values: Array<unknown> };

/**
 * Helper to add a request matcher to match metrics view queries with certain filter
 */
export function metricsViewRequestFilterMatcher(
  response: Response,
  includeFilters: Array<RequestMatcherFilter>,
  excludeFilters: Array<RequestMatcherFilter>
) {
  const filterRequest = response.request().postDataJSON()
    .filter as V1MetricsViewFilter;
  const includeFilterRequest = new Map<string, MetricsViewFilterCond>();
  filterRequest.include.forEach((cond) =>
    includeFilterRequest.set(cond.name, cond)
  );
  const excludeFilterRequest = new Map<string, MetricsViewFilterCond>();
  filterRequest.exclude.forEach((cond) =>
    excludeFilterRequest.set(cond.name, cond)
  );

  return (
    includeFilters.every(
      ({ label, values }) =>
        includeFilterRequest
          .get(label)
          ?.in.every((val) => values.indexOf(val) >= 0) ?? false
    ) &&
    excludeFilters.every(
      ({ label, values }) =>
        excludeFilterRequest
          .get(label)
          ?.in.every((val) => values.indexOf(val) >= 0) ?? false
    )
  );
}

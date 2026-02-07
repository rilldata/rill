import { expect, type Page, type Response } from "@playwright/test";
import { assertLeaderboards } from "web-local/tests/utils/metricsViewHelpers";
import {
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./commonHelpers";
import { createModel } from "./modelHelpers";
import { uploadFile, waitForSource } from "./sourceHelpers";

export interface TimeSeriesValue {
  ts: string;
  bin?: number;
  records: Record<string, number | null>;
}

export interface TimeSeriesResponse {
  data: TimeSeriesValue[];
}

export const AD_BIDS_METRICS_PATH = "/metrics/AdBids_model_metrics.yaml";
export const AD_BIDS_EXPLORE_PATH =
  "/dashboards/AdBids_model_metrics_explore.yaml";

export async function createAdBidsModel(page: Page) {
  await Promise.all([
    waitForSource(page, "/sources/AdBids.yaml", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    uploadFile(page, "AdBids.csv"),
  ]);

  await createModel(page, "AdBids_model.sql");
  await Promise.all([
    waitForProfiling(page, "AdBids_model", [
      "publisher",
      "domain",
      "timestamp",
    ]),
    updateCodeEditor(page, `select * from "AdBids"`),
  ]);
}

export async function assertAdBidsDashboard(page: Page) {
  await wrapRetryAssertion(() =>
    assertLeaderboards(page, [
      {
        label: "Publisher",
        values: ["null", "Facebook", "Google", "Yahoo", "Microsoft"],
      },
      {
        label: "Domain",
        values: [
          "facebook.com",
          "msn.com",
          "google.com",
          "news.yahoo.com",
          "instagram.com",
          "sports.yahoo.com",
          "news.google.com",
        ],
      },
    ]),
  );

  // Assert timeseries chart is rendered
  await assertTimeseriesChartRendered(page);
}

/**
 * Waits for a timeseries API response and returns the parsed data.
 * Must be called BEFORE the action that triggers the request.
 */
export function interceptTimeseriesResponse(
  page: Page,
): Promise<TimeSeriesResponse> {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      page.off("response", handler);
      reject(new Error("Timeout waiting for timeseries response"));
    }, 30000);

    const handler = async (response: Response) => {
      if (
        response.url().includes("/queries/metrics-views/") &&
        response.url().includes("/timeseries") &&
        response.request().method() === "POST"
      ) {
        try {
          const data = (await response.json()) as TimeSeriesResponse;
          clearTimeout(timeout);
          page.off("response", handler);
          resolve(data);
        } catch {
          // Response body might not be available, ignore and wait for next
        }
      }
    };

    page.on("response", handler);
  });
}

/**
 * Gets the chart container element for timeseries
 */
export function getChartContainer(page: Page) {
  // The chart SVG has an aria-label and contains path elements for the line
  return page
    .locator('svg[aria-label*="Measure Chart"]')
    .filter({ has: page.locator("path") })
    .first();
}

/**
 * Asserts that the timeseries chart is rendered with data
 */
export async function assertTimeseriesChartRendered(page: Page) {
  const chart = getChartContainer(page);
  await expect(chart).toBeVisible();

  const paths = chart.locator("path");
  const pathCount = await paths.count();
  expect(pathCount).toBeGreaterThan(0);
}

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
  // The chart SVG has role="application" and contains path elements for the line
  return page.locator('svg[role="application"]').filter({ has: page.locator("path") }).first();
}

/**
 * Hovers over a specific x-position on the chart to trigger tooltip
 * @param page Playwright page
 * @param xPercent Position as percentage of chart width (0-1), default 0.5 (middle)
 */
export async function hoverOnChart(page: Page, xPercent: number = 0.5) {
  const chart = getChartContainer(page);
  const box = await chart.boundingBox();
  if (!box) throw new Error("Chart container not found");

  const x = box.x + box.width * xPercent;
  const y = box.y + box.height * 0.5;

  // Move to position and wait for tooltip to appear
  await page.mouse.move(x, y, { steps: 5 });
  await page.waitForTimeout(500);
}

/**
 * Gets the displayed timestamp label from the chart tooltip
 */
export async function getDisplayedTimestampLabel(
  page: Page,
): Promise<string | null> {
  const timestampLabel = page
    .locator("svg text.fill-gray-700.stroke-surface")
    .first();
  const isVisible = await timestampLabel.isVisible().catch(() => false);
  if (!isVisible) return null;
  return timestampLabel.textContent();
}

/**
 * Gets the displayed value from the chart tooltip
 */
export async function getDisplayedValue(page: Page): Promise<string | null> {
  const valueTspan = page.locator("svg tspan.widths").first();
  const isVisible = await valueTspan.isVisible().catch(() => false);
  if (!isVisible) return null;
  return valueTspan.textContent();
}

/**
 * Parses a formatted number string back to a number
 * Handles formats like "19.3k", "1.2M", "100", "1,234"
 */
export function parseFormattedNumber(formatted: string): number | null {
  if (!formatted) return null;

  const trimmed = formatted.trim();
  if (!trimmed) return null;

  const suffixMultipliers: Record<string, number> = {
    k: 1_000,
    K: 1_000,
    M: 1_000_000,
    B: 1_000_000_000,
    T: 1_000_000_000_000,
  };

  const lastChar = trimmed.slice(-1);
  if (suffixMultipliers[lastChar]) {
    const numPart = parseFloat(trimmed.slice(0, -1).replace(/,/g, ""));
    return numPart * suffixMultipliers[lastChar];
  }

  return parseFloat(trimmed.replace(/,/g, ""));
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

/**
 * Asserts that hovering on the chart displays a tooltip with timestamp and value
 */
export async function assertTimeseriesHoverTooltip(
  page: Page,
  xPercent: number = 0.5,
) {
  await hoverOnChart(page, xPercent);

  const timestamp = await getDisplayedTimestampLabel(page);
  const value = await getDisplayedValue(page);

  expect(timestamp || value).toBeTruthy();

  if (timestamp) {
    expect(timestamp.length).toBeGreaterThan(3);
  }

  if (value) {
    const parsed = parseFormattedNumber(value);
    expect(parsed).not.toBeNull();
  }
}

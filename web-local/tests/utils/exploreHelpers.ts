import type { Page } from "playwright";
import {
  clickMenuButton,
  openFileNavEntryContextMenu,
} from "web-local/tests/utils/commonHelpers";
import { waitForFileNavEntry } from "web-local/tests/utils/waitHelpers";

export async function createExploreFromSource(
  page: Page,
  sourcePath = "/sources/AdBids.yaml",
  metricsViewPath = "/metrics/AdBids_metrics.yaml",
) {
  await openFileNavEntryContextMenu(page, sourcePath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await page.getByText("Create explore").click();
}

export async function createExploreFromModel(
  page: Page,
  modelPath = "/models/AdBids_model.sql",
  metricsViewPath = "/metrics/AdBids_model_metrics.yaml",
  explorePath = "/explore-dashboards/AdBids_model_metrics_explore.yaml",
) {
  await openFileNavEntryContextMenu(page, modelPath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await Promise.all([
    waitForFileNavEntry(page, explorePath, true),
    page.getByText("Create explore").click(),
  ]);
}

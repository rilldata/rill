import type { Page } from "playwright";
import {
  clickMenuButton,
  openFileNavEntryContextMenu,
} from "web-local/tests/utils/commonHelpers";
import { waitForFileNavEntry } from "web-local/tests/utils/waitHelpers";

export async function createExploreFromSource(
  page: Page,
  sourcePath = "/models/AdBids.yaml",
  metricsViewPath = "/metrics/AdBids_metrics.yaml",
) {
  await openFileNavEntryContextMenu(page, sourcePath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await page.getByText("Generate Explore Dashboard").click();
}

export async function createExploreFromModel(
  page: Page,
  navigateToPreview = false,
  modelPath = "/models/AdBids_model.sql",
  metricsViewPath = "/metrics/AdBids_model_metrics.yaml",
) {
  await openFileNavEntryContextMenu(page, modelPath);
  await clickMenuButton(page, "Generate metrics");
  await waitForFileNavEntry(page, metricsViewPath, true);
  await page.getByText("Generate Explore Dashboard").click();

  if (navigateToPreview) {
    await page.getByRole("button", { name: "Preview" }).click();
  }

  await page.waitForTimeout(1000);
}

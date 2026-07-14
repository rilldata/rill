import { expect } from "@playwright/test";
import type { Page } from "playwright";
import {
  clickMenuButton,
  openFileNavEntryContextMenu,
} from "web-local/tests/utils/commonHelpers";
import { waitForFileNavEntry } from "web-local/tests/utils/waitHelpers";

/**
 * Waits for the Preview button to be enabled and clicks it.
 * Reconciliation can take a while in CI, so we use a generous timeout.
 */
export async function clickPreviewButton(page: Page, timeout = 10_000) {
  const previewButton = page.getByRole("button", { name: "Preview" });
  await previewButton.waitFor({ state: "visible" });
  await expect(previewButton).toBeEnabled({ timeout });
  await previewButton.click();
}

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
    await clickPreviewButton(page);
  }

  await page.waitForTimeout(1000);
}

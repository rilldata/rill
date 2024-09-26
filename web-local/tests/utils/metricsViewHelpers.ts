import { expect } from "@playwright/test";
import type { Page } from "playwright";
import { clickMenuButton, openFileNavEntryContextMenu } from "./commonHelpers";

export async function createMetricsViewFromSource(
  page: Page,
  sourcePath = "/sources/AdBids.yaml",
) {
  await openFileNavEntryContextMenu(page, sourcePath);
  await clickMenuButton(page, "Generate metrics");
}

export async function createMetricsViewFromModel(
  page: Page,
  modelPath = "/models/AdBids_model.sql",
) {
  await openFileNavEntryContextMenu(page, modelPath);
  await clickMenuButton(page, "Generate metrics");
}

export async function assertLeaderboards(
  page: Page,
  leaderboards: Array<{
    label: string;
    values: Array<string>;
  }>,
) {
  for (const { label, values } of leaderboards) {
    const leaderboardBlock = page.getByRole("table", {
      name: `${label} leaderboard`,
    });
    await expect(leaderboardBlock).toBeVisible();

    const actualValues = await leaderboardBlock
      .locator("tr > td:nth-child(2)")
      .allInnerTexts();
    expect(actualValues).toEqual(values);
  }
}

// Helper that opens the time range menu, calls your interactions, and then waits until the menu closes
export async function interactWithTimeRangeMenu(
  page: Page,
  cb: () => void | Promise<void>,
) {
  // Open the menu
  await page.getByLabel("Select time range").click();
  // Run the defined interactions
  await cb();
  // Wait for menu to close
  await expect(
    page.getByRole("menu", { name: "Select time range" }),
  ).not.toBeVisible();
}

import { expect } from "@playwright/test";
import type { Page } from "playwright";

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

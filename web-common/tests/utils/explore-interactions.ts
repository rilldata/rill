import { expect, type Locator, type Page } from "@playwright/test";

// Helper that opens the time range menu, calls your interactions, and then waits until the menu closes
export async function interactWithTimeRangeMenu(
  page: Page | Locator,
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

export async function setDashboardTimezone(page: Page, timezone: string) {
  const currentUrl = new URL(page.url());
  currentUrl.searchParams.set("tz", timezone);
  await page.goto(currentUrl.toString());
}

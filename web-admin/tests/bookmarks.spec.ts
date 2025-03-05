import { expect, type Page } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Bookmarks", () => {
  test("Create and check home bookmark", async ({ page }) => {
    await page.goto("/e2e/openrtb/explore/auction_explore");

    // Select "Last 14 days" as time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 14 Days" }).click();
    });

    // Filter to "Not Available" "App Site Domain" via leaderboard
    await page.getByRole("row", { name: "Not Available 158.5M" }).click();
    // Filter to "Not Available" "Pub Name" via leaderboard
    await page.getByRole("row", { name: "Not Available 98.7M" }).click();

    // Open the bookmarks dropdown
    await page.getByLabel("Bookmark dropdown").click();
    // Create the current filter as home filter
    await page
      .getByRole("menuitem", { name: "Bookmark current view as Home." })
      .click();
    // Wait for the notification that home bookmark was created
    await expect(page.getByText("Home bookmark created")).toBeVisible();

    await page.goto("/e2e/openrtb");
    // Navigate to the explore
    await page
      .getByRole("link", { name: "Programmatic Ads Auction" })
      .first()
      .click();

    // saved home bookmark is restored
    await expect(page.getByText("Last 14 Days")).toBeVisible();
    await expect(page.getByText("App Site Domain Not Available")).toBeVisible();
    await expect(page.getByText("Pub Name Not Available")).toBeVisible();
  });
});

// Helper that opens the time range menu, calls your interactions, and then waits until the menu closes
// TODO: move to common place
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

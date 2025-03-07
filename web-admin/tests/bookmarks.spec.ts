import { expect, type Page } from "@playwright/test";
import { assertUrlParams } from "@rilldata/web-common/tests/utils/assertUrlParams";
import { test } from "./setup/base";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/exploreInteractions";

test.describe("Bookmarks", () => {
  test("Create and verify home bookmark", async ({ page }) => {
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
    // Create the current filter as home bookmark
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
    // make sure the url has the correct params
    assertUrlParams(
      page,
      `tr=P14D&f=app_site_domain IN ('Not Available') AND pub_name IN ('Not Available')`,
    );

    // Open the bookmarks dropdown
    await page.getByLabel("Bookmark dropdown").click();
    // Verify that a home bookmark was created
    await expect(page.getByLabel("Home Bookmark Entry")).toHaveText(
      "Home Main view for this dashboard",
    );
    // Verify that the bookmark has the correct icon
    await expect(
      page.getByLabel("Home Bookmark Entry").getByLabel("Home Bookmark Icon"),
    ).toBeVisible();
  });

  test("Create and use filter-only bookmarks", async ({ page }) => {
    await page.goto("/e2e/openrtb/explore/auction_explore");

    // Select "Last 4 weeks" as time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 4 weeks" }).click();
    });

    // Filter to "FuboTV" and "Sling" "App Site Name" via leaderboard
    await page.getByRole("row", { name: "FuboTV 8.0M" }).click();
    await page.getByRole("row", { name: "Sling 5.1M" }).click();

    // Enter dimension table "App Site Name"
    await page.getByText("App Site Domain").click();
    // Enable time comparison
    await page.getByLabel("Toggle time comparison").click();

    // Open the bookmarks dropdown
    await page.getByLabel("Bookmark dropdown").click();
    // Create a new bookmark
    await page
      .getByRole("menuitem", { name: "Bookmark current view", exact: true })
      .click();

    // Assert the selected filters
    await expect(page.getByLabel("Readonly Filter Chips")).toHaveText(
      ` Last 4 Weeks    App Site Name FuboTV  +1 other   `,
    );
    // Create a personal bookmark
    await enterBookmarkDetails(
      page,
      "Filter-Only",
      "My filter-only bookmark.",
      false,
      true,
    );
    // Save bookmark
    await page.getByRole("button", { name: "Save" }).click();
    // Wait for the notification that bookmark was created
    await expect(page.getByText("Bookmark created")).toBeVisible();

    await page.goto("/e2e/openrtb/explore/auction_explore");

    // Open the bookmarks dropdown
    await page.getByLabel("Bookmark dropdown").click();
    const filterOnlyBookmarkLocator = page.getByLabel(
      "Filter-Only Bookmark Entry",
    );
    // Verify that a home bookmark was created
    await expect(filterOnlyBookmarkLocator).toHaveText(
      "Filter-Only My filter-only bookmark.",
    );
    // Verify that the bookmark has the correct icon
    await expect(
      filterOnlyBookmarkLocator.getByLabel("Filter Icon"),
    ).toBeVisible();
    await filterOnlyBookmarkLocator.click();

    // saved home bookmark is restored
    await expect(page.getByText("Last 4 Weeks")).toBeVisible();
    await expect(page.getByText("App Site Name FuboTV +1 other")).toBeVisible();
    // make sure the url has the correct params
    // NOTE: comparison time range is not added for filter-only as per requirement
    assertUrlParams(
      page,
      `tr=P4W&tz=UTC&f=app_site_name IN ('FuboTV','Sling')`,
    );
  });
});

async function enterBookmarkDetails(
  page: Page,
  label: string,
  description: string,
  isManaged: boolean,
  isFiltersOnly: boolean,
) {
  await page.getByTitle("Label").fill(label);
  await page.getByTitle("Description").fill(description);

  await page.getByLabel("Category").click();
  await page
    .getByRole("option", {
      name: isManaged ? "Managed bookmarks" : "Your bookmarks",
    })
    .click();

  if (isFiltersOnly) {
    await page.getByLabel("Filters only").click();
  }
}

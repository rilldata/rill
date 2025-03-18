import { expect, type Page } from "@playwright/test";
import { assertUrlParams } from "@rilldata/web-common/tests/utils/assert-url-params";
import { test } from "./setup/base";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";

test.describe("Bookmarks", () => {
  // TODO: use a separate explore to isolate bookmarks to avoid conflicts.
  //       ideally we would have an explore per feature with a suffix like "auction_explore_bookmarks"

  test.describe.serial("Home bookmarks", () => {
    test("Should create a home bookmark", async ({ page }) => {
      // This would ideally be done in a beforeAll hook. But page fixture is not supported in that hook.
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 24 Hours" as time range
      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: "Last 24 Hours" }).click();
      });

      // Filter to "Not Available" "App Site Domain" via leaderboard
      await page.getByRole("row", { name: "Not Available 197.0k" }).click();
      // Filter to "Not Available" "Pub Name" via leaderboard
      await page.getByRole("row", { name: "Not Available 128.4k" }).click();

      // Open the bookmarks dropdown
      await page.getByLabel("Bookmark dropdown").click();
      // Create the current filter as home bookmark
      await page
        .getByRole("menuitem", { name: "Bookmark current view as Home." })
        .click();
      // Wait for the notification that home bookmark was created
      await expect(page.getByText("Home bookmark created")).toBeVisible();
    });

    test("Visiting home should restore home bookmark", async ({ page }) => {
      await page.goto("/e2e/openrtb");
      // Navigate to the explore
      await page
        .getByRole("link", { name: "Programmatic Ads Auction For Bookmarks" })
        .first()
        .click();

      // saved home bookmark is restored
      await expect(page.getByText("Last 24 Hours")).toBeVisible();
      await expect(
        page.getByText("App Site Domain Not Available"),
      ).toBeVisible();
      await expect(page.getByText("Pub Name Not Available")).toBeVisible();
      // make sure the url has the correct params
      assertUrlParams(
        page,
        `tr=PT24H&grain=hour&f=app_site_domain IN ('Not Available') AND pub_name IN ('Not Available')`,
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

    test("Visiting dashboard with params should not apply home bookmark", async ({
      page,
    }) => {
      // Add random params. Home bookmark shouldnt apply
      await page.goto(
        "/e2e/openrtb/explore/auction_explore_bookmarks?compare_tr=rill-PW&expand_dim=app_site_name",
      );
      // Default time range is present
      await expect(page.getByText("Last 7 Days")).toBeVisible();
      // Comparison is previous week
      await expect(page.getByText("Previous week")).toBeVisible();
      // No filter applied
      await expect(page.getByText("No filters selected")).toBeVisible();
      // In "App Site Name" dimension table
      await expect(page.getByLabel("Dimension Display")).toBeVisible();
      await expect(page.getByText("App Site Name")).toBeVisible();

      // Remove all filters. Home Bookmarks shouldn't apply.
      // This is as per requirement where clearing filters should take the user to the default set of values.
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");
      await expect(page.getByText("Last 7 Days")).toBeVisible();
      await expect(page.getByText("No filters selected")).toBeVisible();
      await expect(page.getByText("no comparison period")).toBeVisible();
      await expect(
        page.getByLabel("Leaderboards", { exact: true }),
      ).toBeVisible();
    });

    // TODO: verify editing home bookmark. since these are changing in a future feature, these tests should be part of that PR
  });

  test.describe.serial("Filter-only bookmark", () => {
    test("Create filter-only bookmark", async ({ page }) => {
      // This would ideally be done in a beforeAll hook. But page fixture is not supported in that hook.
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 6 Hours" as time range
      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
      });

      // Filter to "FuboTV" and "Philo" "App Site Name" via leaderboard
      await page.getByRole("row", { name: "FuboTV 2.6k" }).click();
      await page.getByRole("row", { name: "Philo 2.2k" }).click();

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
        ` Last 6 Hours    App Site Name FuboTV  +1 other   `,
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
    });

    test("Applying filter-only bookmark should not change other settings", async ({
      page,
    }) => {
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Visit TDD
      await page.getByRole("button", { name: "Requests 6.60M" }).click();

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

      // saved bookmark is restored
      await expect(page.getByText("Last 6 Hours")).toBeVisible();
      await expect(
        page.getByText("App Site Name FuboTV +1 other"),
      ).toBeVisible();
      await expect(
        page.getByLabel("requests Time Dimension Display"),
      ).toBeVisible();
      // non-filter state is retained
      // make sure the url has the correct params
      // NOTE: comparison time range is not added for filter-only as per requirement
      assertUrlParams(
        page,
        `view=tdd&tr=PT6H&grain=hour&f=app_site_name IN ('FuboTV','Philo')&measure=requests`,
      );
    });
  });

  test.describe.serial("Complete bookmarks", () => {
    test("Create a complete bookmark.", async ({ page }) => {
      // This would ideally be done in a beforeAll hook. But page fixture is not supported in that hook.
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 6 Hours" as time range
      await interactWithTimeRangeMenu(page, async () => {
        await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
      });

      // Filter to "FuboTV" and "Philo" "App Site Name" via leaderboard
      await page.getByRole("row", { name: "FuboTV 2.6k" }).click();
      await page.getByRole("row", { name: "Philo 2.2k" }).click();

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
        ` Last 6 Hours    App Site Name FuboTV  +1 other   `,
      );
      // Create a personal bookmark
      await enterBookmarkDetails(
        page,
        "Complete",
        "My complete bookmark.",
        false,
        false,
      );
      // Save bookmark
      await page.getByRole("button", { name: "Save" }).click();
      // Wait for the notification that bookmark was created
      await expect(page.getByText("Bookmark created")).toBeVisible();
    });

    test("Applying complete bookmark replaces every setting", async ({
      page,
    }) => {
      await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Visit TDD
      await page.getByRole("button", { name: "Requests 6.60M" }).click();

      // Open the bookmarks dropdown
      await page.getByLabel("Bookmark dropdown").click();
      const filterOnlyBookmarkLocator = page.getByLabel(
        "Complete Bookmark Entry",
      );
      // Verify that a home bookmark was created
      await expect(filterOnlyBookmarkLocator).toHaveText(
        "Complete My complete bookmark.",
      );
      // Verify that the bookmark has the correct icon
      await expect(
        filterOnlyBookmarkLocator.getByLabel("Bookmark Icon"),
      ).toBeVisible();
      await filterOnlyBookmarkLocator.click();

      // saved bookmark is restored
      await expect(page.getByText("Last 6 Hours")).toBeVisible();
      await expect(page.getByText("Previous period")).toBeVisible();
      await expect(
        page.getByText("App Site Name FuboTV +1 other"),
      ).toBeVisible();
      // In "App Site Name" dimension table
      await expect(page.getByLabel("Dimension Display")).toBeVisible();
      await expect(page.getByText("App Site Name")).toBeVisible();
      // Previous view TDD is not present
      await expect(
        page.getByLabel("requests Time Dimension Display"),
      ).not.toBeVisible();
      // make sure the url has the correct params
      assertUrlParams(
        page,
        `tr=PT6H&compare_tr=rill-PP&grain=hour&f=app_site_name IN ('FuboTV','Philo')&expand_dim=app_site_domain`,
      );
    });
  });

  // More exhaustive bookmark tests should either be component tests or unit tests on `getBookmarkDataForDashboard`
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

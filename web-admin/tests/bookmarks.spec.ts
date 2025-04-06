import { expect, type Page } from "@playwright/test";
import { assertUrlParams } from "@rilldata/web-common/tests/utils/assert-url-params";
import { test } from "./setup/base";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";

test.describe("Bookmarks", () => {
  // TODO: use a separate explore to isolate bookmarks to avoid conflicts.
  //       ideally we would have an explore per feature with a suffix like "auction_explore_bookmarks"

  test.describe.serial("Filter-only bookmark", () => {
    test("Create filter-only bookmark", async ({ adminPage }) => {
      // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
      await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 6 Hours" as time range
      await interactWithTimeRangeMenu(adminPage, async () => {
        await adminPage.getByRole("menuitem", { name: "Last 6 Hours" }).click();
      });

      // Filter to "FuboTV" and "Philo" "App Site Name" via leaderboard
      await adminPage.getByRole("row", { name: "FuboTV 2.6k" }).click();
      await adminPage.getByRole("row", { name: "Philo 2.2k" }).click();

      // Enter dimension table "App Site Name"
      await adminPage.getByText("App Site Domain").click();
      // Enable time comparison
      await adminPage.getByLabel("Toggle time comparison").click();

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      // Create a new bookmark
      await adminPage
        .getByRole("menuitem", { name: "Bookmark current view", exact: true })
        .click();

      // Assert the selected filters
      await expect(adminPage.getByLabel("Readonly Filter Chips")).toHaveText(
        ` Last 6 Hours    App Site Name FuboTV  +1 other   `,
      );
      // Create a personal bookmark
      await enterBookmarkDetails(
        adminPage,
        "Filter-Only",
        "My filter-only bookmark.",
        false,
        true,
      );
      // Save bookmark
      await adminPage.getByRole("button", { name: "Save" }).click();
      // Wait for the notification that bookmark was created
      await expect(adminPage.getByText("Bookmark created")).toBeVisible();
    });

    test("Applying filter-only bookmark should not change other settings", async ({
      adminPage,
    }) => {
      await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Visit TDD
      await adminPage.getByRole("button", { name: "Requests 6.60M" }).click();

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      const filterOnlyBookmarkLocator = adminPage.getByLabel(
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
      await expect(adminPage.getByText("Last 6 Hours")).toBeVisible();
      await expect(
        adminPage.getByText("App Site Name FuboTV +1 other"),
      ).toBeVisible();
      await expect(
        adminPage.getByLabel("requests Time Dimension Display"),
      ).toBeVisible();
      // non-filter state is retained
      // make sure the url has the correct params
      // NOTE: comparison time range is not added for filter-only as per requirement
      assertUrlParams(
        adminPage,
        `view=tdd&tr=PT6H&grain=hour&f=app_site_name IN ('FuboTV','Philo')&measure=requests`,
      );
    });
  });

  test.describe.serial("Complete bookmarks", () => {
    test("Create a complete bookmark.", async ({ adminPage }) => {
      // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
      await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 6 Hours" as time range
      await interactWithTimeRangeMenu(adminPage, async () => {
        await adminPage.getByRole("menuitem", { name: "Last 6 Hours" }).click();
      });

      // Filter to "FuboTV" and "Philo" "App Site Name" via leaderboard
      await adminPage.getByRole("row", { name: "FuboTV 2.6k" }).click();
      await adminPage.getByRole("row", { name: "Philo 2.2k" }).click();

      // Enter dimension table "App Site Name"
      await adminPage.getByText("App Site Domain").click();
      // Enable time comparison
      await adminPage.getByLabel("Toggle time comparison").click();

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      // Create a new bookmark
      await adminPage
        .getByRole("menuitem", { name: "Bookmark current view", exact: true })
        .click();

      // Assert the selected filters
      await expect(adminPage.getByLabel("Readonly Filter Chips")).toHaveText(
        ` Last 6 Hours    App Site Name FuboTV  +1 other   `,
      );
      // Create a personal bookmark
      await enterBookmarkDetails(
        adminPage,
        "Complete",
        "My complete bookmark.",
        false,
        false,
      );
      // Save bookmark
      await adminPage.getByRole("button", { name: "Save" }).click();
      // Wait for the notification that bookmark was created
      await expect(adminPage.getByText("Bookmark created")).toBeVisible();
    });

    test("Applying complete bookmark replaces every setting", async ({
      adminPage,
    }) => {
      await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Visit TDD
      await adminPage.getByRole("button", { name: "Requests 6.60M" }).click();

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      const filterOnlyBookmarkLocator = adminPage.getByLabel(
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
      await expect(adminPage.getByText("Last 6 Hours")).toBeVisible();
      await expect(adminPage.getByText("Previous period")).toBeVisible();
      await expect(
        adminPage.getByText("App Site Name FuboTV +1 other"),
      ).toBeVisible();
      // In "App Site Name" dimension table
      await expect(adminPage.getByLabel("Dimension Display")).toBeVisible();
      await expect(adminPage.getByText("App Site Name")).toBeVisible();
      // Previous view TDD is not present
      await expect(
        adminPage.getByLabel("requests Time Dimension Display"),
      ).not.toBeVisible();
      // make sure the url has the correct params
      assertUrlParams(
        adminPage,
        `tr=PT6H&compare_tr=rill-PP&grain=hour&f=app_site_name IN ('FuboTV','Philo')&expand_dim=app_site_domain`,
      );
    });
  });

  // Home bookmark interferes with other bookmark creation since we are adding some filters.
  // So adding it to the end.
  test.describe.serial("Home bookmarks", () => {
    test("Should create a home bookmark", async ({ adminPage }) => {
      // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
      await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

      // Select "Last 24 Hours" as time range
      await interactWithTimeRangeMenu(adminPage, async () => {
        await adminPage
          .getByRole("menuitem", { name: "Last 24 Hours" })
          .click();
      });

      // Filter to "Not Available" "App Site Domain" via leaderboard
      await adminPage
        .getByRole("row", { name: "Not Available 197.0k" })
        .click();
      // Filter to "Not Available" "Pub Name" via leaderboard
      await adminPage
        .getByRole("row", { name: "Not Available 128.4k" })
        .click();

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      // Create the current filter as home bookmark
      await adminPage
        .getByRole("menuitem", { name: "Bookmark current view as Home." })
        .click();
      // Wait for the notification that home bookmark was created
      await expect(adminPage.getByText("Home bookmark created")).toBeVisible();
    });

    test("Visiting home should restore home bookmark", async ({
      adminPage,
    }) => {
      await adminPage.goto("/e2e/openrtb");
      // Navigate to the explore
      await adminPage
        .getByRole("link", { name: "Programmatic Ads Auction For Bookmarks" })
        .first()
        .click();

      // saved home bookmark is restored
      await expect(adminPage.getByText("Last 24 Hours")).toBeVisible();
      await expect(
        adminPage.getByText("App Site Domain Not Available"),
      ).toBeVisible();
      await expect(adminPage.getByText("Pub Name Not Available")).toBeVisible();
      // make sure the url has the correct params
      assertUrlParams(
        adminPage,
        `tr=PT24H&grain=hour&f=app_site_domain IN ('Not Available') AND pub_name IN ('Not Available')`,
      );

      // Open the bookmarks dropdown
      await adminPage.getByLabel("Bookmark dropdown").click();
      // Verify that a home bookmark was created
      await expect(adminPage.getByLabel("Home Bookmark Entry")).toHaveText(
        "Home Main view for this dashboard",
      );
      // Verify that the bookmark has the correct icon
      await expect(
        adminPage
          .getByLabel("Home Bookmark Entry")
          .getByLabel("Home Bookmark Icon"),
      ).toBeVisible();
    });

    test("Visiting dashboard with params should not apply home bookmark", async ({
      adminPage,
    }) => {
      // Add random params. Home bookmark shouldnt apply
      await adminPage.goto(
        "/e2e/openrtb/explore/auction_explore_bookmarks?compare_tr=rill-PW&expand_dim=app_site_name",
      );
      // Default time range is present
      await expect(adminPage.getByText("Last 7 Days")).toBeVisible();
      // Comparison is previous week
      await expect(adminPage.getByText("Previous week")).toBeVisible();
      // No filter applied
      await expect(adminPage.getByText("No filters selected")).toBeVisible();
      // In "App Site Name" dimension table
      await expect(adminPage.getByLabel("Dimension Display")).toBeVisible();
      await expect(adminPage.getByText("App Site Name")).toBeVisible();
    });

    // TODO: verify editing home bookmark. since these are changing in a future feature, these tests should be part of that PR
  });
  // More exhaustive bookmark tests should either be component tests or unit tests on `getBookmarkDataForDashboard`
});

async function enterBookmarkDetails(
  adminPage: Page,
  label: string,
  description: string,
  isManaged: boolean,
  isFiltersOnly: boolean,
) {
  await adminPage.getByTitle("Label").fill(label);
  await adminPage.getByTitle("Description").fill(description);

  await adminPage.getByLabel("Category").click();
  await adminPage
    .getByRole("option", {
      name: isManaged ? "Managed bookmarks" : "Your bookmarks",
    })
    .click();

  if (isFiltersOnly) {
    await adminPage.getByLabel("Filters only").click();
  }
}

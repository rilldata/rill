import { expect, type Page } from "@playwright/test";
import { assertUrlParams } from "@rilldata/web-common/tests/utils/assert-url-params";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { test } from "./setup/base";

test.describe("Bookmarks", () => {
  // TODO: use a separate explore to isolate bookmarks to avoid conflicts.
  //       ideally we would have an explore per feature with a suffix like "auction_explore_bookmarks"

  test.describe.serial("Explore bookmarks", () => {
    test.describe.serial("Filter-only explore bookmark", () => {
      test("Create filter-only bookmark", async ({ adminPage }) => {
        // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

        // Select "Last 6 Hours" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 6 Hours" })
            .click();
        });

        // Filter to "FuboTV" and "My Little Universe" "App Site Name" via leaderboard
        await adminPage.getByRole("row", { name: "FuboTV 2.6k" }).click();
        await adminPage
          .getByRole("row", { name: "My Little Universe 4.6k" })
          .click();

        // Enter dimension table "App Site Name"
        await adminPage.getByText("App Site Domain").click();
        // Enable time comparison
        await adminPage.getByLabel("Toggle time comparison").click();

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Other bookmark dropdown").click();
        // Create a new bookmark
        await adminPage
          .getByRole("menuitem", { name: "Bookmark current view", exact: true })
          .click();

        // Assert the selected filters
        await expect(adminPage.getByLabel("Readonly Filter Chips")).toHaveText(
          ` Last 6 hours    App Site Name FuboTV  +1 other   `,
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
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const filterOnlyBookmarkLocator = adminPage.getByLabel(
          "Filter-Only Bookmark Entry",
        );
        // Verify that a home bookmark was created
        await expect(filterOnlyBookmarkLocator).toHaveText(
          "Filter-Only My filter-only bookmark.",
        );
        // Verify that the bookmark has the correct icon
        await expect(
          filterOnlyBookmarkLocator.getByLabel("Filter outline icon"),
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
          `view=tdd&tr=6h+as+of+latest%2Fh%2B1h&f=app_site_name IN ('FuboTV','My+Little+Universe')&measure=requests&chart_type=line`,
        );

        // Open bookmark dropdown and verify the "filled" state for the bookmark
        await adminPage.getByLabel("Other bookmark dropdown").click();
        await expect(
          adminPage
            .getByLabel("Filter-Only Bookmark Entry")
            .getByLabel("Filter filled icon"),
        ).toBeVisible();
      });

      test("Should delete filter-only bookmark", async ({ adminPage }) => {
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const menuItem = adminPage.getByLabel("Filter-Only Bookmark Entry");
        await menuItem.hover();
        await menuItem.getByRole("button", { name: "Delete bookmark" }).click();
        await expect(
          adminPage.getByText("Bookmark Filter-Only deleted"),
        ).toBeVisible();
      });
    });

    test.describe.serial("Complete explore bookmarks", () => {
      test("Create a complete bookmark.", async ({ adminPage }) => {
        // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

        // Select "Last 6 Hours" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 6 Hours" })
            .click();
        });

        // Filter to "FuboTV" and "My Little Universe" "App Site Name" via leaderboard
        await adminPage.getByRole("row", { name: "FuboTV 2.6k" }).click();
        await adminPage
          .getByRole("row", { name: "My Little Universe 4.6k" })
          .click();

        // Enter dimension table "App Site Name"
        await adminPage.getByText("App Site Domain").click();
        // Enable time comparison
        await adminPage.getByLabel("Toggle time comparison").click();

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Other bookmark dropdown").click();
        // Create a new bookmark
        await adminPage
          .getByRole("menuitem", { name: "Bookmark current view", exact: true })
          .click();

        // Assert the selected filters
        await expect(adminPage.getByLabel("Readonly Filter Chips")).toHaveText(
          ` Last 6 hours    App Site Name FuboTV  +1 other   `,
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
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const filterOnlyBookmarkLocator = adminPage.getByLabel(
          "Complete Bookmark Entry",
        );
        // Verify that a home bookmark was created
        await expect(filterOnlyBookmarkLocator).toHaveText(
          "Complete My complete bookmark.",
        );
        // Verify that the bookmark has the correct icon
        await expect(
          filterOnlyBookmarkLocator.getByLabel("Bookmark outline icon"),
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
          `tr=6h+as+of+latest%2Fh%2B1h&compare_tr=rill-PP&f=app_site_name IN ('FuboTV','My+Little+Universe')&expand_dim=app_site_domain`,
        );

        // Open bookmark dropdown and verify the "filled" state for the bookmark
        await adminPage.getByLabel("Other bookmark dropdown").click();
        await expect(
          adminPage
            .getByLabel("Complete Bookmark Entry")
            .getByLabel("Bookmark filled icon"),
        ).toBeVisible();
      });

      test("Should delete complete bookmark", async ({ adminPage }) => {
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const menuItem = adminPage.getByLabel("Complete Bookmark Entry");
        await menuItem.hover();
        await menuItem.getByRole("button", { name: "Delete bookmark" }).click();
        await expect(
          adminPage.getByText("Bookmark Complete deleted"),
        ).toBeVisible();
      });
    });

    // Home bookmark interferes with other bookmark creation since we are adding some filters.
    // So adding it to the end.
    test.describe.serial("Home explore bookmarks", () => {
      test("Should create a home bookmark", async ({ adminPage }) => {
        // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

        // Select "Last 24 Hours" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 7 days" })
            .click();
        });

        // Filter to "Not Available" "App Site Domain" via leaderboard
        await adminPage
          .getByRole("row", { name: "Not Available 1.7M" })
          .click();
        // Filter to "Not Available" "Pub Name" via leaderboard
        await adminPage
          .getByRole("row", { name: "Not Available 1.0M" })
          .click();

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Home bookmark dropdown").click();
        // Create the current filter as home bookmark
        await adminPage
          .getByRole("menuitem", { name: "Bookmark current view as Home." })
          .click();
        // Wait for the notification that home bookmark was created
        await expect(
          adminPage.getByText("Home bookmark created"),
        ).toBeVisible();
      });

      test("Visiting home should restore home bookmark", async ({
        adminPage,
      }) => {
        // Navigate to the explore
        await adminPage.goto("/e2e/openrtb/-/dashboards");
        await adminPage
          .getByRole("link", { name: "Programmatic Ads Auction For Bookmarks" })
          .first()
          .click();

        // saved home bookmark is restored
        await expect(adminPage.getByText("Last 7 days")).toBeVisible();
        await expect(
          adminPage.getByText("App Site Domain Not Available"),
        ).toBeVisible();
        await expect(
          adminPage.getByText("Pub Name Not Available"),
        ).toBeVisible();
        // make sure the url has the correct params
        assertUrlParams(
          adminPage,
          `tr=7D+as+of+latest%2FD%2B1D&grain=day&f=app_site_domain IN ('Not Available') AND pub_name IN ('Not Available')`,
        );
      });

      test("Visiting dashboard with params should not apply home bookmark", async ({
        adminPage,
      }) => {
        // Add random params. Home bookmark shouldnt apply
        await adminPage.goto(
          "/e2e/openrtb/explore/auction_explore_bookmarks?compare_tr=rill-PW&expand_dim=app_site_name",
        );
        // Default time range is present
        await expect(adminPage.getByText("Last 24 Hours")).toBeVisible();
        // Comparison is previous week
        await expect(adminPage.getByText("Previous week")).toBeVisible();
        // No filter applied
        await expect(adminPage.getByText("No filters selected")).toBeVisible();
        // In "App Site Name" dimension table
        await expect(adminPage.getByLabel("Dimension Display")).toBeVisible();
        await expect(adminPage.getByText("App Site Name")).toBeVisible();

        // Click on "Go to home bookmark"
        await adminPage.getByLabel("Home bookmark dropdown").click();
        await adminPage.getByLabel("Home Bookmark Entry").click();
        // saved home bookmark is restored
        await expect(adminPage.getByText("Last 7 days")).toBeVisible();
        await expect(
          adminPage.getByText("App Site Domain Not Available"),
        ).toBeVisible();
        await expect(
          adminPage.getByText("Pub Name Not Available"),
        ).toBeVisible();
        // make sure the url has the correct params
        assertUrlParams(
          adminPage,
          `tr=7D+as+of+latest%2FD%2B1D&grain=day&f=app_site_domain IN ('Not Available') AND pub_name IN ('Not Available')`,
        );
      });

      test("Should delete home bookmark", async ({ adminPage }) => {
        await adminPage.goto("/e2e/openrtb/explore/auction_explore_bookmarks");
        await adminPage.getByLabel("Home bookmark dropdown").click();
        const menuItem = adminPage.getByLabel("Home Bookmark Entry");
        await menuItem.hover();
        await menuItem.getByRole("button", { name: "Delete bookmark" }).click();
        await expect(
          adminPage.getByText("Bookmark Go to home deleted"),
        ).toBeVisible();
      });

      // TODO: verify editing home bookmark. since these are changing in a future feature, these tests should be part of that PR
    });
    // More exhaustive bookmark tests should either be component tests or unit tests on `getBookmarkDataForDashboard`
  });

  test.describe.serial("Canvas bookmarks", () => {
    test.describe.serial("Complete canvas bookmarks", () => {
      test("Create a complete bookmark.", async ({ adminPage }) => {
        // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
        await adminPage.goto("/e2e/openrtb/canvas/bids_canvas_bookmarks");

        // Select "Last 6 Hours" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 6 Hours" })
            .click();
        });

        // Filter to "Instacart" and "Leafly" "Advertiser Name" via leaderboard
        await adminPage
          .getByRole("row", { name: "Instacart $252.33" })
          .scrollIntoViewIfNeeded();
        await adminPage.getByRole("row", { name: "Instacart $252.33" }).click();
        await adminPage.getByRole("row", { name: "Leafly $195.89" }).click();

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Other bookmark dropdown").click();
        // Create a new bookmark
        await adminPage
          .getByRole("menuitem", { name: "Bookmark current view", exact: true })
          .click();

        // Assert the selected filters
        await expect(adminPage.getByLabel("Readonly Filter Chips")).toHaveText(
          ` Last 6 hours    Advertiser Name Instacart  +1 other   `,
        );
        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$448.22\s*\+\$417.48 \+1k%\s*vs previous day/,
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
        await adminPage.goto("/e2e/openrtb/canvas/bids_canvas_bookmarks");

        // Select "Last 12 Hours" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 12 Hours" })
            .click();
        });

        // Filter to "Site" "App or Site" via leaderboard
        await adminPage
          .getByRole("row", { name: "Site $1.6k" })
          .scrollIntoViewIfNeeded();
        await adminPage.getByRole("row", { name: "Site $1.6k" }).click();

        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$1,632\s*\+\$770.59 \+89%\s*vs previous day/,
        );

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const filterOnlyBookmarkLocator = adminPage.getByLabel(
          "Complete Bookmark Entry",
        );
        // Verify that a bookmark was created
        await expect(filterOnlyBookmarkLocator).toHaveText(
          "Complete My complete bookmark.",
        );
        // Verify that the bookmark has the correct icon
        await expect(
          filterOnlyBookmarkLocator.getByLabel("Bookmark outline icon"),
        ).toBeVisible();
        await filterOnlyBookmarkLocator.click();

        // saved bookmark is restored
        await expect(adminPage.getByText("Last 6 Hours")).toBeVisible();
        await expect(
          adminPage.getByText("Advertiser Name Instacart  +1 other"),
        ).toBeVisible();
        // make sure the url has the correct params
        assertUrlParams(
          adminPage,
          `tr=6h+as+of+latest%2Fh%2B1h&compare_tr=rill-PD&f=advertiser_name IN ('Instacart','Leafly')`,
        );
        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$448.22\s*\+\$417.48 \+1k%\s*vs previous day/,
        );

        // Open bookmark dropdown and verify the "filled" state for the bookmark
        await adminPage.getByLabel("Other bookmark dropdown").click();
        await expect(
          adminPage
            .getByLabel("Complete Bookmark Entry")
            .getByLabel("Bookmark filled icon"),
        ).toBeVisible();
      });

      test("Should delete complete bookmark", async ({ adminPage }) => {
        await adminPage.goto("/e2e/openrtb/canvas/bids_canvas_bookmarks");
        await adminPage.getByLabel("Other bookmark dropdown").click();
        const menuItem = adminPage.getByLabel("Complete Bookmark Entry");
        await menuItem.hover();
        await menuItem.getByRole("button", { name: "Delete bookmark" }).click();
        await expect(
          adminPage.getByText("Bookmark Complete deleted"),
        ).toBeVisible();
      });
    });

    // Home bookmark interferes with other bookmark creation since we are adding some filters.
    // So adding it to the end.
    test.describe.serial("Home canvas bookmarks", () => {
      test("Should create a home bookmark", async ({ adminPage }) => {
        // This would ideally be done in a beforeAll hook. But adminPage fixture is not supported in that hook.
        await adminPage.goto("/e2e/openrtb/canvas/bids_canvas_bookmarks");

        // Select "Last 7 days" as time range
        await interactWithTimeRangeMenu(adminPage, async () => {
          await adminPage
            .getByRole("menuitem", { name: "Last 7 days" })
            .click();
        });

        // Filter to "instacart.com" and "hyundaiusa.com" "Adomain" via leaderboard
        await adminPage
          .getByRole("row", { name: "hyundaiusa.com $3.6k" })
          .scrollIntoViewIfNeeded();
        await adminPage
          .getByRole("row", { name: "hyundaiusa.com $3.6k" })
          .click();
        await adminPage
          .getByRole("row", { name: "instacart.com $2.2k" })
          .click();

        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$5,802\s*\+\$1,250 \+28%\s*vs previous day/,
        );

        // "Go to home" resets to default when there is no home bookmark
        await adminPage.getByLabel("Home bookmark dropdown").click();
        await adminPage.getByLabel("Home Bookmark Entry").click();
        // saved home bookmark is restored
        await expect(adminPage.getByText("Last 24 hours")).toBeVisible();
        await expect(adminPage.getByText("No filters selected")).toBeVisible();
        // make sure the url has the correct params
        assertUrlParams(adminPage, `tr=PT24H&compare_tr=rill-PD`);
        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$3,900\s*\+\$1,858 \+91%\s*vs previous day/,
        );

        // Go back to previous state
        await adminPage.goBack();
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$5,802\s*\+\$1,250 \+28%\s*vs previous day/,
        );

        // Open the bookmarks dropdown
        await adminPage.getByLabel("Home bookmark dropdown").click();
        // Create the current filter as home bookmark
        await adminPage
          .getByRole("menuitem", { name: "Bookmark current view as Home." })
          .click();
        // Wait for the notification that home bookmark was created
        await expect(
          adminPage.getByText("Home bookmark created"),
        ).toBeVisible();
      });

      test("Visiting home should restore home bookmark", async ({
        adminPage,
      }) => {
        // Navigate to the canvas
        await adminPage.goto("/e2e/openrtb/-/dashboards");
        await adminPage
          .getByRole("link", { name: "Bids Canvas Dashboard For Bookmarks" })
          .first()
          .click();

        // saved home bookmark is restored
        await expect(adminPage.getByText("Last 7 days")).toBeVisible();
        await expect(
          adminPage.getByText("Adomain hyundaiusa.com  +1 other"),
        ).toBeVisible();
        // Filters have applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$5,802\s*\+\$1,250 \+28%\s*vs previous day/,
        );
        // make sure the url has the correct params
        assertUrlParams(
          adminPage,
          `tr=7D+as+of+latest%2Fh%2B1h&compare_tr=rill-PD&f=adomain IN ('hyundaiusa.com','instacart.com')`,
        );
      });

      test("Visiting dashboard with params should not apply home bookmark", async ({
        adminPage,
      }) => {
        // Add random params. Home bookmark shouldnt apply
        await adminPage.goto(
          "/e2e/openrtb/canvas/bids_canvas_bookmarks?compare_tr=rill-PW",
        );
        // Default time range is present
        await expect(adminPage.getByText("Last 24 Hours")).toBeVisible();
        // Comparison is previous week
        await expect(
          adminPage.getByRole("button", {
            name: "Select time comparison option",
          }),
        ).toContainText("Previous week");
        // No filter applied
        await expect(adminPage.getByText("No filters selected")).toBeVisible();
        // Assert filters applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$3,900\s*\+\$3,877 \+17k%\s*vs previous week/,
        );

        // Click on "Go to home bookmark"
        await adminPage.getByLabel("Home bookmark dropdown").click();
        await adminPage.getByLabel("Home Bookmark Entry").click();
        // saved home bookmark is restored
        await expect(adminPage.getByText("Last 7 days")).toBeVisible();
        await expect(
          adminPage.getByText("Adomain hyundaiusa.com  +1 other"),
        ).toBeVisible();
        // Filters have applied
        await expect(
          adminPage.getByLabel("overall_spend KPI data"),
        ).toContainText(
          /Advertising Spend Overall\s*\$5,802\s*\+\$1,250 \+28%\s*vs previous day/,
        );
        // make sure the url has the correct params
        assertUrlParams(
          adminPage,
          `tr=7D+as+of+latest%2Fh%2B1h&compare_tr=rill-PD&f=adomain IN ('hyundaiusa.com','instacart.com')`,
        );
      });

      test("Should delete home bookmark", async ({ adminPage }) => {
        await adminPage.goto("/e2e/openrtb/canvas/bids_canvas_bookmarks");
        await adminPage.getByLabel("Home bookmark dropdown").click();
        const menuItem = adminPage.getByLabel("Home Bookmark Entry");
        await menuItem.hover();
        await menuItem.getByRole("button", { name: "Delete bookmark" }).click();
        await expect(
          adminPage.getByText("Bookmark Go to home deleted"),
        ).toBeVisible();
      });

      // TODO: verify editing home bookmark. since these are changing in a future feature, these tests should be part of that PR
    });
  });
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

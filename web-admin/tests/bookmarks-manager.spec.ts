import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { test } from "./setup/base";
import { ADMIN_STORAGE_STATE } from "./setup/constants";

const BOOKMARK_NAME = "Manager bookmark";
const BOOKMARK_DESCRIPTION = "Created for the bookmark manager test.";
const RENAMED_BOOKMARK_NAME = "Manager bookmark renamed";

test.describe.serial("Bookmark manager", () => {
  // The adminPage fixture is test-scoped, so the hook builds its own admin context.
  test.beforeAll(async ({ browser }) => {
    const context = await browser.newContext({
      storageState: ADMIN_STORAGE_STATE,
    });
    const page = await context.newPage();
    await page.goto("/e2e/openrtb/explore/auction_explore_bookmarks");

    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    await page.getByLabel("Other bookmark dropdown").click();
    await page
      .getByRole("menuitem", { name: "Bookmark current view", exact: true })
      .click();
    await page.getByTitle("Label").fill(BOOKMARK_NAME);
    await page.getByTitle("Description").fill(BOOKMARK_DESCRIPTION);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Bookmark created")).toBeVisible();

    await context.close();
  });

  test("Project home lists bookmarks and sorts them", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb");

    await expect(
      adminPage.getByRole("heading", { name: /^Bookmarks/ }),
    ).toBeVisible();
    const entry = adminPage.getByLabel(`${BOOKMARK_NAME} Bookmark Entry`);
    await expect(entry).toBeVisible();
    await expect(entry).toContainText(BOOKMARK_DESCRIPTION);
    await expect(entry).toContainText("Updated");

    // The bookmarks sort is independent of the dashboards sort on the same page.
    const bookmarksHeading = adminPage.getByRole("heading", {
      name: /^Bookmarks/,
    });
    await bookmarksHeading
      .getByRole("button", { name: "Sort by Last used" })
      .click();
    await adminPage.getByRole("menuitemcheckbox", { name: "Name" }).click();
    await expect(adminPage).toHaveURL(/bookmarks_sort=name_asc/);
    await expect(
      bookmarksHeading.getByRole("button", { name: "Sort by Name" }),
    ).toBeVisible();
    await expect(
      adminPage
        .getByRole("heading", { name: /^Dashboards/ })
        .getByRole("button", { name: "Sort by Last Used" }),
    ).toBeVisible();
  });

  test("Bookmarks page supports search, edit, open and delete", async ({
    adminPage,
  }) => {
    await adminPage.goto("/e2e/openrtb");
    await adminPage
      .getByRole("link", { name: "Bookmarks", exact: true })
      .click();
    await expect(adminPage.getByLabel("Container title")).toHaveText(
      "Project bookmarks",
    );

    const entry = adminPage.getByLabel(`${BOOKMARK_NAME} Bookmark Entry`);
    await expect(entry).toBeVisible();

    // Search matches the name and reports when nothing matches.
    const search = adminPage.getByPlaceholder("Search...");
    await search.fill("no such bookmark");
    await expect(
      adminPage.getByText("No bookmarks match your search"),
    ).toBeVisible();
    await search.fill("manager");
    await expect(entry).toBeVisible();
    await search.fill("");

    // Edit the name from the manager.
    await entry.hover();
    await adminPage.getByRole("button", { name: "Edit bookmark" }).click();
    await adminPage.getByTitle("Label").fill(RENAMED_BOOKMARK_NAME);
    await adminPage.getByRole("button", { name: "Save" }).click();
    await expect(adminPage.getByText("Bookmark updated")).toBeVisible();
    const renamedEntry = adminPage.getByLabel(
      `${RENAMED_BOOKMARK_NAME} Bookmark Entry`,
    );
    await expect(renamedEntry).toBeVisible();

    // Opening the bookmark restores its state on the dashboard.
    await renamedEntry.click();
    await expect(adminPage).toHaveURL(
      /\/e2e\/openrtb\/explore\/auction_explore_bookmarks\?/,
    );
    await expect(adminPage.getByText("Last 6 Hours")).toBeVisible();

    // Opening it recorded a last used time for this browser.
    await adminPage.goto("/e2e/openrtb/-/bookmarks");
    await expect(renamedEntry).toContainText("Last used");

    // Delete with confirmation.
    await renamedEntry.hover();
    await adminPage.getByRole("button", { name: "Delete bookmark" }).click();
    await adminPage
      .getByRole("button", { name: "Delete", exact: true })
      .click();
    await expect(
      adminPage.getByText(`Bookmark ${RENAMED_BOOKMARK_NAME} deleted`),
    ).toBeVisible();
    await expect(renamedEntry).not.toBeVisible();

    // The dashboard's bookmark dropdown no longer lists it either.
    // Navigate within the app so the dropdown reads the cached (invalidated) list instead of a fresh page load.
    await adminPage
      .getByRole("link", { name: "Dashboards", exact: true })
      .click();
    await adminPage
      .getByRole("link", { name: "Programmatic Ads Auction For Bookmarks" })
      .click();
    await expect(adminPage).toHaveURL(
      /\/e2e\/openrtb\/explore\/auction_explore_bookmarks/,
    );
    await adminPage.getByLabel("Other bookmark dropdown").click();
    await expect(
      adminPage.getByRole("menuitem", {
        name: "Bookmark current view",
        exact: true,
      }),
    ).toBeVisible();
    await expect(renamedEntry).not.toBeVisible();
  });
});

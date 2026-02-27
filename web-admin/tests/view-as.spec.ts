import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("View As", () => {
  test.describe.configure({ mode: "serial" });

  const TEST_PROJECT_URL = "/e2e/openrtb";
  const TEST_DASHBOARD_URL = "/e2e/openrtb/explore/auction_explore";
  const DIFFERENT_PROJECT_URL = "/e2e/AdBids";

  test("Admin can access View As menu from project home", async ({
    adminPage,
  }) => {
    // Navigate to project home page
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu
    await adminPage.getByRole("img", { name: "avatar" }).click();

    // Verify View As menu item is visible (requires manageProject permission)
    const viewAsMenuItem = adminPage.getByRole("menuitem", { name: "View as" });
    await expect(viewAsMenuItem).toBeVisible();

    // Open the View As submenu
    await viewAsMenuItem.click();

    // Verify the View As popover content is visible
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    await expect(
      adminPage.getByText("Test your security policies"),
    ).toBeVisible();
  });

  test("Admin can access View As menu from dashboard page", async ({
    adminPage,
  }) => {
    // Navigate to dashboard
    await adminPage.goto(TEST_DASHBOARD_URL);
    await expect(adminPage.getByText("Requests")).toBeVisible({
      timeout: 15000,
    });

    // Open the avatar menu
    await adminPage.getByRole("img", { name: "avatar" }).click();

    // Verify View As menu item is visible
    const viewAsMenuItem = adminPage.getByRole("menuitem", { name: "View as" });
    await expect(viewAsMenuItem).toBeVisible();

    // Open the View As submenu
    await viewAsMenuItem.click();

    // Verify the View As popover content is visible
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
  });

  test("Admin can select a user to view as", async ({ adminPage }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu
    await adminPage.getByRole("img", { name: "avatar" }).click();

    // Open the View As submenu
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();

    // Select the first user in the list (the admin user itself)
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    const userEmail = await firstUser.textContent();
    await firstUser.click();

    // Verify the View As chip is displayed
    await expect(adminPage.getByText(`Viewing as`)).toBeVisible();
    await expect(adminPage.getByText(userEmail?.trim() ?? "")).toBeVisible();
  });

  test("View As state persists across page refresh", async ({ adminPage }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu and select a user
    await adminPage.getByRole("img", { name: "avatar" }).click();
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load and select the first one
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    const userEmail = await firstUser.textContent();
    await firstUser.click();

    // Verify the View As chip is displayed
    await expect(adminPage.getByText("Viewing as")).toBeVisible();

    // Refresh the page
    await adminPage.reload();

    // Verify the View As state persists after refresh
    await expect(adminPage.getByLabel("Project title")).toBeVisible();
    await expect(adminPage.getByText("Viewing as")).toBeVisible();
    await expect(adminPage.getByText(userEmail?.trim() ?? "")).toBeVisible();
  });

  test("View As state persists when navigating within the same project", async ({
    adminPage,
  }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu and select a user
    await adminPage.getByRole("img", { name: "avatar" }).click();
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load and select the first one
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    const userEmail = await firstUser.textContent();
    await firstUser.click();

    // Verify the View As chip is displayed
    await expect(adminPage.getByText("Viewing as")).toBeVisible();

    // Navigate to a dashboard within the same project
    await adminPage.goto(TEST_DASHBOARD_URL);
    await expect(adminPage.getByText("Requests")).toBeVisible({
      timeout: 15000,
    });

    // Verify the View As state persists
    await expect(adminPage.getByText("Viewing as")).toBeVisible();
    await expect(adminPage.getByText(userEmail?.trim() ?? "")).toBeVisible();
  });

  test("View As state clears when navigating to a different project", async ({
    adminPage,
  }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu and select a user
    await adminPage.getByRole("img", { name: "avatar" }).click();
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load and select the first one
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    await firstUser.click();

    // Verify the View As chip is displayed
    await expect(adminPage.getByText("Viewing as")).toBeVisible();

    // Navigate to a different project
    await adminPage.goto(DIFFERENT_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Verify the View As state is cleared
    await expect(adminPage.getByText("Viewing as")).not.toBeVisible();
  });

  test("Admin can clear View As state using the chip remove button", async ({
    adminPage,
  }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu and select a user
    await adminPage.getByRole("img", { name: "avatar" }).click();
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load and select the first one
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    await firstUser.click();

    // Verify the View As chip is displayed
    await expect(adminPage.getByText("Viewing as")).toBeVisible();

    // Click the remove button on the chip
    const removeButton = adminPage.getByLabel("Clear view");
    await removeButton.click();

    // Verify the View As chip is removed
    await expect(adminPage.getByText("Viewing as")).not.toBeVisible();
  });

  test("Admin can change the viewed user using the chip dropdown", async ({
    adminPage,
  }) => {
    // Navigate to project home
    await adminPage.goto(TEST_PROJECT_URL);
    await expect(adminPage.getByLabel("Project title")).toBeVisible();

    // Open the avatar menu and select a user
    await adminPage.getByRole("img", { name: "avatar" }).click();
    await adminPage.getByRole("menuitem", { name: "View as" }).click();

    // Wait for users to load and select the first one
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();
    const userItems = adminPage.locator('[role="option"]');
    const firstUser = userItems.first();
    await expect(firstUser).toBeVisible({ timeout: 10000 });
    const firstUserEmail = await firstUser.textContent();
    await firstUser.click();

    // Verify the View As chip is displayed with the first user
    await expect(adminPage.getByText("Viewing as")).toBeVisible();
    await expect(
      adminPage.getByText(firstUserEmail?.trim() ?? ""),
    ).toBeVisible();

    // Click on the chip to open the dropdown
    await adminPage.getByText("Viewing as").click();

    // Verify the dropdown is open with user search
    await expect(adminPage.getByPlaceholder("Search for users")).toBeVisible();

    // Get users list again - there might be different number of users now
    const newUserItems = adminPage.locator('[role="option"]');
    const userCount = await newUserItems.count();

    if (userCount > 1) {
      // Select a different user (second one)
      const secondUser = newUserItems.nth(1);
      const secondUserEmail = await secondUser.textContent();
      await secondUser.click();

      // Verify the chip now shows the new user
      await expect(adminPage.getByText("Viewing as")).toBeVisible();
      await expect(
        adminPage.getByText(secondUserEmail?.trim() ?? ""),
      ).toBeVisible();
    }
  });

  // This test requires a viewer user to be set up in the e2e environment.
  // Currently, the viewerPage fixture is not fully configured (viewer.json is not created in setup.ts).
  test.skip("Effective permissions update correctly when viewing as a viewer", async ({
    adminPage,
  }) => {
    // Navigate to dashboard as admin
    await adminPage.goto(TEST_DASHBOARD_URL);
    await expect(adminPage.getByText("Requests")).toBeVisible({
      timeout: 15000,
    });

    // Verify the Share button is visible for admin
    const shareButton = adminPage.getByRole("button", { name: "Share" });
    await expect(shareButton).toBeVisible();

    // TODO: Select a viewer user via View As
    // Implementation depends on having a viewer user in the project

    // Verify the Share button is hidden for viewer
    await expect(shareButton).not.toBeVisible();
  });
});

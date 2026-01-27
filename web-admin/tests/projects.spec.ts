import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Projects", () => {
  test("admins should see the admin-only pages", async ({ adminPage }) => {
    await adminPage.goto("/e2e/openrtb");
    await expect(adminPage.getByRole("link", { name: "Status" })).toBeVisible();
    await expect(
      adminPage.getByRole("link", { name: "Settings" }),
    ).toBeVisible();
  });

  test("status page should show Local Development section", async ({
    adminPage,
  }) => {
    await adminPage.goto("/e2e/openrtb/-/status");

    // Check Local Development header is visible
    await expect(adminPage.getByText("Local Development")).toBeVisible();

    // Click the Download project button to open popover
    await adminPage.getByRole("button", { name: "Download project" }).click();

    // Check Learn more link is visible in popover (filter by surrounding text to avoid ambiguity)
    await expect(
      adminPage
        .locator("span")
        .filter({ hasText: "Clone this project to develop locally" })
        .getByRole("link", { name: "Learn more ->" }),
    ).toBeVisible();

    // Check clone command is visible (for non-GitHub connected project)
    await expect(
      adminPage.getByText("rill project clone openrtb"),
    ).toBeVisible();
  });
});

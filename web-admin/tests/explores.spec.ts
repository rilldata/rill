import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Explores", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/e2e/openrtb/-/dashboards");
    await page
      .getByRole("link", { name: "Programmatic Ads Auction" })
      .first()
      .click();
  });

  test("should see the Big Number", async ({ page }) => {
    await expect(
      page.getByRole("button", { name: "Requests 6.60M" }),
    ).toBeVisible();
  });

  test("should see the Rows Viewer", async ({ page }) => {
    await page.getByRole("button", { name: "Toggle rows viewer" }).click();
    await expect(
      page.getByText("app_site_domain", { exact: true }),
    ).toBeVisible();
    await expect(
      page.locator("td").filter({ hasText: "businessinsider.com" }),
    ).toBeVisible();
  });
});

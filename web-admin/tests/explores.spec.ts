import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Explores", () => {
  test("should have data", async ({ page }) => {
    await page.goto("/e2e/openrtb");

    // Navigate to the explore
    await page
      .getByRole("link", { name: "Programmatic Ads Bids" })
      .first()
      .click();

    // Check the Big Number
    await expect(
      page.getByRole("button", { name: "Advertising Spend Overall $1.41M" }),
    ).toBeVisible();
  });
});

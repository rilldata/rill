import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Explores", () => {
  test("should have data", async ({ page }) => {
    await page.goto("/e2e/openrtb");

    // Navigate to the explore
    await page
      .getByRole("link", { name: "Programmatic Ads Auction" })
      .first()
      .click();

    // Set the time zone to UTC
    await page.getByLabel("Timezone selector").click();
    await page.getByRole("menuitem", { name: "UTC UTC UTC+00:00" }).click();

    // Check the Big Number
    await expect(
      page.getByRole("button", { name: "Requests 547M" }),
    ).toBeVisible();
  });
});

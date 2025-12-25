import { expect } from "@playwright/test";
import { test } from "./setup/base";

test.describe("Canvases", () => {
  test("should have data", async ({ page }) => {
    // Navigate to canvas
    await page.goto("/e2e/openrtb/-/dashboards");
    await page
      .getByRole("link", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    // Check the KPI Grid data
    await expect(
      page.getByText("Advertising Spend Overall $3,900"),
    ).toBeVisible();
  });

  test("can switch projects from canvas", async ({ page }) => {
    // Navigate to canvas
    await page.goto("/e2e/openrtb/-/dashboards");
    await page
      .getByRole("link", { name: "Bids Canvas Dashboard" })
      .first()
      .click();

    // navigate via breadcrumbs to another project
    await page
      .getByRole("button", { name: "Breadcrumb dropdown" })
      .first()
      .click();

    await page.getByRole("link", { name: "AdBids" }).click();

    await expect(page.getByText("Adbids Canvas Dashboard")).toBeVisible();
  });
});

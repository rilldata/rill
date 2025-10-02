import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas time filters", () => {
  test.use({ project: "AdBids" });

  test("can update time filters", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page
      .locator("#AdBids_metrics_canvas--component-0-0 div")
      .filter({ hasText: "Total records 1,122 -5 ~0% vs previous day" })
      .first()
      .click();

    await page.getByRole("button", { name: "Filters" }).click();
    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByRole("switch")
      .first()
      .click();

    // Set local time range
    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Select time range")
      .click();
    await page.getByRole("menuitem", { name: "Last 7 days" }).click();

    await page.waitForTimeout(500);

    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Toggle time comparison")
      .click();

    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Select time comparison option")
      .click();

    await page.getByRole("menuitem", { name: "Previous week" }).click();

    await expect(
      page.getByText("Total records 7,863 -15 ~0% vs previous week"),
    ).toBeVisible();

    await page.getByRole("button", { name: "Options" }).click();

    // Change global time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    await expect(page.getByText("7,863")).toBeVisible();
  });

  test("can update domain filters", async ({ page }) => {
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByText("Total records 1,122 -5 ~0% vs previous day").click();

    await page.getByRole("button", { name: "Options" }).click();
    await page.getByRole("button", { name: "Add filter button" }).click();
    await page.getByRole("menuitem", { name: "Domain" }).click();

    await page.getByLabel("domain results").getByText("facebook.com").click();
    await page
      .getByLabel("domain results")
      .getByText("google.com", { exact: true })
      .click();
    await page.getByLabel("domain results").getByText("msn.com").click();

    // Close the dropdown to apply the selections (Select mode applies on close)
    await page
      .getByRole("button", { name: "Open domain filter" })
      .first()
      .click();

    await expect(page.locator(".kpi-wrapper").getByText("797")).toBeVisible();

    await page.getByRole("button", { name: "Filters", exact: true }).click();
    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByRole("switch")
      .nth(1)
      .click();
    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Add filter button")
      .click();
    await page.getByRole("menuitem", { name: "Domain" }).click();
    await page.getByLabel("domain results").getByText("msn.com").click();

    // Close the dropdown to apply the selection (Select mode applies on close)
    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByRole("button", { name: "Open domain filter" })
      .click();

    await expect(page.getByText("375")).toBeVisible();
  });
});

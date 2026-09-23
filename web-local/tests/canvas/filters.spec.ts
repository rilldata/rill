import { expect } from "@playwright/test";
import { interactWithTimeRangeMenu } from "@rilldata/web-common/tests/utils/explore-interactions";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas time filters", () => {
  test.use({ project: "AdBids" });

  test("can update time filters", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByLabel("total_records KPI data").first().click();

    await page.getByRole("button", { name: "Time & filters" }).click();

    // The widget follows the canvas until a range is picked.
    const timeRangeSelect = page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Select time range");
    await expect(timeRangeSelect).toContainText("Inherit from canvas");

    // Set local time range
    await timeRangeSelect.click();
    await page.getByRole("menuitem", { name: "Last 7 days" }).click();

    // Wait for the local time range to apply before choosing a comparison,
    // otherwise the comparison change can race the range change.
    await expect(
      page
        .getByRole("complementary", { name: "Inspector Panel" })
        .getByLabel("Select time range"),
    ).toContainText("Last 7");

    await page
      .getByRole("complementary", { name: "Inspector Panel" })
      .getByLabel("Select time comparison option")
      .click();

    await page.getByRole("menuitem", { name: "Previous week" }).click();

    await expect(
      page.getByLabel("total_records KPI data").first(),
    ).toContainText("vs previous week");

    await page.getByRole("button", { name: "Options" }).click();

    // Change global time range
    await interactWithTimeRangeMenu(page, async () => {
      await page.getByRole("menuitem", { name: "Last 6 Hours" }).click();
    });

    await expect(page.getByText("7,863")).toBeVisible();

    // Handing the range back to the canvas drops the local override,
    // so the widget follows the 6 hour canvas range again.
    await page.getByLabel("total_records KPI data").first().click();
    await page
      .getByRole("button", { name: "Time & filters", exact: true })
      .click();
    await timeRangeSelect.click();
    await page.getByRole("menuitem", { name: "Inherit from canvas" }).click();
    await expect(timeRangeSelect).toContainText("Inherit from canvas");
    await expect(page.getByText("7,863")).not.toBeVisible();
  });

  test("can disable comparison for a single widget", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    const kpi = page.getByLabel("total_records KPI data").first();
    // The canvas can take a while to render under parallel workers.
    await expect(kpi).toBeVisible({ timeout: 30_000 });

    // Make sure comparison is on at the canvas level.
    const globalToggle = page.getByLabel("Toggle time comparison").first();
    const globalSwitch = globalToggle.getByRole("switch");
    if (!(await globalSwitch.isChecked())) {
      await globalToggle.click();
    }
    await expect(globalSwitch).toBeChecked();
    await expect(kpi).toContainText("vs");

    await kpi.click();
    await page
      .getByRole("button", { name: "Time & filters", exact: true })
      .click();

    const inspector = page.getByRole("complementary", {
      name: "Inspector Panel",
    });
    const comparisonSelect = inspector.getByLabel(
      "Select time comparison option",
    );
    await expect(comparisonSelect).toContainText("Inherit from canvas");

    // The widget toggle only affects this widget; the canvas comparison stays on.
    await inspector.getByLabel("Toggle time comparison").click();
    await expect(kpi).not.toContainText("vs");
    await expect(globalSwitch).toBeChecked();

    await comparisonSelect.click();
    await page.getByRole("menuitem", { name: "Inherit from canvas" }).click();
    await expect(comparisonSelect).toContainText("Inherit from canvas");
    await expect(kpi).toContainText("vs");
  });

  test("can update domain filters", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");

    await page.getByLabel("total_records KPI data").first().click();

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

    await page
      .getByRole("button", { name: "Time & filters", exact: true })
      .click();
    // The first switch in the panel is the widget comparison toggle.
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
    // Use Escape instead of clicking the trigger; bits-ui v2's dismiss layer
    // intercepts pointer events on the underlying trigger button.
    await page.keyboard.press("Escape");

    await expect(page.getByText("375")).toBeVisible();
  });
});

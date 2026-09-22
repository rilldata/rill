import { expect } from "@playwright/test";
import { gotoNavEntry } from "web-local/tests/utils/waitHelpers";
import { test } from "../setup/base";

test.describe("canvas kpi grid measure selector", () => {
  test.use({ project: "AdBids" });

  test("dropdown reflects selection state after toggling", async ({ page }) => {
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_canvas.yaml");
    await page
      .getByRole("button", { name: "Add widget" })
      .waitFor({ state: "visible" });
    await page.getByRole("button", { name: "Add widget" }).click();
    await page.getByRole("menuitem", { name: "KPI", exact: true }).click();

    await page.getByLabel("Add measure fields").click();

    const item = page.getByRole("menuitemcheckbox", {
      name: "Sum of Bid Price",
    });
    await expect(item).toHaveAttribute("aria-checked", "true");
    // The menu stays open in multi-select mode and must reflect deselection...
    await item.click();
    await expect(item).toHaveAttribute("aria-checked", "false");
    // ...and reselection.
    await item.click();
    await expect(item).toHaveAttribute("aria-checked", "true");
  });
});

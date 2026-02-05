import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("canvas generation", () => {
  test.use({ project: "AdBids" });

  test("Generate canvas dashboard from metrics view", async ({ page }) => {
    // Expand the metrics folder first
    await page.getByLabel("/metrics").click();

    // Navigate to the metrics view file
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");

    // Open the dropdown menu and click "Create Canvas dashboard"
    await page.getByLabel("Create resource menu").click();
    await page
      .getByRole("menuitem", { name: "Generate Canvas Dashboard" })
      .click();

    // Wait for navigation to a canvas file
    // The name will be unique (e.g., AdBids_metrics_canvas_1 since AdBids_metrics_canvas already exists)
    await page.waitForURL(
      /\/files\/dashboards\/AdBids_metrics_canvas.*\.yaml/,
      {
        timeout: 10000,
      },
    );

    // Verify the Preview button is available (indicates we're on a valid canvas file)
    await expect(page.getByRole("button", { name: "Preview" })).toBeVisible({
      timeout: 5000,
    });

    // Click Preview to verify the canvas renders correctly
    await page.getByRole("button", { name: "Preview" }).click();

    // Wait for preview to load - canvas components should be visible
    // The generated canvas should have components based on the metrics view
    await expect(page.locator('[id*="--component"]').first()).toBeVisible({
      timeout: 5000,
    });
  });
});

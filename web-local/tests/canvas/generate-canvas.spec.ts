import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { gotoNavEntry } from "../utils/waitHelpers";

test.describe("canvas generation", () => {
  test.use({ project: "AdBids" });

  // Agent-based canvas generation involves multiple LLM round-trips
  test.setTimeout(120_000);

  test("Generate canvas dashboard from metrics view", async ({ page }) => {
    // Expand the metrics folder first
    await page.getByLabel("/metrics").click();

    // Navigate to the metrics view file
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");

    // Open the dropdown menu and click "Generate Canvas Dashboard"
    await page.getByLabel("Create resource menu").click();
    await page
      .getByRole("menuitem", { name: "Generate Canvas Dashboard" })
      .click();

    // Wait for navigation to a canvas file
    // The developer agent generates the canvas and then calls the "navigate" tool,
    // which triggers client-side navigation. This can take a while due to LLM round-trips.
    await page.waitForURL(
      /\/files\/dashboards\/AdBids_metrics_canvas.*\.yaml/,
      {
        timeout: 90_000,
      },
    );

    // Verify the Preview button is available (indicates we're on a valid canvas file)
    await expect(page.getByRole("button", { name: "Preview" })).toBeVisible({
      timeout: 10_000,
    });

    // Click Preview to verify the canvas renders correctly
    await page.getByRole("button", { name: "Preview" }).click();

    // Wait for preview to load - canvas components should be visible
    // The generated canvas should have components based on the metrics view
    await expect(page.locator('[id*="--component"]').first()).toBeVisible({
      timeout: 10_000,
    });
  });
});

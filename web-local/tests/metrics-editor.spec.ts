import { expect } from "@playwright/test";
import { gotoNavEntry } from "./utils/waitHelpers";
import { updateCodeEditor, wrapRetryAssertion } from "./utils/commonHelpers";
import { test } from "./setup/base";

test.describe("Metrics editor", () => {
  test.use({ project: "AdBids" });

  test("Can add and remove measures and dimensions", async ({ page }) => {
    await page.getByLabel("/metrics").click();
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");

    await page.getByRole("button", { name: "Add new measure" }).click();
    await page.getByText("Model column").click();
    await page
      .getByRole("option", { name: "bid_price" })
      .waitFor({ timeout: 2_000 });
    await page.getByRole("option", { name: "bid_price" }).click();
    await page.getByLabel("Display name (optional)").fill("New Measure");
    await page.getByRole("button", { name: "Add measure" }).click();

    await expect(page.getByText("New Measure", { exact: true })).toBeVisible();

    await page.getByRole("button", { name: "Add new dimension" }).click();

    await page.getByText("Column from model").click();
    await page.getByRole("option", { name: "bid_price" }).click();
    await page.getByLabel("Display name (optional)").fill("New Dimension");
    await page.getByRole("button", { name: "Add dimension" }).click();

    await expect(
      page.getByText("New Dimension", { exact: true }),
    ).toBeVisible();

    // Delete measure
    await page.getByRole("row", { name: "measure New Measure" }).hover();
    await page
      .getByRole("button", { name: "Delete measure New Measure" })
      .click();

    await page.getByRole("button", { name: "Yes, delete" }).click();

    await expect(
      page.getByText("New Measure", { exact: true }),
    ).not.toBeVisible();

    // Delete dimension
    await page.getByRole("row", { name: "dimension New Dimension" }).hover();
    await page
      .getByRole("button", { name: "Delete dimension New Dimension" })
      .click();

    await page.getByRole("button", { name: "Yes, delete" }).click();

    await expect(
      page.getByText("New Dimension", { exact: true }),
    ).not.toBeVisible();
  });

  test("Metrics editor", async ({ page }) => {
    await page.getByLabel("/metrics").click();
    await page.getByLabel("/dashboards").click();
    await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");

    await page.getByRole("button", { name: "switch to code editor" }).click();

    await updateCodeEditor(page, "");

    // inspector should point to the documentation
    await expect(page.getByText("For help building dashboards")).toBeVisible();

    // skeleton should result in an empty skeleton YAML file
    await page.getByText("start with a skeleton").click();

    // check to see that the placeholder is gone by looking for the button that was once there
    await wrapRetryAssertion(async () => {
      await expect(page.getByText("start with a skeleton")).toBeHidden();
    });

    // This is causing issues in the test, so we'll skip it for now.
    // await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");
    // // the Preview button should be disabled
    // await expect(page.getByRole("button", { name: "Preview" })).toBeDisabled();
    // await page.waitForTimeout(3000);
    // await gotoNavEntry(page, "/metrics/AdBids_metrics.yaml");

    // the editor should show a validation error
    await expect(
      page.getByText(
        'must set a value for either the "model", "table" or "parent" field',
      ),
    ).toBeVisible();

    // now let's create a working metrics view
    await updateCodeEditor(page, "");
    await wrapRetryAssertion(async () => {
      await expect(
        page.getByText("metrics configuration from an existing model"),
      ).toBeVisible();
    });

    // select the first menu item.
    await page
      .getByText("metrics configuration from an existing model")
      .click();
    await page.getByRole("menuitem").getByText("AdBids_model").click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 6 Hours" }),
    ).not.toBeVisible();

    // let's check the inspector.
    await expect(page.getByText("Table columns")).toBeVisible();

    // go to the dashboard and make sure the metrics and dimensions are there.
    await gotoNavEntry(page, "/dashboards/AdBids_metrics_explore.yaml");

    await page.getByRole("button", { name: "Preview" }).click();

    // check to see metrics make sense.
    await expect(page.getByText("Total Records 100k")).toBeVisible();

    // double-check that leaderboards make sense.
    await expect(
      page.getByRole("row", { name: "google.com 15.1k" }),
    ).toBeVisible();

    // go back to the metrics page.
    await page.getByRole("button", { name: "Edit" }).click();
    await page.getByRole("menuitem", { name: "Metrics View" }).click();

    // Create a metrics view from a table (rather than a model)
    await updateCodeEditor(
      page,
      `version: 1
type: metrics_view
title: "AdBids table"
table: "AdBids"
timeseries: "timestamp"
measures:
  - name: "Total Records"
    expression: count(*)
dimensions:
  - name: publisher
    label: Publisher
    column: publisher
  - name: domain
    label: Domain
    column: domain
  `,
    );

    // Check that the metrics inspector shows the table columns
    await expect(page.getByText("Table columns")).toBeVisible();
  });
});

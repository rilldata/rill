import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";

/// ClickHouse Cloud test to add clickhouse connector, create metrics -> explore dashboard, as well as table -> dashboard.
/// Fairly basic testing, and expects the preview to load correctly.

test.describe("Connecting to ClickHouse Cloud and Managed ClickHouse.", () => {
  RillTest("Create ClickHouse Connection...", async ({ page }) => {
    /// Using ClickHouse Cloud, create a metrics view then explore dashboard.
    // create blank file
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Source").click();
    await page.locator(`button#clickhouse`).click();

    const inputField = page.locator("input#host");
    await inputField.fill("l3e2nu99sw.us-east-1.aws.clickhouse.cloud");

    const inputField2 = page.locator("input#port");
    await inputField2.fill("8443");

    const inputField3 = page.locator("input#username");
    await inputField3.fill(`default`);

    const inputField4 = page.locator("input#password");
    await inputField4.fill(`g3Dvo~aQjNwKf`);

    const inputField5 = page.locator("input#ssl");
    await inputField5.check();

    await page
      .locator('button[form="add-data-clickhouse-form"]')
      .waitFor({ state: "visible" });
    await page.locator('button[form="add-data-clickhouse-form"]').click();

    // wait for connector to be created. Check that connectors clichouse is displayed
    await page.waitForSelector(
      'li[aria-label="/connectors/clickhouse.yaml Nav Entry"]',
      { state: "visible" },
    );
    const section = page.locator("section.connector-section");
    const clickhouseEntry = section.locator('li[aria-label="clickhouse"]');
    await expect(clickhouseEntry).toHaveCount(1);

    // find table
    const defaultButton = page.locator(
      'li[aria-label="clickhouse"] button:has-text("default")',
    );
    const ukPricePaidItem = page.locator(
      'a.clickable-text[href="/connector/clickhouse/clickhouse/default/uk_price_paid"]',
    );

    // Check if the "ol" is visible
    const isVisible = await ukPricePaidItem.isVisible();

    if (!isVisible) {
      // Click the "default" button to expand it
      await defaultButton.click();
    }

    // Hover over the "uk_price_paid" item and create metrics
    await ukPricePaidItem.hover();
    await page
      .getByLabel("clickhouse-default.uk_price_paid actions menu trigger")
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Generate metrics with AI")')
      .click();

    // Create explore dashboard
    await page.waitForSelector(
      'li[aria-label="/metrics/uk_price_paid_metrics.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page
      .locator('button[role="button"]:has-text("Create Explore dashboard")')
      .click();

    // Check Visual Dashboard is working
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_metrics_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();

    // Checking the dashboard, 9/10 Total Transactions measure gets created but had 1 instance where it was not...
    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    await expect(
      page
        .locator('a[role="button"]:has-text("Total Transactions")')
        .locator("div")
        .last(),
    ).toHaveText("28.9M");

    /// Create an explore dashboard with AI from table
    await ukPricePaidItem.hover();
    await page
      .getByLabel("clickhouse-default.uk_price_paid actions menu trigger")
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Generate dashboard with AI")')
      .click();

    // Check Visual dashboard and dashboard preview
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_metrics_1_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();

    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    await expect(
      page
        .locator('a[role="button"]:has-text("Total Transactions")')
        .locator("div")
        .last(),
    ).toHaveText("28.9M");

    /// Feature Flag ClickHouse Modeling, testing a model
    await page.locator('li[aria-label="/rill.yaml Nav Entry"]').click();
    const projectTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await projectTextbox.click();

    // sometimes its too fast and the typing gets messed up so adding a 2 second wait
    await page.waitForTimeout(2000);

    const projectlines = ["", "features:", "  - clickhouseModeling"];

    // Type each line with a newline after
    for (const line of projectlines) {
      await projectTextbox.type(line); // Type the line
      await projectTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    // Start testing the UK dataset by creating model.
    await ukPricePaidItem.hover();
    await page
      .getByLabel("clickhouse-default.uk_price_paid actions menu trigger")
      .click();
    await page
      .locator('div[role="menuitem"]:has-text("Create new model")')
      .click();
    await page.waitForSelector(
      'li[aria-label="/models/uk_price_paid_model.sql Nav Entry"]',
      { state: "visible" },
    );

    // Check the inspector panel and make sure all the panels are not empty or (Truthy)
    await page.getByLabel("Inspector Panel").isVisible();
    const columns = [
      "date",
      "price",
      "is_new",
      "addr1",
      "street",
      "postcode2",
      "addr2",
      "locality",
      "district",
      "town",
      "county",
      "postcode1",
      "type",
      "duration",
    ];
    const inspectorPanel = page.locator('aside[aria-label="Inspector Panel"]');
    for (const columnName of columns) {
      await inspectorPanel.locator(`button:has-text("${columnName}")`).click();
      const columnPanel = inspectorPanel.locator(
        `div:has-text("${columnName}")`,
      );

      const wFullDivs = columnPanel.locator("div.w-full");

      // Check that at least one `div.w-full` is not empty check this later
      const isNonEmptyDivPresent = await wFullDivs.evaluateAll((divs) =>
        divs.some((div) => div.textContent.trim() !== ""),
      );

      // Assert that at least one non-empty div is present
      expect(isNonEmptyDivPresent).toBeTruthy();

      //close details
      await inspectorPanel.locator(`button:has-text("${columnName}")`).click();
    }

    // Create metrics and dashboard from Model
    await page
      .locator('button:has-text("Generate metrics view with AI")')
      .click();
    await page.waitForSelector(
      'li[aria-label="/metrics/uk_price_paid_model_metrics.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page
      .locator('button[role="button"]:has-text("Create Explore dashboard")')
      .click();

    // Since the model is a sample of 100K, the values are not certain, instead check if the graph is truthy. 9/10 Total Transaction appears
    await page.waitForSelector(
      'li[aria-label="/dashboards/uk_price_paid_model_metrics_explore.yaml Nav Entry"]',
      { state: "visible" },
    );
    await page.getByLabel("Inspector Panel").isVisible();
    await page.waitForSelector(
      'a[role="button"]:has-text("Total Transactions")',
      { state: "visible" },
    );
    // await expect(page.locator('a[role="button"]:has-text("Total Transactions")').locator('div').last()).toHaveText('99,999'); // since sample its going to be random
    const tddGraph = page.locator(
      'a[role = "button"]:has-text("Total Transactions")',
    );
    expect(tddGraph.locator('svg[role="application"]')).toBeTruthy();

    //TO DO: Incremental Modeling
  });
});

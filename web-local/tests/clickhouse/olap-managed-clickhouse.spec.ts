import { test, expect } from "@playwright/test";
import { test as RillTest } from "../utils/test";
import {
  addFolderWithCheck,
  addFileWithCheck,
  waitForSource,
  uploadFile,
} from "../utils/sourceHelpers";

/// Managed ClickHouse

test.describe("Connecting to ClickHouse Cloud and Managed ClickHouse.", () => {
  RillTest("Create ClickHouse Connection...", async ({ page }) => {
    /// Using Managed ClickHouse , create a metrics view then explore dashboard from a model. include checks on each step.
    addFolderWithCheck(page, "untitled_folder");
    await page.locator('button[id="/untitled_folder-nav-entry"]').hover();
    await page.getByLabel("untitled_folder actions menu trigger").click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const inputField = page.locator("input#folder-name");
    await inputField.fill("connectors");

    await page.locator('button[form="rename-asset-form"]').click();

    // create blank file
    const navContainer = page.locator("#nav-\\/");
    await addFileWithCheck(page, "untitled_file");
    await navContainer
      .locator('li[aria-label="/untitled_file Nav Entry"]')
      .dragTo(page.locator('button[id="/connectors-nav-entry"]'));

    await page
      .locator('li[aria-label="/connectors/untitled_file Nav Entry"]')
      .hover();
    await page
      .getByLabel("/connectors/untitled_file actions menu trigger")
      .click();
    await page.locator('div[role="menuitem"]:has-text("Rename...")').click();

    const inputField1 = page.locator("input#file-name");
    await inputField1.fill("clickhouse.yaml");
    await page.locator('button[form="rename-asset-form"]').click();

    const childTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );
    await childTextbox.click(); // Ensure it's focused for typing

    const lines = ["type: connector", "driver: clickhouse", "managed: true"];

    // Type each line with a newline after
    for (const line of lines) {
      await childTextbox.type(line); // Type the line
      await childTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    const section = page.locator("section.connector-section");
    const clickhouseEntry = section.locator('li[aria-label="clickhouse"]');
    await expect(clickhouseEntry).toHaveCount(1);

    const noTablesMessage = section.locator('span:has-text("No tables found")');
    await expect(noTablesMessage).toBeVisible();

    // Now that Managed ClickHouse has been set up, enable featureflag for CH modeling and start workflow.

    await page.locator('li[aria-label="/rill.yaml Nav Entry"]').click();
    const projectTextbox = await page.locator(
      'div[role="textbox"] div.cm-content',
    );

    await projectTextbox.click();

    const projectlines = ["", "features:", "  - clickhouseModeling"];

    // Type each line with a newline after
    for (const line of projectlines) {
      await projectTextbox.type(line); // Type the line
      await projectTextbox.press("Enter"); // Press Enter for a new line
    }

    await page.locator('button:has-text("Save")').click();

    // await Promise.all([
    //   waitForSource(page, "/sources/AdBids.yaml", [
    //     "publisher",
    //     "domain",
    //     "timestamp",
    //   ]),
    //   uploadFile(page, "AdBids.csv"),
    // ]);

    // TODO: Add source, create model, etc.

    // TODO: Incremental Modeling
  });
});

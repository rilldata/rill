import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Snowflake connector", () => {
  test.use({ project: "Blank" });

  test("submits connector with DSN", async ({ page }) => {
    const dsn = process.env.RILL_RUNTIME_SNOWFLAKE_TEST_DSN;
    if (!dsn) {
      test.skip(true, "RILL_RUNTIME_SNOWFLAKE_TEST_DSN is not set");
    }

    // Open Add Data modal and pick Snowflake
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#snowflake").click();
    await page.waitForSelector('form[id*="snowflake"]');

    // Switch to the DSN tab
    await page.getByRole("button", { name: "Enter connection string" }).click();

    // Fill DSN field
    await page
      .getByRole("textbox", { name: "Snowflake Connection String" })
      .fill(dsn!);

    // Submit connector form
    const submitButton = page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" });
    await expect(submitButton).toBeEnabled();
    await submitButton.click();

    // Expect navigation to the new connector file
    await page.waitForURL("**/files/connectors/snowflake.yaml");

    // Validate connector YAML contents
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: snowflake");
    await expect(codeEditor).toContainText(
      'dsn: "{{ .env.connector.snowflake.dsn }}"',
    );
  });
});

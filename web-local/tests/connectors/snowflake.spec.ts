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
    await page.getByLabel("Connect to snowflake").click();

    // Switch to the DSN tab
    await page.getByRole("tab", { name: "Connection String" }).click();

    // Fill DSN field
    await page.getByRole("textbox", { name: "Connection String" }).fill(dsn!);

    // Validate connector YAML contents
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).toContainText("type: connector");
    await expect(yamlPreview).toContainText("driver: snowflake");
    await expect(yamlPreview).toContainText('dsn: "{{ .env.SNOWFLAKE_DSN }}"');

    // Submit connector form
    const submitButton = page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" });
    await expect(submitButton).toBeEnabled();
    await submitButton.click();

    // Wait for pick a table screen
    await expect(
      page.getByText(
        "Pick a table or input your SQL to power your first dashboard",
      ),
    ).toBeVisible();

    // Skip testing import to avoid putting load on our infrastructure
  });
});

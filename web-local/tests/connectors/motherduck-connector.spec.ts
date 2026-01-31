import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { updateCodeEditor } from "../utils/commonHelpers";

test.describe("MotherDuck welcome flow", () => {
  // Start from an empty workspace to exercise the onboarding path
  test.use({ project: undefined });

  test("initializes MotherDuck from welcome screen and persists secrets before connector", async ({
    page,
  }) => {
    const token = process.env.RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN;
    const path = process.env.RILL_RUNTIME_MOTHERDUCK_TEST_PATH;
    const schema =
      process.env.RILL_RUNTIME_MOTHERDUCK_TEST_SCHEMA ??
      "rilldata_integration_test";

    if (!token || !path) {
      test.skip(
        true,
        "RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN or RILL_RUNTIME_MOTHERDUCK_TEST_PATH not configured",
      );
    }

    // Welcome CTA opens the Add Data modal
    await page.getByRole("button", { name: "Connect your data" }).click();

    // Select MotherDuck connector and wait for the form to render
    await page.locator("#motherduck").click();
    await page.waitForSelector('form[id*="motherduck"]');

    // Fill required fields
    await page.locator("#token").fill(token!);
    await page.locator("#path").fill(path!);
    await page.locator("#schema_name").fill(schema);

    // Test and Connect should create .env first, then the connector file
    await page
      .getByRole("dialog")
      .getByRole("button", { name: /^(Test and Connect|Connect)$/ })
      .click();

    // Navigate to the created connector file
    await page.waitForURL(/.*\/files\/connectors\/motherduck.*\.yaml/);

    // Verify connector file references env-based secrets with new ALL_CAPS naming
    const connectorEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(connectorEditor).toContainText("type: connector");
    await expect(connectorEditor).toContainText("driver: duckdb");
    // Confirm the original {{ .env.MOTHERDUCK_TOKEN }} format exists
    await expect(connectorEditor).toContainText(
      'token: "{{ .env.MOTHERDUCK_TOKEN }}"',
    );
    await expect(connectorEditor).toContainText(`path: "${path}"`);

    // Now update the connector file to use case-insensitive env function syntax
    // Replace entire quoted value to avoid nested quote issues in YAML
    const currentContent = await connectorEditor.textContent();
    const updatedContent = currentContent!.replace(
      '"{{ .env.MOTHERDUCK_TOKEN }}"',
      `'{{ env "motherduck_TOKEN" }}'`,
    );
    await updateCodeEditor(page, updatedContent);
    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForTimeout(1000);

    // Verify the updated syntax is now in the file
    await expect(connectorEditor).toContainText(
      `token: '{{ env "motherduck_TOKEN" }}'`,
    );

    // Verify no errors appear (case-insensitive env function should resolve correctly)
    const errorPane = page.locator(".editor-pane .error");
    const errorCount = await errorPane.count();
    if (errorCount > 0) {
      const errorText = await errorPane.textContent();
      // Should not have error about missing env variable
      expect(errorText).not.toContain("motherduck_TOKEN");
      expect(errorText).not.toContain("environment variable");
    }

    // .env should hold MotherDuck secrets with new ALL_CAPS naming
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText(`MOTHERDUCK_TOKEN=${token}`);

    // rill.yaml should promote MotherDuck as the OLAP connector
    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(rillYamlEditor).toContainText("olap_connector: motherduck");
  });
});

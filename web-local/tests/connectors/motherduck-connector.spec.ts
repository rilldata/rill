import { expect } from "@playwright/test";
import { test } from "../setup/base";

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

    // Verify connector file references env-based secrets
    const connectorEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(connectorEditor).toContainText("type: connector");
    await expect(connectorEditor).toContainText("driver: duckdb");
    await expect(connectorEditor).toContainText(
      'token: "{{ .env.connector.motherduck.token }}"',
    );
    await expect(connectorEditor).toContainText(`path: "${path}"`);

    // .env should hold MotherDuck secrets
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText(
      `connector.motherduck.token=${token}`,
    );

    // rill.yaml should promote MotherDuck as the OLAP connector
    await page.getByRole("link", { name: "rill.yaml" }).click();
    const rillYamlEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(rillYamlEditor).toContainText("olap_connector: motherduck");
  });
});

import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { ClickHouseTestContainer } from "../utils/clickhouse";

test.describe("ClickHouse connector", () => {
  const clickhouseOne = new ClickHouseTestContainer();
  const clickhouseTwo = new ClickHouseTestContainer();

  test.beforeAll(async () => {
    await clickhouseOne.start();
    await clickhouseOne.seedAdBids();
    await clickhouseTwo.start();
    await clickhouseTwo.seedAdImpressions();
  });

  test.afterAll(async () => {
    await clickhouseOne.stop();
    await clickhouseTwo.stop();
  });

  test.describe("Welcome screen", () => {
    test("Create connector using individual fields", async ({ page }) => {
      await enterClickhouseCredentials(page, clickhouseOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, false);

      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", {
            name: "clickhouse",
            exact: true,
          }),
      ).toBeVisible();

      // Go to the `.env` file and verify the CLICKHOUSE_PASSWORD is set
      await page.getByRole("link", { name: ".env" }).click();
      const envEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(envEditor).toContainText(`CLICKHOUSE_PASSWORD=password`);
    });

    test("Create connector using DSN", async ({ page }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();

      // Switch to "self-managed", "cloud" options does not allow non-ssl connections.
      await page.getByLabel("Connection type").click();
      await page.getByText("Self Managed").click();

      // Switch to the DSN tab
      await page.getByRole("tab", { name: "Enter connection string" }).click();

      // Fill in the form correctly
      await page
        .getByRole("textbox", { name: "Connection String" })
        .fill(
          `http://${clickhouseOne.getHost()}:${clickhouseOne.getPort().toString()}?username=default&password=password`,
        );

      // Assert that the preview has correct properties
      const yamlPreview = page.getByLabel("Yaml preview");
      await expect(yamlPreview).toContainText("type: connector");
      await expect(yamlPreview).toContainText("driver: clickhouse");
      await expect(yamlPreview).toContainText(
        'dsn: "{{ .env.CLICKHOUSE_DSN }}"',
      );

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, false);

      // Go to the `.env` file and verify the CLICKHOUSE_DSN is set
      await page.getByRole("link", { name: ".env" }).click();
      const envEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(envEditor).toContainText(
        `CLICKHOUSE_DSN=http://${clickhouseOne.getHost()}:${clickhouseOne.getPort().toString()}?username=default&password=password`,
      );

      // Go to the `rill.yaml` and verify the OLAP connector is set
      await page.getByRole("link", { name: "rill.yaml" }).click();
      const rillYamlEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(rillYamlEditor).toContainText("olap_connector: clickhouse");

      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", {
            name: "clickhouse",
            exact: true,
          }),
      ).toBeVisible();
    });
  });

  test.describe("Home page", () => {
    test.use({ project: "Blank" });

    test("Create connector and metrics view from home screen", async ({
      page,
    }) => {
      await enterClickhouseCredentials(page, clickhouseOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, true);

      // Go to the `rill.yaml` and verify the OLAP connector is set
      await page.getByRole("link", { name: "rill.yaml" }).click();
      const rillYamlEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(rillYamlEditor).toContainText("olap_connector: clickhouse");

      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", {
            name: "clickhouse",
            exact: true,
          }),
      ).toBeVisible();
    });

    test("Create connector from home screen and metrics view from add asset", async ({
      page,
    }) => {
      await enterClickhouseCredentials(page, clickhouseOne);
      // Save without testing connection
      await page.getByLabel("Save connector").click();
      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", {
            name: "clickhouse",
            exact: true,
          }),
      ).toBeVisible();

      // add asset button
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add metrics view").hover();
      await page.getByLabel("Create metrics view for ClickHouse").click();

      await selectAdBidsAndSubmit(page, true);

      // Open the add data modal.
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add Data").click();

      // Create another clickhouse connector.
      await enterClickhouseCredentials(page, clickhouseTwo, false);
      // Save without testing connection
      await page.getByLabel("Save connector").click();
      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", {
            name: "clickhouse_1",
            exact: true,
          }),
      ).toBeVisible();

      // add asset button
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add metrics view").hover();
      await page.getByLabel("Create metrics view for ClickHouse").click();

      // ad_bids from 1st connector is showing.
      await page.getByLabel("Node: default, level 0").click();
      await expect(page.getByLabel("Node: ad_bids, level 1")).toBeVisible();

      // Select the second clickhouse connector
      await page.getByLabel("Select connector").click();
      await page.getByRole("option", { name: "clickhouse_1" }).click();

      // Select `ad_impressions` from the second connector
      await page.getByLabel("Node: default, level 0").click();
      await page.getByLabel("Node: ad_impressions, level 1").click();

      // Click generate dashboard button
      await page
        .getByRole("button", { name: "Generate dashboard with AI" })
        .click();

      // Creating metrics view is triggered.
      await expect(page.getByText("Creating Metrics View...")).toBeVisible();

      // Wait for navigation to the new file
      await page.waitForURL(/\/files\/metrics\/ad_impressions_metrics.yaml/, {
        timeout: 10_000,
      });
    });
  });
});

async function enterClickhouseCredentials(
  page: Page,
  clickhouse: ClickHouseTestContainer,
  assertYaml = true,
) {
  // Open the connect to clickhouse modal
  await page.getByLabel("Connect to clickhouse").click();

  // Verify form validation - empty host - button should be disabled
  await expect(
    page.getByRole("button", { name: "Test and Connect" }),
  ).toBeDisabled();

  // Switch to "self-managed", "cloud" options does not allow non-ssl connections.
  await page.getByLabel("Connection type").click();
  await page.getByText("Self Managed").click();

  // Fill in the form correctly
  await page.getByRole("textbox", { name: "Host" }).fill(clickhouse.getHost());
  await page.getByRole("textbox", { name: "Host" }).press("Tab");
  await page
    .getByRole("textbox", { name: "Port" })
    .fill(clickhouse.getPort().toString());
  await page.getByRole("textbox", { name: "Port" }).press("Tab");
  await page.getByRole("textbox", { name: "Username" }).fill("default");
  await page.getByRole("textbox", { name: "Password" }).fill("password");
  await page.getByRole("checkbox").scrollIntoViewIfNeeded();
  await page.getByRole("checkbox").click();

  if (assertYaml) {
    // Assert that the yaml contains key properties
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).toContainText("type: connector");
    await expect(yamlPreview).toContainText("driver: clickhouse");
    await expect(yamlPreview).toContainText(`host: "${clickhouse.getHost()}"`);
    await expect(yamlPreview).toContainText(
      `port: "${clickhouse.getPort().toString()}"`,
    );
    await expect(yamlPreview).toContainText('username: "default"');
    await expect(yamlPreview).toContainText(
      'password: "{{ .env.CLICKHOUSE_PASSWORD }}"',
    );
    await expect(yamlPreview).toContainText("ssl: false");
  }
}

async function selectAdBidsAndSubmit(page: Page, metricsViewOnly: boolean) {
  // Wait for pick a table screen
  await expect(
    page.getByText("Pick a table to power your first dashboard"),
  ).toBeVisible();

  // Select ad_bids table
  await page.getByLabel("Node: default, level 0").click();
  await page.getByLabel("Node: ad_bids, level 1").click();

  // Click generate dashboard button
  await page
    .getByRole("button", { name: "Generate dashboard with AI" })
    .click();

  // Creating metrics view is triggered.
  await expect(page.getByText("Creating Metrics View...")).toBeVisible();

  if (metricsViewOnly) {
    // Wait for navigation to the new file
    await page.waitForURL(/\/files\/metrics\/ad_bids_metrics.yaml/, {
      timeout: 10_000,
    });
    return;
  }

  // Metrics view is created.
  await expect
    .poll(async () => page.getByText("Created Metrics View").isVisible(), {
      timeout: 10_000,
    })
    .toBeTruthy();

  // Wait for navigation to the new file
  await page.waitForURL(/\/files\/dashboards\/ad_bids_metrics_canvas.yaml/, {
    timeout: 10_000,
  });
}

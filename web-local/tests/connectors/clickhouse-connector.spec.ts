import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { ClickHouseTestContainer } from "../utils/clickhouse";
import { gotoNavEntry } from "../utils/waitHelpers.ts";
import { validateYamlContents } from "../utils/yamlHelpers.ts";

test.describe("ClickHouse connector", () => {
  // There is an edge case where a full page error is triggered by `FileAndResourceWatcher`.
  // This seems to happen when the controller restarts.
  // Since we are changing the olap_connector here, it always restarts and triggers the edge cases sometimes.
  // TODO: fix FileAndResourceWatcher to be more robust.
  test.describe.configure({ retries: 3 });

  const clickhouseOne = new ClickHouseTestContainer();
  const clickhouseTwo = new ClickHouseTestContainer();

  test.beforeAll(async () => {
    await clickhouseOne.start();
    await clickhouseTwo.start();
    await clickhouseOne.seedAdBids();
    await clickhouseTwo.seedAdImpressions();
  });

  test.afterAll(async () => {
    await clickhouseOne.stop();
    await clickhouseTwo.stop();
  });

  test.describe("Welcome screen", () => {
    test("Create connector using individual fields", async ({ page }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
      // Verify form validation - empty host - button should be disabled
      await expect(
        page.getByRole("button", { name: "Test and Connect" }),
      ).toBeDisabled();

      await enterClickhouseCredentials(page, clickhouseOne);
      // Fill the incorrect password first
      await page.getByRole("textbox", { name: "Password" }).fill("invalid");

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Error message should be displayed
      await expect(page.getByLabel("Submission error")).toBeVisible();

      // Fill the correct password
      await page.getByRole("textbox", { name: "Password" }).fill("password");
      // Error message is cleared when fields are updated
      await expect(page.getByLabel("Submission error")).not.toBeVisible();

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
      await validateYamlContents(
        page,
        [`CLICKHOUSE_PASSWORD=password`],
        [`CLICKHOUSE_PASSWORD_1`],
      );
    });

    test("Create connector using DSN", async ({ page }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();

      // Switch to "self-managed", "cloud" options does not allow non-ssl connections.
      await page.getByLabel("Connection type").click();
      await page.getByText("Self Managed").click();

      // Switch to the DSN tab
      await page.getByRole("tab", { name: "Enter connection string" }).click();

      // Fill in incorrect DSN
      await page
        .getByRole("textbox", { name: "Connection String" })
        .fill(
          `http://${clickhouseOne.getHost()}:${clickhouseOne.getPort().toString()}?username=invalid&password=invalid`,
        );

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Error message should be displayed
      await expect(page.getByLabel("Submission error")).toBeVisible();

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
      await validateYamlContents(
        page,
        [
          `CLICKHOUSE_DSN=http://${clickhouseOne.getHost()}:${clickhouseOne.getPort().toString()}?username=default&password=password`,
        ],
        [`CLICKHOUSE_DSN_1`],
      );

      // Go to the `rill.yaml` and verify the OLAP connector is set
      await page.getByRole("link", { name: "rill.yaml" }).click();
      await validateYamlContents(page, [`olap_connector: clickhouse`]);

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
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
      // Verify form validation - empty host - button should be disabled
      await expect(
        page.getByRole("button", { name: "Test and Connect" }),
      ).toBeDisabled();

      await enterClickhouseCredentials(page, clickhouseOne);
      // Fill the incorrect password first
      await page.getByRole("textbox", { name: "Password" }).fill("invalid");

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Error message should be displayed
      await expect(page.getByLabel("Submission error")).toBeVisible();

      // Fill the correct password
      await page.getByRole("textbox", { name: "Password" }).fill("password");

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, true);

      // Go to the `rill.yaml` and verify the OLAP connector is set
      await page.getByRole("link", { name: "rill.yaml" }).click();
      await validateYamlContents(page, [`olap_connector: clickhouse`]);

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
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
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

      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
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
      await page
        .getByLabel("Import Table Form")
        .getByLabel("default.default")
        .click();
      await expect(
        page
          .getByLabel("Import Table Form")
          .getByLabel("clickhouse-default.ad_bids"),
      ).toBeVisible();

      // Select the second clickhouse connector
      await page.getByLabel("Select connector").click();
      await page.getByRole("option", { name: "clickhouse_1" }).click();

      // Select `ad_impressions` from the second connector
      await page
        .getByLabel("Import Table Form")
        .getByLabel("default.default")
        .click();
      await page
        .getByLabel("Import Table Form")
        .getByLabel("clickhouse_1-default.ad_impressions")
        .click();

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

    test("Create one connector and another from header", async ({ page }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
      await enterClickhouseCredentials(page, clickhouseOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Select the add new clickhouse connector option
      await page.getByLabel("Select connector").click();
      await page
        .getByRole("option", { name: "+ ClickHouse connector" })
        .click();

      // Enter details for the 2nd clickhouse instance
      await enterClickhouseCredentials(page, clickhouseTwo, false);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdImpressionsAndSubmit(page, "clickhouse_1");

      // Open the connectors folder
      await page.getByLabel("/connectors").click();

      // Assert that "connector" is created
      await gotoNavEntry(page, "/connectors/clickhouse.yaml");
      await validateYamlContents(page, [
        `port: "${clickhouseOne.getPort().toString()}"`,
        'password: "{{ .env.CLICKHOUSE_PASSWORD }}"',
      ]);

      // Assert that "connector_1" is created
      await gotoNavEntry(page, "/connectors/clickhouse_1.yaml");
      await validateYamlContents(page, [
        `port: "${clickhouseTwo.getPort().toString()}"`,
        'password: "{{ .env.CLICKHOUSE_PASSWORD_1 }}"',
      ]);

      // Go to the `.env` file and verify the CLICKHOUSE_PASSWORD and CLICKHOUSE_PASSWORD_1 is set
      await page.getByRole("link", { name: ".env" }).click();
      await validateYamlContents(page, [
        `CLICKHOUSE_PASSWORD=password`,
        `CLICKHOUSE_PASSWORD_1=password`,
      ]);
    });

    test("Create one connector and go back and create another", async ({
      page,
    }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
      await enterClickhouseCredentials(page, clickhouseOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Wait for pick a table screen
      await expect(
        page.getByText("Pick a table to power your first dashboard"),
      ).toBeVisible();
      // Go back to the connector form
      await page.getByRole("button", { name: "Back" }).click();

      // Enter details for the 2nd clickhouse instance
      await enterClickhouseCredentials(page, clickhouseTwo);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdImpressionsAndSubmit(page, "clickhouse");

      // Open the connectors folder
      await page.getByLabel("/connectors").click();

      // Assert that "connector" is created with the second clickhouse instance
      await gotoNavEntry(page, "/connectors/clickhouse.yaml");
      await validateYamlContents(page, [
        `port: "${clickhouseTwo.getPort().toString()}"`,
        'password: "{{ .env.CLICKHOUSE_PASSWORD }}"',
      ]);

      // Assert that "connector_1" is not created
      await expect(
        page.getByLabel("/connectors/clickhouse_1.yaml"),
      ).not.toBeVisible();

      // Go to the `.env` file and verify the CLICKHOUSE_PASSWORD is set
      await page.getByRole("link", { name: ".env" }).click();
      await validateYamlContents(
        page,
        [`CLICKHOUSE_PASSWORD=password`],
        [`CLICKHOUSE_PASSWORD_1`],
      );
    });

    test("Create a connector cancel all the way to connector selector", async ({
      page,
    }) => {
      // Open the connect to clickhouse modal
      await page.getByLabel("Connect to clickhouse").click();
      await enterClickhouseCredentials(page, clickhouseOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Wait for pick a table screen
      await expect(
        page.getByText("Pick a table to power your first dashboard"),
      ).toBeVisible();
      // Go back to the connector form
      await page.getByRole("button", { name: "Back" }).click();

      // Go back to the connector form
      await page.getByRole("button", { name: "Back" }).click();

      // Open the connectors folder
      await page.getByLabel("/connectors").click();

      // Assert that "connector" is not created
      await expect
        .poll(() => page.getByLabel("/connectors/clickhouse.yaml").count(), {
          timeout: 10_000,
        })
        .toBe(0);

      // Go to the `.env` file and verify the CLICKHOUSE_PASSWORD is unset
      await page.getByRole("link", { name: ".env" }).click();
      await validateYamlContents(page, [], [`CLICKHOUSE_PASSWORD`]);

      // Go to the `rill.yaml` and verify the OLAP connector is unset
      await page.getByRole("link", { name: "rill.yaml" }).click();
      await validateYamlContents(page, [], [`olap_connector: clickhouse`]);
    });
  });
});

async function enterClickhouseCredentials(
  page: Page,
  clickhouse: ClickHouseTestContainer,
  assertYaml = true,
) {
  // Switch to "self-managed", "cloud" options does not allow non-ssl connections.
  await page.getByLabel("Connection type").click();
  await page.getByRole("option").getByText("Self Managed").click();

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
  if (await page.getByRole("checkbox").isChecked()) {
    await page.getByRole("checkbox").click();
  }

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
  await page
    .getByLabel("Import Table Form")
    .getByLabel("default.default")
    .click();
  await page
    .getByLabel("Import Table Form")
    .getByLabel("clickhouse-default.ad_bids")
    .click();

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

async function selectAdImpressionsAndSubmit(page: Page, connectorName: string) {
  // Wait for pick a table screen
  await expect(
    page.getByText("Pick a table to power your first dashboard"),
  ).toBeVisible();

  // Select `ad_impressions` from the second connector
  await page
    .getByLabel("Import Table Form")
    .getByLabel("default.default")
    .click();
  await page
    .getByLabel("Import Table Form")
    .getByLabel(`${connectorName}-default.ad_impressions`)
    .click();

  // Click generate dashboard button
  await page
    .getByRole("button", { name: "Generate dashboard with AI" })
    .click();

  // Wait for navigation to the new file
  await page.waitForURL(/\/files\/metrics\/ad_impressions_metrics.yaml/, {
    timeout: 10_000,
  });
}

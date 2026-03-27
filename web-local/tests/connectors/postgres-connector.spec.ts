import { expect, type Page } from "@playwright/test";
import { test } from "../setup/base";
import { PostgresTestContainer } from "../utils/postgres.ts";

test.describe("Postgres connector", () => {
  const postgresOne = new PostgresTestContainer();
  const postgresTwo = new PostgresTestContainer();

  test.beforeAll(async () => {
    await postgresOne.start();
    await postgresOne.seedAdBids();
    await postgresTwo.start();
    await postgresTwo.seedAdImpressions();
  });

  test.afterAll(async () => {
    await postgresOne.stop();
    await postgresTwo.stop();
  });

  test.describe("Welcome screen", () => {
    test("Create connector using individual fields", async ({ page }) => {
      await page.getByLabel("Connect your data").click();
      await enterPostgresCredentials(page, postgresOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, false);

      // Assert that the connector explorer now has a postgres connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", { name: "postgres" })
          .first(),
      ).toBeVisible();

      // Go to the `.env` file and verify the POSTGRES_PASSWORD is set
      await page.getByRole("link", { name: ".env" }).click();
      const envEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(envEditor).toContainText(`POSTGRES_PASSWORD=password`);
    });

    test("Create connector using DSN", async ({ page }) => {
      // Open the connect to postgres modal
      await page.getByLabel("Connect your data").click();
      await page.getByLabel("Connect to postgres").click();

      // Switch to the DSN tab
      await page.getByRole("tab", { name: "Enter connection string" }).click();

      // Fill in the form correctly
      await page
        .getByRole("textbox", { name: "Connection String" })
        .fill(
          `postgresql://default:password@${postgresOne.getHost()}:${postgresOne.getPort().toString()}/postgres`,
        );

      // Assert that the preview has correct properties
      const yamlPreview = page.getByLabel("Yaml preview");
      await expect(yamlPreview).toContainText("type: connector");
      await expect(yamlPreview).toContainText("driver: postgres");
      await expect(yamlPreview).toContainText('dsn: "{{ .env.POSTGRES_DSN }}"');

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      // Wait for pick a table screen
      await expect(
        page.getByText(
          "Pick a table or input your SQL to power your first dashboard",
        ),
      ).toBeVisible();

      // Switch to the SQL mode
      await page.getByRole("tab", { name: "SQL" }).click();

      // Enter sql query to get data from ad_bids
      await page
        .getByRole("textbox", { name: "SQL" })
        .fill("SELECT * FROM ad_bids");

      // Click generate dashboard button
      await page
        .getByRole("button", { name: "Generate dashboard with AI" })
        .click();

      // Creating model is triggered.
      await expect(page.getByText("Ingesting data...")).toBeVisible();

      // Model is created.
      await expect
        .poll(async () => page.getByText("Ingested data").isVisible(), {
          timeout: 30_000,
        })
        .toBeTruthy();

      // Wait for navigation to the dashboard
      await page.waitForURL(
        /\/files\/dashboards\/ad_bids_metrics_canvas.yaml/,
        {
          timeout: 30_000,
        },
      );

      // Assert that the connector explorer now has a postgres connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", { name: "postgres" })
          .first(),
      ).toBeVisible();

      // Go to the `.env` file and verify the POSTGRES_DSN is set
      await page.getByRole("link", { name: ".env" }).click();
      const envEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(envEditor).toContainText(
        `POSTGRES_DSN=postgresql://default:password@${postgresOne.getHost()}:${postgresOne.getPort().toString()}/postgres`,
      );
    });
  });

  test.describe("Home page", () => {
    test.use({ project: "Blank" });

    test("Create connector and metrics view from home screen", async ({
      page,
    }) => {
      await page.getByLabel("Connect your data").click();
      await enterPostgresCredentials(page, postgresOne);

      // Submit the form
      await page.getByRole("button", { name: "Test and Connect" }).click();

      await selectAdBidsAndSubmit(page, true);

      // Assert that the connector explorer now has a postgres connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", { name: "postgres" })
          .first(),
      ).toBeVisible();

      // Go to the `.env` file and verify the POSTGRES_PASSWORD is set
      await page.getByRole("link", { name: ".env" }).click();
      const envEditor = page
        .getByLabel("codemirror editor")
        .getByRole("textbox");
      await expect(envEditor).toContainText(`POSTGRES_PASSWORD=password`);
    });

    test("Create connector from home screen and metrics view from add asset", async ({
      page,
    }) => {
      await page.getByLabel("Connect your data").click();
      await enterPostgresCredentials(page, postgresOne);
      // Save without testing connection
      await page.getByLabel("Save connector").click();
      // Assert that the connector explorer now has a ClickHouse connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", { name: "postgres" })
          .first(),
      ).toBeVisible();

      // add asset button
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add model").hover();
      await page.getByLabel("Create model for postgres").click();

      await selectAdBidsAndSubmit(page, true);
      await page.getByRole("button", { name: "View this source" }).click();

      // Open the add data modal.
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add Data").click();

      // Create another postgres connector.
      await enterPostgresCredentials(page, postgresTwo, false);
      // Save without testing connection
      await page.getByLabel("Save connector").click();
      // Assert that the connector explorer now has a Postgres connector
      await expect(
        page
          .getByRole("region", { name: "Data explorer" })
          .getByRole("button", { name: "postgres_1" })
          .first(),
      ).toBeVisible();

      // add asset button
      await page.getByLabel("Add Asset").click();
      await page.getByLabel("Add model").hover();
      await page.getByLabel("Create model for postgres").click();

      // ad_bids from 1st connector is showing.
      await page.getByLabel("Node: postgres, level 0").click();
      await page.getByLabel("Node: public, level 1").click();
      await page.getByLabel("Node: ad_bids, level 2").click();

      // Select the second postgres connector
      await page.getByLabel("Select connector").click();
      await page.getByRole("option", { name: "postgres_1" }).click();

      // Select `ad_impressions` from the second connector
      await page.getByLabel("Node: postgres, level 0").click();
      await page.getByLabel("Node: public, level 1").click();
      await page.getByLabel("Node: ad_impressions, level 2").click();

      // Click import data button
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Import Data" })
        .click();

      // Creating model is triggered.
      await expect(page.getByText("Ingesting data...")).toBeVisible();

      // Wait for navigation to the new file
      await page.waitForURL(/\/files\/models\/ad_impressions.yaml/, {
        timeout: 10_000,
      });
    });
  });
});

async function enterPostgresCredentials(
  page: Page,
  postgres: PostgresTestContainer,
  assertYaml = true,
) {
  // Open the connect to postgres modal
  await page.getByLabel("Connect to postgres").click();

  // Fill in the form correctly
  await page.getByRole("textbox", { name: "Host" }).fill(postgres.getHost());
  await page.getByRole("textbox", { name: "Host" }).press("Tab");
  await page
    .getByRole("textbox", { name: "Port" })
    .fill(postgres.getPort().toString());
  await page.getByRole("textbox", { name: "Port" }).press("Tab");
  await page.getByRole("textbox", { name: "Username" }).fill("default");
  await page.getByRole("textbox", { name: "Password" }).fill("password");
  await page.getByRole("textbox", { name: "Database" }).fill("postgres");

  if (assertYaml) {
    // Assert that the yaml contains key properties
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).toContainText("type: connector");
    await expect(yamlPreview).toContainText("driver: postgres");
    await expect(yamlPreview).toContainText(`host: "${postgres.getHost()}"`);
    await expect(yamlPreview).toContainText(
      `port: "${postgres.getPort().toString()}"`,
    );
    await expect(yamlPreview).toContainText(`user: "default"`);
    await expect(yamlPreview).toContainText(
      'password: "{{ .env.POSTGRES_PASSWORD }}"',
    );
  }
}

async function selectAdBidsAndSubmit(page: Page, modelOnly: boolean) {
  // Wait for pick a table screen
  await expect(
    page.getByText(
      "Pick a table or input your SQL to power your first dashboard",
    ),
  ).toBeVisible();

  // Select ad_bids table
  await page.getByLabel("Node: postgres, level 0").click();
  await page.getByLabel("Node: public, level 1").click();
  await page.getByLabel("Node: ad_bids, level 2").click();

  // Click generate dashboard button
  await page
    .getByRole("button", { name: /Generate dashboard with AI|Import Data/ })
    .click();

  // Creating model is triggered.
  await expect(page.getByText("Ingesting data...")).toBeVisible();

  if (modelOnly) {
    // Wait for navigation to the new file
    await page.waitForURL(/\/files\/models\/ad_bids.yaml/, {
      timeout: 30_000,
    });
    return;
  }

  // Model is created.
  await expect
    .poll(async () => page.getByText("Ingested data").isVisible(), {
      timeout: 30_000,
    })
    .toBeTruthy();

  // Wait for navigation to the dashboard
  await page.waitForURL(/\/files\/dashboards\/ad_bids_metrics_canvas.yaml/, {
    timeout: 30_000,
  });
}

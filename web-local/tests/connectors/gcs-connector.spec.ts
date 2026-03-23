import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("GCS connector", () => {
  test.use({ project: "Blank" });

  test("Renders connector step schema via wrapper", async ({ page }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Choose a multi-step connector (GCS).
    await page.getByLabel("Connect to gcs").click();

    // Connector step should show connector preview and connector CTA.
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page
        .getByRole("dialog")
        .getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Auth method controls from the connector schema should render.
    const hmacRadio = page.getByRole("radio", { name: "HMAC keys" });
    await expect(hmacRadio).toBeVisible();
    await expect(page.getByRole("radio", { name: "Public" })).toBeVisible();

    // Select HMAC so its fields are rendered.
    await hmacRadio.click();

    // Connector step fields should be present, while source step fields should not yet render.
    await expect(
      page.getByRole("textbox", { name: "Access Key ID" }),
    ).toBeVisible();
    await expect(
      page.getByRole("textbox", { name: "Secret Access Key" }),
    ).toBeVisible();
    await expect(page.getByRole("textbox", { name: "GS URI" })).toHaveCount(0);
  });

  test("Renders source step schema via wrapper", async ({ page }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Choose a multi-step connector (GCS).
    await page.getByLabel("Connect to gcs").click();

    // Connector step visible with CTA.
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(
      page
        .getByRole("dialog")
        .getByRole("button", { name: "Test and Connect" }),
    ).toBeVisible();

    // Switch to Public auth (no required fields) and continue via CTA.
    await page.getByRole("radio", { name: "Public" }).click();
    const connectorCta = page.getByRole("button", {
      name: "Continue",
    });
    await connectorCta.click();

    // Source step should now render with source schema fields and CTA.
    await expect(page.getByText("Model preview")).toBeVisible();
    const sourceCta = page.getByRole("button", {
      name: "Import Data",
    });
    await expect(sourceCta).toBeVisible();

    // Source fields should be present; connector-only auth fields should not be required to show.
    await expect(page.getByRole("textbox", { name: "GCS URI" })).toBeVisible();
    await expect(
      page.getByRole("textbox", { name: "Model name" }),
    ).toBeVisible();

    // create_secrets_from_connectors will not be present for public buckets.
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).not.toContainText(
      "create_secrets_from_connectors: gcs",
    );
  });

  test("Disables submit until auth requirements met", async ({ page }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.getByLabel("Connect to gcs").click();

    const connectorCta = page.getByRole("button", {
      name: "Test and Connect",
    });

    // Default auth is credentials (file upload); switch to HMAC to check required fields.
    await page.getByRole("radio", { name: "HMAC keys" }).click();
    await expect(connectorCta).toBeDisabled();

    // Fill key only -> still disabled.
    await page
      .getByRole("textbox", { name: "Access Key ID" })
      .fill("AKIA_TEST");
    await expect(connectorCta).toBeDisabled();

    // Fill secret -> enabled.
    await page
      .getByRole("textbox", { name: "Secret Access Key" })
      .fill("SECRET");
    await expect(connectorCta).toBeEnabled();
  });

  test("Public auth option keeps submit enabled and allows advancing", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.getByLabel("Connect to gcs").click();

    const connectorCta = page.getByRole("button", {
      name: "Continue",
    });

    // Switch to Public (no required fields) -> CTA should remain enabled and allow advancing.
    await page.getByRole("radio", { name: "Public" }).click();
    await expect(connectorCta).toBeEnabled();
    await connectorCta.click();

    // Should land on source step without needing connector fields.
    await expect(page.getByText("Model preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Import Data" }),
    ).toBeVisible();
  });

  test("Save button hidden after advancing to model step", async ({ page }) => {
    // Skip test if environment variables are not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY ||
      !process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET environment variable is not set",
      );
    }

    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.getByLabel("Connect to gcs").click();

    // Fill HMAC credentials on connector step.
    await page.getByRole("radio", { name: "HMAC keys" }).click();
    await page
      .getByRole("textbox", { name: "Access Key ID" })
      .fill(process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY!);
    await page
      .getByRole("textbox", { name: "Secret Access Key" })
      .fill(process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET!);

    // Save button should be visible on the connector step.
    const saveButton = page.getByRole("button", { name: "Save", exact: true });
    await expect(saveButton).toBeVisible();

    // Advance to model step via Test and Connect.
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();
    await expect(page.getByText("Model preview")).toBeVisible();

    // Save button should not be visible on the source step.
    await expect(saveButton).toBeHidden();
  });

  test("Submission using HMAC auth method", async ({ page }) => {
    const hmacKey = process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY;
    const hmacSecret = process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET;
    if (!hmacKey || !hmacSecret) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET is not set",
      );
    }
    test.slow();

    const openGcsFlowWithHmac = async () => {
      await page.getByRole("button", { name: "Add Asset" }).click();
      await page.getByRole("menuitem", { name: "Add Data" }).click();
      await page.getByLabel("Connect to gcs").click();
      await page.getByRole("radio", { name: "HMAC keys" }).click();
      await page.getByRole("textbox", { name: "Access Key ID" }).fill(hmacKey!);
      await page
        .getByRole("textbox", { name: "Secret Access Key" })
        .fill(hmacSecret!);
      const connectorCta = page.getByRole("button", {
        name: "Test and Connect",
      });
      await connectorCta.click();
      await expect(page.getByText("Model preview")).toBeVisible();
    };

    // First submission attempt
    await openGcsFlowWithHmac();
    const firstPath =
      "gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_06.parquet";
    const firstModelName = "gcs_model_one";
    await page.getByRole("textbox", { name: "GCS URI" }).fill(firstPath);
    await page
      .getByRole("textbox", { name: "Model name" })
      .fill(firstModelName);

    const submitCta = page.getByRole("button", {
      name: "Import Data",
    });
    await submitCta.click();

    // Ingesting data is triggered.
    await expect(page.getByText("Ingesting data...")).toBeVisible();
    // Wait for navigation to the new file
    await page.waitForURL(`**/files/models/${firstModelName}.yaml`);

    // Re-open and ensure model form is reset
    await openGcsFlowWithHmac();
    await expect(
      page.getByRole("textbox", { name: "GCS URI" }),
    ).not.toHaveValue(firstPath);
    await expect(
      page.getByRole("textbox", { name: "Model name" }),
    ).not.toHaveValue(firstModelName);

    // Create a second model
    const modelName = "gcs_create_secrets_test";
    await page
      .getByRole("textbox", { name: "GCS URI" })
      .fill(
        "gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_06.parquet",
      );
    await page.getByRole("textbox", { name: "Model name" }).fill(modelName);
    await page.getByRole("button", { name: "Import Data" }).click();

    // Ingesting data is triggered.
    await expect(page.getByText("Ingesting data...")).toBeVisible();
    // Wait for navigation to the new model file
    await page.waitForURL(`**/files/models/${modelName}.yaml`);

    // Verify YAML contains the connector reference and create_secrets_from_connectors with the second instance (gcs_1)
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("connector: duckdb");
    await expect(codeEditor).toContainText(
      /create_secrets_from_connectors:\s*gcs_1/,
    );

    // add asset button
    await page.getByLabel("Add Asset").click();
    await page.getByLabel("Add Model").hover();
    await page.getByLabel("Create model for Google Cloud Storage").click();

    // Preview has reference to 1st connect: `gcs`
    const yamlPreview = page.getByLabel("Yaml preview");
    await expect(yamlPreview).toContainText("connector: duckdb");
    await expect(yamlPreview).toContainText(
      /create_secrets_from_connectors:\s*gcs/,
    );

    // Select the second gcs connector
    await page.getByLabel("Select connector").click();
    await page.getByRole("option", { name: "gcs_1" }).click();
    // Preview has reference to 2nd connect: `gcs_1`
    await expect(yamlPreview).toContainText("connector: duckdb");
    await expect(yamlPreview).toContainText(
      /create_secrets_from_connectors:\s*gcs_1/,
    );
  });
});

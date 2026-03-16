import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Multi-step connector wrapper", () => {
  test.use({ project: "Blank" });

  test("GCS connector - renders connector step schema via wrapper", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Choose a multi-step connector (GCS).
    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

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

  test("GCS connector - renders source step schema via wrapper", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Choose a multi-step connector (GCS).
    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

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
      name: /Test and Connect|Continue/i,
    });
    await connectorCta.click();

    // Source step should now render with source schema fields and CTA.
    await expect(page.getByText("Model preview")).toBeVisible();
    const sourceCta = page.getByRole("button", {
      name: /Test and Add data|Importing data|Add data/i,
    });
    await expect(sourceCta).toBeVisible();

    // Source fields should be present; connector-only auth fields should not be required to show.
    await expect(page.getByRole("textbox", { name: "GCS URI" })).toBeVisible(
      {},
    );
    await expect(
      page.getByRole("textbox", { name: "Model name" }),
    ).toBeVisible();
  });

  test("GCS connector - preserves auth selection across steps", async ({
    page,
  }) => {
    const hmacKey = process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY;
    const hmacSecret = process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET;
    if (!hmacKey || !hmacSecret) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET is not set",
      );
    }

    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

    // Pick HMAC auth and fill required fields.
    await page.getByRole("radio", { name: "HMAC keys" }).click();
    await page.getByRole("textbox", { name: "Access Key ID" }).fill(hmacKey!);
    await page
      .getByRole("textbox", { name: "Secret Access Key" })
      .fill(hmacSecret!);

    // Submit connector step via CTA to transition to source step.
    const connectorCta = page.getByRole("button", {
      name: /Test and Connect|Continue/i,
    });
    await expect(connectorCta).toBeEnabled();
    await connectorCta.click();
    await expect(page.getByText("Model preview")).toBeVisible();

    // Go back to connector step.
    await page.getByRole("button", { name: "Back" }).click();

    // Auth selection and values should persist.
    await expect(page.getByText("Connector preview")).toBeVisible();
    await expect(page.getByRole("radio", { name: "HMAC keys" })).toBeChecked({
      timeout: 5000,
    });
    await expect(
      page.getByRole("textbox", { name: "Access Key ID" }),
    ).toHaveValue(hmacKey!);
    await expect(
      page.getByRole("textbox", { name: "Secret Access Key" }),
    ).toHaveValue(hmacSecret!);
  });

  test("GCS connector - disables submit until auth requirements met", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

    const connectorCta = page.getByRole("button", {
      name: /Test and Connect|Continue/i,
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

  test("GCS connector - public auth option keeps submit enabled and allows advancing", async ({
    page,
  }) => {
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

    const connectorCta = page.getByRole("button", {
      name: /Test and Connect|Continue/i,
    });

    // Switch to Public (no required fields) -> CTA should remain enabled and allow advancing.
    await page.getByRole("radio", { name: "Public" }).click();
    await expect(connectorCta).toBeEnabled();
    await connectorCta.click();

    // Should land on source step without needing connector fields.
    await expect(page.getByText("Model preview")).toBeVisible();
    await expect(
      page.getByRole("button", { name: /Test and Add data|Add data/i }),
    ).toBeVisible();
  });

  test("GCS connector - Save button hidden after advancing to model step", async ({
    page,
  }) => {
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

    await page.locator("#gcs").click();
    await page.waitForSelector('form[id*="gcs"]');

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

  test("GCS connector - model form resets after first submission (HMAC)", async ({
    page,
  }) => {
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
      await page.locator("#gcs").click();
      await page.waitForSelector('form[id*="gcs"]');
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

    const dialog = page.getByRole("dialog");
    await dialog
      .waitFor({ state: "detached", timeout: 10000 })
      .catch(async () => {
        // If the modal is still open (e.g., reconciliation failed), close it and continue.
        await page.keyboard.press("Escape");
        await dialog.waitFor({ state: "detached", timeout: 5000 });
      });

    // Re-open and ensure model form is reset
    await openGcsFlowWithHmac();
    await expect(
      page.getByRole("textbox", { name: "GCS URI" }),
    ).not.toHaveValue(firstPath);
    await expect(
      page.getByRole("textbox", { name: "Model name" }),
    ).not.toHaveValue(firstModelName);
  });

  test("GCS connector - model YAML includes create_secrets_from_connectors", async ({
    page,
  }) => {
    const hmacKey = process.env.RILL_RUNTIME_GCS_TEST_HMAC_KEY;
    const hmacSecret = process.env.RILL_RUNTIME_GCS_TEST_HMAC_SECRET;
    if (!hmacKey || !hmacSecret) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_HMAC_KEY or RILL_RUNTIME_GCS_TEST_HMAC_SECRET is not set",
      );
    }

    const startGcsConnector = async () => {
      await page.getByRole("button", { name: "Add Asset" }).click();
      await page.getByRole("menuitem", { name: "Add Data" }).click();
      await page.locator("#gcs").click();
      await page.waitForSelector('form[id*="gcs"]');
      await page.getByRole("radio", { name: "HMAC keys" }).click();
      await page.getByRole("textbox", { name: "Access Key ID" }).fill(hmacKey!);
      await page
        .getByRole("textbox", { name: "Secret Access Key" })
        .fill(hmacSecret!);
      await page
        .getByRole("dialog")
        .getByRole("button", { name: "Test and Connect" })
        .click();
      await expect(page.getByText("Model preview")).toBeVisible();
    };

    // Create first connector instance, then close modal
    await startGcsConnector();
    await page.keyboard.press("Escape");
    await page.getByRole("dialog").waitFor({ state: "detached" });

    // Create second connector instance and proceed to model import
    await startGcsConnector();
    const modelName = "gcs_create_secrets_test";
    await page
      .getByRole("textbox", { name: "GCS URI" })
      .fill(
        "gs://rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_06.parquet",
      );
    await page.getByRole("textbox", { name: "Model name" }).fill(modelName);
    await page.getByRole("button", { name: "Import Data" }).click();

    // Wait for navigation to the new model file
    await page.waitForURL(`**/files/models/${modelName}.yaml`);

    // Verify YAML contains the connector reference and create_secrets_from_connectors with the second instance (gcs_1)
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("connector: duckdb");
    await expect(codeEditor).toContainText(
      /create_secrets_from_connectors:\s*\[gcs_1\]/,
    );
  });
});

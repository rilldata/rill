import { expect } from "@playwright/test";
import { test } from "../setup/base";
import * as path from "path";
import { fileURLToPath } from "url";
import { writeFileSync, unlinkSync, existsSync } from "fs";
import { updateCodeEditor } from "../utils/commonHelpers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

test.describe("BigQuery connector", () => {
  test.use({ project: "Blank" });

  // Get BigQuery credentials from environment variable
  const getCredentialsFromEnv = () => {
    const credentialsJson =
      process.env.RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON;
    if (!credentialsJson) {
      throw new Error(
        "RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON environment variable is required",
      );
    }
    return JSON.parse(credentialsJson);
  };

  test("Create BigQuery connector with JSON credentials upload", async ({
    page,
  }) => {
    // Skip test if environment variable is not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON environment variable is not set",
      );
    }

    // Get credentials from environment variable
    const credentials = getCredentialsFromEnv();

    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select BigQuery connector
    await page.locator("#bigquery").click();

    // Wait for the form to load
    await page.waitForSelector('form[id*="bigquery"]');

    // Upload credentials JSON file
    const credentialsJson = JSON.stringify(credentials, null, 2);

    // Set up file chooser with proper error handling
    const [fileChooser] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByRole("button", { name: "Choose file" }).click(),
    ]);

    // Create a temporary file for upload
    const tempFilePath = path.join(__dirname, "temp-credentials.json");

    try {
      writeFileSync(tempFilePath, credentialsJson);

      // Upload the file using FileChooser
      await fileChooser.setFiles([tempFilePath]);

      // Wait a moment for the file to be processed
      await page.waitForTimeout(500);
    } finally {
      // Clean up temp file
      if (existsSync(tempFilePath)) {
        unlinkSync(tempFilePath);
      }
    }

    // Verify that project_id was automatically extracted and filled
    await expect(page.getByRole("textbox", { name: "Project ID" })).toHaveValue(
      credentials.project_id,
    );

    // Verify that the file was uploaded successfully
    // The CredentialsInput should show the filename after upload
    await expect(page.getByText("temp-credentials.json")).toBeVisible();

    // Submit the form (credentials are already uploaded)
    await page.getByRole("button", { name: "Test and Connect" }).click();

    // Wait for navigation to the new connector file
    await page.waitForURL(`**/files/connectors/bigquery.yaml`);

    // Assert that the file contains key properties with new ALL_CAPS env naming
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");

    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: bigquery");
    await expect(codeEditor).toContainText(
      `project_id: "${credentials.project_id}"`,
    );
    // New ALL_CAPS env variable naming (generic property without driver prefix)
    await expect(codeEditor).toContainText(
      'google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}"',
    );

    // Go to the `.env` file and verify the credentials are stored with new naming
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText("GOOGLE_APPLICATION_CREDENTIALS=");

    // Verify the credentials JSON is properly stored in .env
    const envContent = await envEditor.textContent();
    expect(envContent).toContain(JSON.stringify(credentials));
  });

  test("Duplicate BigQuery connectors get _# suffix for env variables", async ({
    page,
  }) => {
    // Skip test if environment variable is not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON environment variable is not set",
      );
    }

    const credentials = getCredentialsFromEnv();
    const tempFilePath = path.join(__dirname, "temp-credentials.json");
    const credentialsJson = JSON.stringify(credentials, null, 2);

    // Create first BigQuery connector
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#bigquery").click();
    await page.waitForSelector('form[id*="bigquery"]');

    const [fileChooser1] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByRole("button", { name: "Choose file" }).click(),
    ]);

    try {
      writeFileSync(tempFilePath, credentialsJson);
      await fileChooser1.setFiles([tempFilePath]);
      await page.waitForTimeout(500);
    } finally {
      if (existsSync(tempFilePath)) {
        unlinkSync(tempFilePath);
      }
    }

    await page.getByRole("button", { name: "Test and Connect" }).click();
    await page.waitForURL(`**/files/connectors/bigquery.yaml`);

    // Create second BigQuery connector
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#bigquery").click();
    await page.waitForSelector('form[id*="bigquery"]');

    const [fileChooser2] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByRole("button", { name: "Choose file" }).click(),
    ]);

    try {
      writeFileSync(tempFilePath, credentialsJson);
      await fileChooser2.setFiles([tempFilePath]);
      await page.waitForTimeout(500);
    } finally {
      if (existsSync(tempFilePath)) {
        unlinkSync(tempFilePath);
      }
    }

    await page.getByRole("button", { name: "Test and Connect" }).click();
    // Second connector should get _1 suffix in filename
    await page.waitForURL(`**/files/connectors/bigquery_1.yaml`);

    // Second connector should use _1 suffix for env variable
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText(
      'google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS_1 }}"',
    );

    // Verify .env has both variables
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText("GOOGLE_APPLICATION_CREDENTIALS=");
    await expect(envEditor).toContainText("GOOGLE_APPLICATION_CREDENTIALS_1=");
  });

  test("Save Anyway creates connector and navigates to file", async ({
    page,
  }) => {
    // Open the Add Data modal
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();

    // Select BigQuery connector
    await page.locator("#bigquery").click();
    await page.waitForSelector('form[id*="bigquery"]');

    // Fill with invalid/fake credentials to trigger Save Anyway
    await page
      .getByRole("textbox", { name: "Project ID" })
      .fill("fake-project-id");

    // Manually enter credentials text instead of file upload
    const fakeCredentials = JSON.stringify({
      type: "service_account",
      project_id: "fake-project-id",
      private_key_id: "fake-key-id",
      private_key:
        "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----\n",
      client_email: "fake@fake-project-id.iam.gserviceaccount.com",
      client_id: "123456789",
    });

    // Create temp file for fake credentials
    const tempFilePath = path.join(__dirname, "fake-credentials.json");
    try {
      writeFileSync(tempFilePath, fakeCredentials);

      const [fileChooser] = await Promise.all([
        page.waitForEvent("filechooser"),
        page.getByRole("button", { name: "Choose file" }).click(),
      ]);
      await fileChooser.setFiles([tempFilePath]);
      await page.waitForTimeout(500);
    } finally {
      if (existsSync(tempFilePath)) {
        unlinkSync(tempFilePath);
      }
    }

    // Click Test and Connect - should fail with invalid credentials
    await page.getByRole("button", { name: "Test and Connect" }).click();

    // Wait for Save Anyway button to appear
    await expect(page.getByRole("button", { name: "Save Anyway" })).toBeVisible(
      { timeout: 10000 },
    );

    // Click Save Anyway
    await page.getByRole("button", { name: "Save Anyway" }).click();

    // Should navigate to connector file
    await expect(page).toHaveURL(/.*\/files\/connectors\/bigquery.*\.yaml/, {
      timeout: 5000,
    });

    // Verify connector file was created with proper structure
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");
    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: bigquery");
    await expect(codeEditor).toContainText(
      'google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS',
    );
  });

  test("Case insensitive env variable resolution with {{ env }}", async ({
    page,
  }) => {
    // Skip test if environment variable is not set
    if (
      !process.env.RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON
    ) {
      test.skip(
        true,
        "RILL_RUNTIME_GCS_TEST_GOOGLE_APPLICATION_CREDENTIALS_JSON environment variable is not set",
      );
    }

    const credentials = getCredentialsFromEnv();
    const tempFilePath = path.join(__dirname, "temp-credentials.json");
    const credentialsJson = JSON.stringify(credentials, null, 2);

    // Create BigQuery connector first to set up env variables
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#bigquery").click();
    await page.waitForSelector('form[id*="bigquery"]');

    const [fileChooser] = await Promise.all([
      page.waitForEvent("filechooser"),
      page.getByRole("button", { name: "Choose file" }).click(),
    ]);

    try {
      writeFileSync(tempFilePath, credentialsJson);
      await fileChooser.setFiles([tempFilePath]);
      await page.waitForTimeout(500);
    } finally {
      if (existsSync(tempFilePath)) {
        unlinkSync(tempFilePath);
      }
    }

    await page.getByRole("button", { name: "Test and Connect" }).click();
    await page.waitForURL(`**/files/connectors/bigquery.yaml`);

    // Add a custom env variable to .env for testing case insensitivity
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toBeVisible();

    // Get current content and add test variable
    const currentContent = await envEditor.textContent();
    const newEnvContent = `${currentContent}\nMY_TEST_TABLE=test_table_name`;
    await updateCodeEditor(page, newEnvContent);
    await page.getByRole("button", { name: "Save" }).click();
    await page.waitForTimeout(500);

    // Create a model that uses case-insensitive env lookup
    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Model" }).click();

    // Wait for model to be created
    await page.waitForURL(/.*\/files\/models\/.*\.sql/);

    // Update model to use case-insensitive env variable
    // Using mixed case: my_TEST_table should resolve to MY_TEST_TABLE
    const modelContent = `-- Test case insensitive env resolution
-- @connector: duckdb
SELECT '{{ env "my_TEST_table" }}' as table_name`;
    await updateCodeEditor(page, modelContent);
    await page.getByRole("button", { name: "Save" }).click();

    // Wait for model to reconcile - should NOT show error about missing env variable
    await page.waitForTimeout(1000);

    // Check that there's no error about missing environment variable
    // If case insensitive works, the template should resolve correctly
    const errorPane = page.locator(".editor-pane .error");
    const errorCount = await errorPane.count();

    // If there's an error, it should NOT be about missing env variable
    if (errorCount > 0) {
      const errorText = await errorPane.textContent();
      expect(errorText).not.toContain("my_TEST_table");
      expect(errorText).not.toContain("environment variable");
    }
  });
});

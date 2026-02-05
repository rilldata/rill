import { expect } from "@playwright/test";
import { test } from "../setup/base";
import * as path from "path";
import { fileURLToPath } from "url";
import { writeFileSync, unlinkSync, existsSync } from "fs";

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

    // Assert that the file contains key properties
    const codeEditor = page
      .getByLabel("codemirror editor")
      .getByRole("textbox");

    await expect(codeEditor).toContainText("type: connector");
    await expect(codeEditor).toContainText("driver: bigquery");
    await expect(codeEditor).toContainText(
      `project_id: "${credentials.project_id}"`,
    );
    await expect(codeEditor).toContainText(
      'google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"',
    );

    // Go to the `.env` file and verify the credentials are stored
    await page.getByRole("link", { name: ".env" }).click();
    const envEditor = page.getByLabel("codemirror editor").getByRole("textbox");
    await expect(envEditor).toContainText(
      "connector.bigquery.google_application_credentials=",
    );

    // Verify the credentials JSON is properly stored in .env
    const envContent = await envEditor.textContent();
    expect(envContent).toContain(JSON.stringify(credentials));
  });
});

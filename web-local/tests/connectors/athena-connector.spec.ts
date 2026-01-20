import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Athena connector", () => {
  test.use({ project: "Blank" });

  test("explorer step keeps SQL and Model name empty", async ({ page }) => {
    const accessKey = process.env.RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID;
    const secretKey =
      process.env.RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY;
    const outputLocation = process.env.RILL_RUNTIME_ATHENA_TEST_OUTPUT_LOCATION;

    if (!accessKey || !secretKey || !outputLocation) {
      test.skip(
        true,
        "RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID, RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY, or RILL_RUNTIME_ATHENA_TEST_OUTPUT_LOCATION is not set",
      );
    }

    await page.getByRole("button", { name: "Add Asset" }).click();
    await page.getByRole("menuitem", { name: "Add Data" }).click();
    await page.locator("#athena").click();
    await page.waitForSelector('form[id*="athena"]');

    await page
      .getByRole("textbox", { name: "AWS access key ID" })
      .fill(accessKey!);
    await page
      .getByRole("textbox", { name: "AWS secret access key" })
      .fill(secretKey!);
    await page
      .getByRole("textbox", { name: "S3 output location" })
      .fill(outputLocation!);

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();

    await expect(page.getByText("Model preview")).toBeVisible({
      timeout: 120000,
    });

    const sqlField = page.getByRole("textbox", { name: "SQL" });
    await expect(sqlField).toBeVisible();
    await expect(sqlField).toHaveValue("");

    const nameField = page.getByRole("textbox", { name: "Model name" });
    await expect(nameField).toBeVisible();
    await expect(nameField).toHaveValue("");
  });
});

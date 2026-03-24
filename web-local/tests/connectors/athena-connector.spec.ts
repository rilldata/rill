import { expect } from "@playwright/test";
import { test } from "../setup/base";

test.describe("Athena connector", () => {
  test.use({ project: "Blank" });

  test("explorer step keeps SQL and Model name empty", async ({ page }) => {
    const accessKey = process.env.RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID;
    const secretKey =
      process.env.RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY;
    const outputLocation = "s3://integration-test.rilldata.com/athena/";

    if (!accessKey || !secretKey) {
      test.skip(
        true,
        "RILL_RUNTIME_ATHENA_TEST_AWS_ACCESS_KEY_ID or RILL_RUNTIME_ATHENA_TEST_AWS_SECRET_ACCESS_KEY is not set",
      );
    }

    await page.getByLabel("Connect your data").click();
    await page.getByLabel("Connect to athena").click();

    await page
      .getByRole("textbox", { name: "AWS access key ID" })
      .fill(accessKey!);
    await page
      .getByRole("textbox", { name: "AWS secret access key" })
      .fill(secretKey!);
    await page
      .getByRole("textbox", { name: "S3 output location" })
      .fill(outputLocation);

    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Test and Connect" })
      .click();

    await expect(
      page.getByText(
        "Pick a table or input your file SQL to power your first dashboard",
      ),
    ).toBeVisible({
      timeout: 120000,
    });

    // Aws data catalog is visible in explorer
    const awsDataCatalogNode = page.getByLabel("Node: awsdatacatalog, level 0");
    await expect(awsDataCatalogNode).toBeVisible();
    await awsDataCatalogNode.click();

    await expect(page.getByLabel("Node: default, level 1")).toBeVisible();
    await expect(
      page.getByLabel("Node: integration_test, level 1"),
    ).toBeVisible();
  });
});

import { expect } from "@playwright/test";
import { updateCodeEditor, wrapRetryAssertion } from "../utils/commonHelpers";
import { createDashboardFromModel } from "../utils/dashboardHelpers";
import { createAdBidsModel } from "../utils/dataSpecifcHelpers";
import { test } from "../utils/test";

test.describe("Metrics editor", () => {
  test.beforeEach(async ({ page }) => {
    test.setTimeout(60000);
    // disable animations
    await page.addStyleTag({
      content: `
        *, *::before, *::after {
          animation-duration: 0s !important;
          transition-duration: 0s !important;
        }
      `,
    });
    await createAdBidsModel(page);
    await createDashboardFromModel(page, "AdBids_model");

    // Open Edit Metrics
    await page.getByRole("button", { name: "Edit Metrics" }).click();

    // Close the navigation sidebar to give the code editor more space
    await page.getByRole("button", { name: "Close sidebar" }).click();
  });

  test("Metrics editor", async ({ page }) => {
    await updateCodeEditor(page, "");

    // the inspector should be empty.
    await expect(await page.getByText("Let's get started.")).toBeVisible();

    // skeleton should result in an empty skeleton YAML file
    await page.getByText("start with a skeleton").click();

    // check to see that the placeholder is gone by looking for the button
    // that was once there.
    await wrapRetryAssertion(async () => {
      await expect(await page.getByText("start with a skeleton")).toBeHidden();
    });

    // the  button should be disabled.
    await expect(
      await page.getByRole("button", { name: "Preview" }),
    ).toBeDisabled();

    // the inspector should be empty.
    await expect(await page.getByText("Table not defined.")).toBeVisible();

    // now let's scaffold things in
    await updateCodeEditor(page, "");

    await wrapRetryAssertion(async () => {
      await expect(
        await page.getByText("metrics configuration from an existing model"),
      ).toBeVisible();
    });

    // select the first menu item.
    await page
      .getByText("metrics configuration from an existing model")
      .click();
    await page.getByRole("menuitem").getByText("AdBids_model").click();
    // Wait for menu to close
    await expect(
      page.getByRole("menuitem", { name: "Last 6 Hours" }),
    ).not.toBeVisible();

    // let's check the inspector.
    await expect(await page.getByText("Model summary")).toBeVisible();
    await expect(await page.getByText("Model columns")).toBeVisible();

    // go to teh dashboard and make sure the metrics and dimensions are there.

    await page.getByRole("button", { name: "Preview" }).click();

    // check to see metrics make sense.
    await expect(await page.getByText("Total Records 100.0k")).toBeVisible();

    // double-check that leaderboards make sense.
    await expect(
      await page.getByRole("button", { name: "google.com 15.1k" }),
    ).toBeVisible();

    // go back to the metrics page.
    await page.getByRole("button", { name: "Edit metrics" }).click();
  });
});

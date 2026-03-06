import { expect } from "@playwright/test";
import { test } from "../setup/base";
import { updateCodeEditor, waitForProfiling } from "../utils/commonHelpers";
import { createModel } from "../utils/modelHelpers";
import { waitForFileNavEntry } from "../utils/waitHelpers";
test.describe("Incremental models", () => {
  test.use({ project: "Blank" });

  test("partitions browser should display model partitions", async ({
    page,
  }) => {
    // Create a partitioned model
    await createModel(page, "partitioned_model.yaml");
    await waitForFileNavEntry(page, "/models/partitioned_model.yaml", true);
    await updateCodeEditor(
      page,
      `
# This model produces a range of numbers with the current timestamp.
# It is not incremental, which means:
#  - All rows will be replaced on each refresh
#  - You cannot refresh a single partition
type: model
refresh:
  cron: 0 0 * * *
partitions:
  sql: SELECT range AS num FROM range(0,10)
sql: >
  SELECT
    {{ .partition.num }} AS num,
    now() AS inserted_on,
    CASE WHEN {{ .partition.num }} = 2 THEN error('simulated error') ELSE NULL END as error
`,
    );

    // Trigger save via keyboard shortcut — works regardless of auto-save state.
    // (After a rename from .sql to .yaml the editor may keep auto-save enabled,
    // so the "Save" button is not guaranteed to exist.)
    if (process.platform === "darwin") {
      await page.keyboard.press("Meta+s");
    } else {
      await page.keyboard.press("Control+s");
    }

    await waitForProfiling(page, "partitioned_model", ["inserted_on", "num"]);

    // Open the partitions browser
    await page.getByRole("button", { name: "View partitions" }).click();

    // Check that the partitions are displayed
    await expect(page.getByText("num: 0")).toBeVisible();
    await expect(page.getByText("num: 1")).toBeVisible();

    // Filter for the errored partitions
    await page.getByRole("button", { name: "Filter partitions" }).click();
    await page.getByRole("option", { name: "errors" }).click();

    // Check that the errored partition is displayed
    const errorText = page.getByText("failed to incrementally");
    await expect(errorText).toBeVisible();
    await expect(errorText).toContainText("simulated error");
  });
});

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
sql: SELECT {{ .partition.num }} AS num, now() AS inserted_on
`,
    );
    await waitForProfiling(page, "partitioned_model", ["inserted_on", "num"]);

    // Check that the partitions browser displays the model's partitions
    await page.getByRole("button", { name: "View partitions" }).click();
    await expect(page.getByText("num: 0")).toBeVisible();
    await expect(page.getByText("num: 1")).toBeVisible();
  });
});

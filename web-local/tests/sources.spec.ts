import { expect } from "@playwright/test";
import {
  deleteFile,
  renameFileUsingMenu,
  updateCodeEditor,
} from "./utils/commonHelpers";
import {
  TestDataPath,
  createSource,
  uploadFile,
  waitForSource,
} from "./utils/sourceHelpers";
import { test } from "./setup/base";
import { fileNotPresent, waitForFileNavEntry } from "./utils/waitHelpers";

test.describe("sources", () => {
  test.use({ project: "Blank" });

  test("Import sources", async ({ page }) => {
    await Promise.all([
      waitForSource(page, "/models/AdBids.yaml", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      uploadFile(page, "AdBids.csv"),
    ]);

    await Promise.all([
      waitForSource(page, "/models/AdImpressions.yaml", ["city", "country"]),
      uploadFile(page, "AdImpressions.tsv"),
    ]);

    // upload existing table and keep both
    await Promise.all([
      waitForSource(page, "/models/AdBids_1.yaml", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      uploadFile(page, "AdBids.csv", true, true),
    ]);

    // upload existing table and replace
    await uploadFile(page, "AdBids.csv", true, false);
    await fileNotPresent(page, "AdBids_2");
  });

  test("Rename and delete sources", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/models/AdBids.yaml");

    // rename
    await renameFileUsingMenu(page, "/models/AdBids.yaml", "AdBids_new.yaml");
    await waitForFileNavEntry(page, `/models/AdBids_new.yaml`, true);
    await fileNotPresent(page, "/models/AdBids.yaml");

    // delete
    await deleteFile(page, "/models/AdBids_new.yaml");
    await fileNotPresent(page, "/models/AdBids_new");
    await fileNotPresent(page, "/models/AdBids");
  });

  test("Edit source", async ({ page }) => {
    // Upload data & create two sources
    await createSource(page, "AdImpressions.tsv", "/models/AdImpressions.yaml");
    await createSource(page, "AdBids.csv", "/models/AdBids.yaml");

    // Edit source path to a non-existent file
    const nonExistentSource = `connector: local_file
path: ${TestDataPath}/non_existent_file.csv`;
    await updateCodeEditor(page, nonExistentSource);
    await page.getByRole("button", { name: "Save" }).click();

    // Observe error message "file does not exist"
    await expect(page.getByText("file does not exist")).toBeVisible();

    // Edit source path to an existent file
    const adImpressionsSource = `connector: local_file
path: ${TestDataPath}/AdImpressions.tsv`;

    await updateCodeEditor(page, adImpressionsSource);
    await page.getByRole("button", { name: "Save" }).click();

    // Check that the source data is updated
    // (The column "user_id" exists in AdImpressions, but not in AdBids)
    await expect(
      page.getByRole("button").filter({ hasText: "user_id" }).first(),
    ).toBeVisible();
  });
});

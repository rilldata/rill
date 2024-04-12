import { expect } from "@playwright/test";
import {
  deleteEntity,
  renameEntityUsingMenu,
  updateCodeEditor,
} from "./utils/commonHelpers";
import {
  TestDataPath,
  createOrReplaceSource,
  uploadFile,
  waitForSource,
} from "./utils/sourceHelpers";
import { entityNotPresent, waitForFileEntry } from "./utils/waitHelpers";
import { test } from "./utils/test";

test.describe("sources", () => {
  test("Import sources", async ({ page }) => {
    await Promise.all([
      waitForSource(page, "sources/AdBids.yaml", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      uploadFile(page, "AdBids.csv"),
    ]);

    await Promise.all([
      waitForSource(page, "sources/AdImpressions.yaml", ["city", "country"]),
      uploadFile(page, "AdImpressions.tsv"),
    ]);

    // upload existing table and keep both
    await Promise.all([
      waitForSource(page, "sources/AdBids_1.yaml", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      uploadFile(page, "AdBids.csv", true, true),
    ]);

    // upload existing table and replace
    await uploadFile(page, "AdBids.csv", true, false);
    await entityNotPresent(page, "AdBids_2");
  });

  test("Rename and delete sources", async ({ page }) => {
    await createOrReplaceSource(page, "AdBids.csv", "sources/AdBids.yaml");

    // rename
    await renameEntityUsingMenu(page, "AdBids.yaml", "AdBids_new.yaml");
    await waitForFileEntry(page, `sources/AdBids_new.yaml`, true);
    await entityNotPresent(page, "AdBids.yaml");

    // delete
    await deleteEntity(page, "AdBids_new.yaml");
    await entityNotPresent(page, "AdBids_new");
    await entityNotPresent(page, "AdBids");
  });

  test("Edit source", async ({ page }) => {
    // Upload data & create two sources
    await createOrReplaceSource(
      page,
      "AdImpressions.tsv",
      "sources/AdImpressions.yaml",
    );
    await createOrReplaceSource(page, "AdBids.csv", "sources/AdBids.yaml");

    // Edit source path to a non-existent file
    const nonExistentSource = `type: local_file
path: ${TestDataPath}/non_existent_file.csv`;
    await updateCodeEditor(page, nonExistentSource);
    await page.getByRole("button", { name: "Save and refresh" }).click();

    // Observe error message "file does not exist"
    await expect(page.getByText("file does not exist")).toBeVisible();

    // Edit source path to an existent file
    const adImpressionsSource = `type: local_file
path: ${TestDataPath}/AdImpressions.tsv`;
    await updateCodeEditor(page, adImpressionsSource);
    await page.getByRole("button", { name: "Save and refresh" }).click();

    // Check that the source data is updated
    // (The column "user_id" exists in AdImpressions, but not in AdBids)
    await expect(
      page.getByRole("button").filter({ hasText: "user_id" }).first(),
    ).toBeVisible();
  });
});

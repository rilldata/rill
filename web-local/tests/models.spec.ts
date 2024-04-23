import {
  clickMenuButton,
  deleteFile,
  goToFile,
  openFileNavEntryContextMenu,
  renameFileUsingMenu,
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./utils/commonHelpers";
import { createModel, modelHasError } from "./utils/modelHelpers";
import { createSource } from "./utils/sourceHelpers";
import { test } from "./utils/test";
import { fileNotPresent, waitForFileNavEntry } from "./utils/waitHelpers";

test.describe("models", () => {
  test("Create and edit model", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/sources/AdBids.yaml");
    await createSource(
      page,
      "AdImpressions.tsv",
      "/sources/AdImpressions.yaml",
    );

    await createModel(page, "AdBids_model_t.sql");
    await waitForFileNavEntry(page, "/models/AdBids_model_t.sql", true);
    await Promise.all([
      waitForProfiling(page, "AdBids_model_t", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      updateCodeEditor(page, "select * from AdBids"),
    ]);
    await wrapRetryAssertion(() => modelHasError(page, false));

    // Catalog error
    await updateCodeEditor(page, "select * from AdBid");
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));

    // Query parse error
    await updateCodeEditor(page, "select from AdBids");
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));
  });

  test("Rename and delete model", async ({ page }) => {
    // make sure AdBids_rename_delete is present
    await createModel(page, "AdBids_rename_delete.sql");

    // rename
    await renameFileUsingMenu(
      page,
      "/models/AdBids_rename_delete.sql",
      "AdBids_rename_delete_new.sql",
    );
    await waitForFileNavEntry(
      page,
      "/models/AdBids_rename_delete_new.sql",
      true,
    );
    await fileNotPresent(page, "/models/AdBids_rename_delete.sql");

    // delete
    await deleteFile(page, "/models/AdBids_rename_delete_new.sql");
    await fileNotPresent(page, "/models/AdBids_rename_delete_new.sql");
    await fileNotPresent(page, "/models/AdBids_rename_delete.sql");
  });

  test("Create model from source", async ({ page }) => {
    await createSource(page, "AdBids.csv", "/sources/AdBids.yaml");

    await Promise.all([
      waitForProfiling(page, "AdBids_model", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      openFileNavEntryContextMenu(page, "/sources/AdBids.yaml"),
      clickMenuButton(page, "Create New Model"),
    ]);
    await waitForFileNavEntry(page, "/models/AdBids_model.sql", true);

    // navigate to another source
    await createSource(
      page,
      "AdImpressions.tsv",
      "/sources/AdImpressions.yaml",
    );
    // delete the source of model
    await deleteFile(page, "/sources/AdBids.yaml");
    // go to model
    await goToFile(page, "/models/AdBids_model.sql");
    // make sure error has propagated
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));
  });
});

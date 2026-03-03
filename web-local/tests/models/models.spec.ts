import { test } from "../setup/base";
import {
  deleteFile,
  renameFileUsingMenu,
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "../utils/commonHelpers";
import { createModel, modelHasError } from "../utils/modelHelpers";
import { createSource } from "../utils/sourceHelpers";
import { fileNotPresent, waitForFileNavEntry } from "../utils/waitHelpers";

test.describe("models", () => {
  test.use({ project: "Blank" });

  test("Create and edit model", async ({ page }) => {
    // Add the AdBids source
    await createSource(page, "AdBids.csv", "/models/AdBids.yaml");

    // Create a "Hello world" model named AdBids_model.sql
    await createModel(page, "AdBids_model.sql");
    await waitForFileNavEntry(page, "/models/AdBids_model.sql", true);

    // Edit the model to select all columns from the AdBids source
    await updateCodeEditor(page, "select * from AdBids");
    await waitForProfiling(page, "AdBids_model", [
      "publisher",
      "domain",
      "timestamp",
    ]);
    await wrapRetryAssertion(() => modelHasError(page, false));

    // Break the model to see a catalog error
    await updateCodeEditor(page, "select * from AdBid");
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));

    // Break the model to see a query parse error
    await updateCodeEditor(page, "select from AdBids");
    await wrapRetryAssertion(() =>
      modelHasError(page, true, "SELECT clause without selection list"),
    );
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
});

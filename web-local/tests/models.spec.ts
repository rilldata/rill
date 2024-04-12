import {
  clickMenuButton,
  deleteEntity,
  gotoEntity,
  openEntityMenu,
  renameEntityUsingMenu,
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./utils/commonHelpers";
import { createModel, modelHasError } from "./utils/modelHelpers";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { entityNotPresent, waitForFileEntry } from "./utils/waitHelpers";
import { test } from "./utils/test";

test.describe("models", () => {
  test("Create and edit model", async ({ page }) => {
    await createOrReplaceSource(page, "AdBids.csv", "sources/AdBids.yaml");
    await createOrReplaceSource(
      page,
      "AdImpressions.tsv",
      "sources/AdImpressions.yaml",
    );

    await createModel(page, "models/AdBids_model_t.sql");
    await waitForFileEntry(page, "models/AdBids_model_t.sql", true);
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
    await createModel(page, "models/AdBids_rename_delete.sql");

    // rename
    await renameEntityUsingMenu(
      page,
      "AdBids_rename_delete.sql",
      "AdBids_rename_delete_new.sql",
    );
    await waitForFileEntry(page, "models/AdBids_rename_delete_new.sql", true);
    await entityNotPresent(page, "AdBids_rename_delete");

    // delete
    await deleteEntity(page, "AdBids_rename_delete_new.sql");
    await entityNotPresent(page, "AdBids_rename_delete_new");
    await entityNotPresent(page, "AdBids_rename_delete");
  });

  test("Create model from source", async ({ page }) => {
    await createOrReplaceSource(page, "AdBids.csv", "sources/AdBids.yaml");

    await Promise.all([
      waitForProfiling(page, "AdBids_model", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      openEntityMenu(page, "AdBids.yaml"),
      clickMenuButton(page, "Create New Model"),
    ]);
    await waitForFileEntry(page, "sources/AdBids_model.sql", true);

    // navigate to another source
    await createOrReplaceSource(
      page,
      "AdImpressions.tsv",
      "sources/AdImpressions.yaml",
    );
    // delete the source of model
    await deleteEntity(page, "AdBids.yaml");
    // go to model
    await gotoEntity(page, "AdBids_model.sql");
    // make sure error has propagated
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));
  });
});

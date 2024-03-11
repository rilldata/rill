import {
  TestEntityType,
  deleteEntity,
  gotoEntity,
  renameEntityUsingMenu,
  updateCodeEditor,
  waitForProfiling,
  wrapRetryAssertion,
} from "./utils/commonHelpers";
import {
  createModel,
  createModelFromSource,
  modelHasError,
} from "./utils/modelHelpers";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";
import { test } from "./utils/test";

test.describe("models", () => {
  test("Create and edit model", async ({ page }) => {
    await createOrReplaceSource(page, "AdBids.csv", "AdBids");
    await createOrReplaceSource(page, "AdImpressions.tsv", "AdImpressions");

    await createModel(page, "AdBids_model_t");
    await waitForEntity(page, TestEntityType.Model, "AdBids_model_t", true);
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
    await createModel(page, "AdBids_rename_delete");

    // rename
    await renameEntityUsingMenu(
      page,
      "AdBids_rename_delete",
      "AdBids_rename_delete_new",
    );
    await waitForEntity(
      page,
      TestEntityType.Model,
      "AdBids_rename_delete_new",
      true,
    );
    await entityNotPresent(page, "AdBids_rename_delete");

    // delete
    await deleteEntity(page, "AdBids_rename_delete_new");
    await entityNotPresent(page, "AdBids_rename_delete_new");
    await entityNotPresent(page, "AdBids_rename_delete");
  });

  test("Create model from source", async ({ page }) => {
    await createOrReplaceSource(page, "AdBids.csv", "AdBids");

    await Promise.all([
      waitForProfiling(page, "AdBids_model", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      createModelFromSource(page, "AdBids"),
    ]);
    await waitForEntity(page, TestEntityType.Model, "AdBids_model", true);

    // navigate to another source
    await createOrReplaceSource(page, "AdImpressions.tsv", "AdImpressions");
    // delete the source of model
    await deleteEntity(page, "AdBids");
    // go to model
    await gotoEntity(page, "AdBids_model");
    // make sure error has propagated
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));
  });
});

import { test } from "@playwright/test";
import { createAdBidsModel } from "./utils/dataSpecifcHelpers";
import {
  deleteEntity,
  gotoEntity,
  renameEntityUsingMenu,
} from "./utils/commonHelpers";
import {
  TestEntityType,
  waitForProfiling,
  wrapRetryAssertion,
} from "./utils/helpers";
import {
  createModel,
  createModelFromSource,
  modelHasError,
  updateModelSql,
} from "./utils/modelHelpers";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";
import { startRuntimeForEachTest } from "./utils/startRuntimeForEachTest";

test.describe("models", () => {
  startRuntimeForEachTest();

  test("Create and edit model", async ({ page }) => {
    await page.goto("/");

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
      updateModelSql(page, "select * from AdBids"),
    ]);
    await wrapRetryAssertion(() => modelHasError(page, false));

    // Catalog error
    await updateModelSql(page, "select * from AdBid");
    await wrapRetryAssertion(() => modelHasError(page, true, "Catalog Error"));

    // Query parse error
    await updateModelSql(page, "select from AdBids");
    await wrapRetryAssertion(() => modelHasError(page, true, "Parser Error"));
  });

  test("Rename and delete model", async ({ page }) => {
    await page.goto("/");

    // make sure AdBids_rename_delete is present
    await createModel(page, "AdBids_rename_delete");

    // rename
    await renameEntityUsingMenu(
      page,
      "AdBids_rename_delete",
      "AdBids_rename_delete_new"
    );
    await waitForEntity(
      page,
      TestEntityType.Model,
      "AdBids_rename_delete_new",
      true
    );
    await entityNotPresent(page, "AdBids_rename_delete");

    // delete
    await deleteEntity(page, "AdBids_rename_delete_new");
    await entityNotPresent(page, "AdBids_rename_delete_new");
    await entityNotPresent(page, "AdBids_rename_delete");
  });

  test("Create model from source", async ({ page }) => {
    await page.goto("/");

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

  test("Embedded source", async ({ page }) => {
    await page.goto("/");
    await createAdBidsModel(page);
  });
});

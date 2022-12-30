import { describe, it } from "@jest/globals";
import {
  deleteEntity,
  gotoEntity,
  renameEntityUsingMenu,
} from "./utils/commonHelpers";
import { TestEntityType, waitForProfiling } from "./utils/helpers";
import {
  createModel,
  createModelFromSource,
  modelHasError,
  updateModelSql,
} from "./utils/modelHelpers";
import { useRegisteredServer } from "./utils/serverConfigs";
import { createOrReplaceSource } from "./utils/sourceHelpers";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";

describe.skip("models", () => {
  const testBrowser = useRegisteredServer("models");

  it("Create and edit model", async () => {
    await createOrReplaceSource(testBrowser.page, "AdBids.csv", "AdBids");
    await createOrReplaceSource(
      testBrowser.page,
      "AdImpressions.tsv",
      "AdImpressions"
    );

    await createModel(testBrowser.page, "AdBids_model_t");
    await waitForEntity(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_model_t",
      true
    );
    await updateModelSql(testBrowser.page, "select * from AdBids");
    await modelHasError(testBrowser.page, false);

    // Catalog error
    await updateModelSql(testBrowser.page, "select * from AdBid");
    await modelHasError(testBrowser.page, true, "Catalog Error");

    // Query parse error
    await updateModelSql(testBrowser.page, "select from AdBids");
    await modelHasError(testBrowser.page, true, "Parser Error");
  });

  it("Rename and delete model", async () => {
    // make sure AdBids_rename_delete is present
    await createModel(testBrowser.page, "AdBids_rename_delete");

    // rename
    await renameEntityUsingMenu(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete",
      "AdBids_rename_delete_new"
    );
    await waitForEntity(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete_new",
      true
    );
    await entityNotPresent(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete"
    );

    // delete
    await deleteEntity(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete_new"
    );
    await entityNotPresent(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete_new"
    );
    await entityNotPresent(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_rename_delete"
    );
  });

  it("Create model from source", async () => {
    await createOrReplaceSource(testBrowser.page, "AdBids.csv", "AdBids");

    await Promise.all([
      waitForProfiling(testBrowser.page, "AdBids_model", [
        "publisher",
        "domain",
        "timestamp",
      ]),
      createModelFromSource(testBrowser.page, "AdBids"),
    ]);
    await waitForEntity(
      testBrowser.page,
      TestEntityType.Model,
      "AdBids_model",
      true
    );

    // navigate to another source
    await createOrReplaceSource(
      testBrowser.page,
      "AdImpressions.tsv",
      "AdImpressions"
    );
    // delete the source of model
    await deleteEntity(testBrowser.page, TestEntityType.Source, "AdBids");
    // go to model
    await gotoEntity(testBrowser.page, TestEntityType.Model, "AdBids_model");
    // make sure error has propagated
    await modelHasError(testBrowser.page, true, "Catalog Error");
  });
});

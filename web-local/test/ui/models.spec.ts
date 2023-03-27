import { describe, it } from "@jest/globals";
import { asyncWait } from "@rilldata/web-local/lib/util/waitUtils";
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

describe("models", () => {
  const testBrowser = useRegisteredServer("models");

  it("Create and edit model", async () => {
    const { page } = testBrowser;

    await createOrReplaceSource(page, "AdBids.csv", "AdBids");
    await createOrReplaceSource(page, "AdImpressions.tsv", "AdImpressions");

    await createModel(page, "AdBids_model_t");
    await waitForEntity(page, TestEntityType.Model, "AdBids_model_t", true);
    await updateModelSql(page, "select * from AdBids");
    await modelHasError(page, false);

    // Catalog error
    await updateModelSql(page, "select * from AdBid");
    await modelHasError(page, true, "Catalog Error");

    // Query parse error
    await updateModelSql(page, "select from AdBids");
    await modelHasError(page, true, "Parser Error");
  });

  it("Rename and delete model", async () => {
    const { page } = testBrowser;

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

  it("Create model from source", async () => {
    const { page } = testBrowser;

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
    await modelHasError(page, true, "Catalog Error");
  });
});

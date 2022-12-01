import { describe } from "@jest/globals";
import path from "node:path";
import { TestBrowser, TestEntityType } from "./TestBrowser";
import { useTestServer } from "./useTestServer";

const PORT = 8081;
const DataPath = path.join(__dirname, "../data");

// TODO: these tests cannot run in CI until cli supports custom ports for UI
describe.skip("models", () => {
  useTestServer(PORT, "temp/models");
  const testBrowser = TestBrowser.useTestBrowser(
    DataPath,
    `http://localhost:${PORT}`
  );

  it("Create and edit model", async () => {
    await testBrowser.createOrReplaceSource("AdBids.csv", "AdBids");
    await testBrowser.createOrReplaceSource(
      "AdImpressions.tsv",
      "AdImpressions"
    );

    await testBrowser.createModel("AdBids_model_t");
    await testBrowser.waitForEntity(
      TestEntityType.Model,
      "AdBids_model_t",
      true
    );
    await testBrowser.updateModelSql("select * from AdBids");
    await testBrowser.modelHasError(false);

    // Catalog error
    await testBrowser.updateModelSql("select * from AdBid");
    await testBrowser.modelHasError(true, "Catalog Error");

    // Query parse error
    await testBrowser.updateModelSql("select from AdBids");
    await testBrowser.modelHasError(true, "Parser Error");
  });

  it("Rename and delete model", async () => {
    // make sure AdBids_rename_delete is present
    await testBrowser.createModel("AdBids_rename_delete");

    // rename
    await testBrowser.renameEntityUsingMenu(
      TestEntityType.Model,
      "AdBids_rename_delete",
      "AdBids_rename_delete_new"
    );
    await testBrowser.waitForEntity(
      TestEntityType.Model,
      "AdBids_rename_delete_new",
      true
    );
    await testBrowser.entityNotPresent(
      TestEntityType.Model,
      "AdBids_rename_delete"
    );

    // delete
    await testBrowser.deleteEntity(
      TestEntityType.Model,
      "AdBids_rename_delete_new"
    );
    await testBrowser.entityNotPresent(
      TestEntityType.Model,
      "AdBids_rename_delete_new"
    );
    await testBrowser.entityNotPresent(
      TestEntityType.Model,
      "AdBids_rename_delete"
    );
  });

  it("Create model from source", async () => {
    await testBrowser.createOrReplaceSource("AdBids.csv", "AdBids");

    await testBrowser.createModelFromSource("AdBids");
    await testBrowser.waitForEntity(TestEntityType.Model, "AdBids_model", true);

    // navigate to another source
    await testBrowser.createOrReplaceSource(
      "AdImpressions.tsv",
      "AdImpressions"
    );
    // delete the source of model
    await testBrowser.deleteEntity(TestEntityType.Source, "AdBids");
    // go to model
    await testBrowser.gotoEntity(TestEntityType.Model, "AdBids_model");
    // make sure error has propagated
    await testBrowser.modelHasError(true, "Catalog Error");
  });
});

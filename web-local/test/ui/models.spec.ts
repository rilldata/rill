import { describe } from "@jest/globals";
import path from "node:path";
import { useInlineTestServer } from "../utils/useInlineTestServer";
import { TestBrowser } from "./TestBrowser";

const PORT = 8080;
const DataPath = path.join(__dirname, "../data");

// TODO: these tests cannot run in CI until cli supports custom ports for UI
describe.skip("models", () => {
  useInlineTestServer(PORT, "temp/models");
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

    await testBrowser.createModel("AdBids_model");
    await testBrowser.waitForEntity("model", "AdBids_model", true);
    await testBrowser.updateModelSql("select * from AdBids");
    await testBrowser.modelHasError(false);

    await testBrowser.updateModelSql("select * from AdBid");
    await testBrowser.modelHasError(true, "Catalog Error");

    await testBrowser.updateModelSql("select from AdBids");
    await testBrowser.modelHasError(true, "Parser Error");
  });

  it("Rename and delete model", async () => {
    // make sure AdBids_rename_delete is present
    await testBrowser.createModel("AdBids_rename_delete");

    // rename
    await testBrowser.renameEntityUsingMenu(
      "model",
      "AdBids_rename_delete",
      "AdBids_rename_delete_new"
    );
    await testBrowser.waitForEntity("model", "AdBids_rename_delete_new", true);
    await testBrowser.entityNotPresent("model", "AdBids_rename_delete");

    // delete
    await testBrowser.deleteEntity("model", "AdBids_rename_delete_new");
    await testBrowser.entityNotPresent("model", "AdBids_rename_delete_new");
    await testBrowser.entityNotPresent("model", "AdBids_rename_delete");
  });

  it("Create model from source", async () => {
    await testBrowser.createOrReplaceSource("AdBids.csv", "AdBids");

    await testBrowser.createModelFromSource("AdBids");
    await testBrowser.waitForEntity("model", "AdBids_model", true);
  });
});

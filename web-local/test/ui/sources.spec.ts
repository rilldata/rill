import { describe } from "@jest/globals";
import path from "node:path";
import { useInlineTestServer } from "../utils/useInlineTestServer";
import { TestBrowser, TestEntityType } from "./TestBrowser";

const PORT = 8080;
const DataPath = path.join(__dirname, "../data");

// TODO: these tests cannot run in CI until cli supports custom ports for UI
describe.skip("sources", () => {
  useInlineTestServer(PORT, "temp/sources");
  const testBrowser = TestBrowser.useTestBrowser(
    DataPath,
    `http://localhost:${PORT}`
  );

  it("Import sources", async () => {
    await testBrowser.uploadFile("AdBids.csv");
    await testBrowser.waitForEntity(TestEntityType.Source, "AdBids", true);

    await testBrowser.uploadFile("AdImpressions.tsv");
    await testBrowser.waitForEntity(
      TestEntityType.Source,
      "AdImpressions",
      true
    );

    // upload existing table and keep both
    await testBrowser.uploadFile("AdBids.csv", true, true);
    await testBrowser.waitForEntity(TestEntityType.Source, "AdBids", false);
    await testBrowser.waitForEntity(TestEntityType.Source, "AdBids_1", true);

    // upload existing table and replace
    await testBrowser.uploadFile("AdBids.csv", true, false);
    await testBrowser.waitForEntity(TestEntityType.Source, "AdBids", true);
    await testBrowser.entityNotPresent(TestEntityType.Source, "AdBids_2");
  });

  it("Rename and delete sources", async () => {
    // make sure AdBids is present
    await testBrowser.createOrReplaceSource("AdBids.csv", "AdBids");

    // rename
    await testBrowser.renameEntityUsingMenu(
      TestEntityType.Source,
      "AdBids",
      "AdBids_new"
    );
    await testBrowser.waitForEntity(TestEntityType.Source, "AdBids_new", true);
    await testBrowser.entityNotPresent(TestEntityType.Source, "AdBids");

    // delete
    await testBrowser.deleteEntity(TestEntityType.Source, "AdBids_new");
    await testBrowser.entityNotPresent(TestEntityType.Source, "AdBids_new");
    await testBrowser.entityNotPresent(TestEntityType.Source, "AdBids");
  });
});

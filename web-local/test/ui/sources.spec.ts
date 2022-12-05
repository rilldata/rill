import { describe, it } from "@jest/globals";
import { deleteEntity, renameEntityUsingMenu } from "./utils/commonHelpers";
import {
  waitForAdBids,
  waitForAdImpressions,
} from "./utils/dataSpecifcHelpers";
import { TestEntityType } from "./utils/helpers";
import { useRegisteredServer } from "./utils/serverConfigs";
import { createOrReplaceSource, uploadFile } from "./utils/sourceHelpers";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";

describe.skip("sources", () => {
  const testBrowser = useRegisteredServer("models");

  it("Import sources", async () => {
    const { page } = testBrowser;

    await Promise.all([
      waitForAdBids(page, "AdBids"),
      uploadFile(page, "AdBids.csv"),
    ]);

    await Promise.all([
      waitForAdImpressions(page, "AdImpressions"),
      uploadFile(page, "AdImpressions.tsv"),
    ]);

    // upload existing table and keep both
    await Promise.all([
      waitForEntity(page, TestEntityType.Source, "AdBids", false),
      waitForAdBids(page, "AdBids_1"),
      uploadFile(page, "AdBids.csv", true, true),
    ]);

    // upload existing table and replace
    await Promise.all([
      waitForAdBids(page, "AdBids"),
      uploadFile(page, "AdBids.csv", true, false),
    ]);
    await entityNotPresent(page, TestEntityType.Source, "AdBids_2");
  });

  it("Rename and delete sources", async () => {
    const { page } = testBrowser;

    // make sure AdBids is present
    await createOrReplaceSource(page, "AdBids.csv", "AdBids");

    // rename
    await renameEntityUsingMenu(
      page,
      TestEntityType.Source,
      "AdBids",
      "AdBids_new"
    );
    await waitForEntity(page, TestEntityType.Source, "AdBids_new", true);
    await entityNotPresent(page, TestEntityType.Source, "AdBids");

    // delete
    await deleteEntity(page, TestEntityType.Source, "AdBids_new");
    await entityNotPresent(page, TestEntityType.Source, "AdBids_new");
    await entityNotPresent(page, TestEntityType.Source, "AdBids");
  });
});

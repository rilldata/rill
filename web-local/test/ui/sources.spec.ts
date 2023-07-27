import { test } from "@playwright/test";
import { deleteEntity, renameEntityUsingMenu } from "./utils/commonHelpers";
import {
  waitForAdBids,
  waitForAdImpressions,
} from "./utils/dataSpecifcHelpers";
import { TestEntityType } from "./utils/helpers";
import { createOrReplaceSource, uploadFile } from "./utils/sourceHelpers";
import { entityNotPresent, waitForEntity } from "./utils/waitHelpers";
import { startRuntimeForEachTest } from "./utils/startRuntimeForEachTest";

test.describe("sources", () => {
  startRuntimeForEachTest();

  test("Import sources", async ({ page }) => {
    await page.goto("/");

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
    await uploadFile(page, "AdBids.csv", true, false);
    await entityNotPresent(page, "AdBids_2");
  });

  test("Rename and delete sources", async ({ page }) => {
    await page.goto("/");

    await createOrReplaceSource(page, "AdBids.csv", "AdBids");

    // rename
    await renameEntityUsingMenu(page, "AdBids", "AdBids_new");
    await waitForEntity(page, TestEntityType.Source, "AdBids_new", true);
    await entityNotPresent(page, "AdBids");

    // delete
    await deleteEntity(page, "AdBids_new");
    await entityNotPresent(page, "AdBids_new");
    await entityNotPresent(page, "AdBids");
  });
});

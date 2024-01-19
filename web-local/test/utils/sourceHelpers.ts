import { expect } from "@playwright/test";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import path from "node:path";
import type { Page } from "playwright";
import { fileURLToPath } from "url";
import {
  clickModalButton,
  getEntityLink,
  TestEntityType,
  waitForProfiling,
} from "./commonHelpers";
import { waitForEntity } from "./waitHelpers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const TestDataPath = path.join(__dirname, "../../data");

/**
 * Used to upload local file as a source
 * @param page
 * @param file File name relative to test data folder
 * @param isDuplicate
 * @param keepBoth
 */
export async function uploadFile(
  page: Page,
  file: string,
  isDuplicate = false,
  keepBoth = false,
) {
  // add table button
  await page.locator("button#add-table").click();
  // click local file button
  await page.locator("button#local_file").click();
  // wait for file chooser while clicking on upload button
  const [fileChooser] = await Promise.all([
    page.waitForEvent("filechooser"),
    page.getByText("Upload a CSV, JSON or Parquet file").click(),
  ]);
  // input the `file` after joining with `testDataPath`
  const fileUploadPromise = fileChooser.setFiles([
    path.join(TestDataPath, file),
  ]);

  // TODO: infer duplicate
  if (isDuplicate) {
    await fileUploadPromise;
    let duplicatePromise;
    if (keepBoth) {
      // click on `Keep Both` if `isDuplicate`=true and `keepBoth`=true
      duplicatePromise = clickModalButton(page, "Keep Both");
    } else {
      // else click on `Replace Existing Source`
      duplicatePromise = clickModalButton(page, "Replace Existing Source");
    }
    await Promise.all([page.waitForResponse(/files\/-\//), duplicatePromise]);
  } else {
    await Promise.all([page.waitForResponse(/files\/-\//), fileUploadPromise]);
    // if not duplicate wait and make sure `Duplicate source name` modal is not open
    await asyncWait(100);
    await expect(page.getByText("Duplicate source name")).toBeHidden();
  }
}

export async function createOrReplaceSource(
  page: Page,
  file: string,
  name: string,
) {
  try {
    await getEntityLink(page, name).waitFor({
      timeout: 100,
    });
    await uploadFile(page, file, true, false);
  } catch (err) {
    await uploadFile(page, file);
  }
  await Promise.all([
    page.getByText("View this source").click(),
    waitForEntity(page, TestEntityType.Source, name, true),
  ]);
}

export async function waitForSource(
  page: Page,
  name: string,
  columns: Array<string>,
) {
  await Promise.all([
    page.getByText("View this source").click(),
    waitForEntity(page, TestEntityType.Source, name, true),
    waitForProfiling(page, name, columns),
  ]);
}

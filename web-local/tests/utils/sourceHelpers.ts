import { expect } from "@playwright/test";
import {
  extractFileName,
  splitFolderAndName,
} from "@rilldata/web-common/features/sources/extract-file-name";
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
import { waitForEntity, waitForFileEntry } from "./waitHelpers";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const TestDataPath = path.join(__dirname, "../data");

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
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
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
  const fileRespWaitPromise = page.waitForResponse(/files\/-\//);

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
    await Promise.all([fileRespWaitPromise, duplicatePromise]);
  } else {
    await Promise.all([fileRespWaitPromise, fileUploadPromise]);
    // if not duplicate wait and make sure `Duplicate source name` modal is not open
    await asyncWait(100);
    await expect(page.getByText("Duplicate source name")).toBeHidden();
  }
}

export async function createOrReplaceSource(
  page: Page,
  file: string,
  filePath: string,
) {
  const [, fileName] = splitFolderAndName(filePath);
  const name = extractFileName(filePath);
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
    waitForFileEntry(page, filePath, fileName, true),
  ]);
}

export async function waitForSource(
  page: Page,
  filePath: string,
  columns: Array<string>,
) {
  const [, fileName] = splitFolderAndName(filePath);
  const name = extractFileName(filePath);
  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileEntry(page, filePath, fileName, true),
    waitForProfiling(page, name, columns),
  ]);
}

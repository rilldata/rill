import { expect } from "@playwright/test";
import {
  extractFileName,
  splitFolderAndFileName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import path from "node:path";
import type { Page } from "playwright";
import { fileURLToPath } from "url";
import { clickModalButton, waitForProfiling } from "./commonHelpers";
import { waitForFileNavEntry } from "./waitHelpers";

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
  await page.getByLabel("Add Data").click();
  // click local file button
  await page.locator("button#local_file").click();
  // input the `file` after joining with `testDataPath`
  await page.setInputFiles("input[type=file]", [path.join(TestDataPath, file)]);
  // click upload.
  await page.getByRole("button", { name: "Upload" }).click();

  const fileRespWaitPromise = page.waitForResponse(/files\/entry/);

  // TODO: infer duplicate
  if (isDuplicate) {
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
    await Promise.all([fileRespWaitPromise]);
    // if not duplicate wait and make sure `Duplicate source name` modal is not open
    await asyncWait(100);
    await expect(page.getByText("Duplicate source name")).toBeHidden();
  }
}

export async function createSource(page: Page, file: string, filePath: string) {
  await uploadFile(page, file);
  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileNavEntry(page, filePath, true),
  ]);
}

export async function waitForSource(
  page: Page,
  filePath: string,
  columns: Array<string>,
) {
  const [, fileName] = splitFolderAndFileName(filePath);
  const name = extractFileName(fileName);

  await Promise.all([
    page.getByText("View this source").click(),
    waitForFileNavEntry(page, filePath, true),
    waitForProfiling(page, name, columns),
  ]);
}

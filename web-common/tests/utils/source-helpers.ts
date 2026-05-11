import type { Page } from "playwright";
import path from "node:path";
import { waitForFileNavEntry } from "@rilldata/rill/tests/utils/waitHelpers.ts";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
export const TestDataPath = path.join(
  __dirname,
  "../../../web-local/tests/data",
);
console.log(__dirname, TestDataPath);

export async function createLocalFileSource(
  page: Page,
  file: string,
  filePath: string,
  viewSource = true,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Data").click();
  // click local file button
  await page.getByLabel("Connect to local_file").click();
  // wait for file chooser while clicking on upload button
  const [fileChooser] = await Promise.all([
    page.waitForEvent("filechooser"),
    page.getByLabel("Upload file").click(),
  ]);
  // input the `file` after joining with `testDataPath`
  await fileChooser.setFiles([path.join(TestDataPath, file)]);
  // Import and wait for source to be created.
  await page.getByRole("button", { name: "Import Data" }).click();

  if (viewSource) {
    await Promise.all([
      page.getByRole("button", { name: "View this source" }).click(),
      waitForFileNavEntry(page, filePath, true),
    ]);
  }
}

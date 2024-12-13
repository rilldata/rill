import { expect } from "@playwright/test";
import {
  extractFileName,
  splitFolderAndFileName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import path from "node:path";
import type { Page } from "playwright";
import { fileURLToPath } from "url";
import { clickModalButton, waitForProfiling, getFileNavEntry } from "./commonHelpers";
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
  const fileRespWaitPromise = page.waitForResponse(/files\/entry/);

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



//additions by Roy

export async function addFileWithCheck(
  page: Page,
  file: string, //used to check URL for creation
) {
  const filePath = `/files/${file}`
  // add asset button
  await page.getByLabel("Add Asset").click();
  // Hover over "More" to reveal the dropdown menu
  await page.locator('text=More').hover();
  // Click the "Blank file" button
  await page.locator('text=Blank file').click();
  //check the URL for created file.
  await page.waitForURL(`**${filePath}`);
}

export async function addFolderWithCheck(
  page: Page,
  file: string, //used to check URL for creation 
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // Hover over "More" to reveal the dropdown menu
  await page.locator('text=More').hover();
  // Click the "Blank file" button
  await page.getByRole('menuitem', { name: 'Folder' }).click();
  // how to check folder...?
  const entityLocator = getFileNavEntry(page, file);
  await entityLocator.isVisible();
}


export async function waitForTable(
  page: Page,
  filePath: string,
  columns: Array<string>,
) {
  const [, fileName] = splitFolderAndFileName(filePath);
  const name = extractFileName(fileName);
  console.log(name)
  // add checks later, need to figure out how this works
  await Promise.all([
 //   page.getByText("View this source").click(),    Once v52 released, need to select View this Source
      waitForFileNavEntry(page, filePath, true),
      waitForProfiling(page, name, columns), //this one is imported bc failed sources will still navigate
  ]);
}


//S3, GCS, ABS
export async function cloud(
  page: Page,
  file: string,
  comp: string,
//  isDuplicate = false,
//  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator(`button#${comp}`).click();

  // input the needed details for mysql 
// Locate the  SQL text box and modify
console.log("opened the UI")

const inputField = page.locator('input#path');
let bucketPath;
// Modify the text
// Get the current working directory
if (comp === 'gcs') {
  bucketPath = `gs://playwright-${comp}-qa/${file}`;
} else {
  bucketPath = `${comp}://playwright-${comp}-qa/${file}`;
}

await inputField.fill(bucketPath);


// Locate the DSN text box and modify
const inputField2 = page.locator('input#name');
const fileName = file.split('.')[0];

// Modify the text
await inputField2.fill(fileName);

  // add source menu item
  await page.locator(`button[form="add-data-${comp}-form"]`).waitFor({ state: 'visible' });
  await page.locator(`button[form="add-data-${comp}-form"]`).click();
//  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load 
//  await page.reload();

  // TODO: infer duplicate
//currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}


// Additions MySQL

export async function mySQLDataset(
  page: Page,
  file: string,
//  isDuplicate = false,
//  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#mysql").click();

  // input the needed details for mysql 
// Locate the  SQL text box and modify
const inputField = page.locator('input#sql');
// Modify the text
await inputField.fill(`SELECT * FROM ${file};`);


// Locate the DSN text box and modify
const inputField2 = page.locator('input#dsn');

// Modify the text
await inputField2.fill('root:rootpass@tcp(127.0.0.1:3306)/default');

// Locate the source_name text box and modify
const inputField3 = page.locator('input#name');

// Modify the text
await inputField3.fill(`${file}`);

  // add source menu item
  await page.locator('button[form="add-data-mysql-form"]').waitFor({ state: 'visible' });
  await page.locator('button[form="add-data-mysql-form"]').click();
  await page.waitForTimeout(2000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load 
 // await page.reload();

  // TODO: infer duplicate
//currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

// Additions SQLITE

export async function sqlLiteDataset(
  page: Page,
  file: string,
//  isDuplicate = false,
//  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#sqlite").click();

  // input the needed details for mysql 
// Locate the  SQL text box and modify
console.log("opened the UI")

const currentDir = process.cwd(); // Returns the directory where the script is executed
const dbPath = path.join(currentDir, 'mydb.sqlite');

const inputField = page.locator('input#db');
// Modify the text
// Get the current working directory

await inputField.fill(dbPath);


// Locate the DSN text box and modify
const inputField2 = page.locator('input#table');

// Modify the text
await inputField2.fill(`${file}`);

// Locate the source_name text box and modify
const inputField3 = page.locator('input#name');

// Modify the text
await inputField3.fill(`${file}`);

  // add source menu item
  await page.locator('button[form="add-data-sqlite-form"]').waitFor({ state: 'visible' });
  await page.locator('button[form="add-data-sqlite-form"]').click();
  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load 
//  await page.reload();

  // TODO: infer duplicate
//currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

// Additions postgre

export async function pgDataset(
  page: Page,
  file: string,
//  isDuplicate = false,
//  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#postgres").click();

  // input the needed details for mysql 
// Locate the  SQL text box and modify
console.log("opened the UI")

const inputField = page.locator('input#sql');
// Modify the text
// Get the current working directory
await inputField.fill(`SELECT * FROM ${file};`);


// Locate the DSN text box and modify
const inputField2 = page.locator('input#database_url');

// Modify the text
await inputField2.fill('postgresql://postgres:postgrespass@localhost:5432/default');

// Locate the source_name text box and modify
const inputField3 = page.locator('input#name');

// Modify the text
await inputField3.fill(`${file}`);

  // add source menu item
  await page.locator('button[form="add-data-postgres-form"]').waitFor({ state: 'visible' });
  await page.locator('button[form="add-data-postgres-form"]').click();
  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load 
//  await page.reload();

  // TODO: infer duplicate
//currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}

export async function DuckDB(
  page: Page,
  file: string,
//  isDuplicate = false,
//  keepBoth = false,
) {
  // add asset button
  await page.getByLabel("Add Asset").click();
  // add source menu item
  await page.getByLabel("Add Source").click();
  // click local file button
  await page.locator("button#duckdb").click();

  // input the needed details for mysql 
// Locate the  SQL text box and modify
console.log("opened the UI")
const currentDir = process.cwd(); // Returns the directory where the script is executed
const dbPath = path.resolve(currentDir, 'tests/data/playwright.db'); //need to fix this when ready to deploy

const inputField = page.locator('input#db');

// Modify the text
// Get the current working directory
await inputField.fill(dbPath);


// Locate the DSN text box and modify
const inputField2 = page.locator('input#sql');

// Modify the text
await inputField2.fill(`SELECT * FROM ${file};`);

// Locate the source_name text box and modify
const inputField3 = page.locator('input#name');

const bucketPath = '';

// Modify the text
await inputField3.fill(`${file}`);

  // add source menu item
  await page.locator('button[form="add-data-duckdb-form"]').waitFor({ state: 'visible' });
  await page.locator('button[form="add-data-duckdb-form"]').click();
  await page.waitForTimeout(5000); // Waits for 5000 milliseconds (5 seconds) -- refreshing page bc idk whats happnening forever load 
//  await page.reload();

  // TODO: infer duplicate
//currently we dont check anything when duplicate source is added, or sometimes it overwrites.
}
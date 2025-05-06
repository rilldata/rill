import { type Page, expect } from "@playwright/test";

/**
 * Validates the contents of a table on a Playwright page.
 * @param page - The Playwright Page instance.
 * @param tableSelector - The CSS selector for the table.
 * @param expectedData - A 2D array representing the expected table content.
 */
export async function validateTableContents(
  page: Page,
  tableSelector: string,
  expectedData: string[][],
): Promise<void> {
  // Select all rows within the table
  const rows = page.locator(`${tableSelector} > tbody > tr`);

  // Validate row count
  await expect(rows).toHaveCount(expectedData.length);

  // Loop through each row to validate cell contents
  for (let i = 0; i < expectedData.length; i++) {
    const cells = rows.nth(i).locator("td"); // Select cells in the current row

    // Validate cell count in the row
    await expect(cells).toHaveCount(expectedData[i].length);

    // Loop through each cell to verify its content
    for (let j = 0; j < expectedData[i].length; j++) {
      const cellText = await cells.nth(j).innerText();
      expect(cellText.trim()).toBe(expectedData[i][j]);
    }
  }
}

import { google, sheets_v4 } from 'googleapis';
import { getGoogleAuth } from '../auth.js';
import { shareDocument } from './drive.js';

/**
 * Sets up the Google Sheets client.
 */
export async function setupSheetsClient(): Promise<sheets_v4.Sheets> {
  const auth = await getGoogleAuth();
  return google.sheets({ version: 'v4', auth });
}

/**
 * Creates a new Google Sheet with the specified title.
 * @param title - The title of the sheet to be created.
 */
export async function createSheet(
  title: string,
): Promise<sheets_v4.Schema$Spreadsheet | undefined> {
  const sheets = await setupSheetsClient();
  const response = await sheets.spreadsheets.create({
    requestBody: {
      properties: { title },
    },
  });
  return response.data;
}

/**
 * Updates the data in a Google Sheet.
 * @param spreadsheetId - The ID of the spreadsheet to update.
 * @param sheetName - The name of the sheet within the spreadsheet.
 * @param headers - An array of header strings.
 * @param rows - A 2D array representing the rows of data.
 */
export async function updateSheetData(
  spreadsheetId: string,
  sheetName: string,
  headers: string[],
  rows: any[][],
): Promise<void> {
  const sheets = await setupSheetsClient();

  // Write headers
  await sheets.spreadsheets.values.update({
    spreadsheetId,
    range: `${sheetName}!A1`,
    valueInputOption: 'RAW',
    requestBody: {
      values: [headers],
    },
  });

  // Write data rows
  await sheets.spreadsheets.values.update({
    spreadsheetId,
    range: `${sheetName}!A2`,
    valueInputOption: 'RAW',
    requestBody: {
      values: rows,
    },
  });

  // Format headers (bold, frozen)
  await sheets.spreadsheets.batchUpdate({
    spreadsheetId,
    requestBody: {
      requests: [
        {
          repeatCell: {
            range: {
              sheetId: 0,
              startRowIndex: 0,
              endRowIndex: 1,
            },
            cell: {
              userEnteredFormat: {
                textFormat: { bold: true },
                backgroundColor: { red: 0.9, green: 0.9, blue: 0.9 },
              },
            },
            fields: 'userEnteredFormat(textFormat,backgroundColor)',
          },
        },
        {
          updateSheetProperties: {
            properties: {
              sheetId: 0,
              gridProperties: {
                frozenRowCount: 1,
              },
            },
            fields: 'gridProperties.frozenRowCount',
          },
        },
      ],
    },
  });
}

/**
 * Shares the Google Sheet with the specified email or makes it publicly accessible.
 * @param spreadsheetId - The ID of the spreadsheet to share.
 * @param email - Optional email address to share the sheet with.
 * @returns The web view link of the shared sheet.
 */
export async function shareSheet(
  spreadsheetId: string,
  email?: string,
): Promise<string | null | undefined> {
  return shareDocument(spreadsheetId, email);
}

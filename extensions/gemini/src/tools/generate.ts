import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { z } from 'zod';
import { shareDocument } from '../lib/drive/drive.js';
import { createDocument, updateDocument } from '../lib/drive/docs.js';
import {
  writeLocalReport,
  getDefaultLocalPath,
  convertMarkdownToGoogleDocsRequests,
} from '../lib/utils.js';
import { createSheet, shareSheet, updateSheetData } from '../lib/drive/sheets.js';

export function registerGenerateTool(server: McpServer) {
  server.registerTool(
    'generate',
    {
      description:
        'Creates a comprehensive analytics report in Google Docs based on Rill data analysis. Optionally saves a local copy first.',
      inputSchema: {
        title: z.string().describe('The title of the report document.'),
        content: z.string().describe('The main content of the report in markdown format.'),
        shareEmail: z
          .string()
          .optional()
          .describe(
            'Email address to share the document with (optional - if not provided, document will be public).',
          ),
        localFilePath: z
          .string()
          .optional()
          .describe(
            'Optional local file path to save the report before creating Google Doc. If not provided, will generate a default path.',
          ),
        saveLocalOnly: z
          .boolean()
          .optional()
          .describe(
            'If true, only saves locally and skips Google Docs creation. Defaults to false.',
          ),
      },
    },
    async (input) => {
      try {
        const { title, content, shareEmail, localFilePath, saveLocalOnly = false } = input;

        const results: string[] = [];
        let localPath: string | null = null;

        // Write local file if requested or if saveLocalOnly is true
        if (localFilePath || saveLocalOnly) {
          const filePath = localFilePath || getDefaultLocalPath(title);
          localPath = await writeLocalReport(content, filePath);
          results.push(`Local report saved: ${localPath}`);
        }

        // Skip Google Docs creation if saveLocalOnly is true
        if (!saveLocalOnly) {
          // Create the document
          const document = await createDocument(title);
          if (!document?.documentId) {
            throw new Error('Failed to create document');
          }

          // Convert markdown content to Google Docs format
          const requests = convertMarkdownToGoogleDocsRequests(content);

          // Update document with content
          await updateDocument(document.documentId, requests);

          // Share the document
          const shareLink = await shareDocument(document.documentId, shareEmail);

          results.push(`Google Docs report created successfully!`);
          results.push(`Document ID: ${document.documentId}`);
          results.push(`Share Link: ${shareLink}`);
        }

        results.push(`Title: ${title}`);

        return {
          content: [
            {
              type: 'text',
              text: results.join('\n\n'),
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Error creating report: ${error instanceof Error ? error.message : 'Unknown error'}`,
            },
          ],
        };
      }
    },
  );
}

export function registerExportSheetTool(server: McpServer) {
  server.registerTool(
    'export_to_sheet',
    {
      description:
        'Exports Rill query results to a Google Sheet with formatted tables and optional sharing.',
      inputSchema: {
        title: z.string().describe('The title of the Google Sheet.'),
        data: z
          .array(z.record(z.any()))
          .describe('Array of query result objects to export as rows.'),
        sheetName: z.string().optional().describe('Name for the worksheet tab (default: Sheet1).'),
        shareEmail: z.string().optional().describe('Email address to share the sheet with.'),
      },
    },
    async (input) => {
      try {
        const { title, data, sheetName = 'Sheet1', shareEmail } = input;

        // Create new spreadsheet
        const sheet = await createSheet(title);
        if (!sheet?.spreadsheetId) {
          throw new Error('Failed to create spreadsheet');
        }

        // Convert data to rows format
        const headers = Object.keys(data[0] || {});
        const rows = data.map((row) => headers.map((h) => row[h]));

        // Update sheet with data
        await updateSheetData(sheet.spreadsheetId, sheetName, headers, rows);

        // Share the sheet
        const shareLink = await shareSheet(sheet.spreadsheetId, shareEmail);

        return {
          content: [
            {
              type: 'text',
              text: `Sheet exported successfully!\n\nSpreadsheet ID: ${sheet.spreadsheetId}\nTitle: ${title}\nShare Link: ${shareLink}`,
            },
          ],
        };
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Error exporting to sheet: ${error instanceof Error ? error.message : 'Unknown error'}`,
            },
          ],
        };
      }
    },
  );
}

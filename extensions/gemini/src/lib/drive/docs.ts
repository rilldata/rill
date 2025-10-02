import { google, docs_v1, drive_v3 } from 'googleapis';
import { getGoogleAuth } from '../auth.js';
import { setupDriveClient } from './drive.js';

/**
 * Sets up the Google Docs client.
 */
export async function setupDocsClient(): Promise<docs_v1.Docs> {
  const auth = await getGoogleAuth();
  return google.docs({ version: 'v1', auth });
}

/** Creates a new Google Document with the specified title.
 * @param title - The title of the document to be created.
 */
export async function createDocument(title: string): Promise<docs_v1.Schema$Document | undefined> {
  const docs = await setupDocsClient();
  const response = await docs.documents.create({
    requestBody: { title },
  });
  return response.data;
}

/**
 * Updates the content of a Google Document.
 * @param documentId
 * @param content
 */
export async function updateDocument(
  documentId: string,
  content: docs_v1.Schema$Request[],
): Promise<void> {
  const docs = await setupDocsClient();
  await docs.documents.batchUpdate({
    documentId,
    requestBody: { requests: content },
  });
}

/**
 * Shares a Google Document with the specified email or makes it public.
 * @param documentId
 * @param email
 */
export async function shareDocument(
  documentId: string,
  email?: string,
): Promise<string | null | undefined> {
  const drive = await setupDriveClient();

  const permission: drive_v3.Schema$Permission = {
    type: 'anyone',
    role: 'reader',
  };

  if (email) {
    permission.type = 'user';
    permission.emailAddress = email;
  }

  await drive.permissions.create({
    fileId: documentId,
    requestBody: permission,
  });

  const file = await drive.files.get({
    fileId: documentId,
    fields: 'webViewLink',
  });

  return file.data.webViewLink;
}

/**
 * Gets the content of a Google Document.
 * @param documentId - The ID of the document to retrieve
 */
export async function getDocument(
  documentId: string,
): Promise<docs_v1.Schema$Document | undefined> {
  const docs = await setupDocsClient();
  const response = await docs.documents.get({
    documentId,
  });
  return response.data;
}

/**
 * Lists Google Drive documents (reports) created by this extension.
 * @param maxResults - Maximum number of results to return (default: 10)
 */
export async function listDriveReports(maxResults: number = 10): Promise<drive_v3.Schema$File[]> {
  const drive = await setupDriveClient();
  const response = await drive.files.list({
    q: "mimeType='application/vnd.google-apps.document' and trashed=false",
    fields: 'files(id, name, modifiedTime, webViewLink, createdTime)',
    pageSize: maxResults,
    orderBy: 'modifiedTime desc',
  });
  return response.data.files || [];
}

/**
 * Uploads a local file to Google Drive and converts it to a Google Doc.
 * @param filePath - Path to the local file
 * @param title - Title for the Google Doc
 * @param mimeType - MIME type of the source file (default: text/markdown)
 */
export async function uploadFileToGoogleDocs(
  content: string,
  title: string,
  mimeType: string = 'text/markdown',
): Promise<drive_v3.Schema$File | undefined> {
  const drive = await setupDriveClient();

  const fileMetadata = {
    name: title,
    mimeType: 'application/vnd.google-apps.document',
  };

  const media = {
    mimeType: mimeType,
    body: content,
  };

  const response = await drive.files.create({
    requestBody: fileMetadata,
    media: media,
    fields: 'id, name, webViewLink',
  });

  return response.data;
}

/**
 * Deletes all content from a Google Document and replaces it with new content.
 * @param documentId - The ID of the document to update
 * @param _content - The new content in markdown format (to be implemented)
 */
export async function replaceDocumentContent(documentId: string, _content: string): Promise<void> {
  const docs = await setupDocsClient();

  const doc = await getDocument(documentId);
  if (!doc?.body?.content) {
    throw new Error('Could not retrieve document content');
  }

  const endIndex = doc.body.content[doc.body.content.length - 1]?.endIndex || 1;

  const deleteRequest = {
    deleteContentRange: {
      range: {
        startIndex: 1,
        endIndex: endIndex - 1,
      },
    },
  };

  await docs.documents.batchUpdate({
    documentId,
    requestBody: {
      requests: [deleteRequest],
    },
  });
}

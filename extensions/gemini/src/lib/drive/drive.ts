import { google, drive_v3 } from 'googleapis';
import { getGoogleAuth } from '../auth.js';

/**
 * Sets up the Google Drive client.
 */
export async function setupDriveClient(): Promise<drive_v3.Drive> {
  const auth = await getGoogleAuth();
  return google.drive({ version: 'v3', auth });
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

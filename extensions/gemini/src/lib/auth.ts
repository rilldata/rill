import { GoogleAuth } from 'google-auth-library';

// Singleton Google Auth instance
let googleAuthInstance: GoogleAuth | null = null;

/**
 * Gets or creates a single Google Auth instance with all required scopes.
 */
export async function getGoogleAuth(): Promise<GoogleAuth> {
  if (!googleAuthInstance) {
    googleAuthInstance = new GoogleAuth({
      scopes: [
        'https://www.googleapis.com/auth/drive',
        'https://www.googleapis.com/auth/documents',
        'https://www.googleapis.com/auth/spreadsheets',
      ],
    });

    // Verify credentials work with required scopes
    const hasCredentials = await ensureGCPCredentials();
    if (!hasCredentials) {
      googleAuthInstance = null;
      throw new Error('Google Cloud credentials are not properly configured');
    }
  }

  return googleAuthInstance;
}

/**
 * Ensures that Google Cloud credentials are set up correctly with the required scopes.
 */
export async function ensureGCPCredentials(): Promise<boolean> {
  try {
    if (!googleAuthInstance) {
      return false;
    }

    const client = await googleAuthInstance.getClient();
    const accessToken = await client.getAccessToken();

    return accessToken.token !== null && accessToken.token !== undefined;
  } catch (error) {
    console.error('Google Cloud authentication failed:', error);
    return false;
  }
}

// Replica of Playwright's `StorageState` type, which Playwright does not export
interface StorageState {
  cookies: Array<{
    name: string;
    value: string;
    domain: string;
    path: string;
    expires: number;
    httpOnly: boolean;
    secure: boolean;
    sameSite: "Strict" | "Lax" | "None";
  }>;
  origins: Array<{
    origin: string;
    localStorage: Array<{
      name: string;
      value: string;
    }>;
  }>;
}

export function getGitHubStorageState(
  storageStateJson: string | undefined,
): StorageState {
  if (!storageStateJson) {
    throw new Error(
      "Missing environment variable required for GitHub authentication",
    );
  }
  return JSON.parse(storageStateJson) as StorageState;
}

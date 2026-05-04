/**
 * One-shot flag the View-as user pickers can set just before navigating,
 * to tell the project-layout's platform-change watcher *not* to clear
 * the impersonation it just established.
 *
 * Without this, picking a mock/real user from the split-button dropdown
 * triggers a navigation across a platform boundary (editor → preview),
 * the platform watcher fires, and the freshly-set impersonation gets
 * wiped before the new page can render with it.
 */
let skip = false;

export function skipNextPlatformReset(): void {
  skip = true;
}

export function consumePlatformResetSkip(): boolean {
  if (!skip) return false;
  skip = false;
  return true;
}

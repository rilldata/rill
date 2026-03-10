/**
 * Utilities for encoding branch names as `@branch` path segments.
 *
 * URL format:
 *   Production: /org/project/explore/dashboard
 *   Branch:     /org/project/@feature-x/explore/dashboard
 *
 * The `@` prefix disambiguates branch segments from route/dashboard names.
 *
 * Branch names with `/` (e.g., `feature/login`) are encoded by replacing
 * `/` with `~`. This is safe because git disallows `~` in branch names.
 * Standard `encodeURIComponent` is also applied for other special characters.
 */

// Matches /@branch as the 3rd path segment (after org and project)
const BRANCH_SEGMENT_RE = /^(\/[^/]+\/[^/]+)\/@([^/]+)(\/.*)?$/;

/** Encode a branch name for use in a URL path segment. */
function encodeBranch(branch: string): string {
  return encodeURIComponent(branch.replaceAll("/", "~"));
}

/** Decode a branch name from a URL path segment. */
function decodeBranch(encoded: string): string {
  return decodeURIComponent(encoded).replaceAll("~", "/");
}

/** Extract the branch name from a pathname, or undefined if not present. */
export function extractBranchFromPath(pathname: string): string | undefined {
  const match = pathname.match(BRANCH_SEGMENT_RE);
  if (!match) return undefined;
  return decodeBranch(match[2]);
}

/** Insert `@branch` after the project segment. */
export function injectBranchIntoPath(pathname: string, branch: string): string {
  const segments = pathname.split("/");
  // segments[0] = "", segments[1] = org, segments[2] = project
  if (segments.length < 3) return pathname;
  segments.splice(3, 0, `@${encodeBranch(branch)}`);
  return segments.join("/");
}

/** Remove `@branch` from the pathname if present. */
export function removeBranchFromPath(pathname: string): string {
  const match = pathname.match(BRANCH_SEGMENT_RE);
  if (!match) return pathname;
  return match[1] + (match[3] ?? "");
}

/** Returns the branch path prefix for building tab URLs: "" or "/@encoded-branch". */
export function branchPathPrefix(branch: string | undefined): string {
  if (!branch) return "";
  return `/@${encodeBranch(branch)}`;
}

/**
 * Shared flag: when set, the next `beforeNavigate` call in the project layout
 * will skip `@branch` injection. Used by the BranchSelector and the "Back to
 * production" banner to navigate to production without re-injection.
 */
let _skipNext = false;
export function requestSkipBranchInjection(): void {
  _skipNext = true;
}
export function consumeSkipBranchInjection(): boolean {
  if (_skipNext) {
    _skipNext = false;
    return true;
  }
  return false;
}

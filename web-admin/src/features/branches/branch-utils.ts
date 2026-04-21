import type { BeforeNavigate } from "@sveltejs/kit";

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
 * Given a navigation target, returns the branch-injected URL string if
 * the navigation should be redirected, or null if no redirect is needed.
 *
 * Used by `handleBranchNavigation` to transparently preserve the active
 * `@branch` segment on project-internal navigations.
 */
export function getBranchRedirect(
  to: URL,
  activeBranch: string,
  organization: string,
  project: string,
): string | null {
  const prefix = `/${organization}/${project}`;
  if (!to.pathname.startsWith(prefix + "/") && to.pathname !== prefix)
    return null;
  if (to.pathname.includes("/-/share/")) return null;
  if (extractBranchFromPath(to.pathname)) return null;
  return injectBranchIntoPath(to.pathname, activeBranch) + to.search + to.hash;
}

/**
 * Intercept a client-side navigation and inject `@branch` if needed.
 *
 * Many components build project URLs without branch awareness. This function
 * is called from the project layout's `beforeNavigate` hook to transparently
 * add the active branch segment to those navigations.
 *
 * Pass `goto` as `navigateFn` to keep this module decoupled from SvelteKit's
 * `$app/navigation` (which can only be imported inside components).
 */
export function handleBranchNavigation(
  nav: BeforeNavigate,
  activeBranch: string | undefined,
  organization: string,
  project: string,
  navigateFn: (url: string) => Promise<void>,
): void {
  console.log("handleBranchNavigation", activeBranch, nav.to?.url?.toString());
  if (!activeBranch || !nav.to?.url) return;
  if (consumeSkipBranchInjection()) return;
  if (nav.type === "popstate") return;
  const redirect = getBranchRedirect(
    nav.to.url,
    activeBranch,
    organization,
    project,
  );
  console.log("handleBranchNavigation:redirect", redirect);
  if (!redirect) return;
  nav.cancel();
  void navigateFn(redirect);
}

/**
 * Shared flag: when set, the next `beforeNavigate` call in the project layout
 * will skip `@branch` injection. Used by the BranchSelector to navigate
 * to production without the layout re-injecting the current branch.
 *
 * Auto-expires after 500ms to prevent a stale flag from leaking if the
 * expected navigation never fires (e.g., cancelled by another hook).
 */
const SKIP_TTL_MS = 500;
let _skipRequestedAt = 0;
export function requestSkipBranchInjection(): void {
  _skipRequestedAt = Date.now();
}
export function consumeSkipBranchInjection(): boolean {
  if (!_skipRequestedAt) return false;
  const elapsed = Date.now() - _skipRequestedAt;
  _skipRequestedAt = 0;
  return elapsed < SKIP_TTL_MS;
}

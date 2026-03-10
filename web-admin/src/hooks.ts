import type { Reroute } from "@sveltejs/kit";

// Matches /@branch as the 3rd path segment (after org and project)
const BRANCH_SEGMENT_RE = /^(\/[^/]+\/[^/]+)\/@[^/]+(\/.*)?$/;

/**
 * Strip `@branch` from the URL before route matching.
 * Existing route files work unchanged; the original URL (with @branch)
 * is preserved in `$page.url` so the layout can extract the branch.
 */
export const reroute: Reroute = ({ url }) => {
  const match = url.pathname.match(BRANCH_SEGMENT_RE);
  if (match) return match[1] + (match[2] ?? "");
};

import type { Reroute } from "@sveltejs/kit";
import { removeBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";

/**
 * Strip `@branch` from the URL before route matching.
 * Existing route files work unchanged; the original URL (with @branch)
 * is preserved in `$page.url` so the layout can extract the branch.
 */
export const reroute: Reroute = ({ url }) => {
  const stripped = removeBranchFromPath(url.pathname);
  if (stripped !== url.pathname) return stripped;
};

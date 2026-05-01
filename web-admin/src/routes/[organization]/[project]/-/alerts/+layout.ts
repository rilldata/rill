import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";

// Alerts is a cloud-only feature; on a branch view send the user back to
// the branch home so deep links (bookmarks, share URLs, stale tabs) don't
// dead-end at a hidden section.
//
// Scoped to this section's loader on purpose — putting the same check on
// the project-wide `[organization]/[project]/+layout.ts` registers `url`
// as a dependency of that loader, which then re-runs on every in-project
// URL change and clobbers in-flight client-side `goto()`s such as the
// home-bookmark URL restoration in `DashboardStateSync`.
export const load: LayoutLoad = ({ url, params }) => {
  const activeBranch = extractBranchFromPath(url.pathname);
  if (activeBranch) {
    throw redirect(
      307,
      `/${params.organization}/${params.project}${branchPathPrefix(activeBranch)}`,
    );
  }
};

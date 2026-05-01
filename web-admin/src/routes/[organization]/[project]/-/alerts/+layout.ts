import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";

// Alerts is a cloud-only feature: assertions evaluate against the
// production deployment, so it doesn't make sense to expose it inside a
// branch view. On a branch URL we redirect deep links (bookmarks, shared
// URLs, stale tabs) back to the branch home rather than letting the user
// dead-end at a hidden section.
//
// Scoped to this section's loader on purpose — putting the same check on
// the project-wide `[organization]/[project]/+layout.ts` registers `url`
// as a dependency of that loader, which then re-runs on every in-project
// URL change and clobbers in-flight client-side `goto()`s such as the
// home-bookmark URL restoration in `DashboardStateSync`.
export const load = ({ url, params: { organization, project } }) => {
  const activeBranch = extractBranchFromPath(url.pathname);
  if (activeBranch) {
    throw redirect(
      307,
      `/${organization}/${project}${branchPathPrefix(activeBranch)}`,
    );
  }
};

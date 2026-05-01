import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";

// Settings (env vars, public URLs, billing, members) only applies to the
// production deployment, so it's hidden from the project nav on branch
// views. This loader catches deep links — bookmarks, shared URLs, stale
// tabs — and bounces them to the branch home so users don't dead-end at
// a section that's intentionally hidden.
//
// Scoped to this section's loader on purpose — putting the same check on
// the project-wide `[organization]/[project]/+layout.ts` registers `url`
// as a dependency of that loader, which then re-runs on every in-project
// URL change and clobbers in-flight client-side `goto()`s such as the
// home-bookmark URL restoration in `DashboardStateSync`.
//
// Production view also gates on `manageProject` so non-admin users can't
// reach Settings via a direct URL.
export const load = async ({
  url,
  parent,
  params: { organization, project },
}) => {
  const activeBranch = extractBranchFromPath(url.pathname);
  if (activeBranch) {
    throw redirect(
      307,
      `/${organization}/${project}${branchPathPrefix(activeBranch)}`,
    );
  }
  const { projectPermissions } = await parent();
  if (!projectPermissions?.manageProject) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

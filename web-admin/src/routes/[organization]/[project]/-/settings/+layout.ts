import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";

// Branch views: redirect deep links back to the branch home.
// Production view: gate on manageProject so users without permission
// can't open Settings via a direct URL.
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

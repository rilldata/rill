import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";

// Status is hidden on branch views; redirect deep links back to the
// branch home. See alerts/+layout.ts for why this lives in a
// section-scoped loader rather than the project-wide layout.
export const load: LayoutLoad = ({ url, params }) => {
  const activeBranch = extractBranchFromPath(url.pathname);
  if (activeBranch) {
    throw redirect(
      307,
      `/${params.organization}/${params.project}${branchPathPrefix(activeBranch)}`,
    );
  }
};

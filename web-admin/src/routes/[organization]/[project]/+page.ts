import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
import { redirect } from "@sveltejs/kit";

/**
 * Cloud's project page is the production view (main branch only).
 * Visiting a branch URL — `/{org}/{project}/@{branch}` — has no separate
 * preview surface; the consolidated Developer + Preview experience lives
 * under `/-/edit`. Send the user straight to the preview side of that.
 */
export const load = async ({ url }) => {
  const branch = extractBranchFromPath(url.pathname);
  if (!branch) return {};

  const target = `${url.pathname.replace(/\/$/, "")}/-/edit/dashboards${url.search}`;
  throw redirect(303, target);
};

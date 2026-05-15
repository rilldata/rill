import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";
import { maybeRedirectToEditableDeployment } from "@rilldata/web-admin/features/branches/deployment-utils.ts";
import { isEditPage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";
import { error } from "@sveltejs/kit";

export const load = async ({
  params: { organization, project },
  parent,
  route,
  url,
}) => {
  const { organizationPermissions } = await parent();

  if (!organizationPermissions.manageOrg) return;

  // Edit pages handle their own branch routing; everything below is non-edit only.
  if (isEditPage({ route })) return;

  // Branch deployments are only viewable from inside `/-/edit`.
  const branch = extractBranchFromPath(url.pathname);
  if (branch) {
    throw error(404, "Branch deployments are only available from the editor.");
  }

  await maybeRedirectToEditableDeployment(organization, project, url);
};

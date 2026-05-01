import { redirect } from "@sveltejs/kit";

// Status reports on the production deployment but is also accessible from
// branch views. Editors and admins can reach it; viewers (no
// `readProdStatus`) cannot, so this loader bounces them to project home
// to avoid letting them dead-end at a 403.
export const load = async ({ parent, params: { organization, project } }) => {
  const { projectPermissions } = await parent();
  if (!projectPermissions?.readProdStatus) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

import { redirect } from "@sveltejs/kit";

export const load = async ({ parent, params: { organization, project } }) => {
  const { projectPermissions } = await parent();
  if (!projectPermissions?.readProdStatus) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

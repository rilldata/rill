import { redirect } from "@sveltejs/kit";

export const load = async ({ parent, params: { organization, project } }) => {
  const { projectPermissions } = await parent();
  if (!projectPermissions?.manageProject) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

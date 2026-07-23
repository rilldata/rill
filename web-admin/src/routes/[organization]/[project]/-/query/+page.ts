import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({
  parent,
  params: { organization, project },
}) => {
  const { projectPermissions } = await parent();
  if (!projectPermissions?.manageProject) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

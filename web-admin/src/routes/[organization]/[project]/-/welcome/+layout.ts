import { projectWelcomeStatusStores } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
import { redirect } from "@sveltejs/kit";

export const load = ({ params: { organization, project } }) => {
  if (!projectWelcomeStatusStores.getProjectWelcomeBranch(project)) {
    throw redirect(307, `/${organization}/${project}`);
  }
};

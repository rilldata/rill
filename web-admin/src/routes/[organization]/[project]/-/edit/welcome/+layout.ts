import { projectWelcomeStatusStores } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
import { redirect } from "@sveltejs/kit";
import {
  extractBranchFromPath,
  injectBranchIntoPath,
} from "@rilldata/web-admin/features/branches/branch-utils.ts";

export const load = ({ params: { organization, project }, url }) => {
  if (!projectWelcomeStatusStores.isProjectWelcomeStep(project)) {
    const branch = extractBranchFromPath(url.pathname);
    throw redirect(
      307,
      injectBranchIntoPath(`/${organization}/${project}/-/edit`, branch),
    );
  }
};

import { projectWelcomeStatusStores } from "@rilldata/web-admin/features/welcome/project/welcome-status.ts";
import { isProjectWelcomePage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";
import { redirect } from "@sveltejs/kit";
import { CreateProjectBranchName } from "@rilldata/web-admin/features/projects/publish-project.ts";

export const load = ({ params: { organization, project }, route }) => {
  if (
    projectWelcomeStatusStores.isProjectWelcomeStep(project) &&
    !isProjectWelcomePage({ route })
  ) {
    throw redirect(
      307,
      `/${organization}/${project}/@${CreateProjectBranchName}/-/edit/welcome`,
    );
  }
};

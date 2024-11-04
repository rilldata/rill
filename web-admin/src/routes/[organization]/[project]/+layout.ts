import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import { fetchAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ params: { organization }, parent }) => {
  const { organizationPermissions, issues, prodDeployment, jwt } =
    await parent();
  if (!organizationPermissions.manageOrg) {
    return;
  }

  if (prodDeployment) {
    void runtime.setRuntime(
      queryClient,
      fixLocalhostRuntimePort(prodDeployment.runtimeHost),
      prodDeployment.runtimeInstanceId,
      jwt,
      "user", // TODO: others
    );
  }

  let projectHibernating = false;

  try {
    projectHibernating = await fetchAllProjectsHibernating(organization);
  } catch (e) {
    if (e.response?.status !== 403) {
      throw error(e.response?.status, "Error fetching project status");
    }
  }

  // if all projects were hibernated due to a blocker issue on org then take the user to projects page
  if (hasBlockerIssues(issues) && projectHibernating) {
    throw redirect(307, `/${organization}`);
  }
};

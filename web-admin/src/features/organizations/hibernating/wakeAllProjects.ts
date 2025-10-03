import {
  adminServiceListProjectsForOrganization,
  createAdminServiceRedeployProject,
  getAdminServiceGetProjectQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
  type V1Project,
  type V1RedeployProjectResponse,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { get } from "svelte/store";

export async function wakeAllProjects(organization: string) {
  const projectDeployer = createAdminServiceRedeployProject(
    undefined,
    queryClient,
  );
  const promises: Promise<V1RedeployProjectResponse>[] = [];

  let pageToken: string | undefined = undefined;
  do {
    try {
      const projectsResp = await queryClient.fetchQuery({
        queryKey: getAdminServiceListProjectsForOrganizationQueryKey(
          organization,
          {
            pageToken,
          },
        ),
        queryFn: () =>
          adminServiceListProjectsForOrganization(organization, {
            pageToken,
          }),
      });
      if (projectsResp.projects.length === 0) break;

      projectsResp.projects.forEach((project) => {
        if (project.prodDeploymentId) return;
        promises.push(redeployProject(organization, project, projectDeployer));
      });
      pageToken = projectsResp.nextPageToken;
    } catch {
      // TODO
      break;
    }
  } while (pageToken);

  try {
    await Promise.all(promises);
  } catch {
    // TODO
  }

  void queryClient.refetchQueries({
    queryKey: getAdminServiceListProjectsForOrganizationQueryKey(organization),
  });
}

async function redeployProject(
  organization: string,
  project: V1Project,
  projectDeployer: ReturnType<typeof createAdminServiceRedeployProject>,
) {
  const resp = await get(projectDeployer).mutateAsync({
    org: organization,
    project: project.name ?? "",
  });
  void queryClient.refetchQueries({
    queryKey: getAdminServiceGetProjectQueryKey(
      organization,
      project.name ?? "",
    ),

    // avoid invalidating createAdminServiceGetProjectWithBearerToken
    exact: true,
  });
  return resp;
}

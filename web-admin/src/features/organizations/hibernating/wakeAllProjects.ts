import {
  adminServiceListProjectsForOrganization,
  createAdminServiceRedeployProject,
  getAdminServiceGetProjectQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
  type V1Project,
  type V1RedeployProjectResponse,
} from "@rilldata/web-admin/client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { get } from "svelte/store";

export async function wakeAllProjects(organization: string) {
  const projectDeployer = createAdminServiceRedeployProject({
    mutation: {
      queryClient,
    },
  });
  const promises: Promise<V1RedeployProjectResponse>[] = [];

  eventBus.emit("banner", {
    type: "info",
    message: "Waking projects. We’ll notify you when they’re ready.",
    iconType: "loading",
  });

  let pageToken: string | undefined = undefined;
  do {
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
  } while (pageToken);

  try {
    await Promise.all(promises);
  } catch {
    // TODO
  }

  eventBus.emit("banner", {
    type: "success",
    message: "Your projects are awake and ready.",
    iconType: "check",
    cta: {
      type: "link",
      text: "View projects ->",
      url: `/${organization}`,
    },
  });
  eventBus.emit("notification", {
    type: "success",
    message: "Projects are now ready and accessible",
  });
}

async function redeployProject(
  organization: string,
  project: V1Project,
  projectDeployer: ReturnType<typeof createAdminServiceRedeployProject>,
) {
  const resp = await get(projectDeployer).mutateAsync({
    organization,
    project: project.name,
  });
  void queryClient.refetchQueries(
    getAdminServiceGetProjectQueryKey(organization, project.name),
    {
      // avoid refetching createAdminServiceGetProjectWithBearerToken
      exact: true,
    },
  );
  return resp;
}

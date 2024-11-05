import {
  adminServiceGetOrganization,
  adminServiceListProjectsForOrganization,
  createAdminServiceListProjectsForOrganization,
  getAdminServiceGetOrganizationQueryKey,
  getAdminServiceListProjectsForOrganizationQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export function areAllProjectsHibernating(organization: string) {
  return createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organization,
        select: (data) =>
          data.projects?.length &&
          data.projects.every((p) => !p.prodDeploymentId),
      },
    },
  );
}

export async function fetchAllProjectsHibernating(organization: string) {
  const projectsResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListProjectsForOrganizationQueryKey(organization),
    queryFn: () => adminServiceListProjectsForOrganization(organization),
    staleTime: Infinity,
  });
  return projectsResp.projects?.every((p) => !p.prodDeploymentId) ?? false;
}

export async function fetchOrganization(organization: string) {
  return await queryClient.fetchQuery({
    queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    queryFn: () => adminServiceGetOrganization(organization),
    staleTime: Infinity,
  });
}

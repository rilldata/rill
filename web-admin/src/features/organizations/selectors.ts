import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";

export function areAllProjectsHibernating(organization: string) {
  return createAdminServiceListProjectsForOrganization(
    organization,
    undefined,
    {
      query: {
        enabled: !!organization,
        select: (data) => data.projects.every((p) => !p.prodDeploymentId),
      },
    },
  );
}

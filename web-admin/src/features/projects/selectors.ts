import {
  createAdminServiceGetProject,
  createAdminServiceListProjectMembers,
} from "@rilldata/web-admin/client";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

// The TTL is actually set in the Admin server – we just use the value for some frontend logic
export const RUNTIME_ACCESS_TOKEN_DEFAULT_TTL = 30 * 60 * 1000; // 30 minutes

export function useProjectRuntime(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      // Proactively refetch the JWT before it expires
      refetchInterval: RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2,
      select: (data) => {
        // There may not be a prodDeployment if the project was hibernated
        if (!data.prodDeployment) {
          return;
        }

        return {
          // Hack: in development, the runtime host is actually on port 8081
          host: data.prodDeployment.runtimeHost.replace(
            "localhost:9091",
            "localhost:8081",
          ),
          instanceId: data.prodDeployment.runtimeInstanceId,
          jwt: data?.jwt,
        };
      },
    },
  });
}

export function useProjectMembersEmails(organization: string, project: string) {
  return createAdminServiceListProjectMembers(
    organization,
    project,
    undefined,
    {
      query: {
        select: (data) => {
          return data.members
            ?.filter((member) => !!member?.userEmail)
            .map((member) => member.userEmail);
        },
      },
    },
  );
}

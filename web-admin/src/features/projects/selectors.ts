import {
  createAdminServiceGetProject,
  createAdminServiceListProjectMembers,
} from "@rilldata/web-admin/client";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

export function useProjectRuntime(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      // Proactively refetch the JWT because it's only valid for 1 hour
      refetchInterval: 1000 * 60 * 30, // 30 minutes
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

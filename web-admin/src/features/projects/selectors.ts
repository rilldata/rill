import {
  createAdminServiceGetProject,
  createAdminServiceListProjectMembers,
} from "@rilldata/web-admin/client";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

/**
 * Function: getProjectRuntimeQueryKey
 *
 * This function generates a unique query key for `GetProject` requests related to the project *runtime*.
 * This is a workaround to manage a side effect of calling `GetProject` for project *status*:
 * - a new JWT is returned
 * - the Runtime Store is updated, and all dependent stores are refreshed
 * - certain outstanding queries are cancelled and refetched (BAD!)
 *
 * By using a unique query key, requests and responses for the project runtime are managed separately
 * from those for the project status.
 *
 * Note: A better solution for the future would be to break the Runtime Store into more granular
 * stores. Then, we can update the JWT independently from the runtime's instanceID. This will prevent
 * unnecessary query cancellations and refetches.
 */
export function getProjectRuntimeQueryKey(orgName: string, projName: string) {
  return ["projectRuntime", orgName, projName];
}

export function useProjectRuntime(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      queryKey: getProjectRuntimeQueryKey(orgName, projName),
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

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(
    orgName,
    projectName,
    {},
    {
      query: {
        enabled: !!orgName && !!projectName,
        select: (resp) => resp.project?.id,
      },
    },
  );
}

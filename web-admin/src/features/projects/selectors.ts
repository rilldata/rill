import {
  V1DeploymentStatus,
  createAdminServiceGetProject,
  createAdminServiceListProjectMembers,
} from "@rilldata/web-admin/client";
import {
  V1ListResourcesResponse,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

export const PollTimeDuringReconcile = 1000;
export const PollTimeDuringError = 5000;
export const PollTimeWhenProjectReady = 60 * 1000;

export function useProjectDeploymentStatus(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1DeploymentStatus>(orgName, projName, {
    query: {
      select: (data) => {
        // There may not be a prodDeployment if the project was hibernated
        return (
          data?.prodDeployment?.status ||
          V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
        );
      },
      refetchInterval: (data) => {
        switch (data) {
          // case "DEPLOYMENT_STATUS_RECONCILING":
          case "DEPLOYMENT_STATUS_PENDING":
            return PollTimeDuringReconcile;

          case "DEPLOYMENT_STATUS_ERROR":
          case "DEPLOYMENT_STATUS_UNSPECIFIED":
            return PollTimeDuringError;

          case "DEPLOYMENT_STATUS_OK":
            return PollTimeWhenProjectReady;

          default:
            return PollTimeWhenProjectReady;
        }
      },
    },
  });
}

export function useProjectRuntime(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
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
            .map((member) => member.userEmail as string);
        },
      },
    },
  );
}

// This function returns the most recent refreshedOn date of all the project's resources.
// In the future, we really should display the refreshedOn date for all resources individually.
export function useProjectDataLastRefreshed(
  instanceId: string,
): CreateQueryResult<Date> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      enabled: !!instanceId,
      select: (data: V1ListResourcesResponse) => {
        const refreshedOns = data.resources.map((res) => {
          if (res.model?.state?.refreshedOn) {
            return new Date(res.model.state.refreshedOn).getTime();
          }
          if (res.source?.state?.refreshedOn) {
            return new Date(res.source.state.refreshedOn).getTime();
          }
          return 0;
        });
        const max = Math.max(...refreshedOns);
        return new Date(max);
      },
    },
  });
}

import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  createRuntimeServicePing,
  type V1ListResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data: { deployment?: V1Deployment }) => {
          // There may not be a deployment if the project is hibernating
          return data?.deployment;
        },
      },
    },
  );
}

export function useResources(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {},
    {
      query: {
        select: (data: V1ListResourcesResponse) => {
          const filtered = data?.resources?.filter(
            (resource) =>
              resource?.meta?.name?.kind !== ResourceKind.ProjectParser &&
              resource?.meta?.name?.kind !== ResourceKind.RefreshTrigger,
          );
          return {
            ...data,
            resources: filtered,
          };
        },
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

export function useRuntimeVersion() {
  return createRuntimeServicePing({
    query: {
      staleTime: 60000, // Cache for 1 minute
    },
  });
}

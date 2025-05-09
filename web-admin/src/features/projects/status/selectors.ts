import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { refetchIntervalStore } from "../../shared/refetch-interval";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data: { prodDeployment?: V1Deployment }) => {
          // There may not be a prodDeployment if the project is hibernating
          return data?.prodDeployment;
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
        select: (data) => ({
          ...data,
          // Filter out project parser and refresh triggers
          resources: data?.resources?.filter(
            (resource) =>
              resource.meta.name.kind !== ResourceKind.ProjectParser &&
              resource.meta.name.kind !== ResourceKind.RefreshTrigger,
          ),
        }),
        refetchInterval: refetchIntervalStore.calculatePollingInterval,
      },
    },
  );
}

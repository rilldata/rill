import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  createRuntimeServiceListResources,
  type V1ListResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  updateSmartRefetchMeta,
  INITIAL_REFETCH_INTERVAL,
} from "../../shared/refetch-interval-store";
import { get, writable } from "svelte/store";

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
  // Local store for per-query refetch interval
  const refetchIntervalStore = writable<number | false>(
    INITIAL_REFETCH_INTERVAL,
  );
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
          // Update the local refetch interval store
          const meta = updateSmartRefetchMeta(filtered, {
            refetchInterval: get(refetchIntervalStore),
          });
          refetchIntervalStore.set(meta.refetchInterval);
          return {
            ...data,
            resources: filtered,
          };
        },
        refetchInterval: () => get(refetchIntervalStore),
      },
    },
  );
}

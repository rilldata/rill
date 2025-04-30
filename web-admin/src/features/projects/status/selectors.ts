import {
  createAdminServiceGetProject,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  INITIAL_REFETCH_INTERVAL,
  MAX_REFETCH_INTERVAL,
  BACKOFF_FACTOR,
  isResourceReconciling,
  isResourceErrored,
} from "../../shared/refetch-interval";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data) => {
          // There may not be a prodDeployment if the project is hibernating
          return data?.prodDeployment;
        },
      },
    },
  );
}

export function useResources(instanceId: string) {
  let currentRefetchInterval = INITIAL_REFETCH_INTERVAL;

  return createRuntimeServiceListResources(
    instanceId,
    {
      // Ensure admins can see all resources, regardless of the security policy
      skipSecurityChecks: true,
    },
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
        refetchInterval: (query) => {
          if (query.state.error) return false;
          if (!query.state.data) return INITIAL_REFETCH_INTERVAL;

          const hasErrors = query.state.data.resources.some(isResourceErrored);
          const hasReconcilingResources = query.state.data.resources.some(
            isResourceReconciling,
          );

          if (hasErrors || !hasReconcilingResources) {
            currentRefetchInterval = INITIAL_REFETCH_INTERVAL;
            return false;
          }

          currentRefetchInterval = Math.min(
            currentRefetchInterval * BACKOFF_FACTOR,
            MAX_REFETCH_INTERVAL,
          );
          return currentRefetchInterval;
        },
      },
    },
  );
}

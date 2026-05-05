import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import {
  createSmartRefetchInterval,
  isResourceReconciling,
} from "@rilldata/web-admin/lib/refetch-interval-store";
import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

export function useDashboardsLastUpdated(
  client: RuntimeClient,
  organization: string,
  project: string,
) {
  return derived(
    [
      useValidExplores(client),
      createAdminServiceGetProject(organization, project),
    ],
    ([dashboardsResp, projResp]) => {
      if (!dashboardsResp.data?.length) {
        if (!projResp.data?.deployment?.updatedOn) return undefined;

        // return project's last updated if there are no dashboards
        return new Date(projResp.data.deployment.updatedOn);
      }

      const max = Math.max(
        ...dashboardsResp.data.map((res) =>
          new Date(res.meta!.stateUpdatedOn!).getTime(),
        ),
      );
      return new Date(max);
    },
  );
}

// Only poll while canvas/explore resources are reconciling. Without this
// filter, the unfiltered ListResources response includes all resource types;
// on dev/branch deployments the ProjectParser stays RUNNING indefinitely
// (it watches for file changes), causing perpetual polling.
const dashboardRefetchInterval = createSmartRefetchInterval(
  (res) => !!res.canvas || !!res.explore,
);

export function useDashboards(
  client: RuntimeClient,
): CreateQueryResult<V1Resource[]> {
  return createRuntimeServiceListResources(
    client,
    {},
    {
      query: {
        select: (data) => {
          return data.resources.filter((res) => res.canvas || res.explore);
        },
        enabled: !!client.instanceId,
        refetchInterval: dashboardRefetchInterval,
      },
    },
  );
}

/**
 * Returns true when the runtime is still in its initial build phase:
 * no dashboards exist yet AND non-parser resources are still reconciling.
 *
 * Used to show a "building" state instead of "no dashboards yet" during
 * deployment startup. In steady state (e.g., a model refresh on a project
 * that already has dashboards), this returns false because the dashboards
 * exist — even if other resources are reconciling.
 */
export function useIsInitialBuild(client: RuntimeClient) {
  return createRuntimeServiceListResources(
    client,
    {},
    {
      query: {
        select: (data): boolean => {
          const resources = data.resources;
          if (!resources?.length) return true;
          const hasDashboards = resources.some((r) => r.canvas || r.explore);
          if (hasDashboards) return false;
          return resources.some(
            (r) => !r.projectParser && isResourceReconciling(r),
          );
        },
        enabled: !!client.instanceId,
        refetchInterval: dashboardRefetchInterval,
      },
    },
  );
}

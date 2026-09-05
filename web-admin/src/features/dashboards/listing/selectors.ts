import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { createSmartRefetchInterval } from "@rilldata/web-admin/lib/refetch-interval-store";
import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
import type {
  V1ListResourcesResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
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
          res.meta?.stateUpdatedOn
            ? new Date(res.meta.stateUpdatedOn).getTime()
            : 0,
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
        select: (data) =>
          data.resources?.filter((res) => {
            if (res.canvas)
              return isManagedOrShared(res.canvas?.state?.validSpec ?? {});
            return !!res.explore;
          }) ?? [],
        enabled: !!client.instanceId,
        refetchInterval: dashboardRefetchInterval,
      },
    },
  );
}

function isManagedOrShared({
  annotations,
}: {
  annotations?: Record<string, string>;
}) {
  return !annotations?.admin_managed || annotations?.admin_shared === "true";
}

/**
 * Returns true when the runtime is still in its initial build phase, i.e. no dashboards exist yet
 * and the runtime reports that more resources may still appear.
 *
 * Used to show a "building" state instead of "no dashboards yet" during deployment startup. In
 * steady state (e.g., a model refresh on a project that already has dashboards), this returns false
 * because the dashboards exist — even if other resources are reconciling.
 *
 * An empty response is never enough to infer "still building" on its own: ListResources returns 200
 * with an empty list when security policies deny every resource, which a deny-by-default project
 * does for users who are entitled to nothing. Only `initializing` distinguishes the two.
 */
export function useIsInitialBuild(client: RuntimeClient) {
  return createRuntimeServiceListResources(
    client,
    {},
    {
      query: {
        select: isInitialBuild,
        enabled: !!client.instanceId,
        refetchInterval: dashboardRefetchInterval,
      },
    },
  );
}

export function isInitialBuild(data: V1ListResourcesResponse): boolean {
  const resources = data.resources ?? [];
  if (resources.some((r) => r.canvas || r.explore)) return false;
  return data.initializing ?? false;
}

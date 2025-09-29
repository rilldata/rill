import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { createSmartRefetchInterval } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useDashboardsLastUpdated(
  instanceId: string,
  organization: string,
  project: string,
) {
  return derived(
    [
      useValidExplores(instanceId),
      createAdminServiceGetProject(organization, project),
    ],
    ([dashboardsResp, projResp]) => {
      if (!dashboardsResp.data?.length) {
        if (!projResp.data?.prodDeployment?.updatedOn) return undefined;

        // return project's last updated if there are no dashboards
        return new Date(projResp.data.prodDeployment.updatedOn);
      }

      const max = Math.max(
        ...dashboardsResp.data.map((res) =>
          new Date(res.meta.stateUpdatedOn).getTime(),
        ),
      );
      return new Date(max);
    },
  );
}

export function useDashboards(
  instanceId: string,
): CreateQueryResult<V1Resource[]> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        return data.resources.filter((res) => res.canvas || res.explore);
      },
      refetchInterval: createSmartRefetchInterval,
    },
  });
}

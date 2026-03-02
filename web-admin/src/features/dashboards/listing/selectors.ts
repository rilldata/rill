import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { useValidExplores } from "@rilldata/web-common/features/dashboards/selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { derived } from "svelte/store";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

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
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

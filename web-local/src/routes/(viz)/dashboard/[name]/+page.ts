import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
} from "@rilldata/web-common/runtime-client";
import { error } from "@sveltejs/kit";
import type { QueryFunction } from "@tanstack/svelte-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

let ran = false;

export const load = async ({ params, depends, url }) => {
  const { instanceId } = get(runtime);

  const dashboardName = params.name;

  depends(dashboardName, "dashboard");

  const queryParams = {
    "name.kind": ResourceKind.MetricsView,
    "name.name": dashboardName,
  };

  const queryKey = getRuntimeServiceGetResourceQueryKey(
    instanceId,
    queryParams,
  );

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceGetResource>>
  > = ({ signal }) =>
    runtimeServiceGetResource(instanceId, queryParams, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const metricsViewResource = response.resource;

    if (!metricsViewResource?.metricsView) {
      throw error(404, "Dashboard not found");
    }

    const spec = metricsViewResource?.metricsView?.spec;
    const state = metricsViewResource?.metricsView?.state;

    if (!spec || !state || !dashboardName) {
      throw error(404, "Metrics view not found");
    }

    if (ran) {
      return {
        metricsView: metricsViewResource,
        initDimensions: new Map(),
      };
    }

    ran = true;

    const searchParams = new URLSearchParams(url.searchParams);

    const dimensions = spec.dimensions ?? [];

    const initDimensions = new Map<
      string,
      { exclude: boolean; values: string[] }
    >();

    dimensions.forEach(({ name }) => {
      if (!name) return;
      const valueStrings = searchParams.getAll(name);

      if (!valueStrings.length || !valueStrings) return;

      valueStrings.forEach((valueString) => {
        if (valueString) {
          const values = valueString.split(",") ?? [];

          if (values.length) {
            const exclude = values[0] === "!";

            if (exclude) {
              values.shift();
            }

            initDimensions.set(name, { exclude, values });
          }
        }
      });
    });

    return {
      metricsView: metricsViewResource,
      initDimensions,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Dashboard not found");
  }
};

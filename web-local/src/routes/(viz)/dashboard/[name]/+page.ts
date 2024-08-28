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

    const rawFilter = searchParams.get("filter") ?? "";

    let filter = "";

    if (
      (rawFilter.startsWith(`"`) && rawFilter.endsWith(`"`)) ||
      (rawFilter.startsWith(`'`) && rawFilter.endsWith(`'`)) ||
      (rawFilter.startsWith(`“`) && rawFilter.endsWith(`”`))
    ) {
      filter = rawFilter.slice(1, -1);
    } else {
      filter = rawFilter;
    }

    const dimensions = spec.dimensions ?? [];

    const dimensionNames = dimensions.map(({ name }) => name).filter(isDefined);

    const { initDimensions, errorMessage } = parseFilterString(
      filter,
      dimensionNames,
    );

    return {
      metricsView: metricsViewResource,
      initDimensions,
      errorMessage,
      filter,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Dashboard not found");
  }
};

const conditionRegex = /(\w+)\s+(eq|ne)\s+(.+)/;
const dimensionValueRegex = /[‘'’]([^‘'’]*)[‘'’]/g;

function parseFilterString(filterString: string, dimensions: string[]) {
  const conditions = filterString.split(" and ");

  const initDimensions = new Map<
    string,
    { exclude: boolean; values: string[] }
  >();

  let errorMessage: string | null = null;

  conditions.forEach((condition) => {
    const match = condition.match(conditionRegex);
    if (match) {
      const [, dimension, operator, valueString] = match;

      const values: string[] = [];

      if (valueString.startsWith("(") && valueString.endsWith(")")) {
        let match: RegExpExecArray | null;

        while ((match = dimensionValueRegex.exec(valueString)) !== null) {
          values.push(match[1]);
        }
      } else if (valueString.match(dimensionValueRegex)) {
        values.push(valueString.slice(1, -1));
      } else {
        errorMessage = `Value missing quotes: ${valueString}`;
        return;
      }

      if (!dimensions.includes(dimension)) {
        errorMessage = `Invalid value string: ${valueString}`;
        return;
      } else if (values.length === 0) {
        errorMessage = `Invalid values: ${valueString}`;
        return;
      }

      const exclude = operator === "ne" || operator === "nin";

      initDimensions.set(dimension, {
        exclude,
        values,
      });
    } else {
      errorMessage = `Invalid condition. Expected format: <dimension> <eq|ne> ('<value>', '<value>')`;
    }
  });

  return {
    initDimensions,
    errorMessage,
  };
}

function isDefined(value: string | undefined): value is string {
  return value !== undefined;
}

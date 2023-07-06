import {
  createRuntimeServiceGetCatalogEntry,
  type V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import type { BusinessModel } from "../business-model/business-model";
import { derived, writable } from "svelte/store";
import type { CreateQueryResult } from "@tanstack/svelte-query";

export const useMetaQuery = <T = V1MetricsView>(
  ctx: BusinessModel,
  selector?: (meta: V1MetricsView) => T
) => {
  const test = writable(1);
  setTimeout(() => {
    test.set(2);
  }, 5000);
  return derived(
    [ctx.runtime, ctx.metricsViewName, test],
    ([runtime, metricViewName, t], set) => {
      return createRuntimeServiceGetCatalogEntry(
        runtime.instanceId,
        metricViewName,
        {
          query: {
            select: (data) =>
              selector
                ? selector(data?.entry?.metricsView)
                : data?.entry?.metricsView,
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set);
    }
  );
};

export const useModelHasTimeSeries = (ctx: BusinessModel) =>
  useMetaQuery(
    ctx,
    (meta) => !!meta?.timeDimension
  ) as CreateQueryResult<boolean>;

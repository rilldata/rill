import { createRuntimeServiceGetResource } from "@rilldata/web-common/runtime-client";

export enum ResourceKind {
  Source = "source",
  Model = "model",
  MetricsView = "metricsview",
}

export function useSource(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.Source,
      "name.name": name,
    },
    {
      query: {
        select: (data) => data?.resource?.source,
      },
    }
  );
}

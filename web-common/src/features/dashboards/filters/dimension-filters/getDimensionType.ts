import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";

export function getDimensionType(
  instanceId: string,
  metricsViewName: string,
  dimensionName: string,
) {
  return createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
    {},
    {
      query: {
        select: (data) =>
          data.schema?.fields?.find((f) => f.name === dimensionName)?.type
            ?.code as string,
      },
    },
  );
}

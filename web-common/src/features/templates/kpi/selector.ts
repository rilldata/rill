import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";

export function useKPITotals(
  instanceId: string,
  metricViewName: string,
  measure: string,
  timeRange: string,
) {
  return createQueryServiceMetricsViewAggregation(
    instanceId,
    metricViewName,
    {
      measures: [{ name: measure }],
      timeRange: { isoDuration: timeRange },
    },
    {
      query: {
        select: (data) => {
          return data.data?.[0]?.[measure] ?? null;
        },
      },
    },
  );
}

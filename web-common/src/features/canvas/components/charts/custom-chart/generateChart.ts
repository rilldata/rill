import { runtimeServiceGenerateChart } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export interface GenerateChartParams {
  prompt: string;
  previousSql?: string[];
  previousSpec?: string;
}

export interface GenerateChartResult {
  metricsSql: string[];
  vegaSpec: string;
}

export async function generateChart(
  client: RuntimeClient,
  params: GenerateChartParams,
  signal?: AbortSignal,
): Promise<GenerateChartResult> {
  const response = await runtimeServiceGenerateChart(
    client,
    {
      prompt: params.prompt,
      previousSql: params.previousSql ?? [],
      previousSpec: params.previousSpec ?? "",
    },
    { signal },
  );

  const metricsSql = (response.metricsSql as string[]) ?? [];
  const vegaSpec = (response.vegaSpec as string) ?? "";

  if (metricsSql.length === 0 || !vegaSpec) {
    throw new Error("AI returned an empty chart specification");
  }

  return { metricsSql, vegaSpec };
}

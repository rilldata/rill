import type { PartialMessage, Struct } from "@bufbuild/protobuf";
import {
  buildDefaultArgs,
  paramChannels,
  VALUE_COUNT_PROBE_LIMIT,
} from "@rilldata/web-common/features/custom-viz/params";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  MetricsViewSpecDimensionType,
  type V1ComponentParam,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  runtimeServiceGetResource,
  runtimeServiceListResources,
  runtimeServiceQueryResolver,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";

/**
 * Picks the test values a component's preview renders with, from the metrics view's own data.
 *
 * buildDefaultArgs decides which field suits each channel; this supplies what it needs to decide:
 * how many distinct values each dimension has. A metrics view's dimensions can be expressions
 * rather than columns, so the counts come from small metrics SQL probes rather than a column
 * profile.
 */

/** Bounds the probes for metrics views with very many dimensions. */
const MAX_PROBED_DIMENSIONS = 40;

const valueCountCache = new Map<string, Promise<Map<string, number>>>();

/**
 * Distinct values per dimension of a metrics view, capped at VALUE_COUNT_PROBE_LIMIT. Cached for
 * the session; a dimension whose probe fails is left out rather than failing the whole lookup.
 */
export function fetchDimensionValueCounts(
  client: RuntimeClient,
  metricsViewName: string,
  metricsViewSpec: V1MetricsViewSpec,
): Promise<Map<string, number>> {
  const key = `${client.instanceId}/${metricsViewName}`;
  let counts = valueCountCache.get(key);
  if (!counts) {
    counts = probeValueCounts(client, metricsViewName, metricsViewSpec);
    valueCountCache.set(key, counts);
  }
  return counts;
}

async function probeValueCounts(
  client: RuntimeClient,
  metricsViewName: string,
  metricsViewSpec: V1MetricsViewSpec,
): Promise<Map<string, number>> {
  const dimensions = (metricsViewSpec.dimensions ?? [])
    .filter(
      (dimension) =>
        dimension.name &&
        dimension.name !== metricsViewSpec.timeDimension &&
        dimension.type !== MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME,
    )
    .slice(0, MAX_PROBED_DIMENSIONS);

  const counts = new Map<string, number>();
  await Promise.all(
    dimensions.map(async ({ name }) => {
      try {
        // Metrics SQL groups by the selected dimension, so each row is one distinct value.
        const res = await runtimeServiceQueryResolver(client, {
          resolver: "metrics_sql",
          resolverProperties: {
            sql: `SELECT ${name} FROM ${metricsViewName} LIMIT ${VALUE_COUNT_PROBE_LIMIT}`,
          } as unknown as PartialMessage<Struct>,
        });
        counts.set(name!, res.data?.length ?? 0);
      } catch {
        // Left out: the picker treats an unknown count as unsuitable for a channel.
      }
    }),
  );
  return counts;
}

/**
 * The metrics view a new component's preview binds to: the first valid one with a time
 * dimension, else the first valid one. Mirrors useDefaultMetrics for callers outside a component.
 */
export async function fetchDefaultMetricsView(
  client: RuntimeClient,
): Promise<{ name: string; spec: V1MetricsViewSpec } | undefined> {
  const res = await runtimeServiceListResources(client, {
    kind: ResourceKind.MetricsView,
  });
  const valid = (res.resources ?? []).filter(
    (resource) =>
      resource.meta?.name?.name && resource.metricsView?.state?.validSpec,
  );
  const chosen =
    valid.find(
      (resource) => resource.metricsView?.state?.validSpec?.timeDimension,
    ) ?? valid[0];
  if (!chosen) return undefined;
  return {
    name: chosen.meta!.name!.name!,
    spec: chosen.metricsView!.state!.validSpec!,
  };
}

/**
 * Test values for a component's params, suited to the channel each is drawn on. Fields are picked
 * from the metrics view the component will query: its metrics_view param's default when that is a
 * valid metrics view, else the given fallback.
 */
export async function pickTestArgs(
  client: RuntimeClient,
  params: V1ComponentParam[],
  rendererProperties: Record<string, unknown> | undefined,
  fallbackMetricsView: { name: string; spec: V1MetricsViewSpec },
): Promise<Record<string, unknown>> {
  const metricsViewParam = params.find(
    (param) => param.type === "metrics_view",
  );
  const metricsView =
    (await fetchMetricsView(client, metricsViewParam?.default)) ??
    fallbackMetricsView;

  const valueCounts = params.some((param) => param.type === "dimension")
    ? await fetchDimensionValueCounts(
        client,
        metricsView.name,
        metricsView.spec,
      )
    : undefined;
  const spec = rendererProperties?.spec as { chartType?: unknown } | undefined;
  const args = buildDefaultArgs(params, metricsView.name, metricsView.spec, {
    channels: paramChannels(rendererProperties),
    chartType: typeof spec?.chartType === "string" ? spec.chartType : undefined,
    valueCounts,
  });
  // The fields were picked from this metrics view, so bind it even over an unusable default.
  if (metricsViewParam?.name) args[metricsViewParam.name] = metricsView.name;
  return args;
}

async function fetchMetricsView(
  client: RuntimeClient,
  name: unknown,
): Promise<{ name: string; spec: V1MetricsViewSpec } | undefined> {
  if (typeof name !== "string" || !name) return undefined;
  try {
    const res = await runtimeServiceGetResource(client, {
      name: { kind: ResourceKind.MetricsView, name },
    });
    const spec = res.resource?.metricsView?.state?.validSpec;
    return spec ? { name, spec } : undefined;
  } catch {
    return undefined;
  }
}

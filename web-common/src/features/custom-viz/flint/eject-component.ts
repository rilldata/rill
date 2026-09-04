import { compileFlintSpec, type FlintChartSpec } from "./compile";
import { ejectToVegaSpec, type EjectFieldBinding } from "./eject";
import { deriveFlintFields } from "./semantic-types";
import {
  getDeclaredParams,
  type ComponentParamType,
} from "@rilldata/web-common/features/custom-viz/params";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";
import {
  V1TimeGrain,
  type V1ComponentParam,
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { queryServiceResolveComponent } from "@rilldata/web-common/runtime-client/v2/gen/query-service";
import {
  runtimeServiceGetResource,
  runtimeServiceQueryResolver,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { PartialMessage, Struct } from "@bufbuild/protobuf";

/**
 * The grain an ejected chart is compiled at. A component file has no dashboard time controls to
 * inherit a grain from, so this matches what its preview renders at.
 */
const EJECT_TIME_GRAIN = V1TimeGrain.TIME_GRAIN_DAY;

/**
 * Ejects a component's chart spec into a Vega-Lite spec.
 *
 * Ejection needs the compiled Vega-Lite, and the compiler only produces that from real rows: it
 * reads the data to pick scales, sorting, stacking and layout. So this resolves the component
 * against the given args, runs its query, compiles, and turns the resolved bindings back into
 * template references (see ejectToVegaSpec).
 *
 * It returns the string to write to the component's `custom_chart.vega_spec`; the caller owns the
 * file edit.
 */
export async function ejectComponent(
  client: RuntimeClient,
  resource: V1Resource,
  componentName: string,
  args: Record<string, unknown>,
): Promise<string> {
  const params = getDeclaredParams(resource, true);

  // Resolve through the server rather than substituting client-side, so the chart spec that gets
  // compiled is the one a dashboard would render.
  const resolved = await queryServiceResolveComponent(client, {
    component: componentName,
    args: args as PartialMessage<Struct>,
  });
  const props = resolved.rendererProperties as
    | Record<string, unknown>
    | undefined;

  const flintSpec = props?.spec as FlintChartSpec | undefined;
  if (!flintSpec?.chartType) {
    throw new Error("This component has no chart spec to eject.");
  }
  const metricsSQL = normalizeMetricsSQL(props?.metrics_sql);
  if (!metricsSQL) {
    throw new Error("This component has no Metrics SQL query to run.");
  }

  const data = await runtimeServiceQueryResolver(client, {
    resolver: "metrics_sql",
    resolverProperties: {
      sql: metricsSQL,
      additional_time_grain: V1TimeGrainToDateTimeUnit[EJECT_TIME_GRAIN],
    } as unknown as PartialMessage<Struct>,
  });
  const rows = (data.data ?? []) as Record<string, unknown>[];
  const columns = (data.schema?.fields ?? []).map(
    (field) => field.name as string,
  );

  const metricsViewSpec = await fetchMetricsViewSpec(
    client,
    boundMetricsViewName(params, args),
  );
  const fields = deriveFlintFields(columns, metricsViewSpec, EJECT_TIME_GRAIN);
  const measureFields = (metricsViewSpec?.measures ?? [])
    .map((measure) => measure.name as string)
    .filter((name) => name && columns.includes(name));

  // Compiled without a size, so the output does not depend on how large the editor happened to be;
  // the fixed view size the compiler fills in is stripped on the way out anyway.
  const compiled = compileFlintSpec({
    spec: flintSpec,
    rows,
    fields,
    measureFields,
  });
  if (compiled.error || !compiled.spec) {
    throw new Error(compiled.error ?? "The chart spec failed to compile.");
  }

  return ejectToVegaSpec(
    compiled.spec as unknown as Record<string, unknown>,
    fieldBindings(params, args, fields.field_display_names, measureFields),
  );
}

/**
 * Pairs each field-typed param with the field it is bound to and that field's display name, which
 * together are what ejection looks for in the compiled spec. The display names come from what the
 * compiler was handed rather than straight from the metrics view, so a label the compiler
 * substituted (such as "Time" for a timestamp column) is recognized.
 */
function fieldBindings(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
  displayNames: Record<string, string>,
  measureFields: string[],
): EjectFieldBinding[] {
  const bindings: EjectFieldBinding[] = [];
  for (const param of params) {
    const type = param.type as ComponentParamType;
    if (
      !param.name ||
      (type !== "measure" && type !== "dimension" && type !== "time_dimension")
    ) {
      continue;
    }
    const field = args[param.name] ?? param.default;
    if (typeof field !== "string" || !field) continue;
    bindings.push({
      param: param.name,
      field,
      displayName: displayNames[field],
      // Only a measure has a Rill number formatter registered for it, so only a measure's
      // formatter name can appear in the compiled spec. The metrics view has the last word:
      // a component may declare a param loosely and still bind it to a measure.
      type: measureFields.includes(field) ? "measure" : type,
    });
  }
  return bindings;
}

async function fetchMetricsViewSpec(
  client: RuntimeClient,
  name: string | undefined,
): Promise<V1MetricsViewSpec | undefined> {
  if (!name) return undefined;
  const res = await runtimeServiceGetResource(client, {
    name: { kind: ResourceKind.MetricsView, name },
  });
  return res.resource?.metricsView?.state?.validSpec;
}

/** The metrics view bound to the component's first metrics_view param. */
function boundMetricsViewName(
  params: V1ComponentParam[],
  args: Record<string, unknown>,
): string | undefined {
  const param = params.find((candidate) => candidate.type === "metrics_view");
  if (!param?.name) return undefined;
  const name = args[param.name] ?? param.default;
  return typeof name === "string" && name ? name : undefined;
}

function normalizeMetricsSQL(value: unknown): string | undefined {
  if (typeof value === "string") return value.trim() ? value : undefined;
  if (Array.isArray(value)) {
    return value.find(
      (entry): entry is string =>
        typeof entry === "string" && entry.trim().length > 0,
    );
  }
  return undefined;
}

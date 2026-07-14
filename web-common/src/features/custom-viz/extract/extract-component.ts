import type { ComponentRefComponent } from "@rilldata/web-common/features/canvas/components/component-ref";
import type {
  CustomChartComponent,
  QueryFieldMeta,
} from "@rilldata/web-common/features/canvas/components/charts/custom-chart";
import {
  CUSTOM_VIZ_FOLDER,
  getDeclaredParams,
  substituteArgsRecursively,
} from "@rilldata/web-common/features/custom-viz/params";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";
import { Document } from "yaml";

/**
 * Extraction and its inverse: converting an inline custom chart into a standalone,
 * parameterized component ("Save as custom viz"), and materializing a component
 * reference back into an inline custom chart ("Detach copy").
 *
 * Extraction lifts params at the string level: the metrics view is templated by
 * rewriting the FROM clause, and lifted fields by word-boundary replacement in both
 * the metrics SQL and the Vega spec (Vega encodings reference the SQL output columns,
 * so one replacement rule covers both). Anything ambiguous stays hardcoded and can be
 * refined later in the component workspace.
 */

export interface ExtractResult {
  filePath: string;
  componentName: string;
}

export interface LiftedField extends QueryFieldMeta {
  // Whether the field is the metrics view's time dimension.
  isTimeDimension?: boolean;
}

export function extractedComponentYAML(
  spec: { metrics_sql?: string[]; vega_spec?: string },
  metricsViewName: string | undefined,
  displayName: string,
  liftedFields: LiftedField[],
): { yaml: string; bindings: Record<string, unknown> } {
  let metricsSQL = [...(spec.metrics_sql ?? [])];
  let vegaSpec = spec.vega_spec ?? "";

  const params: Record<string, unknown>[] = [
    { name: "metrics_view", type: "metrics_view", required: true },
  ];
  const bindings: Record<string, unknown> = { metrics_view: metricsViewName };

  if (metricsViewName) {
    const fromClause = new RegExp(
      `\\bFROM\\s+${escapeRegExp(metricsViewName)}\\b`,
      "gi",
    );
    metricsSQL = metricsSQL.map((sql) =>
      sql.replace(fromClause, "FROM {{ .params.metrics_view }}"),
    );
  }

  const usedNames = new Set(["metrics_view"]);
  for (const field of liftedFields) {
    const paramName = uniqueParamName(sanitizeParamName(field.name), usedNames);
    usedNames.add(paramName);

    params.push({
      name: paramName,
      type:
        field.type === "measure"
          ? "measure"
          : field.isTimeDimension
            ? "time_dimension"
            : "dimension",
      required: true,
    });
    bindings[paramName] = field.name;

    const token = new RegExp(`\\b${escapeRegExp(field.name)}\\b`, "g");
    const placeholder = `{{ .params.${paramName} }}`;
    metricsSQL = metricsSQL.map((sql) => sql.replace(token, placeholder));
    vegaSpec = vegaSpec.replace(token, placeholder);
  }

  const doc = new Document({
    type: "component",
    display_name: displayName,
    params,
    custom_chart: {
      metrics_sql: metricsSQL,
      vega_spec: vegaSpec,
    },
  });

  return { yaml: doc.toString(), bindings };
}

/**
 * Writes the extracted component to viz_library/<name>.yaml and replaces the inline
 * canvas item with a reference to it, preserving the item's width.
 */
export async function extractCustomChart(
  client: RuntimeClient,
  component: CustomChartComponent,
  componentName: string,
  displayName: string,
  liftedFields: LiftedField[],
): Promise<ExtractResult> {
  const { yaml, bindings } = extractedComponentYAML(
    get(component.specStore),
    component.metricsViewName,
    displayName,
    liftedFields,
  );

  const filePath = `/${CUSTOM_VIZ_FOLDER}/${componentName}.yaml`;
  await runtimeServicePutFile(client, {
    path: filePath,
    blob: yaml,
    create: true,
    createOnly: true,
  });

  replaceItem(component, { component: componentName, params: bindings });

  return { filePath, componentName };
}

/**
 * The inverse of extraction: replaces a component reference with an inline custom
 * chart whose param placeholders are materialized with the item's current bindings
 * (declared defaults included). Other template namespaces (.env, .user) are preserved.
 */
export function detachComponentRef(component: ComponentRefComponent) {
  const resource = get(component.resource);
  const componentSpec =
    resource?.component?.state?.validSpec ?? resource?.component?.spec;
  if (!componentSpec?.rendererProperties) return;

  // Merge declared defaults under the bound values.
  const args: Record<string, unknown> = {};
  for (const param of getDeclaredParams(resource, true)) {
    if (!param.name) continue;
    if (param.default !== undefined && param.default !== null) {
      args[param.name] = param.default;
    }
  }
  Object.assign(args, component.args());

  const inlineProps = substituteArgsRecursively(
    componentSpec.rendererProperties,
    args,
  ) as Record<string, unknown>;

  replaceItem(component, {
    [componentSpec.renderer ?? "custom_chart"]: inlineProps,
  });
}

/**
 * Replaces the canvas item at the component's YAML path with the given item map,
 * preserving its width.
 */
function replaceItem(
  component: CustomChartComponent | ComponentRefComponent,
  item: Record<string, unknown>,
) {
  const parent = component.parent;
  if (!parent.fileArtifact) return;
  const parsedDocument = get(parent.parsedContent);

  const itemPath = component.pathInYAML.slice(0, -1);
  const width = parsedDocument.getIn([...itemPath, "width"]);

  parsedDocument.setIn(itemPath, {
    ...item,
    ...(width !== undefined ? { width } : {}),
  });

  parent.fileArtifact.updateEditorContent(parsedDocument.toString(), false, true);
}

function sanitizeParamName(name: string): string {
  const sanitized = name.replace(/\W/g, "_").replace(/^(\d)/, "_$1");
  return sanitized || "field";
}

function uniqueParamName(name: string, used: Set<string>): string {
  if (!used.has(name)) return name;
  for (let i = 2; ; i++) {
    const candidate = `${name}_${i}`;
    if (!used.has(candidate)) return candidate;
  }
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

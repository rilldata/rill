import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  CanvasEntity,
  ComponentPath,
} from "@rilldata/web-common/features/canvas/stores/canvas-entity";
import {
  getDeclaredParams,
  isOrderByParam,
  orderByOptions,
  paramToInputParam,
  prettyParamLabel,
  reconcileOrderByArg,
} from "@rilldata/web-common/features/custom-viz/params";
import type {
  V1CanvasItem,
  V1ComponentParam,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import CanvasComponentRef from "./CanvasComponentRef.svelte";

/**
 * The flattened spec of a canvas item that references an externally defined component:
 * the component name plus the values bound to its declared params spread at the top
 * level, so the generated inspector inputs can bind to them like ordinary spec keys.
 */
export interface ComponentRefSpec {
  component: string;
  [param: string]: unknown;
}

/**
 * Canvas component for items that reference an externally defined component
 * (`component: <name>` + `params:` in the canvas YAML). Unlike other canvas
 * components, its editable state is the canvas item's param bindings, not the
 * component's renderer properties, and its inspector inputs are generated from
 * the referenced component's declared params.
 */
export class ComponentRefComponent extends BaseCanvasComponent<ComponentRefSpec> {
  minSize = { width: 4, height: 4 };
  defaultSize = { width: 6, height: 4 };
  resetParams = [];
  type: CanvasComponentType = "component_ref";
  component = CanvasComponentRef;

  componentName: string;

  constructor(
    resource: V1Resource,
    parent: CanvasEntity,
    path: ComponentPath,
    item?: V1CanvasItem,
  ) {
    super(resource, parent, path, { component: item?.component ?? "" });
    this.componentName = item?.component ?? "";
    this.specStore.set(this.specFromItem(item));
    this.syncMetricsViewName();
  }

  override update(resource: V1Resource, path: ComponentPath, item?: V1CanvasItem) {
    super.update(resource, path);
    if (item?.component) this.componentName = item.component;
    this.specStore.set(this.specFromItem(item));
    this.syncMetricsViewName();
  }

  /** The referenced component's declared params. */
  get declaredParams(): V1ComponentParam[] {
    return getDeclaredParams(get(this.resource), this.parent.allowUnvalidatedSpec);
  }

  /**
   * The values currently bound to the declared params.
   * Filters to declared names, since the spec may carry derived keys
   * (e.g. the metrics_view surfaced for the canvas filter system).
   */
  args(spec: ComponentRefSpec = get(this.specStore)): Record<string, unknown> {
    const declared = new Set(this.declaredParams.map((p) => p.name));
    return Object.fromEntries(
      Object.entries(spec).filter(
        ([key, value]) => declared.has(key) && value !== undefined,
      ),
    );
  }

  isValid(spec: ComponentRefSpec): boolean {
    return !!spec.component;
  }

  inputParams(): InputParams<ComponentRefSpec> {
    const options: Record<string, ReturnType<typeof paramToInputParam>> = {};
    const args = this.args();
    for (const param of this.declaredParams) {
      if (!param.name) continue;
      // The by-convention order_by param selects among the currently bound
      // field values instead of a free-text input.
      if (isOrderByParam(param)) {
        options[param.name] = {
          type: "select",
          label: prettyParamLabel(param.name),
          optional: !param.required,
          description: param.description,
          meta: { options: orderByOptions(this.declaredParams, args) },
        };
        continue;
      }
      options[param.name] = paramToInputParam(param);
    }
    return { options, filter: {} };
  }

  override updateProperty(
    key: AllKeys<ComponentRefSpec>,
    value: ComponentRefSpec[AllKeys<ComponentRefSpec>],
  ) {
    const currentSpec = get(this.specStore);
    const newSpec = { ...currentSpec, [key]: value };
    if (value === undefined || value === "") {
      delete newSpec[key];
    }

    // When the bound metrics view changes, clear the bound field params that
    // depend on it: they referenced fields of the previous metrics view.
    for (const param of this.declaredParams) {
      if (param.type === "metrics_view" && param.name === key) {
        for (const dependent of this.declaredParams) {
          if (dependent.metricsViewParam === key && dependent.name) {
            delete newSpec[dependent.name];
          }
        }
      }
    }

    // Keep the by-convention order_by binding in step: it follows a re-bound
    // field it pointed at, and is re-seeded from the measure when it no longer
    // matches any bound field (e.g. after a metrics view change).
    reconcileOrderByArg(this.declaredParams, newSpec, {
      name: String(key),
      previousValue: currentSpec[key],
    });

    this.writeParamsToYAML(this.args(newSpec));
    this.specStore.set(newSpec);
    this.syncMetricsViewName();
  }

  override setSpec(newSpec: ComponentRefSpec) {
    this.writeParamsToYAML(this.args(newSpec));
    this.specStore.set(newSpec);
    this.syncMetricsViewName();
  }

  static newComponentSpec(): ComponentRefSpec {
    return { component: "" };
  }

  private specFromItem(item: V1CanvasItem | undefined): ComponentRefSpec {
    return {
      component: item?.component ?? this.componentName,
      ...(item?.params ?? {}),
    };
  }

  /**
   * Persists the param bindings to the canvas YAML at the item's `params:` key.
   * Only the params map is written; the item's other keys (component, width) stay untouched.
   */
  private writeParamsToYAML(params: Record<string, unknown>) {
    if (!this.parent.fileArtifact) return;
    const parsedDocument = get(this.parent.parsedContent);
    const { updateEditorContent } = this.parent.fileArtifact;

    // pathInYAML ends with the component type segment; the item map is its parent.
    const itemPath = this.pathInYAML.slice(0, -1);
    if (Object.keys(params).length > 0) {
      parsedDocument.setIn([...itemPath, "params"], params);
    } else {
      parsedDocument.deleteIn([...itemPath, "params"]);
    }

    updateEditorContent(parsedDocument.toString(), false, true);
  }

  /**
   * Tracks the metrics view bound to the first metrics_view-typed param so the
   * canvas filter and time systems know which metrics view this item uses.
   */
  private syncMetricsViewName() {
    const spec = get(this.specStore);
    const mvParam = this.declaredParams.find((p) => p.type === "metrics_view");
    const name = mvParam?.name
      ? (spec[mvParam.name] ?? mvParam.default)
      : undefined;
    if (typeof name !== "string" || !name || name === this.metricsViewName) {
      return;
    }
    this.metricsViewName = name;
    this.localFilters = this.parent.filterManager.createLocalFilterStore(name);
    // Surface the metrics view under the key the shared canvas systems read.
    this.specStore.update((s) => ({ ...s, metrics_view: name }));
  }
}

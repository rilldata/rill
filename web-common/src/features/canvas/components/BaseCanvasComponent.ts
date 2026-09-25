import type {
  CanvasComponentType,
  ComponentSize,
  ComponentSpec,
} from "@rilldata/web-common/features/canvas/components/types";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import type {
  V1Expression,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { Component, ComponentType, SvelteComponent } from "svelte";
import type { Readable, Unsubscriber } from "svelte/store";
import { derived, get, writable, type Writable } from "svelte/store";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  DEFAULT_INHERIT_URL_PARAMS,
  TimeFilterManager,
} from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { dedupe } from "@rilldata/web-common/lib/arrayUtils.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Widget specific dimension and measure filters
  localExpressionFilters: ExpressionFilterManager;
  // Widget specific time filters
  localTimeFilters: TimeFilterManager;

  metricsViewsProvider: MetricsViewsProvider;
  yamlConfigProvider: YAMLConfigProvider;

  // Final expression filter manager based on parent and local
  expressionFilters: ExpressionFilterManager;
  // Final time filter manager based on parent and local
  timeFilters: TimeFilterManager;

  // Lazy-load latch: flipped true once the component scrolls into view, gating
  // its data query so off-screen components don't fetch.
  visible = writable(false);

  // Whether the component's data query should run: true when the component is
  // visible or while the canvas is exporting to PDF. Export uses this rather than
  // forcing `visible` so it never mutates the live lazy-load state.
  dataEnabled: Readable<boolean>;

  // Tears down the spec subscription opened in the constructor. Without this,
  // a component replaced in CanvasEntity.processRows keeps reacting to spec
  // emissions and mutates the shared filter/time state of a deleted widget.
  private unsubscribeSpec: Unsubscriber;
  private unsubExpressionSync: Unsubscriber;

  abstract type: CanvasComponentType;
  // Component responsible for DOM rendering.
  // The union covers both Svelte 4 class components and Svelte 5 (runes) components,
  // since the display components are being migrated one at a time.
  abstract component: Component<any> | ComponentType<SvelteComponent>;
  // Will be deprecated
  abstract minSize: ComponentSize;
  // Will be deprecated
  abstract defaultSize: ComponentSize;
  // Parameters to reset when the metrics_view changes
  abstract resetParams: string[];
  // Minimum condition needed for the component to be rendered
  abstract isValid(spec: T): boolean;
  // Configuration for the sidebar editor
  abstract inputParams(type?: CanvasComponentType): InputParams<T>;

  getExploreTransformerProperties?(): Partial<ExploreState>;

  metricsViewName: string;

  constructor(
    resource: V1Resource,
    public parent: CanvasEntity,
    path: ComponentPath,
    public defaultSpec: T,
  ) {
    const yamlSpec =
      resource.component?.state?.validSpec?.rendererProperties ??
      (parent.allowUnvalidatedSpec
        ? resource.component?.spec?.rendererProperties
        : undefined);

    const mergedSpec = { ...defaultSpec, ...yamlSpec };
    this.metricsViewName = mergedSpec["metrics_view"] as string;
    this.specStore = writable(mergedSpec);
    this.pathInYAML = path;

    this.resource.set(resource);
    this.id = resource.meta?.name?.name as string;

    this.dataEnabled = derived(
      [this.visible, this.parent.exportMode],
      ([visible, exportMode]) => visible || exportMode,
    );

    this.metricsViewsProvider = new MetricsViewsProvider(this.parent.client, [
      this.metricsViewName,
    ]);
    this.yamlConfigProvider = this.parent.dashboardProvider.yamlConfigProvider;

    this.localExpressionFilters = new ExpressionFilterManager(
      this.metricsViewsProvider,
      this.yamlConfigProvider,
      false,
    );

    this.localTimeFilters = new TimeFilterManager(
      this.parent.client,
      this.metricsViewsProvider,
      this.yamlConfigProvider,
      {
        ...this.parent.timeFilterManager.config,
        skipTimeGrain: true,
        log: true,
      },
      this.parent.timeFilterManager,
    );

    this.expressionFilters = new ExpressionFilterManager(
      this.metricsViewsProvider,
      this.yamlConfigProvider,
    );

    this.timeFilters = this.localTimeFilters;

    this.unsubscribeSpec = this.specStore.subscribe((spec) => {
      this.localExpressionFilters.setParamForMetricsView(
        this.metricsViewName,
        (spec["dimension_filters"] ?? "") as string,
      );
      this.syncExpressionFilters();

      this.localTimeFilters.setUrlParams(
        spec?.["time_filters"]
          ? new URLSearchParams(spec["time_filters"])
          : new URLSearchParams(DEFAULT_INHERIT_URL_PARAMS),
      );
    });

    this.unsubExpressionSync = this.parent.expressionFilterManager.storeSync.on(
      "change",
      () => this.syncExpressionFilters(),
    );
  }

  destroy() {
    this.unsubscribeSpec?.();
    this.unsubExpressionSync?.();
    this.metricsViewsProvider.cleanup();
    this.yamlConfigProvider.cleanup?.();
  }

  update(resource: V1Resource, path: ComponentPath) {
    const yamlSpec = (resource.component?.state?.validSpec
      ?.rendererProperties ??
      (this.parent.allowUnvalidatedSpec
        ? resource.component?.spec?.rendererProperties
        : undefined)) as T;
    this.resource.set(resource);
    this.pathInYAML = path;
    this.specStore.set(yamlSpec);
  }

  public syncExpressionFilters() {
    const globalExpr =
      this.parent.expressionFilterManager.exprByMetricsView[
        this.metricsViewName
      ];
    const localExpr =
      this.localExpressionFilters.exprByMetricsView[this.metricsViewName];

    let resolvedExpr: V1Expression | undefined;
    if (globalExpr && localExpr) {
      resolvedExpr = mergeFilters(globalExpr, localExpr);
    } else {
      resolvedExpr = globalExpr ?? localExpr;
    }

    this.expressionFilters.setExprForMetricsView(
      this.metricsViewName,
      resolvedExpr,
      dedupe([
        ...this.parent.expressionFilterManager.inList,
        ...this.localExpressionFilters.inList,
      ]),
    );
  }

  private updateYAML(newSpec: T) {
    if (!this.parent.fileArtifact) return;
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent } = this.parent.fileArtifact;

    parsedDocument.setIn(this.pathInYAML, newSpec);

    updateEditorContent(parsedDocument.toString(), false, true);
  }

  setSpec(newSpec: T) {
    if (this.isValid(newSpec)) {
      this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  updateProperty(key: AllKeys<T>, value: T[AllKeys<T>]) {
    const currentSpec = get(this.specStore);

    const newSpec = { ...currentSpec, [key]: value };

    if (value === undefined || value == "") {
      delete newSpec[key];
    }

    // If the metrics_view is changed, clear the time_filters and dimension_filters
    if (key === "metrics_view") {
      if ("time_filters" in newSpec) {
        delete newSpec.time_filters;
      }
      if ("dimension_filters" in newSpec) {
        delete newSpec.dimension_filters;
      }
      if (this.resetParams.length > 0) {
        this.resetParams.forEach((param) => {
          delete newSpec[param];
        });
      }
    }

    if (this.isValid(newSpec)) {
      this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }
}

import { applyFieldControls } from "@rilldata/web-common/features/canvas/controls/apply-fields";
import {
  andControlFilters,
  applyTimeControls,
  controlFiltersExpression,
  withoutTimeControls,
} from "@rilldata/web-common/features/canvas/controls/apply-filters";
import type { AppliedControl } from "@rilldata/web-common/features/canvas/controls/controls";
import type {
  CanvasComponentType,
  ComponentSize,
  ComponentSpec,
} from "@rilldata/web-common/features/canvas/components/types";
import type {
  AllKeys,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type {
  V1CanvasItem,
  V1Expression,
  V1Resource,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { Component, ComponentType, SvelteComponent } from "svelte";
import type { Readable, Unsubscriber } from "svelte/store";
import { derived, get, writable, type Writable } from "svelte/store";
import { isMap } from "yaml";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";
import {
  createAndExpression,
  sanitiseExpression,
} from "../../dashboards/stores/filter-utils";
import { ALL_TIME_RANGE_ALIAS } from "../../dashboards/time-controls/new-time-controls";
import type {
  ComparisonTimeRangeState,
  TimeAndFilterStore,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import { normalizeTimeFilters, timeFiltersToParams } from "./time-filters";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import type { ComponentExportPayload } from "@rilldata/web-common/features/exports/xlsx/payload";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copoy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Viewer controls in effect for this component, nearest level first, with their values
  controls: Readable<AppliedControl[]>;
  // The spec as viewers see it: the YAML spec with the field controls applied.
  // Display and query code reads this; the editor reads and writes specStore.
  viewSpecStore: Readable<T>;
  // The component's metrics view, shared by its local filter managers.
  metricsViewsProvider: MetricsViewsProvider;
  // Widget specific dimension and measure filters
  localExpressionFilters: ExpressionFilterManager;
  // Widget specific time filters, following the canvas time filters where `time_filters` inherits them.
  localTimeFilterManager: TimeFilterManager;

  // Lazy-load latch: flipped true once the component scrolls into view, gating
  // its data query so off-screen components don't fetch.
  visible = writable(false);

  // Whether the component's data query should run: true when the component is
  // visible or while the canvas is exporting to PDF. Export uses this rather than
  // forcing `visible` so it never mutates the live lazy-load state.
  dataEnabled: Readable<boolean>;

  // The data the component currently renders, published by its display
  // component for the workbook export (see features/exports/xlsx). Undefined
  // while loading or when the component has nothing to export.
  exportPayload: Writable<ComponentExportPayload | undefined> =
    writable(undefined);

  // True while the component's export render is still drawing after its data
  // arrived, like a map loading its tiles. The canvas exports wait for it
  // before they capture the component (see prepareCanvasForCapture).
  drawing = writable(false);

  // Tears down the spec subscription opened in the constructor. Without this,
  // a component replaced in CanvasEntity.processRows keeps reacting to spec
  // emissions and mutates the shared filter/time state of a deleted widget.
  private unsubscribeSpec: Unsubscriber;
  // The time filter params last given to the widget's time state, with the viewers' time controls applied,
  // and the same params from the YAML alone.
  private timeFilterParams = {
    shown: new URLSearchParams(),
    yaml: new URLSearchParams(),
  };
  private unsubscribeTimeFilterChanges: () => void;

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
    this.controls = parent.controls.forPath(() => this.pathInYAML);
    this.viewSpecStore = derived(
      [this.specStore, this.controls],
      ([spec, controls]) => applyFieldControls(spec, controls),
    );

    this.resource.set(resource);
    this.id = resource.meta?.name?.name as string;

    this.dataEnabled = derived(
      [this.visible, this.parent.exportMode],
      ([visible, exportMode]) => visible || exportMode,
    );

    this.metricsViewsProvider = new MetricsViewsProvider(this.parent.client, [
      this.metricsViewName,
    ]);
    this.localExpressionFilters =
      this.parent.expressionFilterManager.createLocalFilterStore(
        this.metricsViewsProvider,
      );
    this.localTimeFilterManager = new TimeFilterManager(
      this.parent.client,
      this.metricsViewsProvider,
      this.parent.dashboardProvider.yamlConfigProvider,
      {},
      this.parent.timeFilterManager,
    );

    // Writes time filter changes from the controls back to the YAML.
    // Lives here rather than in the inspector so it works without the inspector open.
    // What the viewers' time controls applied stays out of the YAML.
    this.unsubscribeTimeFilterChanges =
      this.localTimeFilterManager.storeSync.on("internal-change", (params) => {
        const { shown, yaml } = this.timeFilterParams;
        this.updateProperty(
          "time_filters" as AllKeys<T>,
          normalizeTimeFilters(
            withoutTimeControls(params, shown, yaml),
          ) as T[AllKeys<T>],
        );
      });

    // The time state follows the YAML's time filters with the viewers' time controls applied.
    this.unsubscribeSpec = derived(
      [this.specStore, this.controls],
      ([spec, controls]) => ({ spec, controls }),
    ).subscribe(({ spec, controls }) => {
      this.localExpressionFilters.setParamForMetricsView(
        this.metricsViewName,
        (spec["dimension_filters"] ?? "") as string,
      );

      // A time range or grain picked through a control applies on top of the YAML.
      const timeFilters = spec?.["time_filters"] as string | undefined;
      this.timeFilterParams = {
        shown: timeFiltersToParams(
          applyTimeControls(timeFilters ?? "", controls).toString(),
        ),
        yaml: timeFiltersToParams(timeFilters),
      };
      this.localTimeFilterManager.storeSync.setUrlParams(
        this.timeFilterParams.shown,
      );
    });
  }

  destroy() {
    this.unsubscribeSpec?.();
    this.unsubscribeTimeFilterChanges?.();
    this.localTimeFilterManager.cleanup();
    this.metricsViewsProvider.cleanup();
  }

  // item is the canvas item this instance renders; only used by components whose
  // editable state lives on the item rather than the resource (see ComponentRefComponent).
  update(resource: V1Resource, path: ComponentPath, item?: V1CanvasItem) {
    void item;
    const yamlSpec = (resource.component?.state?.validSpec
      ?.rendererProperties ??
      (this.parent.allowUnvalidatedSpec
        ? resource.component?.spec?.rendererProperties
        : undefined)) as T;
    this.resource.set(resource);
    this.pathInYAML = path;
    this.specStore.set(yamlSpec);
  }

  // This will be deprecated eventually - bgh
  get timeAndFilterStore(): Readable<TimeAndFilterStore> {
    return derived(
      [
        this.localTimeFilterManager.getTimeControlStore(),
        this.parent.expressionFilterManager.getExprStoreForMetricsView(
          this.metricsViewName,
        ),
        this.specStore,
        this.controls,
        this.parent.metricsView.getMetricsViewFromName(this.metricsViewName),
      ],
      ([
        {
          timeRange: timeRangeName,
          timeGrain,
          timeZone,
          interval,
          showComparison,
          comparisonTimeRange: comparisonTimeRangeName,
          comparisonInterval,
          hasTimeSeries,
        },
        metricsViewFilters,
        componentSpec,
        controls,
        metricsView,
      ]) => {
        // Timestamps sent to the runtime must be UTC:
        // the protobuf JSON codec rejects ISO strings with both milliseconds and a timezone offset,
        // which is what Luxon produces for zoned datetimes (e.g. 2026-08-13T00:00:00.000+09:00).
        const timeRange: V1TimeRange = {
          start: interval?.start.toUTC().toISO(),
          end: interval?.end.toUTC().toISO(),
          timeZone,
        };

        const timeRangeState: TimeRangeState | undefined = {
          selectedTimeRange: interval
            ? {
                name: timeRangeName ?? TimeRangePreset.CUSTOM,
                start: interval.start.toJSDate(),
                end: interval.end.toJSDate(),
                interval: timeGrain,
              }
            : undefined,
          timeStart: interval?.start.toUTC().toISO(),
          timeEnd: interval?.end.toUTC().toISO(),
        };

        // All time has nothing before it to compare against.
        // `comparisonInterval` is only set while the comparison is shown.
        const showTimeComparison =
          showComparison && timeRangeName !== ALL_TIME_RANGE_ALIAS;
        const shownComparisonInterval = showTimeComparison
          ? comparisonInterval
          : undefined;

        const comparisonTimeRange: V1TimeRange = {
          start: shownComparisonInterval?.start.toUTC().toISO(),
          end: shownComparisonInterval?.end.toUTC().toISO(),
          timeZone,
        };

        const comparisonTimeRangeState: ComparisonTimeRangeState | undefined =
          shownComparisonInterval && {
            comparisonTimeStart: shownComparisonInterval.start.toUTC().toISO(),
            comparisonTimeEnd: shownComparisonInterval.end.toUTC().toISO(),
            selectedComparisonTimeRange: {
              start: shownComparisonInterval.start.toJSDate(),
              end: shownComparisonInterval.end.toJSDate(),
              name: comparisonTimeRangeName,
            },
          };

        // Dimension Filters
        // The global filters are absent until the canvas' metrics views resolve, and for a component
        // pointed at a metrics view the canvas does not reference. The component's own filters below
        // still apply in both cases.
        const globalDimensionOnlyWhere =
          sanitiseExpression(
            metricsViewFilters?.dimensionOnlyExpr,
            undefined,
          ) ?? createAndExpression([]);

        let fullWhere: V1Expression | undefined = metricsViewFilters?.expr;
        let dimensionOnlyWhere: V1Expression | undefined =
          globalDimensionOnlyWhere;

        if (componentSpec?.["dimension_filters"]) {
          const { expr: componentWhere } = getFiltersFromText(
            componentSpec?.["dimension_filters"] as string,
          );
          fullWhere = mergeFilters(fullWhere, componentWhere);
          dimensionOnlyWhere = mergeFilters(
            globalDimensionOnlyWhere,
            componentWhere,
          );
        }

        const controlWhere = controlFiltersExpression(
          controls,
          this.metricsViewName,
          (name) =>
            !!metricsView.metricsView?.dimensions?.some((d) => d.name === name),
        );
        if (controlWhere) {
          fullWhere = andControlFilters(fullWhere, controlWhere);
          dimensionOnlyWhere = andControlFilters(
            dimensionOnlyWhere,
            controlWhere,
          );
        }

        return {
          timeRange,
          showTimeComparison,
          comparisonTimeRange,
          dimensionOnlyWhere,
          where: fullWhere,
          timeGrain,
          timeRangeState,
          comparisonTimeRangeState,
          hasTimeSeries,
        };
      },
    );
  }

  // Resolves to false when writing the canvas file failed. Without a file,
  // the spec only changes in memory.
  private updateYAML(newSpec: T): Promise<boolean> {
    if (!this.parent.fileArtifact) return Promise.resolve(true);
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent, saveLocalContent } = this.parent.fileArtifact;

    parsedDocument.setIn(this.pathInYAML, newSpec);

    // Saved explicitly rather than with auto-save, so the result can be awaited.
    updateEditorContent(parsedDocument.toString(), false, false);
    return saveLocalContent(true);
  }

  // Resolves to false when writing the canvas file failed. An invalid spec
  // is not written, only applied in memory.
  setSpec(newSpec: T): Promise<boolean> {
    const saved = this.isValid(newSpec)
      ? this.updateYAML(newSpec)
      : Promise.resolve(true);
    this.specStore.set(newSpec);
    return saved;
  }

  updateProperty(key: AllKeys<T>, value: T[AllKeys<T>]): Promise<boolean> {
    return this.updateProperties({ [key]: value } as unknown as Partial<T>);
  }

  // Writes several properties in one YAML update, so dependent keys (such as
  // a sort field and its direction) never reach the reconciler half-applied.
  // A key whose value is undefined or empty is removed from the spec.
  // Resolves to false when writing the canvas file failed; an invalid spec is
  // not written, only applied in memory. Most callers don't wait, since the
  // file's save state reports failures.
  updateProperties(patch: Partial<T>): Promise<boolean> {
    const currentSpec = this.specInYAML() ?? get(this.specStore);

    const newSpec = { ...currentSpec, ...patch };

    for (const [key, value] of Object.entries(patch)) {
      if (value === undefined || value == "") {
        delete newSpec[key];
      }
    }

    // If the metrics_view is changed, clear the time_filters and dimension_filters
    if ("metrics_view" in patch) {
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

    const saved = this.isValid(newSpec)
      ? this.updateYAML(newSpec)
      : Promise.resolve(true);
    this.specStore.set(newSpec);
    return saved;
  }

  // The component's properties as currently written in the file. While
  // editing, the spec store trails the file by a reconcile: a write built on
  // the store would drop a previous write the runtime has not echoed back yet,
  // and would come back with the runtime's alphabetical key order.
  private specInYAML(): T | undefined {
    if (!this.parent.fileArtifact || !this.parent.parsedContent) return;
    const document = get(this.parent.parsedContent);
    const node: unknown = document.getIn(this.pathInYAML);
    if (!isMap(node)) return;
    const spec: unknown = node.toJS(document);
    return spec as T;
  }
}

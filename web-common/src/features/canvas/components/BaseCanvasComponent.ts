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
import type {
  V1Expression,
  V1Resource,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import type { Component, ComponentType, SvelteComponent } from "svelte";
import type { Readable, Unsubscriber } from "svelte/store";
import { derived, get, writable, type Writable } from "svelte/store";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";
import {
  createAndExpression,
  sanitiseExpression,
} from "../../dashboards/stores/filter-utils";
import type {
  ComparisonTimeRangeState,
  TimeAndFilterStore,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
import type {
  CanvasEntity,
  ComponentPath,
  SearchParamsStore,
} from "../stores/canvas-entity";
import { TimeState } from "../stores/time-state";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";

export abstract class BaseCanvasComponent<T = ComponentSpec> {
  id: string;
  // Local copoy of the canvas component resource
  resource: Writable<V1Resource | null> = writable(null);
  // Local copy of the spec (aka rendererProperties) for the component
  specStore: Writable<T>;
  // Path in the YAML where the component is stored
  pathInYAML: ComponentPath;
  // Widget specific dimension and measure filters
  localExpressionFilters: ExpressionFilterManager;
  // Widget specific time filters
  localTimeControls: TimeState;

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

    const yamlTimeFilterStore: SearchParamsStore = (() => {
      const store = derived(this.specStore, (spec) => {
        return new URLSearchParams(spec?.["time_filters"] ?? "");
      });
      return {
        subscribe: store.subscribe,
        set: (map: Map<string, string | undefined>) => {
          const searchParams = get(store);

          map.forEach((value, key) => {
            if (value === undefined || value === null || value === "") {
              searchParams.delete(key);
            } else {
              searchParams.set(key, value);
            }
          });

          this.updateProperty(
            "time_filters" as AllKeys<T>,
            searchParams.toString() as T[AllKeys<T>],
          );
          return true;
        },
        clearAll: () => {
          const searchParams = get(store);

          searchParams.forEach((_, key) => {
            searchParams.delete(key);
          });

          this.updateProperty(
            "time_filters" as AllKeys<T>,
            searchParams.toString() as T[AllKeys<T>],
          );
        },
      };
    })();

    this.localTimeControls = this.parent.timeManager.createLocalTimeState(
      this.id,
      yamlTimeFilterStore,
      this.metricsViewName,
    );

    this.localExpressionFilters =
      this.parent.expressionFilterManager.createLocalFilterStore(
        this.metricsViewName,
      );

    this.unsubscribeSpec = this.specStore.subscribe((spec) => {
      this.localExpressionFilters.setParamForMetricsView(
        this.metricsViewName,
        (spec["dimension_filters"] ?? "") as string,
      );

      this.localTimeControls.onUrlChange(
        new URLSearchParams(spec?.["time_filters"] ?? ""),
      );
    });
  }

  destroy() {
    this.unsubscribeSpec?.();
    this.localExpressionFilters.metricsViewsProvider.cleanup();
    this.localExpressionFilters.yamlConfigProvider.cleanup?.();
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

  // This will be deprecated eventually - bgh
  get timeAndFilterStore(): Readable<TimeAndFilterStore> {
    return derived(
      [
        this.parent.timeManager.state.interval,
        this.parent.timeManager.state.grainStore,
        this.parent.timeManager.state.showTimeComparisonStore,
        this.parent.timeManager.state.comparisonRangeStore,
        this.parent.timeManager.state.comparisonIntervalStore,
        this.parent.timeManager.state.timeZoneStore,
        this.parent.timeManager.state.rangeStore,
        this.localTimeControls.interval,
        this.localTimeControls.comparisonIntervalStore,
        this.localTimeControls.showTimeComparisonStore,
        this.localTimeControls.grainStore,
        this.localTimeControls.comparisonRangeStore,
        this.localTimeControls.rangeStore,
        this.parent.expressionFilterManager.getExprStoreForMetricsView(
          this.metricsViewName,
        ),
        this.parent.timeManager.hasTimeSeriesMap,
        this.specStore,
      ],
      ([
        globalInterval,
        globalGrainStore,
        globalShowTimeComparison,
        globalComparisonRange,
        globalComparisonInterval,
        timeZone,
        globalRange,
        localInterval,
        localComparisonInterval,
        localShowTimeComparison,
        localGrainStore,
        localComparisonRange,
        localRange,
        metricsViewFilters,
        hasTimeSeriesMap,
        componentSpec,
      ]) => {
        const hasTimeSeries = hasTimeSeriesMap.get(this.metricsViewName);

        let timeGrain = globalGrainStore;

        // Timestamps sent to the runtime must be UTC:
        // the protobuf JSON codec rejects ISO strings with both milliseconds and a timezone offset,
        // which is what Luxon produces for zoned datetimes (e.g. 2026-08-13T00:00:00.000+09:00).
        let timeRange: V1TimeRange = {
          start: globalInterval?.start.toUTC().toISO(),
          end: globalInterval?.end.toUTC().toISO(),
          timeZone,
        };

        let timeRangeState: TimeRangeState | undefined = {
          selectedTimeRange: globalInterval
            ? {
                name: globalRange ?? TimeRangePreset.CUSTOM,
                start: globalInterval.start.toJSDate(),
                end: globalInterval.end.toJSDate(),
                interval: globalGrainStore,
              }
            : undefined,
          timeStart: globalInterval?.start.toUTC().toISO(),
          timeEnd: globalInterval?.end.toUTC().toISO(),
        };

        let comparisonTimeRange: V1TimeRange | undefined = {
          start: globalComparisonInterval?.start.toUTC().toISO(),
          end: globalComparisonInterval?.end.toUTC().toISO(),
          timeZone,
        };

        let showTimeComparison = globalShowTimeComparison;

        let comparisonTimeRangeState: ComparisonTimeRangeState | undefined =
          globalComparisonInterval && {
            comparisonTimeStart: globalComparisonInterval.start.toUTC().toISO(),
            comparisonTimeEnd: globalComparisonInterval.end.toUTC().toISO(),
            selectedComparisonTimeRange: {
              start: globalComparisonInterval.start.toJSDate(),
              end: globalComparisonInterval.end.toJSDate(),
              name: globalComparisonRange,
            },
          };

        if (componentSpec?.["time_filters"]) {
          timeRange = {
            start: localInterval?.start.toUTC().toISO(),
            end: localInterval?.end.toUTC().toISO(),
            timeZone,
          };

          comparisonTimeRange = {
            start: localComparisonInterval?.start.toUTC().toISO(),
            end: localComparisonInterval?.end.toUTC().toISO(),
            timeZone,
          };

          showTimeComparison = localShowTimeComparison;

          timeGrain = localGrainStore ?? globalGrainStore;

          const localTimeRangeState: TimeRangeState = {
            selectedTimeRange: localInterval
              ? {
                  name: localRange ?? TimeRangePreset.CUSTOM,
                  start: localInterval.start.toJSDate(),
                  end: localInterval.end.toJSDate(),
                  interval: localGrainStore ?? globalGrainStore,
                }
              : undefined,
            timeStart: localInterval?.start.toUTC().toISO(),
            timeEnd: localInterval?.end.toUTC().toISO(),
          };
          const localComparisonRangeState:
            | ComparisonTimeRangeState
            | undefined = localComparisonInterval && {
            comparisonTimeStart: localComparisonInterval.start.toUTC().toISO(),
            comparisonTimeEnd: localComparisonInterval.end.toUTC().toISO(),
            selectedComparisonTimeRange: {
              start: localComparisonInterval.start.toJSDate(),
              end: localComparisonInterval.end.toJSDate(),
              name: localComparisonRange,
            },
          };

          timeRangeState = localTimeRangeState;
          comparisonTimeRangeState = localComparisonRangeState;
        }

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

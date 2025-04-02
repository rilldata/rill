import type { ChartType } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { PivotSpec } from "@rilldata/web-common/features/canvas/components/pivot";
import type { TableSpec } from "@rilldata/web-common/features/canvas/components/table";
import type {
  CanvasComponentType,
  ComponentSize,
  ComponentSpec,
} from "@rilldata/web-common/features/canvas/components/types";
import { getParsedDocument } from "@rilldata/web-common/features/canvas/inspector/selectors";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  V1Expression,
  V1MetricsViewSpec,
  V1Resource,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { derived, get, writable, type Writable } from "svelte/store";
import { CanvasComponentState } from "../stores/canvas-component";
import type { CanvasEntity, ComponentPath } from "../stores/canvas-entity";
import type {
  ComparisonTimeRangeState,
  TimeRangeState,
} from "../../dashboards/time-controls/time-control-store";
import {
  buildValidMetricsViewFilter,
  createAndExpression,
} from "../../dashboards/stores/filter-utils";
import { mergeFilters } from "../../dashboards/pivot/pivot-merge-filters";

// A base class that implements all the store logic
export abstract class BaseCanvasComponent<T = ComponentSpec> {
  /**
   * Local copy of the spec as a svelte writable store
   */
  specStore: Writable<T>;
  /**
   * Path in the YAML where the component is stored
   */
  pathInYAML: ComponentPath;
  /**
   * File artifact where the component
   * is stored
   */
  // parent: CanvasEntity;

  id: string;

  state: CanvasComponentState;
  resource: Writable<V1Resource | null> = writable(null);

  abstract type: CanvasComponentType;

  // Let child classes define these
  /**
   * Minimum allowed size for the component
   * container on the canvas
   */
  abstract minSize: ComponentSize;

  /**
   * The default size of the container when the component
   * is added to the canvas
   */
  abstract defaultSize: ComponentSize;

  /**
   * The parameters that should be reset when the metrics_view
   * is changed
   */
  abstract resetParams: string[];

  /**
   * The minimum condition needed for the spec to be valid
   * for the given component and to be rendered on the canvas
   */
  abstract isValid(spec: T): boolean;

  /**
   * A map of input params which will be used in the visual
   * UI builder
   */
  abstract inputParams(): InputParams<T>;

  /**
   * Get the spec when the component is added to the canvas
   */
  // abstract newComponentSpec(
  //   metricsViewName: string,
  //   metricsViewSpec: V1MetricsViewSpec | undefined,
  // ): T;

  constructor(
    resource: V1Resource,
    public parent: CanvasEntity,
    path: ComponentPath,
    public defaultSpec: T,
  ) {
    const yamlSpec = resource.component?.state?.validSpec?.rendererProperties;

    const mergedSpec = { ...defaultSpec, ...yamlSpec };
    this.specStore = writable(mergedSpec);
    this.pathInYAML = path;

    this.resource.set(resource);
    this.id = resource.meta?.name?.name as string;
    // this.parent = parent;
    this.state = new CanvasComponentState(
      this.id,
      this.parent.specStore,
      this.parent.spec,
    );
  }

  update(resource: V1Resource, path: ComponentPath) {
    const yamlSpec = resource.component?.state?.validSpec?.rendererProperties;

    const mergedSpec = { ...this.defaultSpec, ...yamlSpec };
    this.resource.set(resource);
    this.pathInYAML = path;
    this.specStore.set(mergedSpec);
  }

  get timeAndFilterStore() {
    return derived(
      [
        this.parent.timeControls.timeRangeStateStore,
        this.state.localTimeControls.timeRangeStateStore,
        this.parent.timeControls.comparisonRangeStateStore,
        this.state.localTimeControls.comparisonRangeStateStore,
        this.parent.timeControls.selectedTimezone,
        this.parent.filters.whereFilter,
        this.parent.filters.dimensionThresholdFilters,
        this.parent.specStore,
        this.parent.timeControls.hasTimeSeries,
        this.specStore,
      ],
      ([
        globalTimeRangeState,
        localTimeRangeState,
        globalComparisonRangeState,
        localComparisonRangeState,
        timeZone,
        whereFilter,
        dtf,
        canvasData,
        hasTimeSeries,
        componentSpec,
      ]) => {
        const metricsViewName = componentSpec["metrics_view"];
        const metricsView = canvasData.data?.metricsViews?.[metricsViewName];
        const dimensions = metricsView?.state?.validSpec?.dimensions ?? [];
        const measures = metricsView?.state?.validSpec?.measures ?? [];

        // Time Filters
        let timeRange: V1TimeRange = {
          start: globalTimeRangeState?.timeStart,
          end: globalTimeRangeState?.timeEnd,
          timeZone,
        };

        let timeGrain = globalTimeRangeState?.selectedTimeRange?.interval;

        const localShowTimeComparison =
          !!localComparisonRangeState?.showTimeComparison;
        const globalShowTimeComparison =
          !!globalComparisonRangeState?.showTimeComparison;

        let showTimeComparison = globalShowTimeComparison;

        let comparisonTimeRange: V1TimeRange | undefined = {
          start: globalComparisonRangeState?.comparisonTimeStart,
          end: globalComparisonRangeState?.comparisonTimeEnd,
          timeZone,
        };

        let timeRangeState: TimeRangeState | undefined = globalTimeRangeState;
        let comparisonTimeRangeState: ComparisonTimeRangeState | undefined =
          globalComparisonRangeState;

        if (componentSpec?.["time_filters"]) {
          timeRange = {
            start: localTimeRangeState?.timeStart,
            end: localTimeRangeState?.timeEnd,
            timeZone,
          };

          comparisonTimeRange = {
            start: localComparisonRangeState?.comparisonTimeStart,
            end: localComparisonRangeState?.comparisonTimeEnd,
            timeZone,
          };

          showTimeComparison = localShowTimeComparison;

          timeGrain = localTimeRangeState?.selectedTimeRange?.interval;

          timeRangeState = localTimeRangeState;
          comparisonTimeRangeState = localComparisonRangeState;
        }

        // Dimension Filters
        const globalWhere =
          buildValidMetricsViewFilter(whereFilter, dtf, dimensions, measures) ??
          createAndExpression([]);

        let where: V1Expression | undefined = globalWhere;

        if (componentSpec?.["dimension_filters"]) {
          const { expr: componentWhere } =
            component.state.localFilters.getFiltersFromText(
              componentSpec?.["dimension_filters"] as string,
            );
          where = mergeFilters(globalWhere, componentWhere);
        }

        return {
          timeRange,
          showTimeComparison,
          comparisonTimeRange,
          where,
          timeGrain,
          timeRangeState,
          comparisonTimeRangeState,
          hasTimeSeries,
        };
      },
    );
  }

  private async updateYAML(newSpec: T): Promise<void> {
    if (!this.parent.fileArtifact) return;
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent, saveLocalContent } = this.parent.fileArtifact;

    // Update the Item
    parsedDocument.setIn(this.pathInYAML, newSpec);

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), false);
    await saveLocalContent();
  }

  /**
   * Set the spec store and YAML with the new values
   */
  async setSpec(newSpec: T): Promise<void> {
    if (this.isValid(newSpec)) {
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  /**
   * Update the spec store and YAML with the new values
   */
  // TODO: Add stricter type definition for keys and value deriving from spec
  async updateProperty(key: string, value: unknown): Promise<void> {
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
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  /**
   * Update the chart type of chart component in store and YAML
   */
  async updateChartType(key: ChartType) {
    if (!this.parent.fileArtifact) return;
    const currentSpec = get(this.specStore);

    const parentPath = this.pathInYAML.slice(0, -1);

    const parseDocumentStore = getParsedDocument(this.parent.fileArtifact);
    const parsedDocument = get(parseDocumentStore);

    const { updateEditorContent, saveLocalContent } = this.parent.fileArtifact;

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [key]: currentSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }

  async updateTableType(
    newTableType: "pivot" | "table",
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ) {
    if (!this.parent.fileArtifact) return;
    const parentPath = this.pathInYAML.slice(0, -1);
    const parseDocumentStore = getParsedDocument(this.parent.fileArtifact);
    const parsedDocument = get(parseDocumentStore);
    const { updateEditorContent, saveLocalContent } = this.parent.fileArtifact;

    const currentSpec = get(this.specStore);

    const allMeasures =
      metricsViewSpec?.measures?.map((m) => m.name as string) || [];
    const allDimensions =
      metricsViewSpec?.dimensions?.map((d) => d.name || (d.column as string)) ||
      [];

    let newSpec: PivotSpec | TableSpec;
    if (newTableType === "pivot") {
      // Switch to pivot table spec
      const flatTableSpec = currentSpec as TableSpec;
      const { columns = [], ...restFlatTableSpec } = flatTableSpec || {};

      const row_dimensions =
        columns?.filter((c) => allDimensions.includes(c)) || [];
      const measures = columns?.filter((c) => allMeasures.includes(c)) || [];

      newSpec = {
        ...restFlatTableSpec,
        row_dimensions,
        measures,
      };
    } else {
      // Switch to flat table spec
      const pivotTableSpec = currentSpec as PivotSpec;

      const {
        row_dimensions = [],
        col_dimensions = [],
        measures = [],
        ...restPivotTableSpec
      } = pivotTableSpec || {};

      const columns = [...row_dimensions, ...col_dimensions, ...measures];

      newSpec = {
        ...restPivotTableSpec,
        columns,
      };
    }

    const width = parsedDocument.getIn([...parentPath, "width"]);

    parsedDocument.setIn(parentPath, { [newTableType]: newSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
  }
}

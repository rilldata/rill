import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type {
  PivotDataStoreConfig,
  PivotState,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";
import { derived, get, writable, type Writable } from "svelte/store";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentFilterProperties,
} from "../types";
import CanvasPivotDisplay from "./CanvasPivotDisplay.svelte";
import { createPivotConfig, usePivotForCanvas } from "./util";

export interface PivotSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  row_dimensions?: string[];
  col_dimensions?: string[];
  show_row_totals?: boolean;
  show_col_totals?: boolean;
}

export interface TableSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  columns: string[];
  show_row_totals?: boolean;
  show_col_totals?: boolean;
}

export { default as Pivot } from "./CanvasPivotDisplay.svelte";

export class PivotCanvasComponent extends BaseCanvasComponent<
  PivotSpec | TableSpec
> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = ["measures", "row_dimensions", "col_dimensions"];
  type: CanvasComponentType;
  component = CanvasPivotDisplay;
  config: Readable<PivotDataStoreConfig>;
  pivotDataStore: ReturnType<typeof usePivotForCanvas>;
  pivotState: Writable<PivotState>;
  /** Dimensions the pivot itself has filtered via click-to-filter.
   *  These are excluded from the pivot's own data query so all rows remain visible. */
  selfFilteredDimensions: Writable<Set<string>>;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const type = (resource.component?.state?.validSpec?.renderer ??
      (parent.allowUnvalidatedSpec
        ? resource.component?.spec?.renderer
        : undefined)) as CanvasComponentType;

    if (type !== "table" && type !== "pivot") {
      throw new Error(
        `Invalid table type: ${type}. Expected "table" or "pivot".`,
      );
    }

    const defaultPivotSpec: PivotSpec = {
      metrics_view: "",
      measures: [],
      row_dimensions: [],
      col_dimensions: [],
    };

    const defaultFlatSpec: TableSpec = {
      metrics_view: "",
      columns: [],
    };

    super(
      resource,
      parent,
      path,
      type === "pivot" ? defaultPivotSpec : defaultFlatSpec,
    );

    this.type = type;

    this.pivotState = writable(this.getInitPivotState(type));
    this.selfFilteredDimensions = writable(new Set<string>());

    this.config = createPivotConfig(
      this.parent,
      this.specStore,
      this.pivotState,
      this.timeAndFilterStore,
      this.selfFilteredDimensions,
    );

    this.pivotDataStore = usePivotForCanvas(
      this.parent,
      derived(this.specStore, ($specStore) => $specStore.metrics_view),
      this.config,
      this.visible,
    );
  }

  getInitPivotState(type: "pivot" | "table"): PivotState {
    return {
      columns: [],
      rows: [],
      expanded: {},
      sorting: [],
      columnPage: 1,
      rowPage: 1,
      enableComparison: false,
      tableMode: type === "pivot" ? "nest" : "flat",
      activeCell: null,
      showRowTotals: true,
      showColTotals: true,
    };
  }

  isValid(spec: PivotSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    const pivotState = get(this.pivotState);
    return {
      pivot: pivotState,
      activePage: DashboardState_ActivePage.PIVOT,
    };
  }
  inputParams(type: "pivot" | "table"): InputParams<PivotSpec | TableSpec> {
    if (type === "pivot") {
      return {
        options: {
          metrics_view: { type: "metrics", label: "Metrics view" },
          measures: {
            type: "multi_fields",
            meta: { allowedTypes: ["measure"] },
            label: "Measures",
          },
          col_dimensions: {
            type: "multi_fields",
            meta: { allowedTypes: ["time", "dimension"] },
            label: "Column dimensions",
          },
          row_dimensions: {
            type: "multi_fields",
            meta: { allowedTypes: ["time", "dimension"] },
            label: "Row dimensions",
          },
          show_row_totals: {
            type: "boolean",
            label: "Total column",
            meta: { defaultValue: true },
          },
          show_col_totals: {
            type: "boolean",
            label: "Total row",
            meta: { defaultValue: true },
          },
          ...commonOptions,
        },
        filter: getFilterOptions(true, false),
      };
    } else {
      return {
        options: {
          metrics_view: { type: "metrics", label: "Metrics view" },
          columns: {
            type: "multi_fields",
            label: "Columns",
            meta: { allowedTypes: ["time", "dimension", "measure"] },
          },
          show_col_totals: {
            type: "boolean",
            label: "Total row",
            meta: { defaultValue: true },
          },
          show_row_totals: {
            type: "boolean",
            label: "Total column",
            meta: { defaultValue: true },
          },
          ...commonOptions,
        },
        filter: getFilterOptions(true, false),
      };
    }
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): TableSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 3).map((m) => m.name as string) ?? [];

    const dimensions =
      metricsViewSpec?.dimensions
        ?.slice(0, 3)
        .map((d) => d.name || (d.column as string)) ?? [];

    return {
      metrics_view: metricsViewName,
      columns: [...dimensions, ...measures],
    };
  }

  updateTableType(newTableType: "pivot" | "table") {
    if (!this.parent.fileArtifact) return;

    // Clear active component if this pivot was the active one
    if (get(this.parent.activeComponent) === this.id) {
      this.parent.clearActiveComponent();
    }
    this.selfFilteredDimensions.set(new Set());

    this.type = newTableType;

    this.pivotState.set(this.getInitPivotState(newTableType));

    const parentPath = this.pathInYAML.slice(0, -1);
    const parseDocumentStore = this.parent.parsedContent;
    const parsedDocument = get(parseDocumentStore);
    const { updateEditorContent } = this.parent.fileArtifact;

    const currentSpec = get(this.specStore);

    const metricsViewSpecQuery = this.parent.metricsView.getMetricsViewFromName(
      currentSpec.metrics_view,
    );

    const metricsViewSpec = get(metricsViewSpecQuery).metricsView;

    const allMeasures =
      metricsViewSpec?.measures?.map((m) => m.name as string) || [];
    const allDimensions =
      metricsViewSpec?.dimensions?.map((d) => d.name || (d.column as string)) ||
      [];

    let newSpec: PivotSpec | TableSpec;

    const commonProperties: ComponentCommonProperties &
      ComponentFilterProperties &
      Pick<PivotSpec, "show_row_totals" | "show_col_totals"> = {
      title: currentSpec.title,
      description: currentSpec.description,
      dimension_filters: currentSpec.dimension_filters,
      time_filters: currentSpec.time_filters,
      show_row_totals: currentSpec.show_row_totals,
      show_col_totals: currentSpec.show_col_totals,
    };

    if ("columns" in currentSpec) {
      const row_dimensions =
        currentSpec?.columns?.filter((c) => allDimensions.includes(c)) || [];
      const measures =
        currentSpec?.columns?.filter((c) => allMeasures.includes(c)) || [];

      newSpec = {
        ...commonProperties,
        metrics_view: currentSpec.metrics_view,
        row_dimensions,
        measures,
      };
    } else {
      newSpec = {
        ...commonProperties,
        metrics_view: currentSpec.metrics_view,
        columns: [
          ...(currentSpec?.row_dimensions ?? []),
          ...(currentSpec?.col_dimensions ?? []),
          ...(currentSpec?.measures ?? []),
        ],
      };
    }

    const width = parsedDocument.getIn([...parentPath, "width"]);

    this.specStore.set(newSpec);

    parsedDocument.setIn(parentPath, { [newTableType]: newSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), false, true);
  }
}

import {
  ephemeralSpecsToDefs,
  type EphemeralMeasureSpec,
} from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  getCommonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type {
  AllKeys,
  ComponentInputParam,
  InputParams,
} from "@rilldata/web-common/features/canvas/inspector/types";
import {
  clampColumnWidth,
  type ColumnWidthRole,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-column-width-utils";
import { PIVOT_ROW_LIMIT_OPTIONS } from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import type {
  PivotDataStoreConfig,
  PivotFormatRule,
  PivotMeasureFormatting,
  PivotState,
  PivotTotalsRowPosition,
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
import {
  defaultSortDir,
  fieldConfigs,
  fieldName,
  fieldNames,
  normalizeRowDimensionEntries,
  setFieldConfig as setFieldConfigInList,
  sortableFieldNames,
  stripColumnDimensionKeys,
  WRAP_LINES_DEFAULT,
  WRAP_LINES_MAX,
  WRAP_LINES_MIN,
  type PivotFieldConfigPatch,
  type PivotFieldEntry,
  type PivotFieldListKey,
  type PivotSortComparison,
  type PivotSortDir,
} from "./field-config";
import {
  createPivotConfig,
  ROW_LIMIT_ALL_VALUE,
  usePivotForCanvas,
} from "./util";

// Per-measure conditional formatting persisted in the canvas YAML. A list (not
// a map) keeps the YAML readable and mirrors the proto representation.
export interface PivotConditionalFormatSpec {
  measure: string;
  mode: "heatmap" | "data_bar" | "rules";
  // Color scheme; only for "heatmap" and "data_bar" modes.
  scheme?: string;
  // Ordered threshold rules (first match wins); only for "rules" mode.
  rules?: {
    operator: PivotFormatRule["operator"];
    value: number;
    value2?: number;
    color: string;
  }[];
}

const DEFAULT_FORMAT_SCHEME = "theme-sequential";

/**
 * Map the YAML conditional_format list to the per-measure formatting config
 * consumed by the pivot renderer. The YAML is hand-editable and reaches here
 * unvalidated, so malformed values and entries are skipped rather than trusted
 * to match the declared type.
 */
export function conditionalFormatSpecToMeasureFormatting(
  specs: PivotConditionalFormatSpec[] | undefined,
): Record<string, PivotMeasureFormatting> {
  const measureFormatting: Record<string, PivotMeasureFormatting> = {};
  if (!Array.isArray(specs)) return measureFormatting;
  for (const spec of specs) {
    if (!spec || typeof spec !== "object" || typeof spec.measure !== "string") {
      continue;
    }
    if (spec.mode === "heatmap" || spec.mode === "data_bar") {
      measureFormatting[spec.measure] = {
        mode: spec.mode,
        scheme:
          typeof spec.scheme === "string" ? spec.scheme : DEFAULT_FORMAT_SCHEME,
      };
    } else if (spec.mode === "rules" && Array.isArray(spec.rules)) {
      const rules = spec.rules.filter((r) => r && typeof r === "object");
      if (rules.length) {
        measureFormatting[spec.measure] = { mode: "rules", rules };
      }
    }
  }
  return measureFormatting;
}

/**
 * Inverse of conditionalFormatSpecToMeasureFormatting, used when writing the
 * inspector state back to the YAML.
 */
export function measureFormattingToConditionalFormatSpec(
  measureFormatting: Record<string, PivotMeasureFormatting>,
): PivotConditionalFormatSpec[] {
  return Object.entries(measureFormatting).map(([measure, fmt]) =>
    fmt.mode === "rules"
      ? { measure, mode: fmt.mode, rules: fmt.rules }
      : { measure, mode: fmt.mode, scheme: fmt.scheme },
  );
}

/**
 * Presentation properties shared by tables and pivots. Per-column overrides
 * (width, wrap, align, label, number format) live on the entries of the field
 * lists themselves; see field-config.ts.
 */
export interface TablePresentationProperties {
  // Shrink the columns so the table fits its container instead of scrolling.
  fit_to_width?: boolean;
  // Wrap dimension cells over `wrap_lines` lines instead of truncating them.
  wrap?: boolean;
  // Wrap header labels over `wrap_lines` lines instead of truncating them.
  wrap_headers?: boolean;
  wrap_lines?: number;
  // Initial sort: any column of a table; a measure (by its row total) or a row
  // dimension of a pivot.
  sort_by?: string;
  sort_dir?: PivotSortDir;
  // Sort on the measure's delta or percent-change column instead of its
  // value; only takes effect while the component shows a time comparison.
  sort_comparison?: PivotSortComparison;
}

export interface PivotSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties,
    TablePresentationProperties {
  metrics_view: string;
  measures: PivotFieldEntry[];
  // Ad-hoc measures derived from existing measures via an arithmetic expression.
  adhoc_measures?: EphemeralMeasureSpec[];
  row_dimensions?: PivotFieldEntry[];
  col_dimensions?: PivotFieldEntry[];
  hide_totals_row?: boolean;
  hide_totals_col?: boolean;
  totals_row_position?: PivotTotalsRowPosition;
  conditional_format?: PivotConditionalFormatSpec[];
  row_limit?: string;
}

export interface TableSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties,
    TablePresentationProperties {
  metrics_view: string;
  columns: PivotFieldEntry[];
  // Ad-hoc measures derived from existing measures via an arithmetic expression.
  adhoc_measures?: EphemeralMeasureSpec[];
  hide_totals_row?: boolean;
  hide_totals_col?: boolean;
  totals_row_position?: PivotTotalsRowPosition;
  conditional_format?: PivotConditionalFormatSpec[];
}

export { default as Pivot } from "./CanvasPivotDisplay.svelte";

import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

function totalsRowPositionOptions(): {
  value: PivotTotalsRowPosition;
  label: string;
}[] {
  return [
    { value: "top", label: m.dashboard_totals_row_position_top() },
    { value: "bottom", label: m.dashboard_totals_row_position_bottom() },
  ];
}

export class PivotCanvasComponent extends BaseCanvasComponent<
  PivotSpec | TableSpec
> {
  minSize = { width: 2, height: 2 };
  defaultSize = { width: 4, height: 10 };
  resetParams = [
    "measures",
    "row_dimensions",
    "col_dimensions",
    "conditional_format",
    "adhoc_measures",
    "sort_by",
    "sort_dir",
    "sort_comparison",
  ];
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
      this.dataEnabled,
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
      showTotalsColumn: true,
      showTotalsRow: true,
    };
  }

  isValid(spec: PivotSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    return {
      pivot: get(this.pivotState),
      ephemeralMeasures: ephemeralSpecsToDefs(
        get(this.specStore).adhoc_measures,
      ),
      activePage: DashboardState_ActivePage.PIVOT,
    };
  }

  /**
   * Keeps dependent properties consistent on every write: column dimensions
   * never carry width/wrap/align, row dimensions never carry align and only
   * the first one carries width/wrap, `wrap_lines` stays in range, and the
   * sort only names a field that is still in the component (its direction
   * only exists together with a field). Writing a value the runtime would
   * reject must be avoided here, because the component's spec would then
   * silently revert to the last valid one while the editor shows the new YAML.
   */
  updateProperties(patch: Partial<PivotSpec | TableSpec>) {
    const current = get(this.specStore);
    const next = { ...current, ...patch } as PivotSpec | TableSpec;
    const normalized = { ...patch } as Partial<PivotSpec> & Partial<TableSpec>;

    if (Array.isArray(normalized.col_dimensions)) {
      normalized.col_dimensions = normalized.col_dimensions.map(
        stripColumnDimensionKeys,
      );
    }
    if (Array.isArray(normalized.row_dimensions)) {
      normalized.row_dimensions = normalizeRowDimensionEntries(
        normalized.row_dimensions,
      );
    }

    if (typeof normalized.wrap_lines === "number") {
      normalized.wrap_lines = Math.min(
        WRAP_LINES_MAX,
        Math.max(WRAP_LINES_MIN, Math.round(normalized.wrap_lines)),
      );
    }

    const sortBy =
      typeof next.sort_by === "string" && next.sort_by !== ""
        ? next.sort_by
        : undefined;
    if (sortBy && !sortableFieldNames(next).includes(sortBy)) {
      normalized.sort_by = undefined;
      normalized.sort_dir = undefined;
      normalized.sort_comparison = undefined;
    } else if (!sortBy) {
      if (next.sort_dir !== undefined) normalized.sort_dir = undefined;
      if (next.sort_comparison !== undefined) {
        normalized.sort_comparison = undefined;
      }
    } else if (
      next.sort_comparison !== undefined &&
      !this.isMetricsViewMeasure(sortBy)
    ) {
      // Comparison columns exist only for the metrics view's own measures.
      normalized.sort_comparison = undefined;
    }

    super.updateProperties(normalized);
  }

  /** Update a single measure's conditional formatting in the YAML; pass null to clear it. */
  setMeasureFormatting(
    measureName: string,
    fmt: PivotMeasureFormatting | null,
  ) {
    const measureFormatting = conditionalFormatSpecToMeasureFormatting(
      get(this.specStore).conditional_format,
    );
    if (fmt) {
      measureFormatting[measureName] = fmt;
    } else {
      delete measureFormatting[measureName];
    }
    this.updateProperty(
      "conditional_format",
      measureFormattingToConditionalFormatSpec(measureFormatting),
    );
  }

  /**
   * Patch one field's presentation overrides on its entry in the given list
   * (null removes a key). Widths are clamped to the resizer bounds of the
   * field's role, which is also what the runtime validates.
   */
  setFieldConfig(
    listKey: PivotFieldListKey,
    name: string,
    patch: PivotFieldConfigPatch,
  ) {
    const spec = get(this.specStore) as Partial<PivotSpec> & Partial<TableSpec>;
    const list = spec[listKey];
    let normalized = patch;
    if (typeof patch.width === "number") {
      normalized = {
        ...patch,
        width: clampColumnWidth(this.fieldRole(listKey, name), patch.width),
      };
    }
    this.updateProperty(listKey, setFieldConfigInList(list, name, normalized));
  }

  /**
   * Persist a width the user set by dragging a column edge (null: a reset to
   * the automatic width). The column id is the field name in a table, the
   * measure name or the first row dimension in a pivot.
   */
  setColumnWidth(columnId: string, width: number | null) {
    const spec = get(this.specStore);
    let listKey: PivotFieldListKey | undefined;
    if ("columns" in spec) {
      if (fieldNames(spec.columns).includes(columnId)) listKey = "columns";
    } else if (fieldNames(spec.measures).includes(columnId)) {
      listKey = "measures";
    } else if (fieldNames(spec.row_dimensions)[0] === columnId) {
      listKey = "row_dimensions";
    }
    if (!listKey) return;
    this.setFieldConfig(listKey, columnId, { width });
  }

  /** Persist the initial sort; every key is cleared when `sortBy` is undefined. */
  setSort(
    sortBy: string | undefined,
    sortDir: PivotSortDir | undefined,
    sortComparison: PivotSortComparison | undefined,
  ) {
    this.updateProperties({
      sort_by: sortBy,
      sort_dir: sortDir,
      sort_comparison: sortComparison,
    });
  }

  private fieldRole(listKey: PivotFieldListKey, name: string): ColumnWidthRole {
    if (listKey === "measures") return "measure";
    if (listKey === "columns" && this.isMeasureName(name)) return "measure";
    return "dimension";
  }

  /** A measure of the metrics view or an adhoc measure of this component. */
  private isMeasureName(name: string): boolean {
    const spec = get(this.specStore);
    if (spec.adhoc_measures?.some((measure) => measure.name === name)) {
      return true;
    }
    return this.isMetricsViewMeasure(name);
  }

  /** Only the metrics view's own measures get comparison columns in the pivot. */
  private isMetricsViewMeasure(name: string): boolean {
    const spec = get(this.specStore);
    const metricsViewSpec = get(
      this.parent.metricsView.getMetricsViewFromName(spec.metrics_view),
    ).metricsView;
    return metricsViewSpec?.measures?.some((m) => m.name === name) ?? false;
  }

  /**
   * The sidebar options for the presentation properties: the display group
   * (fit, wrapping) and the sort group. The wrap line count is only shown
   * while something wraps; the sort direction only while a sort field is set.
   */
  private presentationOptions(
    spec: PivotSpec | TableSpec,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): Partial<Record<AllKeys<PivotSpec | TableSpec>, ComponentInputParam>> {
    const measureNames = new Set<string>([
      ...(metricsViewSpec?.measures?.map((measure) => measure.name as string) ??
        []),
      ...(spec.adhoc_measures?.map((measure) => measure.name) ?? []),
    ]);
    const isMeasure = (name: string) => measureNames.has(name);
    const fieldLabel = (name: string) =>
      metricsViewSpec?.measures?.find((measure) => measure.name === name)
        ?.displayName ||
      metricsViewSpec?.dimensions?.find(
        (dimension) => dimension.name === name || dimension.column === name,
      )?.displayName ||
      spec.adhoc_measures?.find((measure) => measure.name === name)
        ?.display_name ||
      name;

    const anyFieldWraps = Object.values(fieldConfigs(spec)).some(
      (config) => config.wrap === true,
    );
    const sortable = sortableFieldNames(spec);
    const sortBy =
      typeof spec.sort_by === "string" && sortable.includes(spec.sort_by)
        ? spec.sort_by
        : "";
    const comparableMeasureNames = new Set(
      metricsViewSpec?.measures?.map((measure) => measure.name as string) ?? [],
    );

    return {
      fit_to_width: {
        type: "boolean",
        label: m.canvas_fit_to_width_label(),
        meta: { defaultValue: false },
      },
      wrap: {
        type: "boolean",
        label: m.canvas_wrap_text_label(),
        meta: { defaultValue: false },
      },
      wrap_headers: {
        type: "boolean",
        label: m.canvas_wrap_headers_label(),
        meta: { defaultValue: false },
      },
      wrap_lines: {
        type: "select",
        label: m.canvas_wrap_lines_label(),
        meta: {
          default: String(WRAP_LINES_DEFAULT),
          numeric: true,
          options: Array.from(
            { length: WRAP_LINES_MAX - WRAP_LINES_MIN + 1 },
            (_, i) => String(WRAP_LINES_MIN + i),
          ).map((value) => ({ value, label: value })),
        },
        showInUI:
          spec.wrap === true || spec.wrap_headers === true || anyFieldWraps,
      },
      sort_by: {
        type: "select",
        label: m.canvas_sort_by_label(),
        meta: {
          default: "",
          placeholder: m.canvas_sort_default_option(),
          options: [
            { value: "", label: m.canvas_sort_default_option() },
            ...sortable.map((name) => ({
              value: name,
              label: fieldLabel(name),
            })),
          ],
        },
      },
      sort_comparison: {
        type: "select",
        label: m.canvas_sort_on_label(),
        meta: {
          default: "",
          placeholder: m.canvas_sort_on_value(),
          options: [
            { value: "", label: m.canvas_sort_on_value() },
            { value: "delta", label: m.canvas_sort_on_delta() },
            {
              value: "percent_change",
              label: m.canvas_sort_on_percent_change(),
            },
          ],
        },
        // Comparison columns exist only for the metrics view's own measures.
        showInUI: sortBy !== "" && comparableMeasureNames.has(sortBy),
      },
      sort_dir: {
        type: "select",
        label: m.canvas_sort_direction_label(),
        meta: {
          default: sortBy ? defaultSortDir(sortBy, isMeasure) : "desc",
          options: [
            { value: "asc", label: m.canvas_sort_ascending() },
            { value: "desc", label: m.canvas_sort_descending() },
          ],
        },
        showInUI: sortBy !== "",
      },
    };
  }

  inputParams(type: "pivot" | "table"): InputParams<PivotSpec | TableSpec> {
    const spec = get(this.specStore);
    const metricsViewSpec = get(
      this.parent.metricsView.getMetricsViewFromName(spec.metrics_view),
    ).metricsView;

    if (type === "pivot") {
      const measureCount =
        "measures" in spec ? fieldNames(spec.measures).length : 0;
      const rowDimensionCount =
        "row_dimensions" in spec ? fieldNames(spec.row_dimensions).length : 0;
      const colDimensionCount =
        "col_dimensions" in spec ? fieldNames(spec.col_dimensions).length : 0;

      // Mirror PivotToolbar: totals only apply when their constituent fields exist.
      const canShowTotalRow = rowDimensionCount > 0 && measureCount > 0;
      const canShowTotalColumn =
        rowDimensionCount > 0 && colDimensionCount > 0 && measureCount > 0;

      return {
        options: {
          metrics_view: {
            type: "metrics",
            label: m.canvas_metrics_view_label(),
          },
          measures: {
            type: "multi_fields_format",
            meta: { allowedTypes: ["measure"] },
            label: m.canvas_measures_label(),
          },
          adhoc_measures: {
            type: "adhoc_measures",
            label: m.canvas_ephemeral_measures_label(),
            optional: true,
            // Managed through the measures selector's create/edit dialog.
            showInUI: false,
          },
          col_dimensions: {
            type: "multi_fields_format",
            meta: { allowedTypes: ["time", "dimension"] },
            label: m.canvas_column_dimensions_label(),
          },
          row_dimensions: {
            type: "multi_fields_format",
            meta: { allowedTypes: ["time", "dimension"] },
            label: m.canvas_row_dimensions_label(),
          },
          hide_totals_col: {
            type: "boolean",
            label: m.canvas_hide_total_column_label(),
            meta: { defaultValue: false },
            showInUI: canShowTotalColumn,
          },
          hide_totals_row: {
            type: "boolean",
            label: m.canvas_hide_total_row_label(),
            meta: { defaultValue: false },
            showInUI: canShowTotalRow,
          },
          totals_row_position: {
            type: "select",
            label: m.canvas_totals_row_position_label(),
            meta: {
              default: "top",
              options: totalsRowPositionOptions(),
            },
            showInUI: canShowTotalRow && spec.hide_totals_row !== true,
          },
          row_limit: {
            type: "select",
            label: m.canvas_row_limit(),
            meta: {
              default: ROW_LIMIT_ALL_VALUE,
              options: [
                ...PIVOT_ROW_LIMIT_OPTIONS.map((limit) => ({
                  value: limit.toString(),
                  label: limit.toString(),
                })),
                { value: ROW_LIMIT_ALL_VALUE, label: m.common_all() },
              ],
            },
          },
          ...this.presentationOptions(spec, metricsViewSpec),
          ...getCommonOptions(),
        },
        filter: getFilterOptions(true, false),
      };
    } else {
      const columns = "columns" in spec ? fieldNames(spec.columns) : [];
      const measureNames = new Set(
        metricsViewSpec?.measures?.map((m) => m.name as string) || [],
      );
      const measureCount = columns.filter((c) => measureNames.has(c)).length;
      const dimensionCount = columns.length - measureCount;
      const canShowTotalRow = dimensionCount > 0 && measureCount > 0;

      return {
        options: {
          metrics_view: {
            type: "metrics",
            label: m.canvas_metrics_view_label(),
          },
          columns: {
            type: "multi_fields_format",
            label: m.canvas_columns_label(),
            meta: { allowedTypes: ["time", "dimension", "measure"] },
          },
          adhoc_measures: {
            type: "adhoc_measures",
            label: m.canvas_ephemeral_measures_label(),
            optional: true,
            // Managed through the measures selector's create/edit dialog.
            showInUI: false,
          },
          hide_totals_row: {
            type: "boolean",
            label: m.canvas_hide_total_row_label(),
            meta: { defaultValue: false },
            showInUI: canShowTotalRow,
          },
          totals_row_position: {
            type: "select",
            label: m.canvas_totals_row_position_label(),
            meta: {
              default: "top",
              options: totalsRowPositionOptions(),
            },
            showInUI: canShowTotalRow && spec.hide_totals_row !== true,
          },
          ...this.presentationOptions(spec, metricsViewSpec),
          ...getCommonOptions(),
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

    // Field entries keep their objects (and so their per-column overrides)
    // across the switch.
    const commonProperties: ComponentCommonProperties &
      ComponentFilterProperties &
      TablePresentationProperties &
      Pick<
        PivotSpec,
        | "hide_totals_row"
        | "hide_totals_col"
        | "totals_row_position"
        | "conditional_format"
      > = {
      title: currentSpec.title,
      description: currentSpec.description,
      dimension_filters: currentSpec.dimension_filters,
      time_filters: currentSpec.time_filters,
      hide_totals_row: currentSpec.hide_totals_row,
      hide_totals_col: currentSpec.hide_totals_col,
      totals_row_position: currentSpec.totals_row_position,
      conditional_format: currentSpec.conditional_format,
      fit_to_width: currentSpec.fit_to_width,
      wrap: currentSpec.wrap,
      wrap_headers: currentSpec.wrap_headers,
      wrap_lines: currentSpec.wrap_lines,
      sort_by: currentSpec.sort_by,
      sort_dir: currentSpec.sort_dir,
      sort_comparison: currentSpec.sort_comparison,
    };

    if ("columns" in currentSpec) {
      const entries = currentSpec.columns ?? [];
      // Table columns may carry `align`, which row dimensions do not take.
      const row_dimensions = normalizeRowDimensionEntries(
        entries.filter((entry) => allDimensions.includes(fieldName(entry))),
      );
      const measures = entries.filter((entry) =>
        allMeasures.includes(fieldName(entry)),
      );

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
          ...(currentSpec.row_dimensions ?? []),
          ...(currentSpec.col_dimensions ?? []).map(stripColumnDimensionKeys),
          ...(currentSpec.measures ?? []),
        ],
      };
    }

    // The sorted field may not have survived the switch.
    if (
      newSpec.sort_by &&
      !sortableFieldNames(newSpec).includes(newSpec.sort_by)
    ) {
      delete newSpec.sort_by;
      delete newSpec.sort_dir;
      delete newSpec.sort_comparison;
    }
    if (
      newSpec.sort_comparison !== undefined &&
      !(newSpec.sort_by && allMeasures.includes(newSpec.sort_by))
    ) {
      delete newSpec.sort_comparison;
    }

    const width = parsedDocument.getIn([...parentPath, "width"]);

    this.specStore.set(newSpec);

    parsedDocument.setIn(parentPath, { [newTableType]: newSpec, width });

    // Save the updated document
    updateEditorContent(parsedDocument.toString(), false, true);
  }
}

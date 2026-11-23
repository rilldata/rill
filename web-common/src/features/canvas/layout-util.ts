import type {
  V1CanvasItem,
  V1CanvasRow,
  V1ComponentSpecRendererProperties,
  V1MetricsViewSpec,
  V1ResolveCanvasResponseResolvedComponents,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";
import { YAMLMap, YAMLSeq, type Document } from "yaml";
import { ResourceKind } from "../entity-management/resource-selectors";
import { buildDefaultArgs, getDeclaredParams } from "../custom-viz/params";
import type { CanvasComponentType, ComponentSpec } from "./components/types";
import { COMPONENT_CLASS_MAP } from "./components/util";
import { combineConditions } from "./stores/block-edit";

// TODO: Move this individual component class
export const initialHeights: Record<CanvasComponentType, number> = {
  line_chart: 320,
  bar_chart: 320,
  area_chart: 320,
  stacked_bar: 320,
  stacked_bar_normalized: 320,
  donut_chart: 320,
  pie_chart: 320,
  heatmap: 320,
  custom_chart: 320,
  funnel_chart: 320,
  combo_chart: 320,
  scatter_plot: 320,
  markdown: 40,
  kpi_grid: 128,
  image: 80,
  table: 300,
  pivot: 300,
  leaderboard: 300,
  component_ref: 320,
  map: 400,
};

// Minimum heights a component can shrink to, when smaller than its initial
// height. Used as the CSS floor so components like pivots (whose content can be
// short, e.g. a low row limit) don't reserve their full initial height.
export const minHeights: Partial<Record<CanvasComponentType, number>> = {
  pivot: 224,
};

export const MIN_HEIGHT = 40;
export const MIN_WIDTH = 3;
export const COLUMN_COUNT = 12;
export const DEFAULT_DASHBOARD_WIDTH = 1200;

export const mousePosition = (() => {
  const store = writable({ x: 0, y: 0 });

  function update(event: MouseEvent) {
    store.set({ x: event.clientX, y: event.clientY });
  }

  window.addEventListener("mousemove", update);

  return store;
})();

type YAMLItem = Record<string, unknown> & {
  width?: number;
};

export type YAMLRow = {
  items?: YAMLItem[];
  height?: string;
  // A top-level entry may instead be a block that holds rows: a tab group (`tabs`) or a
  // container (`rows`), with an optional `name`. Row transactions pass blocks through
  // untouched so their content is preserved (see mapGuard).
  tabs?: unknown;
  rows?: unknown;
  name?: string;
  // Condition that decides who sees the row (or block). Row transactions must keep it.
  if?: unknown;
};

// The properties a plain row can have. An entry without items that has any other property
// is a block, including kinds of block this editor doesn't know about yet.
const PLAIN_ROW_KEYS = new Set(["items", "height", "if"]);

// Like the parser, a list under `tabs` or `rows` makes an entry a block, whatever `items` holds.
function isBlockEntry(entry: YAMLRow): boolean {
  if (entry.tabs != null || entry.rows != null) return true;
  return (
    entry.items === undefined &&
    Object.keys(entry).some((key) => !PLAIN_ROW_KEYS.has(key))
  );
}

export type DragItem = {
  position?: { row: number; column: number };
  type?: string;
};

export function rowsGuard(value: unknown): unknown[] {
  if (!value || !(value instanceof YAMLSeq)) {
    return [];
  } else {
    return value.items;
  }
}

export function mapGuard(value: unknown[]): Array<YAMLRow> {
  return value.map((el) => {
    if (el instanceof YAMLMap) {
      const jsonObject = el.toJSON() as YAMLRow;

      // Preserve blocks verbatim. Coercing them to `items: []` would strip their content
      // and the empty-items cleanup would then delete the entry, destroying the block
      // whenever an unrelated top-level row transaction runs.
      if (jsonObject && isBlockEntry(jsonObject)) {
        // The parser ignores an empty `items` next to a block's content. Drop it, so the
        // cleanup doesn't take the block for an empty row either.
        if (Array.isArray(jsonObject.items) && jsonObject.items.length === 0) {
          const { items, ...block } = jsonObject;
          void items;
          return block;
        }
        return jsonObject;
      }

      // Keep the row's other properties, such as its `if` condition.
      return {
        ...jsonObject,
        items: jsonObject?.items ?? [],
        height: jsonObject?.height ?? MIN_HEIGHT + "px",
      };
    } else {
      return {
        items: [],
        height: MIN_HEIGHT + "px",
      };
    }
  });
}

/**
 * Returns the properties of the canvas item at itemPath other than its component definition under rendererKey,
 * such as `width` and `if`. Code that replaces an item's component spreads them into the new item to keep them.
 */
export function getItemPropertiesExcept(
  doc: Document,
  itemPath: (string | number)[],
  rendererKey: string | number | undefined,
): Record<string, unknown> {
  const item = doc.getIn(itemPath);
  if (!(item instanceof YAMLMap)) return {};
  const properties = item.toJSON() as Record<string, unknown>;
  if (rendererKey !== undefined) delete properties[String(rendererKey)];
  return properties;
}

interface Position {
  row: number;
  col: number;
}

export type YAMLPath = (string | number)[];

// Identifies an editable rows list by the YAML path of its rows sequence: ["rows"] for the
// top level, ["rows", b, "tabs", t, "rows"] for a tab and ["rows", b, "rows"] for a
// container. An undefined target is the top level.
export type EditTarget = { rowsPath: YAMLPath };

export const TOP_LEVEL_ROWS_PATH: YAMLPath = ["rows"];

/**
 * Whether the canvas element at the YAML path is hidden from the current user,
 * because it or one of its ancestors (a row, tab group, tab or container) is in hiddenPaths.
 */
export function isHiddenPath(
  hiddenPaths: Set<string>,
  path: (string | number)[],
): boolean {
  for (let i = 1; i <= path.length; i++) {
    if (hiddenPaths.has(path.slice(0, i).join("."))) return true;
  }
  return false;
}

/** YAML path of a tab's rows (the YAML omits the proto's tab_group wrapper). */
export function tabRowsPath(blockIndex: number, tabIndex: number): YAMLPath {
  return ["rows", blockIndex, "tabs", tabIndex, "rows"];
}

/** YAML path of a container's rows. */
export function containerRowsPath(blockIndex: number): YAMLPath {
  return ["rows", blockIndex, "rows"];
}

/** The rows path of the target, defaulting to the top-level rows. */
export function rowsPathOf(target: EditTarget | undefined): YAMLPath {
  return target?.rowsPath ?? TOP_LEVEL_ROWS_PATH;
}

/** True if two targets are the same rows list. */
export function sameTarget(
  a: EditTarget | undefined,
  b: EditTarget | undefined,
): boolean {
  return rowsPathOf(a).join(".") === rowsPathOf(b).join(".");
}

/**
 * The YAML path of the rows list a component lives in: its path without the trailing
 * [row, "items", col, type].
 */
export function rowsPathOfComponent(componentPath: YAMLPath): YAMLPath {
  return componentPath.slice(0, -4);
}

/** The edit target of the rows list a component lives in, or undefined for a top-level component. */
export function targetOfComponent(
  componentPath: YAMLPath,
): EditTarget | undefined {
  const rowsPath = rowsPathOfComponent(componentPath);
  return rowsPath.length > 1 ? { rowsPath } : undefined;
}

/**
 * Component name prefix for the rows at a rows path, matching the parser's position key
 * (see parse_canvas.go): "" at the top level, "g<b>-t<t>-" in a tab and "c<b>-" in a container.
 */
export function positionKeyPrefix(rowsPath: YAMLPath): string {
  let prefix = "";
  // After the leading "rows", each nesting level is [b, "tabs", t, "rows"] or [b, "rows"].
  let i = 1;
  while (i + 1 < rowsPath.length) {
    if (rowsPath[i + 1] === "tabs") {
      prefix += `g${rowsPath[i]}-t${rowsPath[i + 2]}-`;
      i += 4;
    } else if (rowsPath[i + 1] === "rows") {
      prefix += `c${rowsPath[i]}-`;
      i += 2;
    } else {
      break;
    }
  }
  return prefix;
}

/**
 * The spec rows at a YAML rows path. The YAML and the proto shapes differ: a tab's rows are
 * at row.tabGroup.tabs[t].rows in the spec and a container's at row.container.rows.
 * The returned list is the one in `rows`, so mutating it mutates `rows`. When `create` is
 * set, a missing nested list is added first (proto JSON omits empty lists, e.g. for an
 * empty container). Returns undefined if the path doesn't exist.
 */
export function specRowsAt(
  rows: V1CanvasRow[],
  rowsPath: YAMLPath,
  create = false,
): V1CanvasRow[] | undefined {
  let current: V1CanvasRow[] | undefined = rows;
  let i = 1;
  while (current && i < rowsPath.length) {
    const row: V1CanvasRow | undefined = current[Number(rowsPath[i])];
    let holder: { rows?: V1CanvasRow[] } | undefined;
    if (rowsPath[i + 1] === "tabs") {
      holder = row?.tabGroup?.tabs?.[Number(rowsPath[i + 2])];
      i += 4;
    } else if (rowsPath[i + 1] === "rows") {
      holder = row?.container;
      i += 2;
    } else {
      return undefined;
    }
    if (holder && create) holder.rows ??= [];
    current = holder?.rows;
  }
  return current;
}

/** True if the spec row is a block (a tab group or a container) rather than a plain row. */
export function isBlockSpecRow(row: V1CanvasRow): boolean {
  return !!row.tabGroup || !!row.container;
}

/**
 * Calls visit for the top-level rows and for every nested rows list (each tab's and each
 * container's), with the YAML path of the list.
 */
export function forEachRowsList(
  rows: V1CanvasRow[],
  visit: (rows: V1CanvasRow[], rowsPath: YAMLPath) => void,
  rowsPath: YAMLPath = TOP_LEVEL_ROWS_PATH,
) {
  visit(rows, rowsPath);
  rows.forEach((row, index) => {
    row.tabGroup?.tabs?.forEach((tab, tabIndex) =>
      forEachRowsList(tab.rows ?? [], visit, [
        ...rowsPath,
        index,
        "tabs",
        tabIndex,
        "rows",
      ]),
    );
    if (row.container) {
      forEachRowsList(row.container.rows ?? [], visit, [
        ...rowsPath,
        index,
        "rows",
      ]);
    }
  });
}

interface BaseTransaction {
  insertRow?: boolean;
}

interface MoveItemTransaction extends BaseTransaction {
  type: "move";
  source: Position;
  destination: Position;
}

interface CopyItemTransaction extends BaseTransaction {
  type: "copy";
  source: Position;
  destination: Position;
  insertRow: true;
}

interface DeleteItemTransaction {
  type: "delete";
  target: Position;
}

interface AddItemTransaction extends BaseTransaction {
  type: "add";
  componentType: CanvasComponentType;
  // When set, the new item references this externally defined component
  // (componentType is "component_ref") instead of creating an inline component.
  componentName?: string;
  destination: Position;
}

// An addable canvas item: a built-in component type (inline component) or a
// reference to an externally defined component.
export type AddableItem = CanvasComponentType | { componentName: string };

export type TransactionOperation =
  | MoveItemTransaction
  | CopyItemTransaction
  | DeleteItemTransaction
  | AddItemTransaction;

export interface Transaction {
  operations: TransactionOperation[];
}

export function generateArrayRearrangeFunction(transaction: Transaction) {
  return <I, R extends { items?: I[] }>(
    array: R[],
    newItemGenerator: (
      pos: Position,
      type: CanvasComponentType,
      operationIndex: number,
    ) => I,
    rowUpdater: (row: R, index: number, touched: boolean) => R,
    // Called when an item is moved or copied to another row, to keep what it inherited from its source row.
    itemCarrier?: (item: I, source: R, destination: R) => I,
  ) => {
    const newArray = structuredClone(array);
    const touchedRows: Array<boolean | null> = Array.from(
      { length: newArray.length },
      () => false,
    );

    for (const [operationIndex, op] of transaction.operations.entries()) {
      switch (op.type) {
        case "delete": {
          const { row, col } = op.target;
          const targetRow = newArray[row];
          if (targetRow && targetRow?.items?.[col] !== undefined) {
            targetRow.items.splice(col, 1);
            touchedRows[row] = true;
          }
          break;
        }

        case "move": {
          const { source, destination, insertRow } = op;
          const sourceRow = newArray[source.row];
          const rowIndex = destination.row;

          if (insertRow) {
            newArray.splice(rowIndex, 0, { items: [] } as unknown as R);
          }

          const destinationRow = newArray[rowIndex];

          if (!sourceRow || !destinationRow) break;

          const item = sourceRow.items?.[source.col];

          if (item === undefined) break;

          if (
            sourceRow !== destinationRow &&
            (destinationRow.items?.length ?? 0) >= 4
          ) {
            throw new Error("Maximum number of items reached");
          }

          sourceRow.items?.splice(source.col, 1);

          const insertIndex =
            sourceRow === destinationRow && destination.col > source.col
              ? destination.col - 1
              : destination.col;

          destinationRow.items?.splice(
            insertIndex,
            0,
            sourceRow === destinationRow || !itemCarrier
              ? item
              : itemCarrier(item, sourceRow, destinationRow),
          );

          touchedRows[source.row] = true;
          touchedRows[rowIndex] = true;

          break;
        }

        case "copy": {
          const { source, destination, insertRow } = op;

          const rowIndex = destination.row;
          if (insertRow) {
            newArray.splice(rowIndex, 0, { items: [] } as unknown as R);
          }

          const sourceRow = newArray[source.row];
          const destinationRow = newArray[rowIndex];
          if (!sourceRow || !destinationRow) break;

          const itemCount = destinationRow.items?.length ?? 0;

          if (itemCount >= 4) {
            throw new Error("Maximum number of items reached");
          }

          const item = sourceRow.items?.[source.col];
          if (item === undefined) break;

          let copy: I = structuredClone(item);
          if (sourceRow !== destinationRow && itemCarrier) {
            copy = itemCarrier(copy, sourceRow, destinationRow);
          }
          destinationRow?.items?.splice(destination.col, 0, copy);
          touchedRows[rowIndex] = true;
          break;
        }

        case "add": {
          const { destination, componentType, insertRow } = op;

          const rowIndex = destination.row;
          if (insertRow) {
            newArray.splice(rowIndex, 0, { items: [] } as unknown as R);
          }

          const row = newArray[rowIndex];
          if (!row) break;

          const itemCount = row.items?.length ?? 0;

          if (itemCount >= 4) {
            throw new Error("Maximum number of items reached");
          }

          const newItem = newItemGenerator(
            destination,
            componentType,
            operationIndex,
          );
          row.items?.splice(destination.col, 0, newItem);
          touchedRows[rowIndex] = true;
          break;
        }
      }
    }

    const cleaned = newArray.filter((row, index) => {
      if (row.items?.length === 0) {
        touchedRows[index] = null;
        return false;
      }
      return true;
    });
    const cleanedTouched = touchedRows.filter((row) => row !== null);

    return cleaned.map((row, index) =>
      rowUpdater(row, index, cleanedTouched[index] ?? false),
    );
  };
}

// Generates the resource name for a component at a position. namePrefix disambiguates
// components nested in tabs and containers (e.g. "g2-t0-" or "c1-") so they don't collide
// with top-level ones; it mirrors the position key used by the parser (see positionKeyPrefix).
export function generateId(
  row: number | undefined,
  column: number | undefined,
  canvasName: string,
  namePrefix = "",
) {
  return `${canvasName}--component-${namePrefix}${row ?? 0}-${column ?? 0}`;
}

/**
 * The client-side instance id for a canvas item. Inline components use their resource
 * name, which the parser already derives from the item's position. Items referencing an
 * externally defined component get a positional suffix, so multiple items can reference
 * the same component (with different params) without sharing state.
 */
export function canvasItemInstanceId(
  item: V1CanvasItem,
  row: number,
  column: number,
  namePrefix = "",
): string {
  if (!item.component) return "";
  if (item.definedInCanvas) return item.component;
  return `${item.component}::${namePrefix}${row}-${column}`;
}

/** Extracts the referenced component's resource name from an instance id. */
export function componentNameFromInstanceId(id: string): string {
  const separator = id.indexOf("::");
  return separator === -1 ? id : id.slice(0, separator);
}

export function generateNewAssets(params: {
  transaction: Transaction;
  yamlRows: YAMLRow[];
  specRows: V1CanvasRow[];
  resolvedComponents: V1ResolveCanvasResponseResolvedComponents | undefined;
  canvasName: string;
  // Prefix for generated component names, to keep the components of tabs and containers
  // unique (see positionKeyPrefix). Default "" (top level).
  namePrefix?: string;
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  };
  // Project component resources by name; used to compute default param bindings
  // and optimistic resources for added component references.
  componentResources?: Record<string, V1Resource | undefined>;
}) {
  const {
    yamlRows,
    specRows,
    defaultMetrics,
    canvasName,
    namePrefix = "",
    resolvedComponents,
    transaction,
    componentResources,
  } = params;

  const mover = generateArrayRearrangeFunction(transaction);
  const addedComponentSpecs = transaction.operations.map((op) => {
    if (op.type !== "add" || op.componentName) return undefined;

    return {
      type: op.componentType,
      spec: createComponentSpec(op.componentType, defaultMetrics),
    };
  });

  // For added component references: the smart default param bindings.
  const addedRefParams = transaction.operations.map((op) => {
    if (op.type !== "add" || !op.componentName) return undefined;
    return buildDefaultArgs(
      getDeclaredParams(componentResources?.[op.componentName]),
      defaultMetrics.metricsViewName,
      defaultMetrics.metricsViewSpec,
    );
  });

  // Returns the referenced component name and default params for an "add" operation
  // that adds a component reference, or undefined for inline adds.
  const refAdd = (operationIndex: number) => {
    const op = transaction.operations[operationIndex];
    if (op?.type === "add" && op.componentName) {
      return {
        name: op.componentName,
        params: addedRefParams[operationIndex] ?? {},
      };
    }
    return undefined;
  };

  // Block rows (tab groups and containers) arrive with `items: []` in the spec JSON. Give
  // them no items, like their YAML rows (see mapGuard), so the empty-row cleanup keeps them
  // and the spec, YAML and resolved arrays stay index-aligned through it.
  const specRowsWithBlocks = specRows.map((row) =>
    isBlockSpecRow(row) ? { ...row, items: undefined } : row,
  );

  const resolvedComponentsArray = specRowsWithBlocks.map((row) => {
    // Preserve block rows; their nested components remain resolvable via the existing
    // resolvedComponents map that updateAssets merges in.
    if (!row.items)
      return { ...row, items: undefined as V1Resource[] | undefined };
    const items = row.items.map((item) => {
      return resolvedComponents?.[item?.component ?? ""];
    });
    return { ...row, items: items.filter(itemExists) };
  });

  const updatedYamlRows = mover<YAMLItem, YAMLRow>(
    yamlRows,
    (_, type, operationIndex) => {
      const ref = refAdd(operationIndex);
      if (ref) {
        return {
          component: ref.name,
          ...(Object.keys(ref.params).length ? { params: ref.params } : {}),
          width: 0,
        };
      }
      const spec = getAddedComponentSpec(
        addedComponentSpecs,
        operationIndex,
        type,
        defaultMetrics,
      );
      return {
        [type]: spec,
        width: 0,
      };
    },
    (row, _, touched) => {
      if (!touched || !row.items) return row;
      const items = row.items;
      const updatedItems = items.map((item) => {
        return {
          ...item,
          width: touched ? COLUMN_COUNT / items.length : item.width,
        };
      });

      return {
        ...row,
        items: updatedItems,
      };
    },
    // An item moved out of a row with an `if` condition keeps that condition,
    // so content hidden from some viewers doesn't become visible to them.
    (item, source, destination) => {
      if (source.if === undefined || source.if === destination.if) return item;
      const condition = combineConditions(source.if, item.if);
      return condition === undefined ? item : { ...item, if: condition };
    },
  );

  const updatedSpecRows = mover<V1CanvasItem, V1CanvasRow>(
    specRowsWithBlocks,
    (_, __, operationIndex) => {
      const ref = refAdd(operationIndex);
      if (ref) {
        return {
          component: ref.name,
          params: ref.params,
          width: 0,
          widthUnit: "px",
          definedInCanvas: false,
        };
      }
      return {
        component: undefined,
        width: 0,
        widthUnit: "px",
        definedInCanvas: true,
      };
    },
    (row, index, touched) => {
      const updatedItems = row.items?.map((item, col) => {
        // Items referencing an externally defined component keep their name;
        // only inline components have positional names that track their location.
        if (item.definedInCanvas || !item.component) {
          item.component = generateId(index, col, canvasName, namePrefix);
        }

        return {
          ...item,
          width: touched ? COLUMN_COUNT / (row.items?.length ?? 1) : item.width,
        };
      });

      return {
        ...row,
        items: updatedItems,
      };
    },
  );

  const updatedResolvedComponents = mover<V1Resource, { items?: V1Resource[] }>(
    resolvedComponentsArray,
    (pos, type, operationIndex) => {
      const ref = refAdd(operationIndex);
      if (ref) {
        const resource = componentResources?.[ref.name];
        if (resource) return structuredClone(resource);
        return {
          meta: { name: { name: ref.name, kind: ResourceKind.Component } },
        };
      }
      const spec = getAddedComponentSpec(
        addedComponentSpecs,
        operationIndex,
        type,
        defaultMetrics,
      );
      return createOptimisticResource({
        type,
        spec,
        ...defaultMetrics,
      });
    },
    (row, index) => {
      const updatedItems = row.items?.map((item, col) => {
        if (!item?.meta?.name) return item;
        // Referenced components keep their real resource name (see updatedSpecRows).
        const specItem = updatedSpecRows[index]?.items?.[col];
        if (specItem && !specItem.definedInCanvas && specItem.component) {
          return item;
        }
        item.meta.name.name = generateId(index, col, canvasName, namePrefix);
        return item;
      });
      return {
        ...row,
        items: updatedItems,
      };
    },
  );

  const resolvedComponentsMap: Record<string, V1Resource> = {};

  updatedResolvedComponents.forEach((row) => {
    row.items?.forEach((item) => {
      if (item?.meta?.name?.name) {
        resolvedComponentsMap[item?.meta?.name?.name] = item;
      }
    });
  });

  return {
    newSpecRows: updatedSpecRows,
    newYamlRows: updatedYamlRows,
    newResolvedComponents: resolvedComponentsMap,
    mover,
  };
}

function createComponentSpec(
  componentType: CanvasComponentType,
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  },
) {
  return COMPONENT_CLASS_MAP[componentType].newComponentSpec(
    defaultMetrics.metricsViewName,
    defaultMetrics.metricsViewSpec,
  );
}

function getAddedComponentSpec(
  addedComponentSpecs: Array<
    | {
        type: CanvasComponentType;
        spec: ComponentSpec;
      }
    | undefined
  >,
  index: number,
  type: CanvasComponentType,
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  },
) {
  const addedComponentSpec = addedComponentSpecs[index];
  const spec =
    addedComponentSpec?.type === type
      ? addedComponentSpec.spec
      : createComponentSpec(type, defaultMetrics);

  return structuredClone(spec);
}

function createOptimisticResource(options: {
  type: CanvasComponentType;
  metricsViewName: string;
  metricsViewSpec: V1MetricsViewSpec | undefined;
  spec?: ComponentSpec;
}): V1Resource {
  const { type, metricsViewName, metricsViewSpec } = options;

  const spec =
    options.spec ??
    COMPONENT_CLASS_MAP[type].newComponentSpec(
      metricsViewName,
      metricsViewSpec,
    );

  return {
    meta: {
      name: {
        name: undefined,
        kind: ResourceKind.Component,
      },
    },
    component: {
      state: {
        validSpec: {
          renderer: type,
          rendererProperties:
            spec as unknown as V1ComponentSpecRendererProperties,
        },
      },
      spec: {
        renderer: type,
        rendererProperties:
          spec as unknown as V1ComponentSpecRendererProperties,
      },
    },
  };
}

function itemExists<T>(item: T | null | undefined): item is T {
  return item !== undefined && item !== null;
}

export function getInitialHeight(id: string | undefined) {
  return initialHeights[id as CanvasComponentType] ?? MIN_HEIGHT;
}

export function getMinHeight(id: string | undefined) {
  return minHeights[id as CanvasComponentType] ?? getInitialHeight(id);
}

export function initComponentSpec(
  componentType: CanvasComponentType,
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  },
) {
  const newSpec = createComponentSpec(componentType, defaultMetrics);

  return {
    [componentType]: newSpec,
    width: 0,
  };
}

// Very basic normalization
// Will add something more comprehensive later - bgh
export function normalizeSizeArray(array: (number | null)[]): number[] {
  const zeroed = array.map((el) => el ?? 0);
  const sum = zeroed.reduce((acc, val) => acc + (val || 0), 0);
  const count = array.length;

  if (sum !== 12) {
    return Array.from({ length: count }, () => 12 / count);
  }

  return zeroed;
}

export const hideBorder = new Set<string | undefined>(["markdown", "image"]);

import type {
  V1CanvasItem,
  V1CanvasRow,
  V1ComponentSpecRendererProperties,
  V1MetricsViewSpec,
  V1ResolveCanvasResponseResolvedComponents,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";
import { YAMLMap, YAMLSeq } from "yaml";
import { ResourceKind } from "../entity-management/resource-selectors";
import type { CanvasComponentType, ComponentSpec } from "./components/types";
import { COMPONENT_CLASS_MAP } from "./components/util";

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
  // A top-level entry may instead be a tab group (carries `tabs` and an optional `name`).
  // These are passed through row transactions untouched so their content is preserved.
  tabs?: unknown;
  name?: string;
};

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

      // Preserve tab group rows verbatim. Coercing them to `items: []` would strip their
      // tabs and the empty-items cleanup would then delete the row, destroying the group
      // whenever an unrelated top-level row transaction runs.
      if (jsonObject?.tabs !== undefined) {
        return jsonObject;
      }

      return {
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

interface Position {
  row: number;
  col: number;
}

// Identifies an editable rows container. undefined targets the top-level rows; a tab target
// scopes editing to one tab's rows (at YAML path rows[blockIndex].tabs[tabIndex].rows).
export type EditTarget = { blockIndex: number; tabIndex: number };

/** YAML path to a tab's rows sequence. */
export function tabRowsPath(blockIndex: number, tabIndex: number) {
  return ["rows", blockIndex, "tabs", tabIndex, "rows"];
}

/** Component name prefix for a tab, matching the parser's position key (see parse_canvas.go). */
export function tabNamePrefix(blockIndex: number, tabIndex: number) {
  return `g${blockIndex}-t${tabIndex}-`;
}

/** Derive the tab target a component path lives in, or undefined for a top-level component. */
export function tabTargetFromPath(
  path: (string | number)[],
): EditTarget | undefined {
  if (path.length >= 5 && path[0] === "rows" && path[2] === "tabs") {
    return { blockIndex: Number(path[1]), tabIndex: Number(path[3]) };
  }
  return undefined;
}

/** Component name prefix for the tab a path lives in, or "" for a top-level component. */
export function namePrefixFromPath(path: (string | number)[]): string {
  const target = tabTargetFromPath(path);
  return target ? tabNamePrefix(target.blockIndex, target.tabIndex) : "";
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
  destination: Position;
}

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

          destinationRow.items?.splice(insertIndex, 0, item);

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

          const copy = structuredClone(item);
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
// components nested in tabs (e.g. "g2-t0-") so they don't collide with top-level ones;
// it mirrors the position key used by the parser (see parse_canvas.go).
export function generateId(
  row: number | undefined,
  column: number | undefined,
  canvasName: string,
  namePrefix = "",
) {
  return `${canvasName}--component-${namePrefix}${row ?? 0}-${column ?? 0}`;
}

export function generateNewAssets(params: {
  transaction: Transaction;
  yamlRows: YAMLRow[];
  specRows: V1CanvasRow[];
  resolvedComponents: V1ResolveCanvasResponseResolvedComponents | undefined;
  canvasName: string;
  // Prefix for generated component names, to keep tab components unique. Default "" (top level).
  namePrefix?: string;
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  };
}) {
  const {
    yamlRows,
    specRows,
    defaultMetrics,
    canvasName,
    namePrefix = "",
    resolvedComponents,
    transaction,
  } = params;

  const mover = generateArrayRearrangeFunction(transaction);
  const addedComponentSpecs = transaction.operations.map((op) => {
    if (op.type !== "add") return undefined;

    return {
      type: op.componentType,
      spec: createComponentSpec(op.componentType, defaultMetrics),
    };
  });

  const resolvedComponentsArray = specRows.map((row) => {
    // Preserve tab group rows (no items) so this array stays index-aligned with the spec
    // and YAML arrays through the cleanup step; their tab components remain resolvable via
    // the existing resolvedComponents map that updateAssets merges in.
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
  );

  const updatedSpecRows = mover<V1CanvasItem, V1CanvasRow>(
    specRows,
    () => {
      return {
        component: undefined,
        width: 0,
        widthUnit: "px",
        definedInCanvas: true,
      };
    },
    (row, index, touched) => {
      const updatedItems = row.items?.map((item, col) => {
        item.component = generateId(index, col, canvasName, namePrefix);

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

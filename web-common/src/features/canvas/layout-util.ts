import type {
  V1CanvasRow,
  V1CanvasItem,
  V1MetricsViewSpec,
  V1ResolveCanvasResponseResolvedComponents,
  V1Resource,
  V1ComponentSpecRendererProperties,
} from "@rilldata/web-common/runtime-client";
import { YAMLMap, YAMLSeq } from "yaml";
import type { CanvasComponentType } from "./components/types";
import { writable } from "svelte/store";
import { COMPONENT_CLASS_MAP } from "./components/util";
import { ResourceKind } from "../entity-management/resource-selectors";

export const initialHeights: Record<CanvasComponentType, number> = {
  line_chart: 320,
  bar_chart: 320,
  area_chart: 320,
  stacked_bar: 320,
  stacked_bar_normalized: 320,
  markdown: 40,
  kpi_grid: 128,
  image: 80,
  table: 300,
  pivot: 300,
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
  items: YAMLItem[];
  height?: string;
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
      const jsonObject = el.toJSON() as Partial<YAMLRow>;

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
    newItemGenerator: (pos: Position, type: CanvasComponentType) => I,
    rowUpdater: (row: R, index: number, touched: boolean) => R,
  ) => {
    const newArray = structuredClone(array);
    const touchedRows: Array<boolean | null> = Array.from(
      { length: newArray.length },
      () => false,
    );

    for (const op of transaction.operations) {
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

          const rowIndex = destination.row;
          const sourceRow = newArray[source.row];
          if (insertRow) {
            newArray.splice(rowIndex, 0, { items: [] } as unknown as R);
          }

          const destinationRow = newArray[rowIndex];

          if (!sourceRow || !destinationRow) break;

          const item = sourceRow?.items?.[source.col];

          if (item === undefined) break;

          sourceRow.items?.splice(source.col, 1);
          destinationRow.items?.splice(destination.col, 0, item);
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

          const newItem = newItemGenerator(destination, componentType);
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

function generateId(
  row: number | undefined,
  column: number | undefined,
  canvasName: string,
) {
  return `${canvasName}--component-${row ?? 0}-${column ?? 0}`;
}

export function generateNewAssets(params: {
  transaction: Transaction;
  yamlRows: YAMLRow[];
  specRows: V1CanvasRow[];
  resolvedComponents: V1ResolveCanvasResponseResolvedComponents | undefined;
  canvasName: string;
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
    resolvedComponents,
    transaction,
  } = params;
  const mover = generateArrayRearrangeFunction(transaction);

  const resolvedComponentsArray = specRows.map((row) => {
    const items =
      row.items?.map((item) => {
        return resolvedComponents?.[item?.component ?? ""];
      }) ?? [];
    return { ...row, items: items.filter(itemExists) };
  });

  const updatedYamlRows = mover<YAMLItem, YAMLRow>(
    yamlRows,
    (_, type) => {
      return {
        ...initComponentSpec(type, defaultMetrics),
        width: 0,
      };
    },
    (row, _, touched) => {
      if (!touched) return row;
      const updatedItems = row.items.map((item) => {
        return {
          ...item,
          width: touched ? COLUMN_COUNT / row.items.length : item.width,
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
        item.component = generateId(index, col, canvasName);

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

  const updatedResolvedComponents = mover<V1Resource, { items: V1Resource[] }>(
    resolvedComponentsArray,
    (pos, type) => {
      return createOptimisticResource({
        type,
        ...defaultMetrics,
      });
    },
    (row, index) => {
      const updatedItems = row.items.map((item, col) => {
        if (!item?.meta?.name) return item;
        item.meta.name.name = generateId(index, col, canvasName);
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
    row.items.forEach((item) => {
      if (item?.meta?.name?.name) {
        resolvedComponentsMap[item?.meta?.name?.name] = item;
      }
    });
  });

  return {
    newSpecRows: updatedSpecRows,
    newYamlRows: updatedYamlRows,
    newResolvedComponents: resolvedComponentsMap,
  };
}

function createOptimisticResource(options: {
  type: CanvasComponentType;
  metricsViewName: string;
  metricsViewSpec: V1MetricsViewSpec | undefined;
}): V1Resource {
  const { type, metricsViewName, metricsViewSpec } = options;

  const spec = COMPONENT_CLASS_MAP[type].newComponentSpec(
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
  const newSpec = COMPONENT_CLASS_MAP[componentType].newComponentSpec(
    defaultMetrics.metricsViewName,
    defaultMetrics.metricsViewSpec,
  );

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

import type {
  V1CanvasRow as APIV1CanvasRow,
  V1CanvasItem,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { YAMLMap, YAMLSeq } from "yaml";
import type { CanvasComponentType } from "./components/types";
import { getComponentRegistry } from "./components/util";

export const initialHeights: Record<CanvasComponentType, number> = {
  line_chart: 320,
  bar_chart: 320,
  area_chart: 320,
  stacked_bar: 320,
  stacked_bar_normalized: 320,
  markdown: 40,
  kpi: 128,
  kpi_grid: 128,
  image: 80,
  table: 300,
};

const componentRegistry = getComponentRegistry();

export const MIN_HEIGHT = 40;
export const MIN_WIDTH = 3;
export const COLUMN_COUNT = 12;
export const DEFAULT_DASHBOARD_WIDTH = 1200;

type YAMLItem = Record<string, unknown> & {
  width?: number;
};

export type YAMLRow = {
  items: (YAMLItem | null)[];
  height?: string;
};

// Items are nulled out when removed from the canvas
type V1CanvasRow = Omit<APIV1CanvasRow, "items"> & {
  items: (V1CanvasItem | null)[];
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

export function moveToRow<T extends YAMLRow | V1CanvasRow>(
  rows: Array<T | null>,
  items: DragItem[],
  dropPosition?: {
    column?: number;
    row: number;
    copy?: boolean;
  },
  defaultMetrics?: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  },
): Array<T> {
  const rowsClone = structuredClone(rows);

  let height = MIN_HEIGHT;

  const stringHeight =
    defaultMetrics ||
    (dropPosition && typeof rowsClone[dropPosition.row]?.height === "string") ||
    typeof rowsClone[items?.[0]?.position?.row ?? 0]?.height === "string";

  const baseRow = {
    items: [] as T["items"],
    height: stringHeight ? MIN_HEIGHT + "px" : MIN_HEIGHT,
  } as T;

  const destinationRow: T =
    dropPosition?.column === undefined
      ? baseRow
      : (rowsClone[dropPosition.row] ?? baseRow);

  const existingNumericHeight = stringHeight
    ? isNaN(parseInt(String(destinationRow?.height).split("px")[0]))
      ? MIN_HEIGHT
      : parseInt(String(destinationRow?.height).split("px")[0])
    : ((destinationRow?.height as number) ?? MIN_HEIGHT);

  const touchedRows = new Set(
    items.map((el) => el.position?.row).filter(itemExists),
  );

  if (!destinationRow?.items) return [];

  const movedComponents: (YAMLItem | V1CanvasItem)[] = [];

  items.forEach((item) => {
    if (!item.position && item.type) {
      movedComponents.push(
        defaultMetrics
          ? {
              ...createComponent(
                item.type as CanvasComponentType,
                defaultMetrics,
              ),
              width: 0,
            }
          : {
              component: undefined,
              width: 0,
              widthUnit: "px",
              definedInCanvas: true,
            },
      );
    } else if (dropPosition?.copy && item.position) {
      const row = rowsClone[item.position.row];
      if (!row) return;
      const component = row.items?.[item.position.column];
      if (!component) return;

      movedComponents.push(structuredClone(component));
    } else if (item.position) {
      const row = rowsClone[item.position.row];
      if (!row) return;
      const component = row.items?.[item.position.column];
      if (!component) return;

      movedComponents.push(structuredClone(component));
      row.items[item.position.column] = null;
    }

    height = Math.max(existingNumericHeight, getInitialHeight(item.type));
  });

  if (dropPosition) {
    destinationRow.items.splice(
      dropPosition.column ?? 0,
      0,
      ...movedComponents.filter((i) => i !== null),
    );

    destinationRow.height = stringHeight ? height + "px" : height;

    if (destinationRow.items?.filter(itemExists).length > 4) {
      return [];
    }

    const baseWidth = COLUMN_COUNT / destinationRow.items.length;

    destinationRow.items.forEach((_, i) => {
      if (!destinationRow.items[i]) return;
      destinationRow.items[i].width = baseWidth;
    });
  }

  touchedRows.forEach((rowIndex) => {
    const row = rowsClone[rowIndex];

    if (!row?.items) return;

    const validItemsLeft = row.items.filter((i) => i !== null);

    if (!validItemsLeft.length) {
      rowsClone[rowIndex] = null;
    } else {
      const baseWidth = COLUMN_COUNT / validItemsLeft.length;

      validItemsLeft.forEach((_, i) => {
        if (!validItemsLeft[i]) return;
        validItemsLeft[i].width = baseWidth;
      });

      row.items = validItemsLeft;
    }
  });

  if (dropPosition) {
    if (dropPosition?.column === undefined) {
      rowsClone.splice(dropPosition?.row, 0, destinationRow);
    } else {
      rowsClone[dropPosition.row] = destinationRow;
    }
  }

  const filtered = rowsClone.filter((row) => row !== null);

  return filtered;
}

function itemExists<T>(item: T | null | undefined): item is T {
  return item !== undefined && item !== null;
}

export function getInitialHeight(id: string | undefined) {
  return initialHeights[id as CanvasComponentType] ?? MIN_HEIGHT;
}

function createComponent(
  componentType: CanvasComponentType,
  defaultMetrics: {
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
  },
) {
  const newSpec = componentRegistry[componentType].newComponentSpec(
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

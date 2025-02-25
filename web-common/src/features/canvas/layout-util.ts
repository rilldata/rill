import { YAMLMap, YAMLSeq } from "yaml";
import type { CanvasComponentType } from "./components/types";
import type {
  V1CanvasItem,
  V1CanvasRow as APIV1CanvasRow,
} from "@rilldata/web-common/runtime-client";
import { getComponentRegistry } from "./components/util";

const initialHeights: Record<CanvasComponentType, number> = {
  line_chart: 350,
  bar_chart: 400,
  area_chart: 400,
  stacked_bar: 400,
  stacked_bar_normalized: 400,
  markdown: 160,
  kpi: 200,
  kpi_grid: 200,
  image: 420,
  table: 400,
};

const componentRegistry = getComponentRegistry();

const MIN_HEIGHT = 40;
const COLUMN_COUNT = 12;

type YAMLItem = Record<string, unknown> & {
  width?: number;
};

type YAMLRow = {
  items: (YAMLItem | null)[];
  height?: string;
};

type V1CanvasRow = Omit<APIV1CanvasRow, "items"> & {
  items: (V1CanvasItem | null)[];
};

export type DragItem = {
  position?: { row: number; column: number };

  type?: CanvasComponentType;
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
  },
  defaultMetrics?: {
    metricsView: string;
    measure: string;
    dimension: string;
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
              ...createComponent(item.type, defaultMetrics),
              width: 0,
            }
          : {
              component: undefined,
              width: 0,
              widthUnit: "px",
              definedInCanvas: true,
            },
      );
    } else if (item.position) {
      const row = rowsClone[item.position.row];
      if (!row) return;
      const component = row.items?.[item.position.column];
      if (!component) return;

      movedComponents.push(structuredClone(component));
      row.items[item.position.column] = null;
    }

    height = Math.max(existingNumericHeight, getInitalHeight(item.type));
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

function getInitalHeight(id: string | undefined) {
  return initialHeights[id as CanvasComponentType] ?? MIN_HEIGHT;
}

function createComponent(
  componentType: CanvasComponentType,
  defaultMetrics: {
    metricsView: string;
    measure: string;
    dimension: string;
  },
) {
  const newSpec = componentRegistry[componentType].newComponentSpec(
    defaultMetrics.metricsView,
    defaultMetrics.measure,
    defaultMetrics.dimension,
  );

  return {
    [componentType]: newSpec,
    width: 0,
  };
}

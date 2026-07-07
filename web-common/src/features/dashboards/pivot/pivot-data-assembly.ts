import {
  MAX_ROW_EXPANSION_LIMIT,
  SHOW_MORE_BUTTON,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import type { ColumnDef } from "tanstack-table-8-svelte-5";
import {
  getFiltersForCell,
  getPivotConfigKey,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import { reduceTableCellDataIntoRows } from "./pivot-table-transformations";
import type { PivotDataRow, PivotDataStoreConfig, PivotFilter } from "./types";

export interface PivotDataCache {
  expandedTableMap: Record<string, PivotDataRow[]>;
  lastPivotColumnDef: ColumnDef<PivotDataRow>[];
  lastPivotData: PivotDataRow[];
  lastProcessedConfigKey: string;
  lastProcessedRowPage: number;
  lastTotalColumns: number;
}

export interface BasePivotDataResult {
  isCellDataEmpty: boolean;
  pivotData: PivotDataRow[];
  pivotSkeleton: PivotDataRow[];
}

export interface FinalPivotStateDetails {
  activeCellFilters: PivotFilter | undefined;
  data: PivotDataRow[];
  reachedEndForRowData: boolean;
}

export function createPivotDataCache(): PivotDataCache {
  return {
    expandedTableMap: {},
    lastPivotColumnDef: [],
    lastPivotData: [],
    lastProcessedConfigKey: "",
    lastProcessedRowPage: 0,
    lastTotalColumns: 0,
  };
}

export function syncPivotCacheToConfig(
  cache: PivotDataCache,
  configKey: string,
) {
  if (configKey === cache.lastProcessedConfigKey) return;
  cache.lastProcessedRowPage = 0;
  cache.lastProcessedConfigKey = configKey;
}

export function getPivotSkeletonForPage(
  config: PivotDataStoreConfig,
  cache: PivotDataCache,
  rowTotals: PivotDataRow[],
) {
  const rowPage = config.pivot.rowPage;
  if (config.isFlat || rowPage <= 1) return rowTotals;

  if (rowPage > cache.lastProcessedRowPage) {
    return [...cache.lastPivotData, ...rowTotals];
  }
  return cache.lastPivotData;
}

export function assembleBasePivotData(args: {
  anchorDimension: string;
  cache: PivotDataCache;
  cellData: PivotDataRow[];
  columnDimensionAxes: Record<string, string[]>;
  config: PivotDataStoreConfig;
  configKey: string;
  pivotSkeleton: PivotDataRow[];
  rowDimensionValues: string[];
}): BasePivotDataResult {
  const {
    anchorDimension,
    cache,
    cellData,
    columnDimensionAxes,
    config,
    configKey,
    pivotSkeleton,
    rowDimensionValues,
  } = args;

  if (configKey in cache.expandedTableMap) {
    return {
      isCellDataEmpty: cellData.length === 0,
      pivotData: cache.expandedTableMap[configKey],
      pivotSkeleton,
    };
  }

  let tableDataWithCells: PivotDataRow[];
  if (config.isFlat) {
    tableDataWithCells = assembleFlatPivotRows(config, cache, cellData);
  } else {
    tableDataWithCells = reduceTableCellDataIntoRows(
      config,
      anchorDimension,
      rowDimensionValues,
      columnDimensionAxes,
      pivotSkeleton,
      cellData,
    );
  }

  return {
    isCellDataEmpty: cellData.length === 0,
    pivotData: structuredClone(tableDataWithCells),
    pivotSkeleton,
  };
}

function assembleFlatPivotRows(
  config: PivotDataStoreConfig,
  cache: PivotDataCache,
  cellData: PivotDataRow[],
) {
  const rowPage = config.pivot.rowPage;
  if (rowPage > 1 && rowPage > cache.lastProcessedRowPage) {
    return [...cache.lastPivotData, ...cellData];
  }
  if (rowPage > 1) return cache.lastPivotData;
  return cellData;
}

export function cacheExpandedPivotData(
  cache: PivotDataCache,
  config: PivotDataStoreConfig,
  tableDataExpanded: PivotDataRow[],
) {
  const key = getPivotConfigKey(config);
  cache.expandedTableMap = {
    [key]: tableDataExpanded,
  };
}

export function buildFinalPivotStateDetails(args: {
  anchorDimension: string;
  columnDimensionAxes: Record<string, string[]> | undefined;
  config: PivotDataStoreConfig;
  data: PivotDataRow[];
  hasMoreRows: boolean;
  isCellDataEmpty: boolean;
  rowDimensionValues: string[];
  rowOffset: number;
}): FinalPivotStateDetails {
  const {
    anchorDimension,
    columnDimensionAxes,
    config,
    hasMoreRows,
    isCellDataEmpty,
    rowDimensionValues,
    rowOffset,
  } = args;

  const data = addOutermostShowMoreRow(
    args.data,
    anchorDimension,
    config.pivot.outermostRowLimit ?? config.pivot.rowLimit,
    hasMoreRows,
  );

  const activeCellFilters = getActiveCellFilters(
    config,
    columnDimensionAxes,
    data,
  );

  return {
    activeCellFilters,
    data,
    reachedEndForRowData: getReachedEndForRowData(
      config,
      isCellDataEmpty,
      rowDimensionValues,
      rowOffset,
    ),
  };
}

function addOutermostShowMoreRow(
  data: PivotDataRow[],
  anchorDimension: string,
  effectiveOutermostLimit: number | undefined,
  hasMoreRows: boolean,
) {
  if (
    !hasMoreRows ||
    !effectiveOutermostLimit ||
    effectiveOutermostLimit >= MAX_ROW_EXPANSION_LIMIT
  ) {
    return data;
  }

  return [
    ...data,
    {
      [anchorDimension]: SHOW_MORE_BUTTON,
      __currentLimit: effectiveOutermostLimit,
    } as PivotDataRow,
  ];
}

function getActiveCellFilters(
  config: PivotDataStoreConfig,
  columnDimensionAxes: Record<string, string[]> | undefined,
  tableData: PivotDataRow[],
) {
  const activeCell = config.pivot.activeCell;
  if (!activeCell) return undefined;

  return getFiltersForCell(
    config,
    activeCell.rowId,
    activeCell.columnId,
    columnDimensionAxes,
    tableData,
  );
}

function getReachedEndForRowData(
  config: PivotDataStoreConfig,
  isCellDataEmpty: boolean,
  rowDimensionValues: string[],
  rowOffset: number,
) {
  const rowPage = config.pivot.rowPage;
  if (config.isFlat) return isCellDataEmpty && rowPage > 1;

  const rowLimit = config.pivot.rowLimit;
  if (rowLimit !== undefined) {
    return rowOffset + rowDimensionValues.length >= rowLimit;
  }
  return rowDimensionValues.length === 0 && rowPage > 1;
}

export function updatePivotDataCache(
  cache: PivotDataCache,
  args: {
    columnDef: ColumnDef<PivotDataRow>[];
    data: PivotDataRow[];
    rowPage: number;
    totalColumns: number;
  },
) {
  cache.lastPivotData = args.data;
  cache.lastProcessedRowPage = args.rowPage;
  cache.lastPivotColumnDef = args.columnDef;
  cache.lastTotalColumns = args.totalColumns;
}

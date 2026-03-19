/**
 * Factory that creates all click-to-filter orchestration logic for a
 * canvas pivot component.
 *
 * Returns readable stores for selection/row-highlight state, click
 * handlers for cells and column headers, and a destroy function for
 * cleanup.
 */
import {
  type PivotClickSelectionState,
  buildClickSelection,
  cellKey,
  columnHeaderKey,
  createEmptyClickSelectionState,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-click-selection";
import {
  type PivotRowSelectionState,
  computePivotRowSelection,
  extractDimensionFiltersFromExpression,
  extractSelectionDimensionFilters,
  getDimensionValuesForRow,
  getFiltersForColumnHeader,
  getFiltersForRowHeader,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
import { getFiltersForCell } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataStore,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { V1Expression } from "@rilldata/web-common/runtime-client";
import {
  type Readable,
  type Writable,
  derived,
  get,
  writable,
} from "svelte/store";
import type { FilterManager } from "../../stores/filter-manager";

interface PivotClickToFilterArgs {
  pivotConfig: Readable<PivotDataStoreConfig>;
  pivotDataStore: PivotDataStore;
  filterManager: FilterManager;
  metricsViewName: string;
  componentId: string;
  activeComponent: Readable<string | null>;
  selfFilteredDimensions: Writable<Set<string>>;
  whereFilterStore: Readable<V1Expression | undefined>;
  onBecomeActive?: () => void;
  onBecomeInactive?: () => void;
}

export interface PivotClickToFilterResult {
  clickSelection: Readable<PivotClickSelectionState>;
  rowSelectionState: Readable<PivotRowSelectionState | undefined>;
  handleCellClickToFilter: (
    rowId: string,
    columnId: string,
    isRowHeader: boolean,
  ) => void;
  handleColumnHeaderClick: (dimensionPath: Record<string, string>) => void;
  destroy: () => void;
}

export function createPivotClickToFilter(
  args: PivotClickToFilterArgs,
): PivotClickToFilterResult {
  const {
    pivotConfig,
    pivotDataStore,
    filterManager,
    metricsViewName,
    componentId,
    activeComponent,
    selfFilteredDimensions,
    whereFilterStore,
    onBecomeActive,
    onBecomeInactive,
  } = args;

  // --- Internal click selection state ---
  const clickSelectionStore = writable<PivotClickSelectionState>(
    createEmptyClickSelectionState(),
  );

  // Maps rowId → dimension column index for dimension-cell clicks (flat table only)
  const rowDimClickIndex = new Map<string, number>();

  // --- Prune selfFilteredDimensions when filters are cleared externally ---
  const pruneUnsub = whereFilterStore.subscribe(($whereFilter) => {
    const activeDims = new Set(
      $whereFilter?.cond?.exprs
        ?.map((e) => e.cond?.exprs?.[0]?.ident)
        .filter(Boolean) ?? [],
    );
    const dims = get(selfFilteredDimensions);
    let changed = false;
    for (const dim of dims) {
      if (!activeDims.has(dim)) {
        dims.delete(dim);
        changed = true;
      }
    }
    if (changed) {
      selfFilteredDimensions.set(new Set(dims));
    }
  });

  // --- Clear click selection when no self-filtered dimensions remain ---
  const clearUnsub = selfFilteredDimensions.subscribe(($selfFiltered) => {
    if ($selfFiltered.size === 0) {
      rowDimClickIndex.clear();
      clickSelectionStore.set(createEmptyClickSelectionState());
      onBecomeInactive?.();
    }
  });

  // --- Yield active state when another component becomes active ---
  const activeUnsub = activeComponent.subscribe(($activeId) => {
    if ($activeId !== componentId) {
      selfFilteredDimensions.set(new Set());
    }
  });

  // --- Derived row selection state ---
  const dimensionFilterMap = derived(
    [whereFilterStore, selfFilteredDimensions],
    ([$whereFilter, $selfFiltered]) =>
      extractSelectionDimensionFilters($whereFilter, [...$selfFiltered]),
  );

  const rowSelectionState: Readable<PivotRowSelectionState | undefined> =
    derived(
      [pivotDataStore, pivotConfig, dimensionFilterMap],
      ([$pivotData, $config, $dimFilterMap]) => {
        if (!$pivotData?.data || !$config) return undefined;
        return computePivotRowSelection(
          $config,
          $pivotData.data,
          $dimFilterMap,
        );
      },
    );

  // --- Retained value computation for safe deselect ---

  function collectRetainedDimensionValues(
    remainingRowHeaders: Set<string>,
    remainingCells: Set<string>,
    remainingColHeaders: Set<string>,
  ): Map<string, Set<string>> {
    const retainedValues = new Map<string, Set<string>>();

    const addRetainedValue = (dimensionName: string, value: string) => {
      let valueSet = retainedValues.get(dimensionName);
      if (!valueSet) {
        valueSet = new Set();
        retainedValues.set(dimensionName, valueSet);
      }
      valueSet.add(value);
    };

    const $config = get(pivotConfig);
    const $data = get(pivotDataStore);
    if (!$config || !$data?.data) return retainedValues;

    // Values from remaining row header selections
    for (const rowId of remainingRowHeaders) {
      const rowDimValues = getDimensionValuesForRow($config, rowId, $data.data);
      for (const { dimensionName, value } of rowDimValues) {
        addRetainedValue(dimensionName, value);
      }
    }

    // Values from remaining data cell selections (row + column dimensions)
    for (const selectionKey of remainingCells) {
      const separatorIndex = selectionKey.indexOf(":");
      if (separatorIndex === -1) continue;
      const rowId = selectionKey.slice(0, separatorIndex);
      const columnId = selectionKey.slice(separatorIndex + 1);
      // For flat-table dimension cell clicks, only include dims up to the clicked column
      const dimIdx = $config.isFlat
        ? $config.rowDimensionNames.indexOf(columnId)
        : -1;
      const upToDimensionIndex = dimIdx >= 0 ? dimIdx : undefined;
      const cellFilter = getFiltersForCell(
        $config,
        rowId,
        columnId,
        $data.columnDimensionAxes ?? {},
        $data.data,
        upToDimensionIndex,
      );
      if (cellFilter.filters) {
        const extractedFilters = extractDimensionFiltersFromExpression(
          cellFilter.filters,
        );
        for (const { dimensionName, values } of extractedFilters) {
          for (const value of values) addRetainedValue(dimensionName, value);
        }
      }
    }

    // Values from remaining column header selections
    // Key format: JSON-serialised sorted entries, e.g. '[["dim","val"]]'
    for (const headerKey of remainingColHeaders) {
      try {
        const entries: [string, string][] = JSON.parse(headerKey);
        for (const [dimensionName, value] of entries) {
          addRetainedValue(dimensionName, value);
        }
      } catch {
        // Malformed key; skip
      }
    }

    return retainedValues;
  }

  // --- Core filter application ---

  function applyDimensionFilters(
    filters: V1Expression,
    isDeselect: boolean,
    updateSelectionSets: (
      rowHeaders: Set<string>,
      cells: Set<string>,
      colHeaders: Set<string>,
    ) => void,
  ) {
    const dimensionFilters = extractDimensionFiltersFromExpression(filters);
    if (dimensionFilters.length === 0) return;

    // Capture which dimensions were already in the global filter before this click.
    // Dimensions present here were not added by this component and should not be
    // excluded from its own query (they are not "self-filtered").
    const preClickWhere = get(whereFilterStore);
    const preExistingDims = new Set(
      preClickWhere?.cond?.exprs
        ?.map((e) => e.cond?.exprs?.[0]?.ident)
        .filter(Boolean) ?? [],
    );

    const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
    if (!filterClass) return;

    // Update selection sets first so we can compute retained values
    const $clickSelection = get(clickSelectionStore);
    const updatedRowHeaders = new Set($clickSelection.rowHeaderSelections);
    const updatedCells = new Set($clickSelection.cellSelections);
    const updatedColHeaders = new Set($clickSelection.columnHeaderSelections);
    updateSelectionSets(updatedRowHeaders, updatedCells, updatedColHeaders);

    // Clear temporary filter status for all dimensions being changed
    dimensionFilters.forEach(({ dimensionName }) => {
      filterManager.checkTemporaryFilter(dimensionName, [metricsViewName]);
    });

    let filterString: string | null = null;
    if (isDeselect) {
      // Compute which values are still needed by remaining selections
      const retainedValues = collectRetainedDimensionValues(
        updatedRowHeaders,
        updatedCells,
        updatedColHeaders,
      );

      // Only toggle (remove) values that are no longer referenced
      dimensionFilters.forEach(({ dimensionName, values }) => {
        const stillNeeded = retainedValues.get(dimensionName);
        const orphanedValues = stillNeeded
          ? values.filter((v) => !stillNeeded.has(v))
          : values;

        if (orphanedValues.length > 0) {
          filterString = filterClass.toggleDimensionValueSelections(
            dimensionName,
            orphanedValues,
            false,
            false,
          );
        }
      });

      // If no values were actually removed, get current filter string for URL
      if (filterString === null) {
        dimensionFilters.forEach(({ dimensionName }) => {
          filterString = filterClass.addDimensionValueSelections(
            dimensionName,
            [],
          );
        });
      }
    } else {
      dimensionFilters.forEach(({ dimensionName, values }) => {
        filterString = filterClass.addDimensionValueSelections(
          dimensionName,
          values,
        );
      });
    }

    clickSelectionStore.set(
      buildClickSelection(
        updatedRowHeaders,
        updatedCells,
        updatedColHeaders,
        rowDimClickIndex,
      ),
    );

    // Mark only newly-added dimensions as self-filtered. Dimensions that were
    // already in the global filter before this click are not owned by this
    // component and must not be stripped from its own query.
    const wasInactive = get(selfFilteredDimensions).size === 0;
    selfFilteredDimensions.update((dims) => {
      const next = new Set(dims);
      dimensionFilters.forEach(({ dimensionName }) => {
        if (!preExistingDims.has(dimensionName)) {
          next.add(dimensionName);
        }
      });
      return next;
    });
    if (wasInactive && get(selfFilteredDimensions).size > 0) {
      onBecomeActive?.();
    }

    // Single batch URL update with the final filter string
    if (filterString !== null) {
      filterManager.applyFiltersToUrl(
        new Map([[metricsViewName, filterString]]),
      );
    }
  }

  // --- Click handlers ---

  function handleCellClickToFilter(
    rowId: string,
    columnId: string,
    isRowHeader: boolean,
  ) {
    const $config = get(pivotConfig);
    const $data = get(pivotDataStore);
    if (!$config || !$data?.data) return;

    const $clickSelection = get(clickSelectionStore);

    // Determine if this click is deselecting a previously selected element
    const isDeselect = isRowHeader
      ? $clickSelection.isRowHeaderSelected(rowId)
      : $clickSelection.isCellSelected(rowId, columnId);

    // For flat-table dimension cell clicks, only filter up to (and including) the
    // clicked column's dimension index, not all row dimensions.
    const flatDimIdx =
      !isRowHeader && $config.isFlat
        ? $config.rowDimensionNames.indexOf(columnId)
        : -1;
    const upToDimensionIndex = flatDimIdx >= 0 ? flatDimIdx : undefined;

    const cellFilters = isRowHeader
      ? getFiltersForRowHeader($config, rowId, $data.data)
      : getFiltersForCell(
          $config,
          rowId,
          columnId,
          $data.columnDimensionAxes ?? {},
          $data.data,
          upToDimensionIndex,
        );

    if (!cellFilters.filters) return;

    applyDimensionFilters(
      cellFilters.filters,
      isDeselect,
      (nextRowHeaders, nextCells) => {
        if (isRowHeader) {
          if (nextRowHeaders.has(rowId)) nextRowHeaders.delete(rowId);
          else nextRowHeaders.add(rowId);
        } else {
          const key = cellKey(rowId, columnId);
          if (nextCells.has(key)) {
            nextCells.delete(key);
            rowDimClickIndex.delete(rowId);
          } else {
            nextCells.add(key);
            // Track which dimension index was clicked for visual classification
            if (upToDimensionIndex !== undefined) {
              rowDimClickIndex.set(rowId, upToDimensionIndex);
            } else {
              rowDimClickIndex.delete(rowId);
            }
          }
        }
      },
    );
  }

  function handleColumnHeaderClick(dimensionPath: Record<string, string>) {
    const $config = get(pivotConfig);
    if (!$config) return;

    const $clickSelection = get(clickSelectionStore);
    const isDeselect = $clickSelection.isColumnHeaderSelected(dimensionPath);

    const colFilters = getFiltersForColumnHeader($config, dimensionPath);
    if (!colFilters.filters) return;

    applyDimensionFilters(
      colFilters.filters,
      isDeselect,
      (_nextRowHeaders, _nextCells, nextColHeaders) => {
        const key = columnHeaderKey(dimensionPath);
        if (nextColHeaders.has(key)) nextColHeaders.delete(key);
        else nextColHeaders.add(key);
      },
    );
  }

  // --- Cleanup ---

  function destroy() {
    pruneUnsub();
    clearUnsub();
    activeUnsub();
  }

  return {
    clickSelection: { subscribe: clickSelectionStore.subscribe },
    rowSelectionState,
    handleCellClickToFilter,
    handleColumnHeaderClick,
    destroy,
  };
}

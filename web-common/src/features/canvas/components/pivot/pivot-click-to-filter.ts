/**
 * Factory that creates all click-to-filter orchestration logic for a
 * canvas pivot component.
 *
 * Selections are keyed by dimension values (dimKey) rather than
 * positional TanStack row indices, so they survive sorting and
 * data refreshes.
 *
 * Returns readable stores for selection/row-highlight state, click
 * handlers for cells and column headers, and a destroy function for
 * cleanup.
 */
import {
  type PivotClickSelectionState,
  type SelectionEntry,
  buildClickSelection,
  cellKey,
  columnHeaderKey,
  createEmptyClickSelectionState,
  dimKeyFromRow,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-click-selection";
import {
  type PivotRowSelectionState,
  computePivotRowSelection,
  extractDimensionFiltersFromExpression,
  extractSelectionDimensionFilters,
  getActiveDimensionNames,
  getFiltersForColumnHeader,
  getFiltersForRowData,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
import { getFiltersFromRow } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataRow,
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
    rowData: PivotDataRow,
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

  // Svelte store subscriptions fire synchronously on setup. This flag prevents
  // onBecomeInactive from firing during the initial subscription cascade before
  // the factory has been fully wired into the component tree.
  let initialized = false;

  // --- Prune selfFilteredDimensions when filters are cleared externally ---
  const pruneUnsub = whereFilterStore.subscribe(($whereFilter) => {
    const activeDims = getActiveDimensionNames($whereFilter);
    const dims = get(selfFilteredDimensions);
    const pruned = new Set<string>();
    let changed = false;
    for (const dim of dims) {
      if (activeDims.has(dim)) {
        pruned.add(dim);
      } else {
        changed = true;
      }
    }
    if (changed) {
      selfFilteredDimensions.set(pruned);
    }
  });

  // --- Clear click selection when no self-filtered dimensions remain ---
  const clearUnsub = selfFilteredDimensions.subscribe(($selfFiltered) => {
    if ($selfFiltered.size === 0) {
      clickSelectionStore.set(createEmptyClickSelectionState());
      if (initialized) {
        onBecomeInactive?.();
      }
    }
  });

  // --- Yield active state when another component becomes active ---
  const activeUnsub = activeComponent.subscribe(($activeId) => {
    if ($activeId !== componentId) {
      selfFilteredDimensions.set(new Set());
    }
  });

  initialized = true;

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

  // --- Helpers ---

  /** Capture dimension name→value pairs from a PivotDataRow. */
  function captureDimValues(
    rowData: PivotDataRow,
    rowDimensionNames: string[],
  ): Record<string, string> {
    const result: Record<string, string> = {};
    for (const dim of rowDimensionNames) {
      const val = rowData[dim];
      if (typeof val === "string" || typeof val === "number") {
        result[dim] = String(val);
      }
    }
    return result;
  }

  // --- Retained value computation for safe deselect ---

  function collectRetainedDimensionValues(
    remainingRowHeaders: Map<string, SelectionEntry>,
    remainingCells: Map<string, SelectionEntry>,
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

    // Values from remaining row header selections (stored dimValues)
    for (const entry of remainingRowHeaders.values()) {
      for (const [dim, val] of Object.entries(entry.dimValues)) {
        addRetainedValue(dim, val);
      }
    }

    // Values from remaining data cell selections (stored dimValues)
    for (const entry of remainingCells.values()) {
      for (const [dim, val] of Object.entries(entry.dimValues)) {
        addRetainedValue(dim, val);
      }
    }

    // Values from remaining column header selections
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

  function applyDimensionFilters(
    filters: V1Expression,
    isDeselect: boolean,
    updateSelectionSets: (
      rowHeaders: Map<string, SelectionEntry>,
      cells: Map<string, SelectionEntry>,
      colHeaders: Set<string>,
    ) => void,
  ) {
    const dimensionFilters = extractDimensionFiltersFromExpression(filters);
    if (dimensionFilters.length === 0) return;

    // Capture which dimensions were already in the global filter before this click
    const preExistingDims = getActiveDimensionNames(get(whereFilterStore));

    const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
    if (!filterClass) return;

    // Update selection sets first so we can compute retained values
    const $clickSelection = get(clickSelectionStore);
    const updatedRowHeaders = new Map($clickSelection.rowHeaderSelections);
    const updatedCells = new Map($clickSelection.cellSelections);
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
      buildClickSelection(updatedRowHeaders, updatedCells, updatedColHeaders),
    );

    // Mark only newly-added dimensions as self-filtered
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

  /**
   * Atomically replaces one cell selection with another in a single URL update.
   * Used by flat tables where only one cell per row is allowed.
   */
  function applyReplacementFilters(
    oldFilters: V1Expression,
    newFilters: V1Expression,
    updateSelectionSets: (
      rowHeaders: Map<string, SelectionEntry>,
      cells: Map<string, SelectionEntry>,
      colHeaders: Set<string>,
    ) => void,
  ) {
    const oldDimFilters = extractDimensionFiltersFromExpression(oldFilters);
    const newDimFilters = extractDimensionFiltersFromExpression(newFilters);
    if (oldDimFilters.length === 0 && newDimFilters.length === 0) return;

    const allDimFilters = [...oldDimFilters, ...newDimFilters];
    const preExistingDims = getActiveDimensionNames(get(whereFilterStore));

    const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
    if (!filterClass) return;

    // Clone and update selection sets (remove old key, add new key)
    const $clickSelection = get(clickSelectionStore);
    const updatedRowHeaders = new Map($clickSelection.rowHeaderSelections);
    const updatedCells = new Map($clickSelection.cellSelections);
    const updatedColHeaders = new Set($clickSelection.columnHeaderSelections);
    updateSelectionSets(updatedRowHeaders, updatedCells, updatedColHeaders);

    // Clear temporary filter status for all affected dimensions
    allDimFilters.forEach(({ dimensionName }) => {
      filterManager.checkTemporaryFilter(dimensionName, [metricsViewName]);
    });

    let filterString: string | null = null;

    // Phase 1: Remove orphaned values from old cell
    const retainedValues = collectRetainedDimensionValues(
      updatedRowHeaders,
      updatedCells,
      updatedColHeaders,
    );

    oldDimFilters.forEach(({ dimensionName, values }) => {
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

    // Phase 2: Add new cell's values
    newDimFilters.forEach(({ dimensionName, values }) => {
      filterString = filterClass.addDimensionValueSelections(
        dimensionName,
        values,
      );
    });

    clickSelectionStore.set(
      buildClickSelection(updatedRowHeaders, updatedCells, updatedColHeaders),
    );

    // Mark newly-added dimensions as self-filtered
    const wasInactive = get(selfFilteredDimensions).size === 0;
    selfFilteredDimensions.update((dims) => {
      const next = new Set(dims);
      allDimFilters.forEach(({ dimensionName }) => {
        if (!preExistingDims.has(dimensionName)) {
          next.add(dimensionName);
        }
      });
      return next;
    });
    if (wasInactive && get(selfFilteredDimensions).size > 0) {
      onBecomeActive?.();
    }

    if (filterString !== null) {
      filterManager.applyFiltersToUrl(
        new Map([[metricsViewName, filterString]]),
      );
    }
  }

  // --- Click handlers ---

  function handleCellClickToFilter(
    _rowId: string,
    columnId: string,
    isRowHeader: boolean,
    rowData: PivotDataRow,
  ) {
    const $config = get(pivotConfig);
    const $data = get(pivotDataStore);
    if (!$config || !$data?.data) return;

    const dk = dimKeyFromRow(rowData, $config.rowDimensionNames);
    const dimValues = captureDimValues(rowData, $config.rowDimensionNames);
    const $clickSelection = get(clickSelectionStore);

    // Determine if this click is deselecting a previously selected element
    const isDeselect = isRowHeader
      ? $clickSelection.isRowHeaderSelected(dk)
      : $clickSelection.isCellSelected(dk, columnId);

    // For flat-table dimension cell clicks, only filter up to (and including) the
    // clicked column's dimension index, not all row dimensions.
    const flatDimIdx =
      !isRowHeader && $config.isFlat
        ? $config.rowDimensionNames.indexOf(columnId)
        : -1;
    const upToDimensionIndex = flatDimIdx >= 0 ? flatDimIdx : undefined;

    const cellFilters = isRowHeader
      ? getFiltersForRowData($config, rowData)
      : getFiltersFromRow(
          $config,
          rowData,
          columnId,
          $data.columnDimensionAxes ?? {},
          upToDimensionIndex,
        );

    if (!cellFilters.filters) return;

    // Build the dimValues to store; for partial dimension clicks, only keep
    // dimensions up to the clicked index
    const storedDimValues =
      upToDimensionIndex !== undefined
        ? Object.fromEntries(
            Object.entries(dimValues).filter(([dim]) => {
              const idx = $config.rowDimensionNames.indexOf(dim);
              return idx >= 0 && idx <= upToDimensionIndex;
            }),
          )
        : dimValues;

    // Flat table: replace existing cell selection in the same row instead of
    // accumulating. Nested tables allow multi-select within a row.
    if ($config.isFlat && !isRowHeader && !isDeselect) {
      // Find existing cell keys for this dimKey
      const existingEntries: [string, SelectionEntry][] = [];
      for (const [key, entry] of $clickSelection.cellSelections) {
        if (entry.dimKey === dk) {
          existingEntries.push([key, entry]);
        }
      }

      if (existingEntries.length > 0) {
        // Compute old cell's filters so we can remove its orphaned values
        const [, oldEntry] = existingEntries[0];
        const oldDimIdx = $config.rowDimensionNames.indexOf(oldEntry.columnId);
        const oldUpToDimIdx = oldDimIdx >= 0 ? oldDimIdx : undefined;
        const oldCellFilters = getFiltersFromRow(
          $config,
          rowData,
          oldEntry.columnId,
          $data.columnDimensionAxes ?? {},
          oldUpToDimIdx,
        );

        if (oldCellFilters.filters) {
          applyReplacementFilters(
            oldCellFilters.filters,
            cellFilters.filters,
            (_nextRowHeaders, nextCells) => {
              // Remove all existing cell keys for this row
              for (const [key] of existingEntries) {
                nextCells.delete(key);
              }
              // Add the new cell
              const newKey = cellKey(dk, columnId);
              nextCells.set(newKey, {
                dimKey: dk,
                dimValues: storedDimValues,
                columnId,
                dimClickIndex: upToDimensionIndex,
              });
            },
          );
          return;
        }
      }
    }

    applyDimensionFilters(
      cellFilters.filters,
      isDeselect,
      (nextRowHeaders, nextCells) => {
        if (isRowHeader) {
          if (nextRowHeaders.has(dk)) {
            nextRowHeaders.delete(dk);
          } else {
            nextRowHeaders.set(dk, {
              dimKey: dk,
              dimValues,
              columnId,
            });
          }
        } else {
          const key = cellKey(dk, columnId);
          if (nextCells.has(key)) {
            nextCells.delete(key);
          } else {
            nextCells.set(key, {
              dimKey: dk,
              dimValues: storedDimValues,
              columnId,
              dimClickIndex: upToDimensionIndex,
            });
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

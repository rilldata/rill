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
  dimKeyFromDimValues,
  dimKeyFromRow,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-click-selection";
import {
  type ExtractedFilter,
  type PivotRowSelectionState,
  computePivotRowSelection,
  extractDimensionFiltersFromExpression,
  extractSelectionDimensionFilters,
  getActiveDimensionNames,
  getDimensionValuesForRow,
  getFiltersForColumnHeader,
  getFiltersForRowData,
  getFiltersForRowHeader,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-row-selection";
import {
  getFiltersForCell,
  getFiltersFromRow,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type {
  PivotDataRow,
  PivotDataState,
  PivotDataStore,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
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
  ): Record<string, string | null> {
    const result: Record<string, string | null> = {};
    for (const dim of rowDimensionNames) {
      const val = rowData[dim];
      if (val === null || val === undefined) {
        result[dim] = null;
      } else if (typeof val === "string" || typeof val === "number") {
        result[dim] = String(val);
      }
    }
    return result;
  }

  /** Determine the nesting level of a column header from its serialized key. */
  function getColumnHeaderLevel(headerKey: string): number {
    const entries = JSON.parse(headerKey) as [string, string][];
    return entries.length;
  }

  /**
   * Returns the level of currently selected column headers, or -1 if none.
   * All selected headers are at the same level due to level-switch enforcement.
   */
  function getCurrentColumnHeaderLevel(colHeaders: Set<string>): number {
    const first = colHeaders.values().next().value as string | undefined;
    return first !== undefined ? getColumnHeaderLevel(first) : -1;
  }

  // --- Retained value computation for safe deselect ---

  function collectRetainedDimensionValues(
    remainingRowHeaders: Map<string, SelectionEntry>,
    remainingCells: Map<string, SelectionEntry>,
    remainingColHeaders: Set<string>,
  ): Map<string, Set<string | null>> {
    const retainedValues = new Map<string, Set<string | null>>();

    const addRetainedValue = (dimensionName: string, value: string | null) => {
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

  /**
   * Shared skeleton for all filter updates: clones selection state, applies
   * removals and additions to the FilterManager, updates stores, and syncs URL.
   */
  function applyFilterUpdate(opts: {
    removals: ExtractedFilter[];
    additions: ExtractedFilter[];
    updateSelectionSets: (
      rowHeaders: Map<string, SelectionEntry>,
      cells: Map<string, SelectionEntry>,
      colHeaders: Set<string>,
    ) => void;
  }) {
    const { removals, additions, updateSelectionSets } = opts;
    const allDimFilters = [...removals, ...additions];
    if (allDimFilters.length === 0) return;

    const preExistingDims = getActiveDimensionNames(get(whereFilterStore));
    const filterClass = filterManager.metricsViewFilters.get(metricsViewName);
    if (!filterClass) return;

    // Clone and update selection sets
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

    // Remove orphaned values
    if (removals.length > 0) {
      const retainedValues = collectRetainedDimensionValues(
        updatedRowHeaders,
        updatedCells,
        updatedColHeaders,
      );

      for (const { dimensionName, values } of removals) {
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
      }

      // If no orphans were found (all values still retained by other
      // selections), get the current filter string for URL sync.
      if (filterString === null && additions.length === 0) {
        for (const { dimensionName } of removals) {
          filterString = filterClass.addDimensionValueSelections(
            dimensionName,
            [],
          );
        }
      }
    }

    // Add new values
    for (const { dimensionName, values } of additions) {
      filterString = filterClass.addDimensionValueSelections(
        dimensionName,
        values,
      );
    }

    clickSelectionStore.set(
      buildClickSelection(updatedRowHeaders, updatedCells, updatedColHeaders),
    );

    // Mark only newly-added dimensions as self-filtered
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

    applyFilterUpdate({
      removals: isDeselect ? dimensionFilters : [],
      additions: isDeselect ? [] : dimensionFilters,
      updateSelectionSets,
    });
  }

  /**
   * Atomically replaces old selections with new ones in a single URL update.
   * Phase 1: removes orphaned filter values from old selections.
   * Phase 2: adds new selection's filter values.
   * Used by flat table single-cell-per-row and column header level switching.
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

    applyFilterUpdate({
      removals: oldDimFilters,
      additions: newDimFilters,
      updateSelectionSets,
    });
  }

  /**
   * Builds a combined old-filters expression from all existing column header
   * selections, then delegates to applyReplacementFilters.
   */
  function applyColumnHeaderLevelSwitch(
    oldColumnHeaders: Set<string>,
    newDimensionPath: Record<string, string>,
    newFilters: V1Expression,
  ) {
    // Aggregate values per dimension across all old column header selections.
    // Deduping matters: without it, a dimension value shared across multiple
    // old headers (e.g. Y=Y1 in both {X:X1,Y:Y1} and {X:X2,Y:Y1}) emits
    // duplicate IN exprs, which later cause toggleDimensionValueSelections to
    // re-toggle (remove then re-add) the shared value.
    const valuesByDim = new Map<string, Set<string>>();
    for (const oldKey of oldColumnHeaders) {
      const entries = JSON.parse(oldKey) as [string, string][];
      for (const [dim, value] of entries) {
        let set = valuesByDim.get(dim);
        if (!set) {
          set = new Set();
          valuesByDim.set(dim, set);
        }
        set.add(value);
      }
    }
    const oldExprs: V1Expression[] = [...valuesByDim.entries()].map(
      ([dim, values]) => createInExpression(dim, [...values]),
    );

    const oldFilters = createAndExpression(oldExprs);
    const newKey = columnHeaderKey(newDimensionPath);

    applyReplacementFilters(
      oldFilters,
      newFilters,
      (_nextRowHeaders, _nextCells, nextColHeaders) => {
        nextColHeaders.clear();
        nextColHeaders.add(newKey);
      },
    );
  }

  // --- Lineage-aware mutual exclusivity helpers ---

  /**
   * Check if a cell is "under" a header: the header's dimension values are
   * a subset of the cell's dimValues.
   */
  function isDimSubset(
    subset: Record<string, string | null>,
    superset: Record<string, string | null>,
  ): boolean {
    return Object.entries(subset).every(
      ([dim, val]) => dim in superset && superset[dim] === val,
    );
  }

  /**
   * Find cell selection keys whose dimValues contain all of the header's
   * dimension values (i.e., cells "under" the header).
   */
  function findCellsUnderRowHeader(
    selection: PivotClickSelectionState,
    headerDimValues: Record<string, string | null>,
  ): string[] {
    const keys: string[] = [];
    for (const [key, entry] of selection.cellSelections) {
      if (isDimSubset(headerDimValues, entry.dimValues)) {
        keys.push(key);
      }
    }
    return keys;
  }

  /**
   * Find row header dimKeys whose dimValues are a subset of the cell's dimValues
   * (i.e., headers that are "above" the cell).
   */
  function findRowHeadersAboveCell(
    selection: PivotClickSelectionState,
    cellDimValues: Record<string, string | null>,
  ): string[] {
    const keys: string[] = [];
    for (const [dk, entry] of selection.rowHeaderSelections) {
      if (isDimSubset(entry.dimValues, cellDimValues)) {
        keys.push(dk);
      }
    }
    return keys;
  }

  /**
   * Find row header dimKeys that are descendants of the clicked header (its
   * dimValues are a strict superset of the clicked header's dimValues).
   * Strictness excludes the clicked header itself from the eviction set.
   */
  function findRowHeadersUnderRowHeader(
    selection: PivotClickSelectionState,
    headerDimValues: Record<string, string | null>,
    clickedDk: string,
  ): string[] {
    const keys: string[] = [];
    for (const [dk, entry] of selection.rowHeaderSelections) {
      if (dk === clickedDk) continue;
      if (isDimSubset(headerDimValues, entry.dimValues)) {
        keys.push(dk);
      }
    }
    return keys;
  }

  /**
   * Find row header dimKeys that are ancestors of the clicked header (its
   * dimValues are a strict subset of the clicked header's dimValues).
   */
  function findRowHeadersAboveRowHeader(
    selection: PivotClickSelectionState,
    headerDimValues: Record<string, string | null>,
    clickedDk: string,
  ): string[] {
    const keys: string[] = [];
    for (const [dk, entry] of selection.rowHeaderSelections) {
      if (dk === clickedDk) continue;
      if (isDimSubset(entry.dimValues, headerDimValues)) {
        keys.push(dk);
      }
    }
    return keys;
  }

  /**
   * Find column header keys whose dimension path values are all present
   * in the cell's dimValues (i.e., column headers "above" the cell).
   */
  function findColHeadersAboveCell(
    selection: PivotClickSelectionState,
    cellDimValues: Record<string, string | null>,
  ): string[] {
    const keys: string[] = [];
    for (const colKey of selection.columnHeaderSelections) {
      const entries = JSON.parse(colKey) as [string, string][];
      const allMatch = entries.every(
        ([name, value]) =>
          name in cellDimValues && cellDimValues[name] === value,
      );
      if (allMatch) {
        keys.push(colKey);
      }
    }
    return keys;
  }

  /**
   * Find cell selection keys whose dimValues contain all of the column header's
   * dimension path values (i.e., cells "under" the column header).
   */
  function findCellsUnderColHeader(
    selection: PivotClickSelectionState,
    dimensionPath: Record<string, string>,
  ): string[] {
    const keys: string[] = [];
    for (const [key, entry] of selection.cellSelections) {
      const allMatch = Object.entries(dimensionPath).every(
        ([name, value]) =>
          name in entry.dimValues && entry.dimValues[name] === value,
      );
      if (allMatch) {
        keys.push(key);
      }
    }
    return keys;
  }

  /**
   * Collect dimension values from a set of selection entries as removals.
   */
  function collectRemovalsFromEntries(
    entries: SelectionEntry[],
  ): Map<string, Set<string | null>> {
    const byDim = new Map<string, Set<string | null>>();
    for (const entry of entries) {
      for (const [name, value] of Object.entries(entry.dimValues)) {
        let set = byDim.get(name);
        if (!set) {
          set = new Set();
          byDim.set(name, set);
        }
        set.add(value);
      }
    }
    return byDim;
  }

  // --- Click handlers ---

  /**
   * Flat table single-cell-per-row: if the same row already has a selected
   * dimension cell, atomically replace it with the new click. Returns true
   * if the replacement was handled, false to fall through to normal toggle.
   */
  function tryFlatTableCellReplacement(
    config: PivotDataStoreConfig,
    selection: PivotClickSelectionState,
    data: PivotDataState,
    dk: string,
    columnId: string,
    rowData: PivotDataRow,
    newFilters: V1Expression,
    storedDimValues: Record<string, string | null>,
    upToDimensionIndex: number | undefined,
  ): boolean {
    // Find existing cell keys for this dimKey (same row)
    const existingEntries: [string, SelectionEntry][] = [];
    for (const [key, entry] of selection.cellSelections) {
      if (entry.dimKey === dk) {
        existingEntries.push([key, entry]);
      }
    }
    if (existingEntries.length === 0) return false;

    // Compute old cell's filters so we can remove its orphaned values
    const [, oldEntry] = existingEntries[0];
    const oldDimIdx = config.rowDimensionNames.indexOf(oldEntry.columnId);
    const oldUpToDimIdx = oldDimIdx >= 0 ? oldDimIdx : undefined;
    const oldCellFilters = getFiltersFromRow(
      config,
      rowData,
      oldEntry.columnId,
      data.columnDimensionAxes ?? {},
      oldUpToDimIdx,
    );

    if (!oldCellFilters.filters) return false;

    applyReplacementFilters(
      oldCellFilters.filters,
      newFilters,
      (_nextRowHeaders, nextCells) => {
        for (const [key] of existingEntries) {
          nextCells.delete(key);
        }
        const newKey = cellKey(dk, columnId);
        nextCells.set(newKey, {
          dimKey: dk,
          dimValues: storedDimValues,
          columnId,
          dimClickIndex: upToDimensionIndex,
        });
      },
    );
    return true;
  }

  function handleCellClickToFilter(
    rowId: string,
    columnId: string,
    isRowHeader: boolean,
    rowData: PivotDataRow,
  ) {
    const $config = get(pivotConfig);
    const $data = get(pivotDataStore);
    if (!$config || !$data?.data) return;

    const $clickSelection = get(clickSelectionStore);

    // In nested mode, row data stores all values under rowDimensions[0],
    // so we must use positional rowId navigation to get correct dim→value pairs.
    const isNested = !$config.isFlat;
    const dimValues = isNested
      ? Object.fromEntries(
          getDimensionValuesForRow($config, rowId, $data.data).map(
            ({ dimensionName, value }) => [dimensionName, value],
          ),
        )
      : captureDimValues(rowData, $config.rowDimensionNames);

    // For nested child rows (depth > 0), build dimKey from the fully-resolved
    // dimValues (which include parent dimension values); dimKeyFromRow only
    // sees rowDimensions[0] and would produce identical keys across parents.
    const isNestedChild = isNested && rowId.includes(".");
    const dk = isNestedChild
      ? dimKeyFromDimValues(dimValues, $config.rowDimensionNames)
      : dimKeyFromRow(rowData, $config.rowDimensionNames);

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
      ? isNested
        ? getFiltersForRowHeader($config, rowId, $data.data)
        : getFiltersForRowData($config, rowData)
      : isNested
        ? getFiltersForCell(
            $config,
            rowId,
            columnId,
            $data.columnDimensionAxes ?? {},
            $data.data,
          )
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
    const storedDimValues: Record<string, string | null> =
      upToDimensionIndex !== undefined
        ? Object.fromEntries(
            Object.entries(dimValues).filter(([dim]) => {
              const idx = $config.rowDimensionNames.indexOf(dim);
              return idx >= 0 && idx <= upToDimensionIndex;
            }),
          )
        : { ...dimValues };

    // For data cell clicks (not row headers), also capture column dimension
    // values so collectRetainedDimensionValues can detect shared column filters
    // during deselect. Without this, deselecting one cell would orphan column
    // filter values still needed by other cells in the same column.
    if (!isRowHeader) {
      const colDimNames = new Set($config.colDimensionNames);
      const allDimFilters = extractDimensionFiltersFromExpression(
        cellFilters.filters,
      );
      for (const { dimensionName, values } of allDimFilters) {
        if (
          colDimNames.has(dimensionName) &&
          !(dimensionName in storedDimValues)
        ) {
          storedDimValues[dimensionName] = values[0] ?? null;
        }
      }
    }

    // Flat table: replace existing cell selection in the same row instead of
    // accumulating. Nested tables allow multi-select within a row.
    if ($config.isFlat && !isRowHeader && !isDeselect) {
      const handled = tryFlatTableCellReplacement(
        $config,
        $clickSelection,
        $data,
        dk,
        columnId,
        rowData,
        cellFilters.filters,
        storedDimValues,
        upToDimensionIndex,
      );
      if (handled) return;
    }

    if (isDeselect) {
      // Deselect: simple toggle off
      applyDimensionFilters(
        cellFilters.filters,
        true,
        (nextRowHeaders, nextCells) => {
          if (isRowHeader) {
            nextRowHeaders.delete(dk);
          } else {
            nextCells.delete(cellKey(dk, columnId));
          }
        },
      );
      return;
    }

    // Mutual exclusivity: headers and cells under the same lineage cannot
    // coexist. When adding a new selection, evict the opposite domain.
    if (isRowHeader) {
      // Clicking a row header: evict cells under this header, plus any row
      // headers in the same lineage (both descendants and ancestors).
      const cellKeysToEvict = findCellsUnderRowHeader(
        $clickSelection,
        dimValues,
      );
      const descendantHeaderKeysToEvict = findRowHeadersUnderRowHeader(
        $clickSelection,
        dimValues,
        dk,
      );
      const ancestorHeaderKeysToEvict = findRowHeadersAboveRowHeader(
        $clickSelection,
        dimValues,
        dk,
      );
      const rowHeaderKeysToEvict = [
        ...descendantHeaderKeysToEvict,
        ...ancestorHeaderKeysToEvict,
      ];

      const evictedCellEntries = cellKeysToEvict
        .map((k) => $clickSelection.cellSelections.get(k))
        .filter(Boolean) as SelectionEntry[];
      const evictedHeaderEntries = rowHeaderKeysToEvict
        .map((k) => $clickSelection.rowHeaderSelections.get(k))
        .filter(Boolean) as SelectionEntry[];
      const evictedDimValues = collectRemovalsFromEntries([
        ...evictedCellEntries,
        ...evictedHeaderEntries,
      ]);

      const removals: ExtractedFilter[] = [...evictedDimValues.entries()].map(
        ([dimensionName, values]) => ({
          dimensionName,
          values: [...values],
        }),
      );

      const additions = extractDimensionFiltersFromExpression(
        cellFilters.filters,
      );

      applyFilterUpdate({
        removals,
        additions,
        updateSelectionSets: (nextRowHeaders, nextCells) => {
          for (const k of cellKeysToEvict) nextCells.delete(k);
          for (const k of rowHeaderKeysToEvict) nextRowHeaders.delete(k);
          nextRowHeaders.set(dk, { dimKey: dk, dimValues, columnId });
        },
      });
    } else {
      // Clicking a measure cell: evict ancestor row headers and column headers
      const rowHeaderKeysToEvict = findRowHeadersAboveCell(
        $clickSelection,
        storedDimValues,
      );
      const colHeaderKeysToEvict = findColHeadersAboveCell(
        $clickSelection,
        storedDimValues,
      );

      const evictedRowEntries = rowHeaderKeysToEvict
        .map((k) => $clickSelection.rowHeaderSelections.get(k))
        .filter(Boolean) as SelectionEntry[];

      const evictedDimValues = collectRemovalsFromEntries(evictedRowEntries);

      // Also collect column header dimension values for removal
      for (const colKey of colHeaderKeysToEvict) {
        const entries = JSON.parse(colKey) as [string, string][];
        for (const [name, value] of entries) {
          let set = evictedDimValues.get(name);
          if (!set) {
            set = new Set();
            evictedDimValues.set(name, set);
          }
          set.add(value);
        }
      }

      const removals: ExtractedFilter[] = [...evictedDimValues.entries()].map(
        ([dimensionName, values]) => ({
          dimensionName,
          values: [...values],
        }),
      );

      const additions = extractDimensionFiltersFromExpression(
        cellFilters.filters,
      );

      applyFilterUpdate({
        removals,
        additions,
        updateSelectionSets: (nextRowHeaders, nextCells, nextColHeaders) => {
          for (const k of rowHeaderKeysToEvict) nextRowHeaders.delete(k);
          for (const k of colHeaderKeysToEvict) nextColHeaders.delete(k);
          const key = cellKey(dk, columnId);
          nextCells.set(key, {
            dimKey: dk,
            dimValues: storedDimValues,
            columnId,
            dimClickIndex: upToDimensionIndex,
          });
        },
      });
    }
  }

  function handleColumnHeaderClick(dimensionPath: Record<string, string>) {
    const $config = get(pivotConfig);
    if (!$config) return;

    const $clickSelection = get(clickSelectionStore);
    const isDeselect = $clickSelection.isColumnHeaderSelected(dimensionPath);

    const colFilters = getFiltersForColumnHeader($config, dimensionPath);
    if (!colFilters.filters) return;

    // Enforce single-level constraint: if clicking a header at a different
    // level than existing selections, replace all old selections atomically.
    const clickLevel = Object.keys(dimensionPath).length;
    const currentLevel = getCurrentColumnHeaderLevel(
      $clickSelection.columnHeaderSelections,
    );

    if (!isDeselect && currentLevel !== -1 && clickLevel !== currentLevel) {
      applyColumnHeaderLevelSwitch(
        $clickSelection.columnHeaderSelections,
        dimensionPath,
        colFilters.filters,
      );
      return;
    }

    if (isDeselect) {
      applyDimensionFilters(
        colFilters.filters,
        true,
        (_nextRowHeaders, _nextCells, nextColHeaders) => {
          nextColHeaders.delete(columnHeaderKey(dimensionPath));
        },
      );
      return;
    }

    // Mutual exclusivity: evict cells under this column header
    const cellKeysToEvict = findCellsUnderColHeader(
      $clickSelection,
      dimensionPath,
    );
    const evictedEntries = cellKeysToEvict
      .map((k) => $clickSelection.cellSelections.get(k))
      .filter(Boolean) as SelectionEntry[];
    const evictedDimValues = collectRemovalsFromEntries(evictedEntries);

    const removals: ExtractedFilter[] = [...evictedDimValues.entries()].map(
      ([dimensionName, values]) => ({
        dimensionName,
        values: [...values],
      }),
    );

    const additions = extractDimensionFiltersFromExpression(colFilters.filters);

    applyFilterUpdate({
      removals,
      additions,
      updateSelectionSets: (_nextRowHeaders, nextCells, nextColHeaders) => {
        for (const k of cellKeysToEvict) nextCells.delete(k);
        nextColHeaders.add(columnHeaderKey(dimensionPath));
      },
    });
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

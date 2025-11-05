import {
  type PivotChipData,
  PivotChipType,
  type PivotState,
  type PivotTableMode,
} from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import type { ExpandedState, SortingState } from "@tanstack/svelte-table";
import { derived, type Readable, writable, type Writable } from "svelte/store";

export class PivotStore {
  public state: Writable<PivotState>;

  // ##########
  // Selectors
  // ##########

  public columnMeasures: Readable<PivotChipData[]>;
  public columnDimensions: Readable<PivotChipData[]>;

  public constructor(initState: PivotState) {
    this.state = writable(initState);

    this.columnMeasures = derived(this.state, ($state) =>
      $state.columns.filter((c) => c.type === PivotChipType.Measure),
    );

    this.columnDimensions = derived(this.state, ($state) =>
      $state.columns.filter((c) => c.type === PivotChipType.Dimension),
    );
  }

  // ##########
  // Actions
  // ##########

  public setRows = (rows: PivotChipData[]) => {
    this.state.update((pivot) => {
      pivot.rowPage = 1;
      pivot.activeCell = null;

      const dimensions: PivotChipData[] = [];

      rows.forEach((val) => {
        if (val.type !== PivotChipType.Measure) {
          dimensions.push(val);
        }
      });

      if (pivot.sorting.length) {
        const accessor = pivot.sorting[0].id;
        const anchorDimension = dimensions?.[0]?.id;
        if (accessor !== anchorDimension) {
          pivot.sorting = [];
        }
      }

      pivot.rows = dimensions;

      return pivot;
    });
  };

  public setColumns = (columns: PivotChipData[]) => {
    this.state.update((pivot) => {
      pivot.rowPage = 1;
      pivot.activeCell = null;
      pivot.expanded = {};

      if (pivot.sorting.length) {
        const accessor = pivot.sorting[0].id;

        if (pivot.tableMode === "flat") {
          const validAccessors = columns.map((d) => d.id);
          if (!validAccessors.includes(accessor)) {
            pivot.sorting = [];
          }
        } else {
          const anchorDimension = pivot.rows?.[0]?.id;
          if (accessor !== anchorDimension) {
            pivot.sorting = [];
          }
        }
      }
      pivot.columns = columns;

      return pivot;
    });
  };

  public addField = (field: PivotChipData, rows: boolean) => {
    this.state.update((pivot) => {
      pivot.rowPage = 1;
      pivot.activeCell = null;
      if (field.type === PivotChipType.Measure) {
        pivot.columns.push(field);
      } else {
        if (rows) {
          pivot.rows.push(field);
        } else {
          pivot.columns.push(field);
        }
      }

      return pivot;
    });
  };

  public setTableMode = (
    tableMode: PivotTableMode,
    rows: PivotChipData[],
    columns: PivotChipData[],
  ) => {
    this.state.update((pivot) => {
      return {
        ...pivot,
        tableMode,
        rows,
        columns,
        sorting: [],
        expanded: {},
        activeCell: null,
      };
    });
  };

  public setExpanded = (expanded: ExpandedState) => {
    this.state.update((pivot) => {
      pivot.expanded = expanded;
      return pivot;
    });
  };

  public setComparison = (enableComparison: boolean) => {
    this.state.update((pivot) => {
      pivot.enableComparison = enableComparison;
      return pivot;
    });
  };

  public setSort = (sorting: SortingState) => {
    this.state.update((pivot) => {
      return {
        ...pivot,
        sorting,
        rowPage: 1,
        expanded: {},
        activeCell: null,
      };
    });
  };

  public setColumnPage = (pageNumber: number) => {
    this.state.update((pivot) => {
      pivot.columnPage = pageNumber;
      return pivot;
    });
  };

  public setRowPage = (pageNumber: number) => {
    this.state.update((pivot) => {
      pivot.rowPage = pageNumber;
      return pivot;
    });
  };

  public setActiveCell = (rowId: string, columnId: string) => {
    this.state.update((pivot) => {
      pivot.activeCell = { rowId, columnId };
      return pivot;
    });
  };

  public removeActiveCell = () => {
    this.state.update((pivot) => {
      pivot.activeCell = null;
      return pivot;
    });
  };
}

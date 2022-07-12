import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import TableCellInput from "$lib/components/table-editable/TableCellInput.svelte";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";
import SelectableTableCell from "../table-editable/SelectableTableCell.svelte";

export const initMeasuresColumns = (inputChangeHandler) =>
  [
    {
      name: "label",
      tooltip: "a human readable name for this measure",
      renderer: TableCellInput,
    },
    {
      name: "expression",
      tooltip: "a valid SQL aggregation expression for this measure",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
      validation: (row: MeasureDefinitionEntity) => row.expressionIsValid,
    },
    {
      name: "description",
      tooltip: "a human readable description of this measure",
      onchange: inputChangeHandler,
      renderer: TableCellInput,
    },
    {
      name: "formatPreset",
      tooltip: "a format for this measure",
      renderer: SelectableTableCell,
    },
    // FIXME: will be needed later for API
    // {
    //   name: "sqlName",
    //   label: "identifier",
    //   tooltip: "a unique SQL identifier for this measure",
    //   renderer: TableCellInput,
    //   onchange: inputChangeHandler,
    //   validation: (row: MeasureDefinitionEntity) => row.sqlNameIsValid,
    // },

    // FIXME: willbe needed later for sparkline summary
    // {
    //   name: "id",
    //   label: "preview",
    //   tooltip: "a preview of this measure over the selected time dimension",
    //   renderer: TableCellSparkline,
    // },
  ] as ColumnConfig[];

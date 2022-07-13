import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";

import TableCellInput from "$lib/components/table-editable/TableCellInput.svelte";
import TableCellSelector from "../table-editable/TableCellSelector.svelte";

export const initDimensionColumns = (inputChangeHandler, dimensionOptions) =>
  [
    {
      name: "labelSingle",
      label: "label (single)",
      tooltip: "a human readable name for this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },

    {
      name: "dimensionColumn",
      label: "dimension column",
      tooltip:
        "a categorical column from the data model that this metrics set is based on",
      renderer: TableCellSelector,
      onchange: inputChangeHandler,
      options: dimensionOptions,
      validation: (row: DimensionDefinitionEntity) => row.dimensionIsValid,
    },
    {
      name: "description",
      tooltip: "a human readable description of this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },
    {
      name: "sqlName",
      label: "identifier",
      tooltip: "a unique SQL identifier for this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
      validation: (row: DimensionDefinitionEntity) => row.sqlNameIsValid,
    },
    {
      name: "labelPlural",
      label: "label (plural)",
      tooltip: "an optional pluralized human readable name for this dimension",
      renderer: TableCellInput,
      onchange: inputChangeHandler,
    },

    // {
    //   name: "id",
    //   label: "unique values",
    //   tooltip: "the number of unique values present in this dimension",

    //   renderType: // FIXME: need cardinality count cell here
    // },
  ] as ColumnConfig[];

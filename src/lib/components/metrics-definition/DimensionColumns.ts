import EditableTableCell from "$lib/components/table-editable/EditableTableCell.svelte";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfigumnConfig";
import RowActionsCell from "$lib/components/table-editable/RowActionsCell.svelte";

// export const DimensionColumns: Array<ColumnConfig> = [
//   "sqlName",
//   "dimensionColumn",
//   "description",
// ].map((col) => ({
//   name: col,
//   type: "VARCHAR",
//   renderer: EditableTableCell,
// }));
// DimensionColumns[0].validation = (row: DimensionDefinitionEntity) =>
//   row.sqlNameIsValid;
// DimensionColumns[1].validation = (row: DimensionDefinitionEntity) =>
//   row.dimensionIsValid;
// DimensionColumns.push({
//   name: "id",
//   label: "Actions",
//   type: "VARCHAR",
//   renderer: RowActionsCell,
// });

export const DimensionColumns: ColumnConfig[] = [
  {
    name: "labelSingle",
    label: "label (single)",
    tooltip: "a human readable name for this dimension",
    renderer: EditableTableCell,
  },

  {
    name: "dimensionColumn",
    label: "dimension column",
    tooltip:
      "a categorical column from the data model that this metrics set is based on",
    renderer: EditableTableCell,
  },
  {
    name: "description",
    tooltip: "a human readable description of this dimension",
    renderer: EditableTableCell,
  },
  {
    name: "sqlName",
    label: "identifier",
    tooltip: "a unique SQL identifier for this dimension",
    renderer: EditableTableCell,
    validation: (row: DimensionDefinitionEntity) => row.sqlNameIsValid,
  },
  {
    name: "labelPlural",
    label: "label (plural)",
    tooltip: "an optional pluralized human readable name for this dimension",
    renderer: EditableTableCell,
  },

  {
    name: "id",
    label: "unique values",
    tooltip: "the number of unique values present in this dimension",
    // FIXME: need cardinality count cell here
    renderer: EditableTableCell,
  },

  {
    name: "id",
    label: "actions",
    tooltip: "actions affecting this row",
    renderer: RowActionsCell,
  },
];

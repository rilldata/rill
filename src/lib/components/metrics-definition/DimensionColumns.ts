import EditableTableCell from "$lib/components/table-editable/EditableTableCell.svelte";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfigumnConfig";
import RowActionsCell from "$lib/components/table-editable/RowActionsCell.svelte";

export const DimensionColumns: Array<ColumnConfig> = [
  "sqlName",
  "dimensionColumn",
  "description",
].map((col) => ({
  name: col,
  type: "VARCHAR",
  renderer: EditableTableCell,
}));
DimensionColumns[0].validation = (row: DimensionDefinitionEntity) =>
  row.sqlNameIsValid;
DimensionColumns[1].validation = (row: DimensionDefinitionEntity) =>
  row.dimensionIsValid;
DimensionColumns.push({
  name: "id",
  label: "Actions",
  type: "VARCHAR",
  renderer: RowActionsCell,
});

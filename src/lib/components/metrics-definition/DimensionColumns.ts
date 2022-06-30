import EditableTableCell from "$lib/components/table/EditableTableCell.svelte";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { ColumnConfig } from "$lib/components/table/ColumnConfig";

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

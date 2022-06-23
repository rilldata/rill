import type { ColumnConfig } from "$lib/components/table/pinnableUtils";
import EditableTableCell from "$lib/components/table/EditableTableCell.svelte";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export const MeasuresColumns: Array<ColumnConfig> = [
  "label",
  "sqlName",
  "expression",
  "description",
].map((col) => ({
  name: col,
  type: "VARCHAR",
  renderer: EditableTableCell,
}));
MeasuresColumns[1].validation = (row: MeasureDefinitionEntity) =>
  row.sqlNameIsValid;
MeasuresColumns[2].validation = (row: MeasureDefinitionEntity) =>
  row.expressionIsValid;

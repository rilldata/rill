import EditableTableCell from "$lib/components/table/EditableTableCell.svelte";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import MeasureSparkLineCell from "$lib/components/metrics-definition/MeasureSparkLineCell.svelte";
import type { ColumnConfig } from "$lib/components/table/ColumnConfig";
import RowActionsCell from "$lib/components/table/RowActionsCell.svelte";

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

MeasuresColumns.push({
  name: "id",
  label: "spark line",
  type: "VARCHAR",
  renderer: MeasureSparkLineCell,
});
MeasuresColumns.push({
  name: "id",
  label: "Actions",
  type: "VARCHAR",
  renderer: RowActionsCell,
});

import EditableTableCell from "$lib/components/table-editable/EditableTableCell.svelte";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import MeasureSparkLineCell from "$lib/components/metrics-definition/MeasureSparkLineCell.svelte";
import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";

export const MeasuresColumns: ColumnConfig[] = [
  {
    name: "label",
    tooltip: "a human readable name for this measure",
    renderer: EditableTableCell,
  },
  {
    name: "expression",
    tooltip: "a valid SQL aggregation expression for this measure",
    renderer: EditableTableCell,
    validation: (row: MeasureDefinitionEntity) => row.expressionIsValid,
  },
  {
    name: "sqlName",
    label: "identifier",
    tooltip: "a unique SQL identifier for this measure",
    renderer: EditableTableCell,
    validation: (row: MeasureDefinitionEntity) => row.sqlNameIsValid,
  },
  {
    name: "description",
    tooltip: "a human readable description of this measure",
    renderer: EditableTableCell,
  },
  {
    name: "id",
    label: "preview",
    tooltip: "a preview of this measure over the selected time dimension",
    renderer: MeasureSparkLineCell,
  },
];

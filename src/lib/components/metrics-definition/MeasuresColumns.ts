import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
// import MeasureSparkLineCell from "$lib/components/metrics-definition/MeasureSparkLineCell.svelte";
import {
  ColumnConfig,
  RenderType,
} from "$lib/components/table-editable/ColumnConfig";

export const MeasuresColumns: ColumnConfig[] = [
  {
    name: "label",
    tooltip: "a human readable name for this measure",
    // renderer: EditableTableCell,
    renderType: RenderType.INPUT,
  },
  {
    name: "expression",
    tooltip: "a valid SQL aggregation expression for this measure",
    // renderer: EditableTableCell,
    validation: (row: MeasureDefinitionEntity) => row.expressionIsValid,
    renderType: RenderType.INPUT,
  },
  {
    name: "sqlName",
    label: "identifier",
    tooltip: "a unique SQL identifier for this measure",
    // renderer: EditableTableCell,
    validation: (row: MeasureDefinitionEntity) => row.sqlNameIsValid,
    renderType: RenderType.INPUT,
  },
  {
    name: "description",
    tooltip: "a human readable description of this measure",
    renderType: RenderType.INPUT,
    // renderer: EditableTableCell,
  },
  {
    name: "id",
    label: "preview",
    tooltip: "a preview of this measure over the selected time dimension",
    renderType: RenderType.SPARKLINE,
    // renderer: MeasureSparkLineCell,
  },
];

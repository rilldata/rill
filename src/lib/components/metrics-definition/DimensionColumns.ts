import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import {
  ColumnConfig,
  RenderType,
} from "$lib/components/table-editable/ColumnConfig";

export const DimensionColumns: ColumnConfig[] = [
  {
    name: "labelSingle",
    label: "label (single)",
    tooltip: "a human readable name for this dimension",
    renderType: RenderType.INPUT,
  },

  {
    name: "dimensionColumn",
    label: "dimension column",
    tooltip:
      "a categorical column from the data model that this metrics set is based on",
    renderType: RenderType.INPUT,
    validation: (row: DimensionDefinitionEntity) => row.dimensionIsValid,
  },
  {
    name: "description",
    tooltip: "a human readable description of this dimension",
    renderType: RenderType.INPUT,
  },
  {
    name: "sqlName",
    label: "identifier",
    tooltip: "a unique SQL identifier for this dimension",
    renderType: RenderType.INPUT,
    validation: (row: DimensionDefinitionEntity) => row.sqlNameIsValid,
  },
  {
    name: "labelPlural",
    label: "label (plural)",
    tooltip: "an optional pluralized human readable name for this dimension",
    renderType: RenderType.INPUT,
  },

  {
    name: "id",
    label: "unique values",
    tooltip: "the number of unique values present in this dimension",
    // FIXME: need cardinality count cell here
    renderType: RenderType.CARDINALITY,
  },
];

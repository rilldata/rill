import {
  ColumnConfig,
  CellConfigInput,
  CellConfigSelector,
} from "$lib/components/table-editable/ColumnConfig";

export const initDimensionColumns = (inputChangeHandler, dimensionOptions) =>
  <ColumnConfig<CellConfigInput | CellConfigSelector>[]>[
    {
      name: "labelSingle",
      label: "label (single)",
      headerTooltip: "a human readable name for this dimension (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },

    {
      name: "dimensionColumn",
      label: "dimension column",
      headerTooltip:
        "a categorical column from the data model that this metrics set is based on",
      cellRenderer: new CellConfigSelector(
        inputChangeHandler,
        dimensionOptions,
        "select a column..."
      ),
    },
    {
      name: "description",
      headerTooltip:
        "a human readable description of this dimension (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "labelPlural",
      label: "label (plural)",
      headerTooltip:
        "an pluralized human readable name for this dimension (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    // FIXME will be needed later for API
    // {
    //   name: "sqlName",
    //   label: "identifier",
    //   headerTooltip: "a unique SQL identifier for this dimension",
    //   renderer: TableCellInput,
    //   onchange: inputChangeHandler,
    //   validation: (row: DimensionDefinitionEntity) => row.sqlNameIsValid,
    // },

    // FIXME: willbe needed later for cardinality summary
    // {
    //   name: "id",
    //   label: "unique values",
    //   headerTooltip: "the number of unique values present in this dimension",
    //   renderer: TabelCellCardinality
    // },
  ];

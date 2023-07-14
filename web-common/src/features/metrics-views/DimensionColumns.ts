import {
  CellConfigInput,
  CellConfigSelector,
  ColumnConfig,
} from "../../components/table-editable/ColumnConfig";

export const initDimensionColumns = (inputChangeHandler, dimensionOptions) =>
  <ColumnConfig<CellConfigInput | CellConfigSelector>[]>[
    {
      name: "label",
      // FIXME: should this be "label (single)" when we add the plural back in?
      label: "Label",
      ariaLabel: "Dimension label",
      headerTooltip: "A human readable name for this dimension (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },

    {
      name: "column",
      label: "Dimension column",
      ariaLabel: "Dimension column",
      headerTooltip:
        "A categorical column from the data model that this metrics set is based on",
      cellRenderer: new CellConfigSelector(
        inputChangeHandler,
        dimensionOptions,
        "Select a column...",
        "The selected dimension is not present in this model. Please choose a valid dimension."
      ),
    },
    {
      name: "description",
      label: "Description",
      ariaLabel: "Dimension description",
      headerTooltip:
        "A human readable description of this dimension (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    // FIXME: we'll want to  add this back later
    // {
    //   name: "labelPlural",
    //   label: "label (plural)",
    //   headerTooltip:
    //     "an pluralized human readable name for this dimension (optional)",
    //   cellRenderer: new CellConfigInput(inputChangeHandler),
    // },
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

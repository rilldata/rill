import { nicelyFormattedTypesSelectorOptions } from "@rilldata/web-common/features/dashboards/humanize-numbers";
import {
  CellConfigInput,
  CellConfigSelector,
  ColumnConfig,
} from "../../components/table-editable/ColumnConfig";
import { ValidationState } from "./errors";
import type { MeasureEntity } from "./metrics-internal-store";

export const initMeasuresColumns = (
  inputChangeHandler,
  expressionValidationHandler
) =>
  <ColumnConfig<CellConfigInput | CellConfigSelector>[]>[
    {
      name: "label",
      label: "Label",
      headerTooltip: "A human readable name for this measure (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "expression",
      label: "Expression",
      headerTooltip: "A valid SQL aggregation expression for this measure",
      cellRenderer: new CellConfigInput(
        inputChangeHandler,
        (row) => ({
          // TODO: remove the entity record type in validation
          state: !(row as MeasureEntity).__ERROR__
            ? ValidationState.OK
            : ValidationState.ERROR,
          message: (row as MeasureEntity).__ERROR__,
        }),
        expressionValidationHandler
      ),
      customClass: "ui-copy-code",
    },
    {
      name: "description",
      label: "Description",

      headerTooltip: "A human readable description of this measure (optional)",

      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "format_preset",
      label: "Number formatting",
      headerTooltip:
        "The number formatting used for this measure in the Metrics Explorer",
      cellRenderer: new CellConfigSelector(
        inputChangeHandler,
        nicelyFormattedTypesSelectorOptions
      ),
    },
    // FIXME: will be needed later for API
    // {
    //   name: "sqlName",
    //   label: "identifier",
    //   headerTooltip: "a unique SQL identifier for this measure",
    //   cellRenderer: TableCellInput,
    //   onchange: inputChangeHandler,
    //   validation: (row: MeasureDefinitionEntity) => row.sqlNameIsValid,
    // },

    // FIXME: will be needed later for sparkline summary
    // {
    //   name: "id",
    //   label: "preview",
    //   tooltip: "a preview of this measure over the selected time dimension",
    //   cellRenderer: TableCellSparkline,
    // },
  ];

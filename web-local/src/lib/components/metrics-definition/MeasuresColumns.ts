import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

import {
  ColumnConfig,
  CellConfigInput,
  CellConfigSelector,
} from "../table-editable/ColumnConfig";
import { nicelyFormattedTypesSelectorOptions } from "../../util/humanize-numbers";

export const initMeasuresColumns = (
  inputChangeHandler,
  expressionValidationHandler
) =>
  <ColumnConfig<CellConfigInput | CellConfigSelector>[]>[
    {
      name: "label",
      headerTooltip: "a human readable name for this measure (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "expression",
      headerTooltip: "a valid SQL aggregation expression for this measure",
      cellRenderer: new CellConfigInput(
        inputChangeHandler,
        (row: MeasureDefinitionEntity) => ({
          state: row.expressionIsValid,
          message: row.expressionValidationError,
        }),
        expressionValidationHandler
      ),
    },
    {
      name: "description",
      headerTooltip: "a human readable description of this measure (optional)",

      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "format_preset",
      label: "number formatting",
      headerTooltip:
        "the number formatting used for this measure in the Metrics Explorer",
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

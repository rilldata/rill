import type { MeasureEntity } from "@rilldata/web-local/lib/application-state-stores/metrics-internal-store";
import { ValidationState } from "@rilldata/web-local/lib/temp/metrics";
import { nicelyFormattedTypesSelectorOptions } from "../../util/humanize-numbers";

import {
  CellConfigInput,
  CellConfigSelector,
  ColumnConfig,
} from "../table-editable/ColumnConfig";

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
        (row) => ({
          // TODO: remove the entity record type in validation
          state: !(row as MeasureEntity).__ERROR__
            ? ValidationState.OK
            : ValidationState.ERROR,
          message: (row as MeasureEntity).__ERROR__,
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

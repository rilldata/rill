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
      ariaLabel: "Measure label",
      headerTooltip: "A human readable name for this measure (optional)",
      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "expression",
      label: "Expression",
      ariaLabel: "Measure expression",
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
      ariaLabel: "Measure description",
      headerTooltip: "A human readable description of this measure (optional)",

      cellRenderer: new CellConfigInput(inputChangeHandler),
    },
    {
      name: "format_preset",
      label: "Number formatting",
      ariaLabel: "Measure number formatting",
      headerTooltip:
        "The number formatting used for this measure in the Metrics Explorer",
      cellRenderer: new CellConfigSelector(
        inputChangeHandler,
        nicelyFormattedTypesSelectorOptions
      ),
    },
  ];

import type { PERC_DIFF } from "@rilldata/web-common/components/data-types/type-utils";
import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";

/**
 * A DimensionTableRow object represents a single row in a
 * dimension table.
 *
 * Each row will have the following key/value pairs:
 *
 * 1) for the dimension:
 *    - key: dimension name //FIXME: could cause key collisions
 *    - value: dimension value for this row. May be null.
 *
 * 2) for the measures:
 *    - key: the `name` field from the YAML e.g. `measure_1`, `total`, ect
 *    - value: the raw measure value. Number or null.
 *
 * 3) context columns for delta, delta percent, pct of total for the
 * active measure:
 *    - key: the YAML `name` of the active measure with  `_delta`,
 *      `_delta_perc`, and `_percent_of_total` appended.
 *    - value: the raw value of the context column. Number or null.
 *
 * 4) additional formatted strings:
 *    - key: the YAML `name` of each measure with `__formatted_` prepended.
 *      Context columns are included here as well.
 *    - value: the formatted string for the measure or context column.
 * 
 * 5) formatted columns for delta, delta percent, pct of total for the
 * active measure:
 *    - key: the YAML `name` of the active measure with  `__formatted_*_delta`,
 *      `__formatted_*_delta_perc`, and `__formatted_*_percent_of_total`.
 *    - value: the formatted value of the context column. string | NumberParts | PERC_DIFF 

 */
export type DimensionTableRow = Record<
  string,
  null | string | number | NumberParts | PERC_DIFF
>;

// import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
// import type { SvelteComponent } from "svelte";

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
 *    - value: dimension value for this row
 *
 * 2) for the measures:
 *    - key: the `name` field from the YAML e.g. `measure_1`, `total`, ect
 *    - value: the raw measure value
 *
 * 3) context columns for delta, delta percent, pct of total for the
 * active measure:
 *    - key: the YAML `name` of the active measure with  `_delta`,
 *      `_delta_perc`, and `_percent_of_total` appended.
 *    - value: the raw value of the context column,
 *      or a NumberParts or PERC_DIFF for percent of total and delta
 *
 * 4) additional formatted strings:
 *    - key: the YAML `name` of each measure with `__formatted_` prepended.
 *      Context columns are included here as well.
 *    - value: the formatted string for the measure or context column.
 *
 */
export type DimensionTableRow = Record<
  string,
  string | number | NumberParts | PERC_DIFF
>;

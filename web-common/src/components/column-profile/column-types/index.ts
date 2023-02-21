import {
  CATEGORICALS,
  INTERVALS,
  isNested,
  NUMERICS,
  TIMESTAMPS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import NestedProfile from "./NestedProfile.svelte";
import NumericProfile from "./NumericProfile.svelte";
import TimestampProfile from "./TimestampProfile.svelte";
import VarcharProfile from "./VarcharProfile.svelte";

export function getColumnType(type) {
  // deal with nested types before we deal with the others.
  if (isNested(type)) return NestedProfile;
  // strip decimal brackets
  if (type.includes("DECIMAL")) type = "DECIMAL";
  if (CATEGORICALS.has(type)) return VarcharProfile;
  if (NUMERICS.has(type) || INTERVALS.has(type)) return NumericProfile;
  if (TIMESTAMPS.has(type)) return TimestampProfile;
}

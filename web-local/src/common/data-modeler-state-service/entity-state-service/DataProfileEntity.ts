import type {
  DerivedEntityRecord,
  EntityState,
  EntityStateActionArg,
} from "./EntityStateService";
import type { ProfileColumn } from "@rilldata/web-local/lib/types";

export interface DataProfileEntity extends DerivedEntityRecord {
  profile?: ProfileColumn[];
  cardinality?: number;
  /**
   * sizeInBytes is the total size of the previewed table.
   * It is not generated until the user exports the query.
   */
  sizeInBytes?: number;
  nullCounts?: number;
  preview?: Array<unknown>;

  profiled?: boolean;
}
export type DataProfileState = EntityState<DataProfileEntity>;
export type DataProfileStateActionArg = EntityStateActionArg<DataProfileEntity>;

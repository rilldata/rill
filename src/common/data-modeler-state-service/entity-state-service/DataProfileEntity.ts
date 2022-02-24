import type {
    DerivedEntityRecord,
    EntityState, EntityStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ProfileColumn } from "$lib/types";

export interface DataProfileEntity extends DerivedEntityRecord {
    profile?: ProfileColumn[];
    cardinality?: number;
    /**
     * sizeInBytes is the total size of the previewed table.
     * It is not generated until the user exports the query.
     */
    sizeInBytes?: number;
    nullCounts?: any;
    preview?: any;
}
export type DataProfileState = EntityState<DataProfileEntity>;
export type DataProfileStateActionArg = EntityStateActionArg<DataProfileEntity>;

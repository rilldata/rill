import type { AppStore } from "$lib/app-store";
import { dataModelerStateService } from "$lib/app-store";
import type {
    PersistentTableEntity,
    PersistentTableState
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
    DerivedTableEntity,
    DerivedTableState
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export type PersistentTableStore = AppStore<PersistentTableEntity, PersistentTableState>;
export function createPersistentTableStore(): PersistentTableStore {
    return dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Persistent).store;
}

export type DerivedTableStore = AppStore<DerivedTableEntity, DerivedTableState>;
export function createDerivedTableStore(): DerivedTableStore {
    return dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Derived).store;
}

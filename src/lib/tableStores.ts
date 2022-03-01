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
import type {
    DerivedTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type {
    PersistentTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";

export type PersistentTableStore = AppStore<PersistentTableEntity, PersistentTableState>;
export function createPersistentTableStore(): [PersistentTableEntityService, PersistentTableStore] {
    const service = dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Persistent);
    return [service, service.store];
}

export type DerivedTableStore = AppStore<DerivedTableEntity, DerivedTableState>;
export function createDerivedTableStore(): [DerivedTableEntityService, DerivedTableStore] {
    const service = dataModelerStateService
        .getEntityStateService(EntityType.Table, StateType.Derived);
    return [service, service.store];
}

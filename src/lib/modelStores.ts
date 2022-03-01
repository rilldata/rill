import type { AppStore } from "$lib/app-store";
import { dataModelerStateService } from "$lib/app-store";
import type {
    PersistentModelEntity, PersistentModelState
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity, DerivedModelState
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    PersistentModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";

export type PersistentModelStore = AppStore<PersistentModelEntity, PersistentModelState>;
export function createPersistentModelStore(): [PersistentModelEntityService, PersistentModelStore] {
    const service = dataModelerStateService
        .getEntityStateService(EntityType.Model, StateType.Persistent);
    return [service, service.store];
}

export type DerivedModelStore = AppStore<DerivedModelEntity, DerivedModelState>;
export function createDerivedModelStore(): [DerivedModelEntityService, DerivedModelStore] {
    const service = dataModelerStateService
        .getEntityStateService(EntityType.Model, StateType.Derived);
    return [service, service.store];
}

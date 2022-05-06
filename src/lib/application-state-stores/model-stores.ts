/**
 * Contains the Svelte store for the model entities.
 * The persistent model is the basic state gleaned from the asset.
 * The derived model state contains derived information about the model
 * such as the column profiles.
 * The persistent model state tends to be generated from the model SQL files.
 * 
 * The stores in this file reactively respond to updates from the application server
 * through the socket server.
 */
import type { AppStore } from "$lib/application-state-stores/application-store";
import { dataModelerStateService } from "$lib/application-state-stores/application-store";
import type {
    PersistentModelEntity, PersistentModelState
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity, DerivedModelState
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export type PersistentModelStore = AppStore<PersistentModelEntity, PersistentModelState>;
export function createPersistentModelStore(): PersistentModelStore {
    return dataModelerStateService
        .getEntityStateService(EntityType.Model, StateType.Persistent).store;
}

export type DerivedModelStore = AppStore<DerivedModelEntity, DerivedModelState>;
export function createDerivedModelStore(): DerivedModelStore {
    return dataModelerStateService
        .getEntityStateService(EntityType.Model, StateType.Derived).store;
}

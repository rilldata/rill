/**
 * Contains the Svelte store for the source entities.
 * The persistent source is the basic state gleaned from the asset.
 * The derived source state contains derived information about the source
 * such as the source profiles.
 * The persistent source state tends to be generated from the model SQL files.
 *
 * The stores in this file reactively respond to updates from the application server
 * through the socket server.
 */
import type { AppStore } from "$lib/application-state-stores/application-store";
import { dataModelerStateService } from "$lib/application-state-stores/application-store";
import type {
  PersistentSourceEntity,
  PersistentSourceState,
} from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import type {
  DerivedSourceEntity,
  DerivedSourceState,
} from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";
import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export type PersistentSourceStore = AppStore<
  PersistentSourceEntity,
  PersistentSourceState
>;
export function createPersistentSourceStore(): PersistentSourceStore {
  return dataModelerStateService.getEntityStateService(
    EntityType.Source,
    StateType.Persistent
  ).store;
}

export type DerivedSourceStore = AppStore<
  DerivedSourceEntity,
  DerivedSourceState
>;
export function createDerivedSourceStore(): DerivedSourceStore {
  return dataModelerStateService.getEntityStateService(
    EntityType.Source,
    StateType.Derived
  ).store;
}

/**
 * Contains the Svelte store for the table entities.
 * The persistent table is the basic state gleaned from the asset.
 * The derived table state contains derived information about the table
 * such as the table profiles.
 * The persistent table state tends to be generated from the model SQL files.
 *
 * The stores in this file reactively respond to updates from the application server
 * through the socket server.
 */
import type { AppStore } from "./application-store";
import { dataModelerStateService } from "./application-store";
import type {
  PersistentTableEntity,
  PersistentTableState,
} from "../../common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
  DerivedTableEntity,
  DerivedTableState,
} from "../../common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import {
  EntityType,
  StateType,
} from "../../common/data-modeler-state-service/entity-state-service/EntityStateService";

export type PersistentTableStore = AppStore<
  PersistentTableEntity,
  PersistentTableState
>;
export function createPersistentTableStore(): PersistentTableStore {
  return dataModelerStateService.getEntityStateService(
    EntityType.Table,
    StateType.Persistent
  ).store;
}

export type DerivedTableStore = AppStore<DerivedTableEntity, DerivedTableState>;
export function createDerivedTableStore(): DerivedTableStore {
  return dataModelerStateService.getEntityStateService(
    EntityType.Table,
    StateType.Derived
  ).store;
}

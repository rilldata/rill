import type {
  EntityState,
  EntityStateActionArg,
} from "./EntityStateService";
import {
  EntityStateService,
  EntityType,
  StateType,
} from "./EntityStateService";
import type { DataProfileEntity } from "./DataProfileEntity";
import type { SourceTable } from "../../../lib/types";

export interface DerivedModelEntity extends DataProfileEntity {
  type: EntityType.Model;
  /** sanitizedQuery is always a 1:1 function of the query itself */
  sanitizedQuery: string;
  error?: string;
  sources?: SourceTable[];
}
export type DerivedModelState = EntityState<DerivedModelEntity>;
export type DerivedModelStateActionArg = EntityStateActionArg<
  DerivedModelEntity,
  DerivedModelState,
  DerivedModelEntityService
>;

export class DerivedModelEntityService extends EntityStateService<
  DerivedModelEntity,
  DerivedModelState
> {
  public readonly entityType = EntityType.Model;
  public readonly stateType = StateType.Derived;
}

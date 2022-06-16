import { StateActions } from "./StateActions";
import type {
  PersistentSourceEntity,
  PersistentSourceStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import type { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataProfileStateActionArg } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

export class SourceStateActions extends StateActions {
  @StateActions.PersistentSourceAction()
  public addOrUpdateSourceToState(
    { stateService, draftState }: PersistentSourceStateActionArg,
    source: PersistentSourceEntity,
    isNew: boolean
  ): void {
    // we do not create any state here since it is contingent on whether the import step
    // worked as expected.
  }
}
